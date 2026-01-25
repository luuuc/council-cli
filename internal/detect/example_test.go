package detect_test

import (
	"fmt"

	"github.com/luuuc/council-cli/internal/detect"
)

func ExampleScan() {
	// Scan analyzes a directory and returns detected technologies
	d, err := detect.Scan(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Languages: %d detected\n", len(d.Languages))
	fmt.Printf("Frameworks: %d detected\n", len(d.Frameworks))
	fmt.Printf("Summary: %s\n", d.Summary())
}

func ExampleDetection_Summary() {
	d := &detect.Detection{
		Languages: []detect.Language{
			{Name: "Go", Percentage: 85.5},
			{Name: "Shell", Percentage: 14.5},
		},
		Frameworks: []detect.Framework{
			{Name: "Cobra"},
		},
		Testing: []string{"Go testing"},
	}

	fmt.Println(d.Summary())
	// Output: Go, Shell, Cobra, Go testing
}
