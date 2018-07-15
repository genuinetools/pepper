package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/genuinetools/pepper/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

var (
	token  string
	enturl string
	orgs   stringSlice
	nouser bool
	dryrun bool

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
	p.Description = "Tool to set all GitHub repo master branches to be protected"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("ship", flag.ExitOnError)
	p.FlagSet.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub API token (or env var GITHUB_TOKEN)")
	p.FlagSet.StringVar(&enturl, "url", "", "GitHub Enterprise URL")
	p.FlagSet.Var(&orgs, "orgs", "organizations to include")
	p.FlagSet.BoolVar(&nouser, "nouser", false, "do not include your user")
	p.FlagSet.BoolVar(&dryrun, "dry-run", false, "do not change branch settings just print the changes that would occur")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if token == "" {
			return errors.New("GitHub token cannot be empty")
		}

		if nouser && orgs == nil {
			return errors.New("no organizations provided")
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context) error {
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
				logrus.Fatal(err)
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
					logrus.Fatalf("%s Limit: %d; Remaining: %d; Retry After: %s", v.Message, v.Rate.Limit, v.Rate.Remaining, time.Until(v.Rate.Reset.Time).String())
				}

				logrus.Fatal(err)
			}
			username := *user.Login
			// add the current user to orgs
			orgs = append(orgs, username)
		}

		page := 1
		perPage := 100
		logrus.Debugf("Getting repositories...")
		if err := getRepositories(ctx, client, page, perPage, affiliation); err != nil {
			if v, ok := err.(*github.RateLimitError); ok {
				logrus.Fatalf("%s Limit: %d; Remaining: %d; Retry After: %s", v.Message, v.Rate.Limit, v.Rate.Remaining, time.Until(v.Rate.Reset.Time).String())
			}

			logrus.Fatal(err)
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func getRepositories(ctx context.Context, client *github.Client, page, perPage int, affiliation string) error {
	opt := &github.RepositoryListOptions{
		Affiliation: affiliation,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	repos, resp, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		logrus.Debugf("Handling repo %s...", *repo.FullName)
		if err := handleRepo(ctx, client, repo); err != nil {
			logrus.Warn(err)
		}
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getRepositories(ctx, client, page, perPage, affiliation)
}

// handleRepo will return nil error if the user does not have access to something.
func handleRepo(ctx context.Context, client *github.Client, repo *github.Repository) error {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	branches, resp, err := client.Repositories.ListBranches(ctx, *repo.Owner.Login, *repo.Name, opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return nil
	}
	if err != nil {
		return err
	}

	for _, branch := range branches {
		if branch.GetName() == "master" && in(orgs, *repo.Owner.Login) {
			// we must get the individual branch for the branch protection to work
			b, _, err := client.Repositories.GetBranch(ctx, *repo.Owner.Login, *repo.Name, branch.GetName())
			if err != nil {
				return err
			}

			// return early if it is already protected
			if b.GetProtected() {
				fmt.Printf("[OK] %s:%s is already protected\n", *repo.FullName, b.GetName())
				return nil
			}

			fmt.Printf("[UPDATE] %s:%s will be changed to protected\n", *repo.FullName, b.GetName())
			if dryrun {
				// return early
				return nil
			}

			// set the branch to be protected
			if _, _, err := client.Repositories.UpdateBranchProtection(ctx, *repo.Owner.Login, *repo.Name, b.GetName(), &github.ProtectionRequest{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{},
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func in(a stringSlice, s string) bool {
	for _, b := range a {
		if b == s {
			return true
		}
	}
	return false
}
