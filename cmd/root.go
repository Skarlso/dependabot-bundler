package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/Skarlso/dependabot-bundler/pkg"
	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	ghau "github.com/Skarlso/dependabot-bundler/pkg/providers/ghaupdater"
	mu "github.com/Skarlso/dependabot-bundler/pkg/providers/mupdater"
	"github.com/Skarlso/dependabot-bundler/pkg/providers/runner"
)

var (
	rootCmd = &cobra.Command{
		Use:   "root",
		Short: "Dependabot bundler action",
		Run:   runRootCmd,
	}
	rootArgs struct {
		botName      string
		token        string
		owner        string
		repo         string
		labels       []string
		targetBranch string
		authorName   string
		authorEmail  string
		prTitle      string
		verbose      bool
	}
)

func init() {
	flag := rootCmd.Flags()
	// Server Configs
	flag.StringVar(&rootArgs.token, "token", "", "--token github token")
	flag.StringVar(&rootArgs.owner, "owner", "", "--owner github organization / owner")
	flag.StringVar(&rootArgs.repo, "repo", "", "--repo github repository")
	flag.StringSliceVar(&rootArgs.labels, "labels", nil, "--labels a list of labels to apply to the PR")
	flag.StringVar(&rootArgs.botName, "bot-name", "app/dependabot", "--bot-name the name of the bot, default is app/dependabot")
	flag.StringVar(&rootArgs.authorName, "author-name", "Github Action", "--author-name name of the committer, default to Github Action")
	flag.StringVar(&rootArgs.authorEmail, "author-email", "41898282+github-actions[bot]@users.noreply.github.com", "--author-email email address of the committer, defaults to github action's email address")
	flag.StringVar(&rootArgs.targetBranch, "target-branch", "main", "--target-branch the branch to open the PR against")
	flag.StringVar(&rootArgs.prTitle, "pr-title", "Dependabot Bundler PR", "--pr-title the title of the PR that will be created")
	flag.BoolVarP(&rootArgs.verbose, "verbose", "v", false, "--verbose|-v if enabled, will output extra debug information")
}

// runRootCmd runs the main notifier command.
func runRootCmd(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: rootArgs.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// setup logger
	var log logger.Logger = &logger.QuiteLogger{}
	if rootArgs.verbose {
		log = &logger.VerboseLogger{}
	}

	// setup GitHub actions updater
	actionsUpdater := ghau.NewGithubActionUpdater(client.Git)

	// setup modules updater
	osRunner := runner.NewOsRunner()
	updater := mu.NewGoUpdater(log, actionsUpdater, osRunner)

	bundler := pkg.NewBundler(pkg.Config{
		Labels:       rootArgs.labels,
		TargetBranch: rootArgs.targetBranch,
		Owner:        rootArgs.owner,
		Repo:         rootArgs.repo,
		BotName:      rootArgs.botName,
		AuthorEmail:  rootArgs.authorEmail,
		AuthorName:   rootArgs.authorName,
		PRTitle:      rootArgs.prTitle,
		Issues:       client.Issues,
		Pulls:        client.PullRequests,
		Git:          client.Git,
		Repositories: client.Repositories,
		Updater:      updater,
		Logger:       log,
		Runner:       osRunner,
	})
	if err := bundler.Bundle(); err != nil {
		fmt.Printf("failed to bundle PRs: %s\n", err)
		os.Exit(1)
	}
}

// Execute runs the main bundler command.
func Execute() error {
	return rootCmd.Execute()
}
