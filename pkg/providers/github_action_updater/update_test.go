package github_action_updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameUpdate(t *testing.T) {
	gau := NewGithubActionUpdater()
	files, err := gau.Update("Bumps [actions/checkout](https://github.com/actions/checkout) from 2 to 3", "github_actions")
	assert.NoError(t, err)
	assert.Equal(t, []string{".github/workflows/test.yaml"}, files)
}
