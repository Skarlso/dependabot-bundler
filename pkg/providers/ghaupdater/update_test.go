package ghaupdater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameUpdate(t *testing.T) {
	testFile := filepath.Join(".github", "workflows", "test.yaml")
	previousContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	// put back previous content
	defer func() {
		err = os.WriteFile(testFile, previousContent, 0777)
		assert.NoError(t, err)
	}()

	gau := NewGithubActionUpdater()
	files, err := gau.Update("Bumps [actions/checkout](https://github.com/actions/checkout) from 2 to 3", "github_actions")
	assert.NoError(t, err)
	assert.Equal(t, []string{".github/workflows/test.yaml"}, files)
	newContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(newContent), "uses: actions/checkout@v3"))
}

func TestNameUpdateInvalidBranch(t *testing.T) {
	gau := NewGithubActionUpdater()
	_, err := gau.Update("Bumps [actions/checkout](https://github.com/actions/checkout) from 2 to 3", "invalid")
	assert.EqualError(t, err, "github_actions was not in the branch name: invalid")
}

func TestNameUpdateInvalidDescription(t *testing.T) {
	gau := NewGithubActionUpdater()
	_, err := gau.Update("invalid", "github_actions")
	assert.EqualError(t, err, "failed to extract action name and from -> to version from description: invalid")
}
