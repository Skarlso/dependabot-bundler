package pkg_test

import (
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/assert"

	"github.com/Skarlso/dependabot-bundler-action/pkg"
	"github.com/Skarlso/dependabot-bundler-action/pkg/fakes"
)

func TestBundler(t *testing.T) {
	fakeGit := &fakes.FakeGit{}
	fakeRepositories := &fakes.FakeRepositories{}
	fakeIssues := &fakes.FakeIssues{}
	fakePulls := &fakes.FakePullRequests{}
	fakeUpdater := &fakes.FakeUpdater{}
	bundler := pkg.NewBundler(pkg.Config{
		Labels:       "",
		TargetBranch: "main",
		Owner:        "owner",
		Repo:         "repo",
		BotName:      "app/dependabot",
		AuthorName:   "author",
		AuthorEmail:  "author@git.com",
		Issues:       fakeIssues,
		Pulls:        fakePulls,
		Git:          fakeGit,
		Updater:      fakeUpdater,
		Repositories: fakeRepositories,
		Test:         true,
	})

	// setup
	fakeIssues.ListByRepoReturns([]*github.Issue{
		{
			ID:     github.Int64(12355),
			Title:  github.String("Title"),
			Number: github.Int(1),
			State:  github.String("open"),
			Body:   github.String("Bumps [github.com/test/test](github.com/test/test)"),
			User: &github.User{
				Login: github.String("app/dependabot"),
			},
			PullRequestLinks: &github.PullRequestLinks{
				URL: github.String("https://api.github.com/repos/test/test/pulls/5170"),
			},
		},
	}, &github.Response{}, nil)

	baseRef := &github.Reference{
		Ref: github.String("refs/heads/main"),
		URL: github.String("https://github.com/test/test/git/ref/main"),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
			URL:  github.String("https://api.github.com/repos/test/test/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"),
		},
		NodeID: nil,
	}
	commitRef := &github.Reference{
		Ref: github.String("refs/heads/commit-12345"),
		URL: github.String("https://github.com/test/test/git/ref/main"),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
			URL:  github.String("https://api.github.com/repos/test/test/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"),
		},
		NodeID: nil,
	}
	fakeGit.GetRefReturns(baseRef, &github.Response{}, nil)

	fakeGit.CreateRefReturns(commitRef, &github.Response{}, nil)

	fakeGit.CreateTreeReturns(&github.Tree{
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
		Entries: []*github.TreeEntry{
			{
				SHA:     github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
				Path:    github.String("go.mod"),
				Mode:    github.String("100644"),
				Type:    github.String("blob"),
				Content: github.String("change"),
			},
		},
	}, &github.Response{}, nil)

	fakeRepositories.GetCommitReturns(&github.RepositoryCommit{
		SHA:    github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
		Commit: &github.Commit{},
	}, nil, nil)

	fakeGit.CreateCommitReturns(&github.Commit{
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}, nil, nil)

	fakeGit.UpdateRefReturns(nil, nil, nil)

	fakePulls.CreateReturns(&github.PullRequest{
		HTMLURL: github.String("https://github.com/test/test/pulls/1"),
	}, nil, nil)

	assert.NoError(t, bundler.Bundle())

	moduleName := fakeUpdater.UpdateArgsForCall(0)
	assert.Equal(t, "github.com/test/test", moduleName)
}
