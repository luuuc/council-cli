package cmd

import (
	"fmt"
	"os/exec"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/sync"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check council health and diagnose issues",
	Long:  `Verifies your council setup and reports any issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor()
	},
}

func runDoctor() error {
	fmt.Println("Council Doctor")
	fmt.Println("==============")
	fmt.Println()

	allGood := true

	// Check 1: .council/ directory exists
	if config.Exists() {
		printCheck(true, ".council/ directory exists")
	} else {
		printCheck(false, ".council/ directory exists")
		fmt.Println("     Run 'council init' to create it")
		allGood = false
	}

	// Check 2: Config file
	cfg, err := config.Load()
	if err == nil {
		printCheck(true, "config.yaml is valid")
	} else {
		printCheck(false, "config.yaml is valid")
		fmt.Printf("     %v\n", err)
		allGood = false
	}

	// Check 3: Experts loaded
	result, err := expert.ListWithWarnings()
	if err != nil {
		printCheck(false, "Experts directory readable")
		fmt.Printf("     %v\n", err)
		allGood = false
	} else {
		count := len(result.Experts)
		if count > 0 {
			printCheck(true, fmt.Sprintf("%d expert(s) loaded", count))
			for _, e := range result.Experts {
				fmt.Printf("       - %s (%s)\n", e.Name, e.ID)
			}
		} else {
			printCheck(false, "No experts found")
			fmt.Println("     Run 'council setup --interactive' or 'council add <name>'")
			allGood = false
		}

		// Show warnings for any files that couldn't be loaded
		for _, w := range result.Warnings {
			printCheck(false, w)
			allGood = false
		}
	}

	// Check 4: Sync targets
	if cfg != nil {
		fmt.Println()
		fmt.Println("Sync targets:")
		for _, targetName := range cfg.Targets {
			target, ok := sync.Targets[targetName]
			if !ok {
				printCheck(false, fmt.Sprintf("%s (unknown target)", targetName))
				allGood = false
				continue
			}

			// Check if target location exists
			exists := target.Check()
			if exists {
				printCheck(true, fmt.Sprintf("%s (%s)", target.Name, target.Location))
			} else {
				printOptional(fmt.Sprintf("%s (%s) - not synced yet", target.Name, target.Location))
			}
		}
	}

	// Check 5: AI CLI (optional)
	if cfg != nil && cfg.AI.Command != "" {
		fmt.Println()
		fmt.Println("AI integration:")
		if _, err := exec.LookPath(cfg.AI.Command); err == nil {
			printCheck(true, fmt.Sprintf("'%s' command available", cfg.AI.Command))
		} else {
			printOptional(fmt.Sprintf("'%s' not found (optional, for 'council setup --apply')", cfg.AI.Command))
		}
	}

	// Summary
	fmt.Println()
	if allGood {
		fmt.Println("Your council is healthy!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  council sync    Sync to AI tool configs")
		fmt.Println("  /council        Use in Claude Code or Cursor")
	} else {
		fmt.Println("Some issues found. See above for details.")
	}

	return nil
}

func printCheck(ok bool, msg string) {
	if ok {
		fmt.Printf("  [ok] %s\n", msg)
	} else {
		fmt.Printf("  [!!] %s\n", msg)
	}
}

func printOptional(msg string) {
	fmt.Printf("  [--] %s\n", msg)
}
