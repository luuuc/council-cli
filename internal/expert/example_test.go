package expert_test

import (
	"fmt"

	"github.com/luuuc/council-cli/internal/expert"
)

func ExampleParse() {
	data := []byte(`---
id: kent-beck
name: Kent Beck
focus: Test-driven development
principles:
  - Red-green-refactor
  - Simple design
---

# Kent Beck

Expert in testing and simple design.`)

	e, err := expert.Parse(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", e.ID)
	fmt.Printf("Name: %s\n", e.Name)
	fmt.Printf("Focus: %s\n", e.Focus)
	fmt.Printf("Principles: %d\n", len(e.Principles))
	// Output:
	// ID: kent-beck
	// Name: Kent Beck
	// Focus: Test-driven development
	// Principles: 2
}

func ExampleToID() {
	fmt.Println(expert.ToID("Kent Beck"))
	fmt.Println(expert.ToID("DHH"))
	fmt.Println(expert.ToID("José Valim"))
	// Output:
	// kent-beck
	// dhh
	// jos-valim
}

func ExampleExpert_ApplyDefaults() {
	e := &expert.Expert{
		ID:   "my-expert",
		Name: "My Expert",
	}
	e.ApplyDefaults()

	fmt.Printf("Category: %s\n", e.Category)
	fmt.Printf("Priority: %s\n", e.Priority)
	// Output:
	// Category: custom
	// Priority: normal
}
