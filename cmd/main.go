package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/orderapp/deps-analyzer/internal/analyzer"
	"github.com/orderapp/deps-analyzer/internal/graph"
)

func main() {
	module := flag.String("module", "", "The module to analyze (e.g., :account:account-domain)")
	depth := flag.Int("depth", 0, "Maximum depth to analyze (0 for no limit)")
	output := flag.String("output", "module_dependencies.svg", "Output SVG file path")
	direction := flag.String("direction", "LR", "Graph direction: LR (horizontal) or TB (vertical)")
	duplicateNodes := flag.Bool("duplicate", false, "Duplicate nodes to ensure each node has at most one incoming edge")
	flag.Parse()

	if *module == "" {
		fmt.Println("Error: module flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if !strings.HasPrefix(*module, ":") {
		*module = ":" + *module
	}

	// Validate direction
	var graphDirection graph.Direction
	switch strings.ToUpper(*direction) {
	case "LR":
		graphDirection = graph.DirectionHorizontal
	case "TB":
		graphDirection = graph.DirectionVertical
	default:
		fmt.Println("Error: direction must be either LR (horizontal) or TB (vertical)")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("\nAnalyzing dependencies for %s\n", *module)
	if *depth > 0 {
		fmt.Printf("Max depth: %d\n", *depth)
	}
	fmt.Printf("Graph direction: %s\n", *direction)
	fmt.Printf("Node duplication: %v\n", *duplicateNodes)

	projectRoot := "."
	analyzer := analyzer.NewDependencyAnalyzer(projectRoot)
	projectDeps, libDeps := analyzer.Analyze(*module, *depth)

	// Print dependencies
	printDependencies(projectDeps, libDeps)

	// Generate SVG visualization
	graph.GenerateOutput(projectDeps, libDeps, *module, *output, graphDirection, *duplicateNodes)
	fmt.Printf("\nSVG visualization generated at: %s\n", *output)
}

func printDependencies(projectDeps, libDeps map[string][]string) {
	fmt.Println("\nProject Dependencies:")
	for module, deps := range projectDeps {
		fmt.Printf("\n%s depends on projects:\n", module)
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep)
		}
	}

	fmt.Println("\nLibrary Dependencies:")
	for module, deps := range libDeps {
		fmt.Printf("\n%s depends on libs:\n", module)
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep)
		}
	}
}
