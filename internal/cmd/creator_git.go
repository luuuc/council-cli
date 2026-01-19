package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/luuuc/council-cli/internal/creator"
	"github.com/spf13/cobra"
)

var (
	publishGitHub  bool
	publishPrivate bool
	publishName    string
)

func init() {
	creatorPublishCmd.Flags().BoolVar(&publishGitHub, "github", false, "Publish to GitHub using gh CLI")
	creatorPublishCmd.Flags().BoolVar(&publishPrivate, "private", false, "Create private repository")
	creatorPublishCmd.Flags().StringVar(&publishName, "name", "my-council", "Repository name")
}

var creatorRemoteCmd = &cobra.Command{
	Use:   "remote [add origin <url>]",
	Short: "Manage git remote",
	Long: `View or configure the git remote for your personal council.

Examples:
  council creator remote                    # Show current remote
  council creator remote add origin <url>   # Add remote`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// No args - show current remote
		if len(args) == 0 {
			url, err := repo.RemoteGet("origin")
			if err != nil {
				fmt.Println("No remote configured")
				fmt.Println()
				fmt.Println("Add one with:")
				fmt.Println("  council creator remote add origin <url>")
				return nil
			}
			fmt.Printf("origin  %s\n", url)
			return nil
		}

		// "add origin <url>"
		if len(args) >= 3 && args[0] == "add" {
			name := args[1]
			url := args[2]

			if repo.RemoteExists(name) {
				return fmt.Errorf("remote '%s' already exists", name)
			}

			if err := repo.RemoteAdd(name, url); err != nil {
				return err
			}

			fmt.Printf("Added remote '%s': %s\n", name, url)
			return nil
		}

		return fmt.Errorf("usage: council creator remote [add <name> <url>]")
	},
}

var creatorPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to remote",
	Long:  `Pushes your personal council to the configured remote repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote exists
		if !repo.RemoteExists("origin") {
			return fmt.Errorf("no remote configured\n\nAdd one with:\n  council creator remote add origin <url>")
		}

		fmt.Println("Pushing to origin...")
		if err := repo.Push(); err != nil {
			return err
		}

		fmt.Println("Done")
		return nil
	},
}

var creatorPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from remote",
	Long:  `Pulls updates from the configured remote repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote exists
		if !repo.RemoteExists("origin") {
			return fmt.Errorf("no remote configured\n\nAdd one with:\n  council creator remote add origin <url>")
		}

		fmt.Println("Pulling from origin...")
		if err := repo.Pull(); err != nil {
			return err
		}

		fmt.Println("Done")
		return nil
	},
}

var creatorPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish your personal council to GitHub",
	Long: `Creates a GitHub repository and pushes your personal council.

Requires the GitHub CLI (gh) to be installed and authenticated.

Examples:
  council creator publish --github           # Create public repo on GitHub
  council creator publish --github --private # Create private repo on GitHub`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote already exists
		if repo.RemoteExists("origin") {
			url, _ := repo.RemoteGet("origin")
			fmt.Printf("Remote already configured: %s\n", url)
			fmt.Println()
			fmt.Println("To push changes:")
			fmt.Println("  council creator push")
			return nil
		}

		if !publishGitHub {
			return fmt.Errorf("specify --github to publish")
		}

		return publishToGitHub(myCouncil, repo)
	},
}

// publishToGitHub creates a GitHub repository and pushes the council.
func publishToGitHub(myCouncil string, repo *creator.GitRepo) error {
	// Check if gh CLI is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) not found\n\nInstall it:\n  brew install gh      # macOS\n  sudo apt install gh  # Ubuntu\n\nThen authenticate:\n  gh auth login")
	}

	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI not authenticated\n\nRun: gh auth login")
	}

	// Build gh repo create command
	visibility := "--public"
	if publishPrivate {
		visibility = "--private"
	}

	fmt.Printf("Creating %s repository '%s'...\n", visibility[2:], publishName)

	cmd = exec.Command("gh", "repo", "create", publishName, visibility, "--source=.", "--push")
	cmd.Dir = myCouncil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	fmt.Println()
	fmt.Println("Published! Your council is now available at:")
	fmt.Printf("  https://github.com/<username>/%s\n", publishName)

	return nil
}
