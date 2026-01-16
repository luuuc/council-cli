package cmd

import (
	"fmt"
	"os"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
)

var rootCmd = &cobra.Command{
	Use:   "council",
	Short: "AI-agnostic expert council setup for coding assistants",
	Long: `council-cli helps you create an expert council for AI coding assistants.

The council pattern establishes quality standards through expert personas
that represent excellence in specific domains. The AI suggests experts
based on your project's tech stack.

Quick start:
  council init           Initialize .council/ directory
  council setup --apply  Analyze project and create council with AI assistance
  council sync           Sync council to AI tool configs`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("council %s (%s)\n", version, commit)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new .council directory",
	Long:  `Creates the .council/ directory structure in the current project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initCouncil()
	},
}

func initCouncil() error {
	// Check if already initialized
	if config.Exists() {
		return fmt.Errorf(".council/ already exists")
	}

	// Create directory structure
	dirs := []string{
		config.CouncilDir,
		config.Path(config.ExpertsDir),
		config.Path(config.CommandsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	// Create default config
	cfg := config.Default()
	if err := cfg.Save(); err != nil {
		return err
	}

	// Create .gitkeep files
	for _, subdir := range []string{config.ExpertsDir, config.CommandsDir} {
		path := config.Path(subdir, ".gitkeep")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep: %w", err)
		}
	}

	fmt.Println("Initialized .council/ directory")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  council setup --apply   Analyze project and create council")
	fmt.Println("  council sync            Sync to AI tool configs")

	return nil
}
