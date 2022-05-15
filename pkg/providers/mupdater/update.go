package mupdater

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Skarlso/dependabot-bundler/pkg/providers"
)

var moduleNameRegexp = regexp.MustCompile(`Bumps \[(.*)\]`)

// GoUpdater uses `go get -u module` to update a specific module.
type GoUpdater struct {
	Next providers.Updater
}

func NewGoUpdater(next providers.Updater) *GoUpdater {
	return &GoUpdater{
		Next: next,
	}
}

// Update updates a dependency using go get in the current working directory.
func (g *GoUpdater) Update(body, branch string) ([]string, error) {
	if !strings.Contains(branch, "go_modules") {
		fmt.Printf("go_modules not found in branch name, calling next in line %s\n", branch)
		if g.Next == nil {
			return nil, fmt.Errorf("no Next updater defined")
		}
		return g.Next.Update(body, branch)
	}
	module := g.extractModuleName(body)
	fmt.Printf("updating dependency for %s\n", module)
	cmd := exec.Command("go", "get", "-u", module)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("update failed, output from command: %s; error: %s", string(output), err)
		return nil, err
	}
	return []string{"go.mod", "go.sum"}, nil
}

func (g *GoUpdater) extractModuleName(description string) string {
	matches := moduleNameRegexp.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return ""
	}
	subMatch := matches[0]
	if len(subMatch) < 2 {
		return ""
	}
	return subMatch[1]
}
