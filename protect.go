package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
)

const protectHelp = `Protect the master branch.`

func (cmd *protectCommand) Name() string      { return "protect" }
func (cmd *protectCommand) Args() string      { return "[OPTIONS]" }
func (cmd *protectCommand) ShortHelp() string { return protectHelp }
func (cmd *protectCommand) LongHelp() string  { return protectHelp }
func (cmd *protectCommand) Hidden() bool      { return false }

func (cmd *protectCommand) Register(fs *flag.FlagSet) {}

type protectCommand struct{}

func (cmd *protectCommand) Run(ctx context.Context, args []string) error {
	return runCommand(ctx, handleRepoProtectBranch)
}

// handleRepo will return nil error if the user does not have access to something.
func handleRepoProtectBranch(ctx context.Context, client *github.Client, repo *github.Repository) error {
	if !in(orgs, *repo.Owner.Login) {
		return nil
	}

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
		if branch.GetName() == "master" {
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

			if dryrun {
				fmt.Printf("[UPDATE] %s:%s will be changed to protected\n", *repo.FullName, b.GetName())
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
			fmt.Printf("[OK] %s:%s is protected\n", *repo.FullName, b.GetName())
		}
	}

	return nil
}
