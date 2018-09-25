package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
)

const auditHelp = `Audit collaborators, branches, hooks, deploy keys etc.`

func (cmd *auditCommand) Name() string      { return "audit" }
func (cmd *auditCommand) Args() string      { return "[OPTIONS]" }
func (cmd *auditCommand) ShortHelp() string { return auditHelp }
func (cmd *auditCommand) LongHelp() string  { return auditHelp }
func (cmd *auditCommand) Hidden() bool      { return false }

func (cmd *auditCommand) Register(fs *flag.FlagSet) {}

type auditCommand struct{}

func (cmd *auditCommand) Run(ctx context.Context, args []string) error {
	return runCommand(ctx, handleAudit)
}

// handleAudit will return nil error if the user does not have access to something.
func handleAudit(ctx context.Context, client *github.Client, repo *github.Repository) error {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	teams, resp, err := client.Repositories.ListTeams(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	collabs, resp, err := client.Repositories.ListCollaborators(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.ListCollaboratorsOptions{ListOptions: *opt})
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	keys, resp, err := client.Repositories.ListKeys(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	hooks, resp, err := client.Repositories.ListHooks(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	branches, _, err := client.Repositories.ListBranches(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if err != nil {
		return err
	}
	protectedBranches := []string{}
	unprotectedBranches := []string{}
	for _, branch := range branches {
		// we must get the individual branch for the branch protection to work
		b, _, err := client.Repositories.GetBranch(ctx, repo.GetOwner().GetLogin(), repo.GetName(), branch.GetName())
		if err != nil {
			return err
		}
		if b.GetProtected() {
			protectedBranches = append(protectedBranches, b.GetName())
		} else {
			unprotectedBranches = append(unprotectedBranches, b.GetName())
		}
	}

	// only print whole status if we have more that one collaborator
	if len(collabs) <= 1 && len(keys) < 1 && len(hooks) < 1 && len(protectedBranches) < 1 && len(unprotectedBranches) < 1 {
		return nil
	}

	output := fmt.Sprintf("%s -> \n", repo.GetFullName())

	if len(collabs) > 1 {
		push := []string{}
		pull := []string{}
		admin := []string{}
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
		output += fmt.Sprintf("\tCollaborators (%d):\n", len(collabs))
		output += fmt.Sprintf("\t\tAdmin (%d):\n%s\n", len(admin), strings.Join(admin, "\n"))
		output += fmt.Sprintf("\t\tWrite (%d):\n%s\n", len(push), strings.Join(push, "\n"))
		output += fmt.Sprintf("\t\tRead (%d):\n%s\n", len(pull), strings.Join(pull, "\n"))
	}

	if len(keys) > 0 {
		kstr := []string{}
		for _, k := range keys {
			kstr = append(kstr, fmt.Sprintf("\t\t%s - ro:%t (%s)", k.GetTitle(), k.GetReadOnly(), k.GetURL()))
		}
		output += fmt.Sprintf("\tKeys (%d):\n%s\n", len(kstr), strings.Join(kstr, "\n"))
	}

	if len(hooks) > 0 {
		hstr := []string{}
		for _, h := range hooks {
			hstr = append(hstr, fmt.Sprintf("\t\t%s - active:%t (%s)", h.GetName(), h.GetActive(), h.GetURL()))
		}
		output += fmt.Sprintf("\tHooks (%d):\n%s\n", len(hstr), strings.Join(hstr, "\n"))
	}

	if len(protectedBranches) > 0 {
		output += fmt.Sprintf("\tProtected Branches (%d): %s\n", len(protectedBranches), strings.Join(protectedBranches, ", "))
	}

	if len(unprotectedBranches) > 0 {
		output += fmt.Sprintf("\tUnprotected Branches (%d): %s\n", len(unprotectedBranches), strings.Join(unprotectedBranches, ", "))
	}

	repo, _, err = client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		return err
	}

	mergeMethods := "\tMerge Methods:"
	if repo.GetAllowMergeCommit() {
		mergeMethods += " mergeCommit"
	}
	if repo.GetAllowSquashMerge() {
		mergeMethods += " squash"
	}
	if repo.GetAllowRebaseMerge() {
		mergeMethods += " rebase"
	}
	output += mergeMethods + "\n"

	fmt.Printf("%s--\n\n", output)

	return nil
}
