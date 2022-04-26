package pkg

import (
	"fmt"
	"os/exec"
)

// Updater updates a specific module.
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
	cmd := exec.Command("go", "get", "-u", module)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("update failed, output from command: %s; error: %s", string(output), err)
		return err
	}
	return nil
}
