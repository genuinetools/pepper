package main

import (
	"context"
	"encoding/json"
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

type outCollaborator struct {
	Login string   `json:"login"`
	Teams []string `json:"teams"`
}

type outCollaborators struct {
	TotalCount int               `json:"totalCount"`
	Admin      []outCollaborator `json:"admin"`
	Write      []outCollaborator `json:"write"`
	Read       []outCollaborator `json:"read"`
}

type outDeployKey struct {
	Title    string `json:"title"`
	ReadOnly bool   `json:"readOnly"`
	URL      string `json:"url"`
}

type outHook struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
	URL    string `json:"url"`
}

type output struct {
	Name                string           `json:"name"`
	Collaborators       outCollaborators `json:"collaborators"`
	DeployKeys          []outDeployKey   `json:"deployKeys"`
	Hooks               []outHook        `json:"hooks"`
	ProtectedBranches   []string         `json:"protectedBranches"`
	UnprotectedBranches []string         `json:"unprotectedBranches"`
	MergeMethods        []string         `json:"mergeMethods"`
}

func outputJSON(o output) {
	b, err := json.Marshal(o)
	if err == nil {
		fmt.Printf("%s\n", b)
	}
}

func outputText(o output) {
	output := fmt.Sprintf("%s -> \n", o.Name)

	if o.Collaborators.TotalCount > 1 {
		push := []string{}
		pull := []string{}
		admin := []string{}
		for _, c := range o.Collaborators.Admin {
			admin = append(admin, fmt.Sprintf("\t\t\t%s (teams: %s)", c.Login, strings.Join(c.Teams, ", ")))
		}
		for _, c := range o.Collaborators.Write {
			push = append(push, fmt.Sprintf("\t\t\t%s", c.Login))
		}
		for _, c := range o.Collaborators.Read {
			pull = append(pull, fmt.Sprintf("\t\t\t%s", c.Login))
		}
		output += fmt.Sprintf("\tCollaborators (%d):\n", o.Collaborators.TotalCount)
		output += fmt.Sprintf("\t\tAdmin (%d):\n%s\n", len(admin), strings.Join(admin, "\n"))
		output += fmt.Sprintf("\t\tWrite (%d):\n%s\n", len(push), strings.Join(push, "\n"))
		output += fmt.Sprintf("\t\tRead (%d):\n%s\n", len(pull), strings.Join(pull, "\n"))
	}

	if len(o.DeployKeys) > 0 {
		kstr := []string{}
		for _, k := range o.DeployKeys {
			kstr = append(kstr, fmt.Sprintf("\t\t%s - ro:%t (%s)", k.Title, k.ReadOnly, k.URL))
		}
		output += fmt.Sprintf("\tKeys (%d):\n%s\n", len(kstr), strings.Join(kstr, "\n"))
	}

	if len(o.Hooks) > 0 {
		hstr := []string{}
		for _, h := range o.Hooks {
			hstr = append(hstr, fmt.Sprintf("\t\t%s - active:%t (%s)", h.Name, h.Active, h.URL))
		}
		output += fmt.Sprintf("\tHooks (%d):\n%s\n", len(hstr), strings.Join(hstr, "\n"))
	}

	if len(o.ProtectedBranches) > 0 {
		output += fmt.Sprintf("\tProtected Branches (%d): %s\n", len(o.ProtectedBranches), strings.Join(o.ProtectedBranches, ", "))
	}

	if len(o.UnprotectedBranches) > 0 {
		output += fmt.Sprintf("\tUnprotected Branches (%d): %s\n", len(o.UnprotectedBranches), strings.Join(o.UnprotectedBranches, ", "))
	}

	output += fmt.Sprintf("\tMerge Methods: %s\n", strings.Join(o.MergeMethods, " "))

	fmt.Printf("%s--\n\n", output)
}

// handleAudit will return nil error if the user does not have access to something.
func handleAudit(ctx context.Context, client *github.Client, repo *github.Repository) error {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	ghteams, resp, err := client.Repositories.ListTeams(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	ghcollabs, resp, err := client.Repositories.ListCollaborators(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.ListCollaboratorsOptions{ListOptions: *opt})
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	ghkeys, resp, err := client.Repositories.ListKeys(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	ghhooks, resp, err := client.Repositories.ListHooks(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		return nil
	}
	if err != nil {
		return err
	}

	ghbranches, _, err := client.Repositories.ListBranches(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if err != nil {
		return err
	}
	protectedBranches := []string{}
	unprotectedBranches := []string{}
	for _, branch := range ghbranches {
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
	if len(ghcollabs) <= 1 && len(ghkeys) < 1 && len(ghhooks) < 1 && len(protectedBranches) < 1 && len(unprotectedBranches) < 1 {
		return nil
	}

	push := []outCollaborator{}
	pull := []outCollaborator{}
	admin := []outCollaborator{}

	if len(ghcollabs) > 1 {
		for _, c := range ghcollabs {
			userTeams := []github.Team{}
			for _, t := range ghteams {
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
				admin = append(admin, outCollaborator{c.GetLogin(), permTeams})
			case perms["push"]:
				permTeams := []string{}
				for _, t := range userTeams {
					if t.GetPermission() == "push" {
						permTeams = append(permTeams, t.GetName())
					}
				}
				push = append(push, outCollaborator{c.GetLogin(), permTeams})
			case perms["pull"]:
				permTeams := []string{}
				for _, t := range userTeams {
					if t.GetPermission() == "pull" {
						permTeams = append(permTeams, t.GetName())
					}
				}
				pull = append(pull, outCollaborator{c.GetLogin(), permTeams})
			}
		}
	}

	keys := []outDeployKey{}
	if len(ghkeys) > 0 {
		for _, k := range ghkeys {
			keys = append(keys, outDeployKey{k.GetTitle(), k.GetReadOnly(), k.GetURL()})
		}
	}

	hooks := []outHook{}
	if len(ghhooks) > 0 {
		for _, h := range ghhooks {
			hooks = append(hooks, outHook{h.GetName(), h.GetActive(), h.GetURL()})
		}
	}

	repo, _, err = client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		return err
	}

	mergeMethods := []string{}
	if repo.GetAllowMergeCommit() {
		mergeMethods = append(mergeMethods, "mergeCommit")
	}
	if repo.GetAllowSquashMerge() {
		mergeMethods = append(mergeMethods, "squash")
	}
	if repo.GetAllowRebaseMerge() {
		mergeMethods = append(mergeMethods, "rebase")
	}

	o := output{
		repo.GetFullName(),
		outCollaborators{
			len(ghcollabs),
			admin,
			push,
			pull,
		},
		keys,
		hooks,
		protectedBranches,
		unprotectedBranches,
		mergeMethods,
	}

	if jsonout {
		outputJSON(o)
	} else {
		outputText(o)
	}

	return nil
}
