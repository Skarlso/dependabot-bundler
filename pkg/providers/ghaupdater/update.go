package ghaupdater

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// don't forget to trim the `.` at the end.
var (
	actionNameAndVersionRegexp = regexp.MustCompile(`Bumps \[(.*)\].*from (.*) to (.*)`)
	actionNamePatter           = "uses: %s@v%s"
)

// GithubActionUpdater gets the version for the github action being updated and replaces
// every occurrence in every .github/workflows file that the version occurs in.
type GithubActionUpdater struct {
}

func NewGithubActionUpdater() *GithubActionUpdater {
	return &GithubActionUpdater{}
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
				// skip if it does not contain the action we are updating
				// so that we don't stage this file.
				if !bytes.Contains(content, []byte(actionName)) {
					return nil
				}
				content = bytes.ReplaceAll(
					content,
					[]byte(fmt.Sprintf(actionNamePatter, actionName, from)),
					[]byte(fmt.Sprintf(actionNamePatter, actionName, to)),
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
	if len(subMatch) < 3 {
		return "", "", ""
	}
	return subMatch[1], subMatch[2], subMatch[3]
}
