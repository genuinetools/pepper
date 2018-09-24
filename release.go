package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

const (
	releaseHelp = `Update the release body information.`

	releaseTmpl = `Below are easy install instructions by OS and Architecture. As always there are always the standard instructions in the [README.md](README.md) as well.

<< range $os, $v := . >>
#### << $os  >>

<< range $arch, $r := $v >>
##### << $arch >> - << $os >>

` + "```" + `console
# Export the sha256sum for verification.
$ export << $r.Repository.Name | ToUpper >>_SHA256="<< $r.BinarySHA256 >>"

# Download and check the sha256sum.
$ curl -fSL "<< $r.BinaryURL >>" -o "/usr/local/bin/<< $r.Repository.Name >>" \
	&& echo "` + "${" + `<< $r.Repository.Name | ToUpper >>_SHA256` + "}" + `  /usr/local/bin/<< $r.Repository.Name >>" | sha256sum -c - \
	&& chmod a+x "/usr/local/bin/<< $r.Repository.Name >>"

$ echo "<< $r.Repository.Name >> installed!"

# Run it!
$ << $r.Repository.Name >> -h
` + "```" + `
<<end>>
<<end>>
`
)

func (cmd *releaseCommand) Name() string      { return "release" }
func (cmd *releaseCommand) Args() string      { return "[OPTIONS]" }
func (cmd *releaseCommand) ShortHelp() string { return releaseHelp }
func (cmd *releaseCommand) LongHelp() string  { return releaseHelp }
func (cmd *releaseCommand) Hidden() bool      { return false }

func (cmd *releaseCommand) Register(fs *flag.FlagSet) {}

type releaseCommand struct{}

func (cmd *releaseCommand) Run(ctx context.Context, args []string) error {
	return runCommand(ctx, handleRelease)
}

// handleRelease will return nil error if the user does not have access to something.
func handleRelease(ctx context.Context, client *github.Client, repo *github.Repository) error {
	opt := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	releases, resp, err := client.Repositories.ListReleases(ctx, repo.GetOwner().GetLogin(), repo.GetName(), opt)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden || err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return err
		}

		// Skip it because there is no release.
		return nil
	}
	if err != nil || len(releases) < 1 {
		return err
	}

	rl := release{
		Repository: repo,
	}
	// Get information about the binary assets.
	arch := "linux-amd64"
	for i := 0; i < len(releases); i++ {
		r := releases[i]

		isLatest := false
		if rl.Release == nil && r.GetDraft() {
			// If this is the latest release and it's not a draft make it the one
			// to return
			rl.Release = r
			isLatest = true
		}

		// This holds data like os -> arch -> release and we will use it for rendering our
		// release body template.
		allReleases := map[string]map[string]release{}

		// Iterate over the assets.
		for _, asset := range r.Assets {
			rl.BinaryDownloadCount += asset.GetDownloadCount()

			if !strings.Contains(asset.GetName(), ".") {
				// We know we are on a binary and not a hashsum.
				suffix := strings.SplitN(strings.TrimPrefix(asset.GetName(), repo.GetName()+"-"), "-", 2)
				if len(suffix) == 2 {
					// Add this to our overall releases map.
					osn := suffix[0]
					arch := suffix[1]

					// Prefill the map to avoid a panic.
					if _, ok := allReleases[osn]; !ok {
						allReleases[osn] = map[string]release{}
					}

					tr, ok := allReleases[osn][arch]
					if !ok {
						allReleases[osn][arch] = release{
							BinaryURL:  asset.GetBrowserDownloadURL(),
							BinaryName: asset.GetName(),
							Repository: repo,
						}
					} else {
						tr.BinaryURL = asset.GetBrowserDownloadURL()
						tr.BinaryName = asset.GetName()
						allReleases[osn][arch] = tr
					}
				}
			}

			if strings.HasSuffix(asset.GetName(), ".sha256") {
				// We know we are on a sha256sum.
				suffix := strings.SplitN(strings.TrimSuffix(strings.TrimPrefix(asset.GetName(), repo.GetName()+"-"), ".sha256"), "-", 2)
				if len(suffix) == 2 {
					// Add this to our overall releases map.
					osn := suffix[0]
					arch := suffix[1]

					c, err := getURLContent(asset.GetBrowserDownloadURL())
					if err != nil {
						return err
					}

					// Prefill the map to avoid a panic.
					if _, ok := allReleases[osn]; !ok {
						allReleases[osn] = map[string]release{}
					}

					tr, ok := allReleases[osn][arch]
					if !ok {
						allReleases[osn][arch] = release{
							BinarySHA256: c,
							Repository:   repo,
						}
					} else {
						tr.BinarySHA256 = c
						allReleases[osn][arch] = tr
					}
				}
			}

			if isLatest && strings.HasSuffix(asset.GetName(), arch) {
				rl.BinaryURL = asset.GetBrowserDownloadURL()
				rl.BinaryName = asset.GetName()
				rl.BinarySince = units.HumanDuration(time.Since(asset.GetCreatedAt().Time))
			}

			if isLatest && strings.HasSuffix(asset.GetName(), arch+".sha256") {
				c, err := getURLContent(asset.GetBrowserDownloadURL())
				if err != nil {
					return err
				}
				rl.BinarySHA256 = c
			}

			if isLatest && strings.HasSuffix(asset.GetName(), arch+".md5") {
				c, err := getURLContent(asset.GetBrowserDownloadURL())
				if err != nil {
					return err
				}
				rl.BinaryMD5 = c
			}
		}

		if err := updateRelease(ctx, client, repo, r, allReleases); err != nil {
			return err
		}

		fmt.Printf("Updated release %s/%s for repo: %s", r.GetName(), r.GetTagName(), repo.GetFullName())
		if isLatest {
			// If we updated the latest release, stop.
			break
		}
	}

	return nil
}

type release struct {
	Repository          *github.Repository
	Release             *github.RepositoryRelease
	BinaryName          string
	BinaryURL           string
	BinarySHA256        string
	BinaryMD5           string
	BinaryDownloadCount int
	BinarySince         string
}

func updateRelease(ctx context.Context, client *github.Client, repo *github.Repository, r *github.RepositoryRelease, releases map[string]map[string]release) error {
	var (
		b bytes.Buffer
	)

	// Parse the template.
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	t := template.Must(template.New("").Funcs(funcMap).Delims("<<", ">>").Parse(releaseTmpl))
	w := io.Writer(&b)

	// Execute the template.
	if err := t.Execute(w, releases); err != nil {
		return err
	}

	s := b.String()
	r.Body = &s
	r.Name = r.TagName

	// Send the new body to GitHub to update the release.
	logrus.Debugf("Updating release for %s -> %s...", repo.GetFullName(), r.GetTagName())
	_, resp, err := client.Repositories.EditRelease(ctx, repo.GetOwner().GetLogin(), repo.GetName(), r.GetID(), r)
	if resp.StatusCode == http.StatusForbidden {
		return nil
	}
	return err
}

func getURLContent(uri string) (string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.Split(string(b), " ")[0], nil
}
