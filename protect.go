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

func (cmd *protectCommand) Register(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.review, "review", false, "Require pull request reviews before merging")
}

type protectCommand struct {
	review bool
}

func (cmd *protectCommand) Run(ctx context.Context, args []string) error {
	return runCommand(ctx, cmd.handleRepoProtectBranch)
}

// handleRepo will return nil error if the user does not have access to something.
func (cmd *protectCommand) handleRepoProtectBranch(ctx context.Context, client *github.Client, repo *github.Repository) error {
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
				if !cmd.review {
					fmt.Printf("[OK] %s:%s is already protected\n", *repo.FullName, b.GetName())
					return nil
				} else {
					// we need to check if pull request reviews are required
					protection, _, err := client.Repositories.GetBranchProtection(ctx, *repo.Owner.Login, *repo.Name, b.GetName())
					if err != nil {
						return err
					}
					if protection.RequiredPullRequestReviews != nil && protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
						fmt.Printf("[OK] %s:%s is already protected and pull request reviews are required\n", *repo.FullName, b.GetName())
						return nil
					}
				}
			}

			if dryrun {
				fmt.Printf("[UPDATE] %s:%s will be changed to protected (require reviews: %v)\n", *repo.FullName, b.GetName(), cmd.review)
				return nil
			}

			// settings for pull request reviews
			var reviews *github.PullRequestReviewsEnforcementRequest
			if cmd.review {
				reviews = &github.PullRequestReviewsEnforcementRequest{
					RequiredApprovingReviewCount: 1,
				}
			}

			// set the branch to be protected
			if _, _, err := client.Repositories.UpdateBranchProtection(ctx, *repo.Owner.Login, *repo.Name, b.GetName(), &github.ProtectionRequest{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{},
				},
				RequiredPullRequestReviews: reviews,
			}); err != nil {
				return err
			}
			fmt.Printf("[OK] %s:%s is protected\n", *repo.FullName, b.GetName())
		}
	}

	return nil
}
