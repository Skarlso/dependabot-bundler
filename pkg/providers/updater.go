package providers

// Updater updates a specific module. Returns a list of modified files.
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_updater.go . Updater
type Updater interface {
	Update(module, branch string) ([]string, error)
}
