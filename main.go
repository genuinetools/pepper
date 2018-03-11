/*
Copyright 2016 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/oauth2"

	"github.com/genuinetools/pepper/version"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = ` _ __   ___ _ __  _ __   ___ _ __
| '_ \ / _ \ '_ \| '_ \ / _ \ '__|
| |_) |  __/ |_) | |_) |  __/ |
| .__/ \___| .__/| .__/ \___|_|
|_|        |_|   |_|

 Set all your GitHub repos master branches to be protected.
 Version: %s
 Build: %s

`
)

var (
	token  string
	enturl string
	orgs   stringSlice
	nouser bool
	dryrun bool

	debug bool
	vrsn  bool
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

func init() {
	// parse flags
	flag.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub API token (or env var GITHUB_TOKEN)")
	flag.StringVar(&enturl, "url", "", "GitHub Enterprise URL")
	flag.Var(&orgs, "orgs", "organizations to include")
	flag.BoolVar(&nouser, "nouser", false, "do not include your user")
	flag.BoolVar(&dryrun, "dry-run", false, "do not change branch settings just print the changes that would occur")

	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION, version.GITCOMMIT))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("pepper version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if token == "" {
		usageAndExit("GitHub token cannot be empty.", 1)
	}

	if nouser && orgs == nil {
		usageAndExit("no organizations provided", 1)
	}
}

func main() {
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	ctx := context.Background()

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
		logrus.Fatal(err)
	}
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

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
