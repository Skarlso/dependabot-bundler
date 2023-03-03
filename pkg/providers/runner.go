package providers

// Runner can run commands.
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_runner.go . Runner
type Runner interface {
	Run(command string, args ...string) ([]byte, error)
}
