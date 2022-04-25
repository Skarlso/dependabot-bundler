package cmd

import (
	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "root",
		Short: "Dependabot bundler action",
		Run:   runRootCmd,
	}
	rootArgs struct {
		token        string
		labels       string
		targetBranch string
	}
)

func init() {
	flag := rootCmd.Flags()
	// Server Configs
	flag.StringVar(&rootArgs.token, "token", "", "--token github token")
	flag.StringVar(&rootArgs.labels, "labels", "", "--labels a list of labels to apply to the PR")
	flag.StringVar(&rootArgs.targetBranch, "target-branch", "", "--target-branch the branch to open the PR against")
}

// runRootCmd runs the main notifier command.
func runRootCmd(cmd *cobra.Command, args []string) {
	// Create the github client here.
	client := github.NewClient()
}

// Execute runs the main krok command.
func Execute() error {
	return rootCmd.Execute()
}
