package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/genuinetools/pepper/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

var (
	token      string
	enturl     string
	orgs       stringSlice
	singleRepo string
	nouser     bool
	dryrun     bool

	debug bool
)

// stringSlice is a slice of strings
type stringSlice []string

// implement the flag interface for stringSlice
func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "pepper"
	p.Description = "A tool for performing actions on GitHub repos or a single repo"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Build the list of available commands.
	p.Commands = []cli.Command{
		&auditCommand{},
		&collaboratorsCommand{},
		&mergeCommand{},
		&protectCommand{},
		&releaseCommand{},
	}

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("pepper", flag.ExitOnError)
	p.FlagSet.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub API token (or env var GITHUB_TOKEN)")
	p.FlagSet.StringVar(&token, "t", os.Getenv("GITHUB_TOKEN"), "GitHub API token (or env var GITHUB_TOKEN)")

	p.FlagSet.StringVar(&enturl, "url", "", "GitHub Enterprise URL")
	p.FlagSet.StringVar(&enturl, "u", "", "GitHub Enterprise URL")

	p.FlagSet.Var(&orgs, "orgs", "organizations to include")
	p.FlagSet.StringVar(&singleRepo, "repo", "", "specific repo (e.g. 'genuinetools/img')")
	p.FlagSet.StringVar(&singleRepo, "r", "", "specific repo (e.g. 'genuinetools/img')")

	p.FlagSet.BoolVar(&nouser, "nouser", false, "do not include your user")
	p.FlagSet.BoolVar(&dryrun, "dry-run", false, "do not change settings just print the changes that would occur")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if token == "" {
			return errors.New("GitHub token cannot be empty")
		}

		if nouser && orgs == nil && len(singleRepo) < 1 {
			return errors.New("no organizations, user, or repo provided")
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func runCommand(ctx context.Context, cmd func(context.Context, *github.Client, *github.Repository) error) error {
	// On ^C, or SIGTERM handle exit.
	signals := make(chan os.Signal, 0)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		for sig := range signals {
			cancel()
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Create the http client.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create the github client.
	client := github.NewClient(tc)
	if enturl != "" {
		var err error
		client.BaseURL, err = url.Parse(enturl + "/api/v3/")
		if err != nil {
			return fmt.Errorf("Parsing URL for enterprise failed: %v", err)
		}
	}

	// Affiliation must be set before we add the user to the "orgs".
	affiliation := "owner,collaborator"
	if len(orgs) > 0 {
		affiliation += ",organization_member"
	}

	if !nouser {
		// Get the current user
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			if v, ok := err.(*github.RateLimitError); ok {
				return fmt.Errorf("%s Limit: %d; Remaining: %d; Retry After: %s", v.Message, v.Rate.Limit, v.Rate.Remaining, time.Until(v.Rate.Reset.Time).String())
			}

			return fmt.Errorf("Getting user failed: %v", err)
		}
		username := *user.Login
		// add the current user to orgs
		orgs = append(orgs, username)
	}

	page := 1
	perPage := 100
	logrus.Debugf("Getting repositories...")
	if err := getRepositories(ctx, client, page, perPage, affiliation, cmd); err != nil {
		if v, ok := err.(*github.RateLimitError); ok {
			return fmt.Errorf("%s Limit: %d; Remaining: %d; Retry After: %s", v.Message, v.Rate.Limit, v.Rate.Remaining, time.Until(v.Rate.Reset.Time).String())
		}

		return err
	}

	return nil
}

func getRepositories(ctx context.Context, client *github.Client, page, perPage int, affiliation string, cmd func(context.Context, *github.Client, *github.Repository) error) error {
	opt := &github.RepositoryListOptions{
		Affiliation: affiliation,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	var (
		repos []*github.Repository
		resp  *github.Response
		err   error
	)

	if len(singleRepo) < 1 {
		// Get all the repos.
		repos, resp, err = client.Repositories.List(ctx, "", opt)
	} else {
		// Find the one repo.
		repos, err = searchRepos(ctx, client, singleRepo)
	}
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if !in(orgs, *repo.Owner.Login) && len(singleRepo) < 1 {
			continue
		}

		logrus.Debugf("Handling repo %s...", *repo.FullName)
		if err := cmd(ctx, client, repo); err != nil {
			logrus.Warn(err)
		}
	}

	// Return early if we are on the last page.
	if resp == nil || page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getRepositories(ctx, client, page, perPage, affiliation, cmd)
}

func in(a stringSlice, s string) bool {
	for _, b := range a {
		if b == s {
			return true
		}
	}
	return false
}

func searchRepos(ctx context.Context, client *github.Client, searchRepo string) ([]*github.Repository, error) {
	optSearch := &github.SearchOptions{
		Sort:  "forks",
		Order: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 1,
		},
	}

	search := strings.SplitN(searchRepo, "/", 2)
	repos, _, err := client.Search.Repositories(ctx, fmt.Sprintf("org:%s in:name %s fork:true", search[0], search[1]), optSearch)
	if err != nil {
		return nil, err
	}

	if len(repos.Repositories) < 1 {
		return nil, fmt.Errorf("found no repositories matching: %s", searchRepo)
	}

	r := []*github.Repository{}
	for _, repo := range repos.Repositories {
		r = append(r, &repo)
	}
	return r, nil
}
