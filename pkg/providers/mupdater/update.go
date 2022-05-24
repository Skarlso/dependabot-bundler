package mupdater

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	"github.com/Skarlso/dependabot-bundler/pkg/providers"
)

var moduleNameRegexp = regexp.MustCompile(`Bumps \[(.*)\]`)

// GoUpdater uses `go get -u module` to update a specific module.
type GoUpdater struct {
	Next   providers.Updater
	Logger logger.Logger
	Runner providers.Runner
}

func NewGoUpdater(log logger.Logger, next providers.Updater, runner providers.Runner) *GoUpdater {
	return &GoUpdater{
		Next:   next,
		Logger: log,
		Runner: runner,
	}
}

// Update updates a dependency using go get in the current working directory.
func (g *GoUpdater) Update(body, branch string) ([]string, error) {
	if !strings.Contains(branch, "go_modules") {
		if g.Next == nil {
			return nil, fmt.Errorf("no Next updater defined")
		}
		return g.Next.Update(body, branch)
	}
	module := g.extractModuleName(body)
	g.Logger.Log("updating dependency for %s\n", module)
	if output, err := g.Runner.Run("go", "get", "-u", module); err != nil {
		g.Logger.Debug("update failed, output from command: %s; error: %s", string(output), err)
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
