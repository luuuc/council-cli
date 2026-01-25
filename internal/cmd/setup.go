package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/detect"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	setupApply  bool
	setupOutput string
	setupYes    bool
)

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupApply, "apply", false, "Send prompt to AI and apply suggestions")
	setupCmd.Flags().StringVarP(&setupOutput, "output", "o", "", "Write prompt to file instead of stdout")
	setupCmd.Flags().BoolVarP(&setupYes, "yes", "y", false, "Skip confirmation when applying")
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Analyze project and generate AI prompt for expert suggestions",
	Long: `Scans your project, detects the tech stack, and generates a prompt
for an AI assistant to suggest appropriate expert personas.

Modes:
  (default)       Output prompt for you to copy to any AI
  --apply         Send prompt to configured AI CLI and create experts (deprecated)

NOTE: 'council setup --apply' is deprecated. Use 'council start' instead
for zero-config setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		// Run detection
		d, err := detect.Scan(dir)
		if err != nil {
			return fmt.Errorf("detection failed: %w", err)
		}

		// Show detection summary
		fmt.Fprintf(os.Stderr, "Detected: %s\n\n", d.Summary())

		// Generate prompt
		promptText, err := prompt.Generate(d)
		if err != nil {
			return fmt.Errorf("prompt generation failed: %w", err)
		}

		if setupApply {
			// Print deprecation warning
			fmt.Fprintln(os.Stderr, "Warning: 'council setup --apply' is deprecated.")
			fmt.Fprintln(os.Stderr, "         Use 'council start' for zero-config setup.")
			fmt.Fprintln(os.Stderr, "         This command will be removed in a future version.")
			fmt.Fprintln(os.Stderr, "")
			return runSetupApply(promptText)
		}

		// Output prompt
		if setupOutput != "" {
			if err := os.WriteFile(setupOutput, []byte(promptText), 0644); err != nil {
				return fmt.Errorf("failed to write prompt: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Prompt written to %s\n", setupOutput)
			return nil
		}

		fmt.Println("Copy this prompt to your AI assistant:")
		fmt.Println("---")
		fmt.Println(promptText)
		fmt.Println("---")
		fmt.Println()
		fmt.Println("Then run: council setup --apply < response.yaml")
		fmt.Println("Or use:   council setup --apply  (to send to configured AI)")

		return nil
	},
}

func runSetupApply(promptText string) error {
	// Check for stdin input first (piped YAML response)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		return applyFromStdin()
	}

	// Load config to get AI command
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHint: run 'council init' first", err)
	}

	// Detect or use configured AI command
	aiCmd, err := cfg.DetectAICommand()
	if err != nil {
		return err
	}

	// Check if command exists
	if _, err := exec.LookPath(aiCmd); err != nil {
		return fmt.Errorf("AI command '%s' not found\n\nInstall it or configure a different command in .council/config.yaml", aiCmd)
	}

	// Execute AI command
	fmt.Fprintf(os.Stderr, "Sending to %s...\n", aiCmd)

	timeout := cfg.AI.Timeout
	if timeout == 0 {
		timeout = 120
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	args := append(cfg.AI.Args, "-p", promptText)
	cmd := exec.CommandContext(ctx, aiCmd, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("AI command timed out after %d seconds\n\nIncrease timeout in .council/config.yaml or use manual mode", timeout)
		}
		return fmt.Errorf("AI command failed: %w\n%s", err, stderr.String())
	}

	return applyResponse(stdout.Bytes())
}

func applyFromStdin() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	return applyResponse(data)
}

func applyResponse(data []byte) error {
	// Parse AI response
	experts, err := expert.ParseAIResponse(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse AI response as YAML.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Raw response:")
		fmt.Fprintln(os.Stderr, "---")
		fmt.Fprintln(os.Stderr, string(data))
		fmt.Fprintln(os.Stderr, "---")
		return fmt.Errorf("parsing failed: %w", err)
	}

	if len(experts) == 0 {
		return fmt.Errorf("no experts found in response")
	}

	// Show preview
	fmt.Printf("\nSuggested council (%d experts):\n", len(experts))
	for i, e := range experts {
		fmt.Printf("  %d. %s - %s\n", i+1, e.Name, e.Focus)
	}
	fmt.Println()

	// Confirm
	if !setupYes {
		if !Confirm("Apply this council?") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Create expert files
	for _, e := range experts {
		if err := e.Save(); err != nil {
			return fmt.Errorf("failed to save expert %s: %w", e.ID, err)
		}
		fmt.Printf("Created %s\n", e.Path())
	}

	fmt.Println()
	fmt.Println("Council created! Next steps:")
	fmt.Println("  council list    View your council")
	fmt.Println("  council sync    Sync to AI tool configs")

	return nil
}
