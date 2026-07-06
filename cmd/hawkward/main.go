package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/ui"
)

func main() {
	// Initialize the root model (checks for first-run)
	rootModel := ui.NewRootModel()

	// Create the Bubble Tea program
	p := tea.NewProgram(rootModel)

	// Run the application
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
