package mupdater

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	"github.com/Skarlso/dependabot-bundler/pkg/providers/fakes"
)

func TestNewGoUpdater(t *testing.T) {
	fakeRunner := &fakes.FakeRunner{}
	mockNext := &mockNext{}
	mu := NewGoUpdater(&logger.QuiteLogger{}, mockNext, fakeRunner)
	files, err := mu.Update("Bumps [github.com/Skarlso/dependabot](https://github.com/Skarlso/dependabot) from 2 to 3", "go_modules", "chore(deps): Bump golang.org/x/sys from 0.0.0-20211013075003-97ac67df715c to 0.1.0")
	assert.NoError(t, err)
	assert.Equal(t, []string{"go.mod", "go.sum"}, files)
	arg, workdir, args := fakeRunner.RunArgsForCall(0)
	assert.Equal(t, "go", arg)
	assert.Equal(t, []string{"get", "-u", "github.com/Skarlso/dependabot"}, args)
	assert.Equal(t, ".", workdir)
}

func TestNewGoUpdaterWithLongerVersion(t *testing.T) {
	fakeRunner := &fakes.FakeRunner{}
	mockNext := &mockNext{}
	mu := NewGoUpdater(&logger.QuiteLogger{}, mockNext, fakeRunner)
	files, err := mu.Update("Bumps [golang.org/x/sys](https://github.com/golang/sys) from 0.0.0-20200323222414-85ca7c5b95cd to 0.1.0.", "go_modules", "chore(deps): Bump golang.org/x/sys from 0.0.0-20211013075003-97ac67df715c to 0.1.0")
	assert.NoError(t, err)
	assert.Equal(t, []string{"go.mod", "go.sum"}, files)
	arg, workdir, args := fakeRunner.RunArgsForCall(0)
	assert.Equal(t, "go", arg)
	assert.Equal(t, []string{"get", "-u", "golang.org/x/sys"}, args)
	assert.Equal(t, ".", workdir)
}

func TestNewGoUpdaterWithLongerVersionWithInFolder(t *testing.T) {
	fakeRunner := &fakes.FakeRunner{}
	mockNext := &mockNext{}
	mu := NewGoUpdater(&logger.QuiteLogger{}, mockNext, fakeRunner)
	files, err := mu.Update("Bumps [golang.org/x/sys](https://github.com/golang/sys) from 0.0.0-20200323222414-85ca7c5b95cd to 0.1.0.", "go_modules", "chore(deps): Bump golang.org/x/sys from 0.0.0-20211013075003-97ac67df715c to 0.1.0 in /hack/tools")
	assert.NoError(t, err)
	assert.Equal(t, []string{"hack/tools/go.mod", "hack/tools/go.sum"}, files)
	arg, workdir, args := fakeRunner.RunArgsForCall(0)
	assert.Equal(t, "go", arg)
	assert.Equal(t, []string{"get", "-u", "golang.org/x/sys"}, args)
	assert.Equal(t, "hack/tools", workdir)
}

type mockNext struct {
	err error
}

func (m *mockNext) Update(body, branch, title string) ([]string, error) {
	return nil, nil
}
