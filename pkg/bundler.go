package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/google/go-github/v43/github"
)

var moduleNameRegexp = regexp.MustCompile(`Bumps \[(.*)\]`)

// Todo: consider using git command line instead, since that's already set up?

// PullRequests defines the GitHub client's pullRequest service.
//go:generate counterfeiter -o fakes/fake_github_client_pull_requests.go . PullRequests
type PullRequests interface {
	Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

type Issues interface {
	ListByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
}

type Git interface {
	GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateCommit(ctx context.Context, owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
	CreateTree(ctx context.Context, owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error)
	UpdateRef(ctx context.Context, owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error)
}

type Repositories interface {
	GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
}

// Bundler bundles.
type Bundler struct {
	Config
}

type Config struct {
	Labels       string
	TargetBranch string
	Owner        string
	Repo         string
	BotName      string
	AuthorName   string
	AuthorEmail  string

	Issues       Issues
	Pulls        PullRequests
	Git          Git
	Updater      Updater
	Repositories Repositories
}

// NewBundler creates a new Slack notifier.
func NewBundler(cfg Config) *Bundler {
	return &Bundler{
		Config: cfg,
	}
}

func (n *Bundler) Bundle() error {
	fmt.Printf("attempting to bundle PRs")
	issues, response, err := n.Issues.ListByRepo(context.Background(), n.Owner, n.Repo, &github.IssueListByRepoOptions{
		State:   "open",
		Creator: n.BotName,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return n.logErrorWithBody(err, response.Body)
	}

	// filter pull requests
	var (
		count     int
		prNumbers string
	)
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			moduleName := n.extractModuleName(*issue.Body)
			if moduleName == "" {
				fmt.Printf("skipping issue %s as no module name was found in description", *issue.Title)
				continue
			}
			if err := n.Updater.Update(moduleName); err != nil {
				fmt.Printf("failed to update %s issue; failure was: %s, skipping...", *issue.Title, err)
				continue
			}
			count++
			prNumbers += fmt.Sprintf("#%d\n", *issue.ID)
		}
	}

	if count == 0 {
		fmt.Println("no pull requests found to bundle, exiting...")
		return nil
	}
	fmt.Printf("gathered %d pull requests, opening PR...", count)
	// open a PR with the modifications
	branch, ref, err := n.getRef()
	if err != nil {
		fmt.Printf("failed to create ref")
		return err
	}

	tree, err := n.getTree(ref)
	if err != nil {
		fmt.Printf("failed to get tree")
		return err
	}

	if err := n.pushCommit(ref, tree); err != nil {
		fmt.Printf("failed to push commit")
		return err
	}

	if err := n.createPR(branch, "Bundling together prs: "+prNumbers, "Bundling dependabot PRs"); err != nil {
		fmt.Printf("failed to create PR")
		return err
	}

	fmt.Printf("PR opened. Thank you for using Bundler, goodbye.")
	return nil
}

func (n *Bundler) getRef() (branch string, ref *github.Reference, err error) {
	var baseRef *github.Reference
	if baseRef, _, err = n.Git.GetRef(context.Background(), n.Owner, n.Repo, "refs/heads/"+n.TargetBranch); err != nil {
		return "", nil, err
	}
	// random generate commit Branch
	var commitBranch string
	newRef := &github.Reference{Ref: github.String("refs/heads/" + commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = n.Git.CreateRef(context.Background(), n.Owner, n.Repo, newRef)
	return commitBranch, ref, err
}

func (n *Bundler) getTree(ref *github.Reference) (*github.Tree, error) {
	// Create a tree with what to commit.
	var entries []*github.TreeEntry

	// We only ever add the mod and sum file. We never commit anything else.
	// This prevents us from creating prs which contain unrelated changes to the update.
	for _, file := range []string{"go.mod", "go.sum"} {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(b)), Mode: github.String("100644")})
	}

	tree, _, err := n.Git.CreateTree(context.Background(), n.Owner, n.Repo, *ref.Object.SHA, entries)
	return tree, err
}

// pushCommit creates the commit in the given reference using the given tree.
func (n *Bundler) pushCommit(ref *github.Reference, tree *github.Tree) (err error) {
	// Get the parent commit to attach the commit to.
	parent, _, err := n.Repositories.GetCommit(context.Background(), n.Owner, n.Repo, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	commitMessage := "Bundling updated dependencies."
	author := &github.CommitAuthor{Date: &date, Name: &n.AuthorName, Email: &n.AuthorEmail}
	commit := &github.Commit{Author: author, Message: &commitMessage, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := n.Git.CreateCommit(context.Background(), n.Owner, n.Repo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = n.Git.UpdateRef(context.Background(), n.Owner, n.Repo, ref, false)
	return err
}

func (n *Bundler) createPR(commitBranch string, description string, title string) error {
	newPR := &github.NewPullRequest{
		Title:               &title,
		Head:                &commitBranch,
		Base:                &n.TargetBranch,
		Body:                &description,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := n.Pulls.Create(context.Background(), n.Owner, n.Repo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("PR created: %s\n", pr.GetHTMLURL())
	return nil
}

func (n *Bundler) extractModuleName(description string) string {
	matches := moduleNameRegexp.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return ""
	}
	subMatch := matches[0]
	if len(subMatch) < 2 {
		return ""
	}
	return subMatch[1]
}

func (n *Bundler) logErrorWithBody(err error, body io.ReadCloser) error {
	content, err := io.ReadAll(body)
	if err != nil {
		fmt.Printf("failed to read body from github response")
		return err
	}
	defer func() {
		if err := body.Close(); err != nil {
			fmt.Printf("failed to close body")
		}
	}()

	fmt.Printf("got response from github: %s", string(content))
	return err
}
