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
	files, err := mu.Update("Bumps [github.com/Skarlso/dependabot](https://github.com/Skarlso/dependabot) from 2 to 3", "go_modules")
	assert.NoError(t, err)
	assert.Equal(t, []string{"go.mod", "go.sum"}, files)
	arg, args := fakeRunner.RunArgsForCall(0)
	assert.Equal(t, "go", arg)
	assert.Equal(t, []string{"get", "-u", "github.com/Skarlso/dependabot"}, args)
}

type mockNext struct {
	err error
}

func (m *mockNext) Update(body, branch string) ([]string, error) {
	return nil, nil
}
