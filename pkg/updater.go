package pkg

import (
	"fmt"
	"os/exec"
)

// Updater updates a specific module.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_updater.go . Updater
type Updater interface {
	Update(module string) error
}

// GoUpdater uses `go get -u module` to update a specific module.
type GoUpdater struct {
}

func NewGoUpdater() *GoUpdater {
	return &GoUpdater{}
}

// Update updates a dependency using go get in the current working directory.
func (g *GoUpdater) Update(module string) error {
	fmt.Printf("updating dependency for %s\n", module)
	cmd := exec.Command("go", "get", "-u", module)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("update failed, output from command: %s; error: %s", string(output), err)
		return err
	}
	return nil
}
