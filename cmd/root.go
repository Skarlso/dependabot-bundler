package cmd

import (
	"context"
	"fmt"

	"github.com/Skarlso/dependabot-bundler/pkg"
	"github.com/Skarlso/dependabot-bundler/pkg/logger"
	ghau "github.com/Skarlso/dependabot-bundler/pkg/providers/ghaupdater"
	mu "github.com/Skarlso/dependabot-bundler/pkg/providers/mupdater"
	"github.com/Skarlso/dependabot-bundler/pkg/providers/pgp"
	"github.com/Skarlso/dependabot-bundler/pkg/providers/runner"
	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const defaultKeyBitLength = 4096

type rootArgsStruct struct {
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
	pgp          struct {
		name       string
		email      string
		publicKey  string
		privateKey string
		bitLength  int
		passphrase string
	}
}

func CreateRootCommand() *cobra.Command {
	rootArgs := &rootArgsStruct{}

	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Dependabot bundler action",
	}

	flag := rootCmd.Flags()

	// Server Configs
	flag.StringVar(&rootArgs.token, "token", "", "--token github token")
	flag.StringVar(&rootArgs.owner, "owner", "", "--owner github organization / owner")
	flag.StringVar(&rootArgs.repo, "repo", "", "--repo github repository")
	flag.StringSliceVar(
		&rootArgs.labels,
		"labels",
		nil,
		"--labels a list of labels to apply to the PR",
	)
	flag.StringVar(
		&rootArgs.botName,
		"bot-name",
		"app/dependabot",
		"--bot-name the name of the bot, default is app/dependabot",
	)
	flag.StringVar(
		&rootArgs.authorName,
		"author-name",
		"Github Action",
		"--author-name name of the committer, default to Github Action",
	)
	flag.StringVar(
		&rootArgs.authorEmail,
		"author-email",
		"41898282+github-actions[bot]@users.noreply.github.com",
		"--author-email email address of the committer, defaults to github action's email address",
	)
	flag.StringVar(
		&rootArgs.targetBranch,
		"target-branch",
		"main",
		"--target-branch the branch to open the PR against",
	)
	flag.StringVar(
		&rootArgs.prTitle,
		"pr-title",
		"Dependabot Bundler PR",
		"--pr-title the title of the PR that will be created",
	)
	flag.BoolVarP(
		&rootArgs.verbose,
		"verbose",
		"v",
		false,
		"--verbose|-v if enabled, will output extra debug information",
	)
	flag.StringVar(
		&rootArgs.pgp.name,
		"signing-name",
		"",
		"--signing-name the name used for the signing key",
	)
	flag.StringVar(
		&rootArgs.pgp.email,
		"signing-email",
		"",
		"--signing-email the email of the signing key",
	)
	flag.StringVar(
		&rootArgs.pgp.publicKey,
		"signing-public-key",
		"",
		"--signing-public-key the public key of the pgp signing key",
	)
	flag.StringVar(
		&rootArgs.pgp.privateKey,
		"signing-private-key",
		"",
		"--signing-private-key the private key of the pgp signing key",
	)
	flag.IntVar(
		&rootArgs.pgp.bitLength,
		"signing-key-bit-length",
		defaultKeyBitLength,
		"--signing-key-bit-length the length of the key",
	)
	flag.StringVar(
		&rootArgs.pgp.passphrase,
		"signing-key-passphrase",
		"",
		"--signing-key-passphrase the passphrase to use for the signing key",
	)

	rootCmd.RunE = rootRunE(rootArgs)

	return rootCmd
}

func rootRunE(rootArgs *rootArgsStruct) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
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

		if rootArgs.pgp.publicKey != "" {
			signer := &pgp.Entity{
				Name:       rootArgs.pgp.name,
				Email:      rootArgs.pgp.email,
				BitSize:    rootArgs.pgp.bitLength,
				PublicKey:  []byte(rootArgs.pgp.publicKey),
				PrivateKey: []byte(rootArgs.pgp.privateKey),
				Passphrase: []byte(rootArgs.pgp.passphrase),
			}
			bundler.Signer = signer
		}

		if err := bundler.Bundle(); err != nil {
			return fmt.Errorf("failed to execute bundler: %w", err)
		}

		return nil
	}
}
