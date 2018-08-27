package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
)

const mergeHelp = `Update all merge settings to allow specific types only.`

func (cmd *mergeCommand) Name() string      { return "merge" }
func (cmd *mergeCommand) Args() string      { return "[OPTIONS]" }
func (cmd *mergeCommand) ShortHelp() string { return mergeHelp }
func (cmd *mergeCommand) LongHelp() string  { return mergeHelp }
func (cmd *mergeCommand) Hidden() bool      { return true }

func (cmd *mergeCommand) Register(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.commits, "commits", false, "Allow merge commits, add all commits from the head branch to the base branch with a merge commit")
	fs.BoolVar(&cmd.squash, "squash", false, "Allow squash merging, combine all commits from the head branch into a single commit in the base branch")
	fs.BoolVar(&cmd.rebase, "rebase", false, "Allow rebase merging, add all commits from the head branch onto the base branch individually")
}

type mergeCommand struct {
	commits bool
	squash  bool
	rebase  bool
}

func (cmd *mergeCommand) Run(ctx context.Context, args []string) error {
	return runCommand(ctx, cmd.handleRepoMergeOpt)
}

// handleRepo will return nil error if the user does not have access to something.
func (cmd *mergeCommand) handleRepoMergeOpt(ctx context.Context, client *github.Client, repo *github.Repository) error {
	if !cmd.commits && !cmd.squash && !cmd.rebase {
		return errors.New("you must choose from commits, squash, and/or rebase")
	}

	repo, resp, err := client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	willBeUpdated := false
	if repo.GetAllowMergeCommit() != cmd.commits {
		willBeUpdated = true
	}
	if repo.GetAllowSquashMerge() != cmd.squash {
		willBeUpdated = true
	}
	if repo.GetAllowRebaseMerge() != cmd.rebase {
		willBeUpdated = true
	}

	opt := []string{}
	if cmd.commits {
		opt = append(opt, "mergeCommits")
	}
	if cmd.squash {
		opt = append(opt, "squash")
	}
	if cmd.rebase {
		opt = append(opt, "rebase")
	}

	if dryrun && willBeUpdated {
		fmt.Printf("[UPDATE] %s will be changed to %s\n", *repo.FullName, strings.Join(opt, " | "))
		return nil
	}

	if !willBeUpdated {
		fmt.Printf("[OK] %s is already set to %s\n", *repo.FullName, strings.Join(opt, " | "))
		return nil
	}

	// TODO: actually update the merge settings when the API allows it.

	return nil
}
