package pkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Skarlso/dependabot-bundler/pkg/api"
	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	"github.com/Skarlso/dependabot-bundler/pkg/providers"
	"github.com/google/go-github/v43/github"
)

const defaultNumberOfItemsPerPage = 100

// Bundler bundles.
type Bundler struct {
	Config
}

// Config contains dependencies and configuration for the Bundler.
type Config struct {
	Labels       []string
	TargetBranch string
	Owner        string
	Repo         string
	BotName      string
	AuthorName   string
	AuthorEmail  string
	PRTitle      string
	Logger       logger.Logger

	Issues       api.Issues
	Pulls        api.PullRequests
	Git          api.Git
	Updater      providers.Updater
	Repositories api.Repositories
	Runner       providers.Runner
	Signer       providers.Entity
}

// NewBundler creates a new Bundler.
func NewBundler(cfg Config) *Bundler {
	return &Bundler{
		Config: cfg,
	}
}

// Bundle performs the action which bundles together dependabot PRs.
func (n *Bundler) Bundle() error {
	n.Logger.Log("attempting to bundle PRs\n")

	issues, response, err := n.Issues.ListByRepo(context.Background(), n.Owner, n.Repo, &github.IssueListByRepoOptions{
		State:   "open",
		Creator: n.BotName,
		ListOptions: github.ListOptions{
			PerPage: defaultNumberOfItemsPerPage,
		},
	})
	if err != nil {
		return n.logErrorWithBody(err, response.Body)
	}

	var (
		count         int
		prNumbers     string
		modifiedFiles = make(map[string]struct{}) // used for deduplication
	)

	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			pr, _, err := n.Pulls.Get(context.Background(), n.Owner, n.Repo, issue.GetNumber())
			if err != nil {
				n.Logger.Debug("failed to get pull request for number %d with error %s, skipping \n", issue.GetNumber(), err)

				continue
			}
			// The head ref is something like this:
			// dependabot/github_actions/actions/github-script-6.0.0
			// dependabot/go_modules/github.com/aws/aws-sdk-go-v2/service/ssm-1.27.0
			// Which we can use to detect what kind of update we would like to perform.
			files, err := n.Updater.Update(issue.GetBody(), pr.GetHead().GetRef(), pr.GetTitle())
			if err != nil {
				n.Logger.Debug("failed to update %s issue; failure was: %s, skipping...\n", issue.GetTitle(), err)

				continue
			}

			for _, f := range files {
				modifiedFiles[f] = struct{}{}
			}
			count++

			prNumbers += fmt.Sprintf("#%d\n", *issue.Number)
		}
	}

	if count == 0 {
		n.Logger.Log("no pull requests found to bundle, exiting...")

		return nil
	}

	n.Logger.Log("gathered %d pull requests, opening PR...\n", count)
	// open a PR with the modifications
	branch, ref, err := n.getRef()
	if err != nil {
		n.Logger.Log("failed to create ref\n")

		return fmt.Errorf("failed to create ref: %w", err)
	}

	tree, err := n.getTree(modifiedFiles, ref)
	if err != nil {
		n.Logger.Log("failed to get tree\n")

		return fmt.Errorf("failed to get tree: %w", err)
	}

	if err := n.pushCommit(ref, tree); err != nil {
		n.Logger.Log("failed to push commit\n")

		return fmt.Errorf("failed to push commit: %w", err)
	}

	number, err := n.createPR(branch, "Contains the following PRs: \n"+prNumbers, n.PRTitle)
	if err != nil {
		n.Logger.Log("failed to create PR\n")

		return fmt.Errorf("failed to create pr: %w", err)
	}

	if err := n.addLabel(number); err != nil {
		n.Logger.Log("failed to apply labels to the PR: %s\n", n.Labels)

		return fmt.Errorf("failed to add labels: %w", err)
	}

	// clean up each modified file
	for k := range modifiedFiles {
		if output, err := n.Runner.Run("git", ".", "checkout", k); err != nil {
			n.Logger.Log("failed to run clean, skipping... return error and output of clean command: %s; %s",
				err.Error(), string(output))
		}
	}

	n.Logger.Log("PR opened. Thank you for using Bundler, goodbye.\n")

	return nil
}

