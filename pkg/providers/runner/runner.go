package runner

import "os/exec"

// OsRunner runs commands using the operating system.
type OsRunner struct {
	Dir string
}

// NewOsRunner creates an operating system based runner.
func NewOsRunner() *OsRunner {
	return &OsRunner{}
}

// Run takes a command and it's arguments and runs it locally.
func (r *OsRunner) Run(command, workdir string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workdir

	return cmd.CombinedOutput()
}
