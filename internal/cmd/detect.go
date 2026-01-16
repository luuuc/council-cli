package cmd

import (
	"fmt"
	"os"

	"github.com/luuuc/council-cli/internal/detect"
	"github.com/spf13/cobra"
)

var detectJSON bool

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output as JSON")
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect project languages and frameworks",
	Long:  `Scans the current directory to detect languages, frameworks, testing tools, and patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		d, err := detect.Scan(dir)
		if err != nil {
			return err
		}

		if detectJSON {
			data, err := d.JSON()
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable output
		fmt.Println("Detected stack:")
		fmt.Println()

		if len(d.Languages) > 0 {
			fmt.Println("Languages:")
			for _, lang := range d.Languages {
				fmt.Printf("  %s (%.1f%%)\n", lang.Name, lang.Percentage)
			}
			fmt.Println()
		}

		if len(d.Frameworks) > 0 {
			fmt.Println("Frameworks:")
			for _, fw := range d.Frameworks {
				if fw.Version != "" {
					fmt.Printf("  %s %s\n", fw.Name, fw.Version)
				} else {
					fmt.Printf("  %s\n", fw.Name)
				}
			}
			fmt.Println()
		}

		if len(d.Testing) > 0 {
			fmt.Println("Testing:")
			for _, t := range d.Testing {
				fmt.Printf("  %s\n", t)
			}
			fmt.Println()
		}

		if len(d.Patterns) > 0 {
			fmt.Println("Patterns:")
			for _, p := range d.Patterns {
				fmt.Printf("  %s\n", p)
			}
			fmt.Println()
		}

		return nil
	},
}
