package ghaupdater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/assert"

	"github.com/Skarlso/dependabot-bundler/pkg/api/fakes"
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
	git := &fakes.FakeGit{}
	gau := NewGithubActionUpdater(git)
	files, err := gau.Update("Bumps [actions/checkout](https://github.com/actions/checkout) from 2 to 3", "github_actions")
	assert.NoError(t, err)
	assert.Equal(t, []string{".github/workflows/test.yaml"}, files)
	newContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Contains(t, string(newContent), "uses: actions/checkout@v3")
}

func TestWithGitSHAPin(t *testing.T) {
	testFile := filepath.Join(".github", "workflows", "test.yaml")
	previousContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	// put back previous content
	defer func() {
		err = os.WriteFile(testFile, previousContent, 0777)
		assert.NoError(t, err)
	}()

	git := &fakes.FakeGit{}
	git.GetRefReturns(&github.Reference{
		Ref: github.String("refs/tags/4.0.1"),
		URL: github.String("https://github.com/skarlso/test"),
		Object: &github.GitObject{
			SHA: github.String("69f6fc9d46f2f8bf0d5491e4aabe0bb8c6a4678a"),
		},
	}, &github.Response{}, nil)
	gau := NewGithubActionUpdater(git)
	files, err := gau.Update("Bumps [docker/metadata-action](https://github.com/docker/metadata-action) from 3.3.0 to 4.0.1.", "github_actions")
	assert.NoError(t, err)
	assert.Equal(t, []string{".github/workflows/test.yaml"}, files)
	newContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Contains(t, string(newContent), "uses: docker/metadata-action@69f6fc9d46f2f8bf0d5491e4aabe0bb8c6a4678a")
}

func TestNameUpdateInvalidBranch(t *testing.T) {
	git := &fakes.FakeGit{}
	gau := NewGithubActionUpdater(git)
	_, err := gau.Update("Bumps [actions/checkout](https://github.com/actions/checkout) from 2 to 3", "invalid")
	assert.EqualError(t, err, "github_actions was not in the branch name: invalid")
}

func TestNameUpdateInvalidDescription(t *testing.T) {
	git := &fakes.FakeGit{}
	gau := NewGithubActionUpdater(git)
	_, err := gau.Update("invalid", "github_actions")
	assert.EqualError(t, err, "failed to extract action name and from -> to version from description: invalid")
}
