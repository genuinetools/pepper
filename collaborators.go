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

const collaboratorsHelp = `Add a collaborator to all the repositories.`

func (cmd *collaboratorsCommand) Name() string      { return "collaborators" }
func (cmd *collaboratorsCommand) Args() string      { return "[OPTIONS] COLLABORATOR" }
func (cmd *collaboratorsCommand) ShortHelp() string { return collaboratorsHelp }
func (cmd *collaboratorsCommand) LongHelp() string  { return collaboratorsHelp }
func (cmd *collaboratorsCommand) Hidden() bool      { return false }

func (cmd *collaboratorsCommand) Register(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.pull, "pull", false, "Team members can pull, but not push to or administer this repository")
	fs.BoolVar(&cmd.push, "push", false, "Team members can pull and push, but not administer this repository")
	fs.BoolVar(&cmd.admin, "admin", false, "Team members can pull, push and administer this repository")
}

type collaboratorsCommand struct {
	pull  bool
	push  bool
	admin bool

	nick string
}

func (cmd *collaboratorsCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("must pass a collaborator")
	}
	cmd.nick = args[0]

	if !cmd.pull && !cmd.push && !cmd.admin {
		return errors.New("you must choose from push, pull, and/or admin")
	}

	return runCommand(ctx, cmd.handleRepoAddCollaborator)
}

// handleRepo will return nil error if the user does not have access to something.
func (cmd *collaboratorsCommand) handleRepoAddCollaborator(ctx context.Context, client *github.Client, repo *github.Repository) error {
	opt := []string{}
	if cmd.admin {
		opt = append(opt, "admin")
	}
	if cmd.pull {
		opt = append(opt, "pull")
	}
	if cmd.push {
		opt = append(opt, "push")
	}
	if len(opt) > 1 {
		return fmt.Errorf("cannot specify multiple values of %s, choose one", strings.Join(opt, " | "))
	}

	lsopt := &github.ListOptions{
		PerPage: 100,
	}

	teams, resp, err := client.Repositories.ListTeams(ctx, repo.GetOwner().GetLogin(), repo.GetName(), lsopt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	collabs, resp, err := client.Repositories.ListCollaborators(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.ListCollaboratorsOptions{ListOptions: *lsopt})
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	push := []string{}
	pull := []string{}
	admin := []string{}
	if len(collabs) > 1 {
		for _, c := range collabs {
			userTeams := []github.Team{}
			for _, t := range teams {
				isMember, resp, err := client.Teams.GetTeamMembership(ctx, t.GetID(), c.GetLogin())
				if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden && err == nil && isMember.GetState() == "active" {
					userTeams = append(userTeams, *t)
				}
			}

			perms := c.GetPermissions()

			switch {
			case perms["admin"]:
				permTeams := []string{}
				for _, t := range userTeams {
					if t.GetPermission() == "admin" {
						permTeams = append(permTeams, t.GetName())
					}
				}
				admin = append(admin, fmt.Sprintf("\t\t\t%s (teams: %s)", c.GetLogin(), strings.Join(permTeams, ", ")))
			case perms["push"]:
				push = append(push, fmt.Sprintf("\t\t\t%s", c.GetLogin()))
			case perms["pull"]:
				pull = append(pull, fmt.Sprintf("\t\t\t%s", c.GetLogin()))
			}
		}
	}

	willBeUpdated := false
	if in(admin, cmd.nick) != cmd.admin {
		willBeUpdated = true
	}
	if in(pull, cmd.nick) != cmd.pull {
		willBeUpdated = true
	}
	if in(push, cmd.nick) != cmd.push {
		willBeUpdated = true
	}

	if willBeUpdated && dryrun {
		fmt.Printf("[UPDATE] %s will have %s added as a collaborator (%s)\n", *repo.FullName, cmd.nick, strings.Join(opt, " | "))
		return nil
	}

	if !willBeUpdated {
		fmt.Printf("[OK] %s already has %s added as a collaborator (%s)\n", *repo.FullName, cmd.nick, strings.Join(opt, " | "))
		return nil
	}

	// Add the collaborator.
	resp, err = client.Repositories.AddCollaborator(ctx, repo.GetOwner().GetLogin(), repo.GetName(), cmd.nick, &github.RepositoryAddCollaboratorOptions{
		Permission: strings.Join(opt, ""),
	})
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}
	fmt.Printf("[OK] %s has %s added as a collaborator (%s)\n", *repo.FullName, cmd.nick, strings.Join(opt, " | "))

	return nil
}
