package api

import (
	"context"

	"github.com/google/go-github/v43/github"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// PullRequests defines the GitHub client's pullRequest service.
// s
//
//counterfeiter:generate -o fakes/fake_github_client_pulls.go . PullRequests
type PullRequests interface {
	Create(
		ctx context.Context,
		owner string,
		repo string,
		pull *github.NewPullRequest,
	) (*github.PullRequest, *github.Response, error)
	Get(
		ctx context.Context,
		owner string,
		repo string,
		number int,
	) (*github.PullRequest, *github.Response, error)
}

// Issues defines the GitHub client's issues service.
//
//counterfeiter:generate -o fakes/fake_github_client_issues.go . Issues
type Issues interface {
	AddLabelsToIssue(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		labels []string,
	) ([]*github.Label, *github.Response, error)
	ListByRepo(
		ctx context.Context,
		owner string,
		repo string,
		opts *github.IssueListByRepoOptions,
	) ([]*github.Issue, *github.Response, error)
}

// Git defines the GitHub client's git service.
//
//counterfeiter:generate -o fakes/fake_github_client_git.go . Git
type Git interface {
	GetRef(
		ctx context.Context,
		owner string,
		repo string,
		ref string,
	) (*github.Reference, *github.Response, error)
	CreateCommit(
		ctx context.Context,
		owner string,
		repo string,
		commit *github.Commit,
	) (*github.Commit, *github.Response, error)
	CreateRef(
		ctx context.Context,
		owner string,
		repo string,
		ref *github.Reference,
	) (*github.Reference, *github.Response, error)
	CreateTree(
		ctx context.Context,
		owner string,
		repo string,
		baseTree string,
		entries []*github.TreeEntry,
	) (*github.Tree, *github.Response, error)
	UpdateRef(
		ctx context.Context,
		owner string,
		repo string,
		ref *github.Reference,
		force bool,
	) (*github.Reference, *github.Response, error)
}

// Repositories defines the GitHub client's repositories service.
//
//counterfeiter:generate -o fakes/fake_github_client_repository.go . Repositories
type Repositories interface {
	GetCommit(
		ctx context.Context,
		owner, repo, sha string,
		opts *github.ListOptions,
	) (*github.RepositoryCommit, *github.Response, error)
}
