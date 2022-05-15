package mupdater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGoUpdater(t *testing.T) {
	t.Skip("mock the executer...")
	mu := NewGoUpdater(nil)
	files, err := mu.Update("Bumps [github.com/actions/checkout](https://github.com/Skarlso/dependabot) from 2 to 3", "go_modules")
	assert.NoError(t, err)
	assert.Equal(t, []string{"go.mod", "go.sum"}, files)
}

type mockNext struct {
	err error
}

func (m *mockNext) Update(body, branch string) ([]string, error) {
	return nil, nil
}