func (n *Bundler) getRef() (string, *github.Reference, error) {
	var (
		ref     *github.Reference
		err     error
		baseRef *github.Reference
	)

	if baseRef, _, err = n.Git.GetRef(context.Background(), n.Owner, n.Repo, "refs/heads/"+n.TargetBranch); err != nil {
		return "", nil, fmt.Errorf("failed to get ref: %w", err)
	}

	// random generate commit Branch
	commitBranch := n.generateCommitBranch()

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/" + commitBranch),
		Object: &github.GitObject{SHA: baseRef.Object.SHA},
	}

	ref, _, err = n.Git.CreateRef(context.Background(), n.Owner, n.Repo, newRef)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create ref: %w", err)
	}

	return commitBranch, ref, nil
}

func (n *Bundler) generateCommitBranch() string {
	return fmt.Sprintf("bundler-%d", time.Now().UTC().Unix())
}

func (n *Bundler) getTree(files map[string]struct{}, ref *github.Reference) (*github.Tree, error) {
	// Create a tree with what to commit.
	var entries []*github.TreeEntry

	for file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		entries = append(
			entries,
			&github.TreeEntry{
				Path:    github.String(file),
				Type:    github.String("blob"),
				Content: github.String(string(content)),
				Mode:    github.String("100644"),
			},
		)
	}

	tree, _, err := n.Git.CreateTree(context.Background(), n.Owner, n.Repo, *ref.Object.SHA, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree: %w", err)
	}

	return tree, nil
}

// pushCommit creates the commit in the given reference using the given tree.
func (n *Bundler) pushCommit(ref *github.Reference, tree *github.Tree) (err error) {
	// Get the parent commit to attach the commit to.
	parent, _, err := n.Repositories.GetCommit(context.Background(), n.Owner, n.Repo, *ref.Object.SHA, nil)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	commitMessage := "Bundling updated dependencies."
	author := &github.CommitAuthor{Date: &date, Name: &n.AuthorName, Email: &n.AuthorEmail}
	commit := &github.Commit{
		Author:  author,
		Message: &commitMessage,
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}

	// if signing key is provided...
	if n.Signer != nil {
		entity, err := n.Signer.GetEntity()
		if err != nil {
			return fmt.Errorf("failed to get entity for signing details: %w", err)
		}

		commit.SigningKey = entity
	}

	newCommit, _, err := n.Git.CreateCommit(context.Background(), n.Owner, n.Repo, commit)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	ref.Object.SHA = newCommit.SHA
	if _, _, err = n.Git.UpdateRef(context.Background(), n.Owner, n.Repo, ref, false); err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	return nil
}

func (n *Bundler) createPR(commitBranch string, description string, title string) (*int, error) {
	newPR := &github.NewPullRequest{
		Title:               &title,
		Head:                &commitBranch,
		Base:                &n.TargetBranch,
		Body:                &description,
		MaintainerCanModify: github.Bool(true),
	}

	createdPR, _, err := n.Pulls.Create(context.Background(), n.Owner, n.Repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	fmt.Printf("PR created: %s\n", createdPR.GetHTMLURL())

	return createdPR.Number, nil
}

func (n *Bundler) logErrorWithBody(err error, body io.ReadCloser) error {
	content, bodyErr := io.ReadAll(body)
	if bodyErr != nil {
		n.Logger.Log("failed to read body from github response\n")

		return bodyErr
	}

	defer func() {
		if err := body.Close(); err != nil {
			n.Logger.Log("failed to close body\n")
		}
	}()

	n.Logger.Log("got response from github: %s\n", string(content))

	return fmt.Errorf("failed to log response body: %w", err)
}

func (n *Bundler) addLabel(number *int) error {
	// splitting an empty string will result in a 1 len slice with the empty string in it.
	// thus we check early.
	if len(n.Labels) == 0 {
		return nil
	}

	if _, _, err := n.Issues.AddLabelsToIssue(context.Background(), n.Owner, n.Repo, *number, n.Labels); err != nil {
		return fmt.Errorf("failed to add lables to issue: %w", err)
	}

	return nil
}
