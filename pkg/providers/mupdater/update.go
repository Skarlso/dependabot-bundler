package mupdater

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	"github.com/Skarlso/dependabot-bundler/pkg/providers"
)

var (
	moduleNameRegexp = regexp.MustCompile(`Bumps \[(.*)\]`)
	titleRegexp      = regexp.MustCompile(`Bump .* in (.*)`)
)

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
func (g *GoUpdater) Update(body, branch, title string) ([]string, error) {
	if !strings.Contains(branch, "go_modules") {
		if g.Next == nil {
			return nil, fmt.Errorf("no Next updater defined")
		}

		update, err := g.Next.Update(body, branch, title)
		if err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}

		return update, nil
	}

	module := g.extractModuleName(body)
	workdir := g.extractWorkdirLocation(title)

	if workdir == "" {
		workdir = "."
	} else {
		workdir = filepath.Join(".", workdir)
	}

	g.Logger.Log("updating dependency for %s at location %s\n", module, workdir)

	if output, err := g.Runner.Run("go", workdir, "get", "-u", module); err != nil {
		g.Logger.Debug("update failed, output from command: %s; error: %s", string(output), err)

		return nil, fmt.Errorf("failed to run go get: %w", err)
	}

	return []string{"go.mod", "go.sum"}, nil
}

func (g *GoUpdater) extractModuleName(description string) string {
	matches := moduleNameRegexp.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return ""
	}

	subMatch := matches[0]

	const moduleNameIndex = 2
	if len(subMatch) < moduleNameIndex {
		return ""
	}

	return subMatch[1]
}

func (g *GoUpdater) extractWorkdirLocation(title string) string {
	matches := titleRegexp.FindAllStringSubmatch(title, -1)
	if len(matches) == 0 {
		return ""
	}

	subMatch := matches[0]

	const moduleNameIndex = 2
	if len(subMatch) < moduleNameIndex {
		return ""
	}

	return subMatch[1]
}
