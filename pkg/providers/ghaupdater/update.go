package ghaupdater

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Skarlso/dependabot-bundler/pkg/api"
)

var (
	actionNameAndVersionRegexp = regexp.MustCompile(`Bumps \[(.*)\].*from (.*) to ([a-z|A-Z|0-9\.]+)`)
	// this does not include the `v` since in case of a ref, there is no leading `v` after the @ sign.
	actionNamePatter = "uses: %s@%s"
)

// GithubActionUpdater gets the version for the github action being updated and replaces
// every occurrence in every .github/workflows file that the version occurs in.
type GithubActionUpdater struct {
	git api.Git
}

func NewGithubActionUpdater(git api.Git) *GithubActionUpdater {
	return &GithubActionUpdater{
		git: git,
	}
}

// Update updates a dependency using go get in the current working directory.
func (g *GithubActionUpdater) Update(body, branch string) ([]string, error) {
	if !strings.Contains(branch, "github_actions") {
		return nil, fmt.Errorf("github_actions was not in the branch name: %s", branch)
	}

	actionName, from, to := g.extractActionNameAndFromToVersion(body)
	if actionName == "" && from == "" && to == "" {
		return nil, fmt.Errorf("failed to extract action name and from -> to version from description: %s", body)
	}

	to = strings.TrimSuffix(to, ".")

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working folder: %w", err)
	}

	var modifiedFiles []string

	err = filepath.Walk(filepath.Join(cwd, ".github", "workflows"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// don't care about folders
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read file to replace content: %w", err)
				}

				// Gather what the action is pinning to. A SHA or a Tag.
				modifiedFrom, modifiedTo, err := g.getShaOrTag(from, to, actionName, string(content))
				if err != nil {
					return fmt.Errorf("failed to get commit for tag %w", err)
				}

				content = bytes.ReplaceAll(
					content,
					[]byte(fmt.Sprintf(actionNamePatter, actionName, modifiedFrom)),
					[]byte(fmt.Sprintf(actionNamePatter, actionName, modifiedTo)),
				)
				if err := os.WriteFile(path, content, info.Mode()); err != nil {
					return fmt.Errorf("failed to modify file content %w", err)
				}

				// This is the full path. Trim the current working directory from it
				path = strings.TrimPrefix(path, cwd)
				path = strings.TrimPrefix(path, "/")
				modifiedFiles = append(modifiedFiles, path)
			}

			return nil
		})

	if err != nil {
		log.Println(err)
	}

	return modifiedFiles, nil
}

func (g *GithubActionUpdater) extractActionNameAndFromToVersion(description string) (string, string, string) {
	matches := actionNameAndVersionRegexp.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return "", "", ""
	}

	subMatch := matches[0]

	const actionNameAndFromToLength = 3
	if len(subMatch) < actionNameAndFromToLength {
		return "", "", ""
	}

	return subMatch[1], subMatch[2], subMatch[3]
}

// returns the from and to of an action by checking if the action pins to a sha rather than a version.
// it returns the sha of To by fetching the Tag from the description of the dependabot PR and
// gathering the sha which defined that tag.
func (g *GithubActionUpdater) getShaOrTag(from, to, actionName, content string) (string, string, error) {
	fetchPinnedShaOrTag := regexp.MustCompile(fmt.Sprintf(`uses: %s@(.*)`, actionName))

	matches := fetchPinnedShaOrTag.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return "v" + from, "v" + to, nil
	}

	subMatch := matches[0]

	const shaOrVersion = 2
	if len(subMatch) < shaOrVersion {
		return "v" + from, "v" + to, nil
	}

	match := subMatch[1]
	if i := strings.Index(match, " "); i > -1 {
		match = match[:i]
	}

	const shaLength = 40
	if len(match) == shaLength {
		split := strings.Split(actionName, "/")

		const ownerRepoSeparator = 2
		if len(split) < ownerRepoSeparator {
			return "", "", fmt.Errorf("couldn't determine owner and repo from action name: %s", actionName)
		}

		owner, repo := split[0], split[1]

		sha, err := g.extractShaWithOptionalV(owner, repo, to)
		if err != nil {
			return "", "", fmt.Errorf("failed to get ref for tag: %w", err)
		}

		return match, sha, nil
	}

	return "v" + from, "v" + to, nil
}

func (g *GithubActionUpdater) extractShaWithOptionalV(owner, repo, to string) (string, error) {
	ref, resp, err := g.git.GetRef(context.Background(), owner, repo, "tags/"+to)
	if err != nil {
		// we try with a `v` in front of the `to` as well.
		if resp.StatusCode == http.StatusNotFound {
			ref, _, err := g.git.GetRef(context.Background(), owner, repo, "tags/v"+to)
			if err != nil {
				return "", fmt.Errorf("failed to get tag: %w", err)
			}

			return ref.GetObject().GetSHA(), nil
		}

		return "", fmt.Errorf("failed to get tag: %w", err)
	}

	return ref.GetObject().GetSHA(), nil
}
