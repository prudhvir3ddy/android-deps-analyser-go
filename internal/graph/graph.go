package graph

import (
	"fmt"
	"os/exec"
	"strings"
)

type Colors struct {
	Root    NodeColors
	Project NodeColors
	Library NodeColors
}

type NodeColors struct {
	Fill string
	Text string
	Edge string
}

var DefaultColors = Colors{
	Root: NodeColors{
		Fill: "#4CAF50",
		Text: "white",
		Edge: "#2E7D32",
	},
	Project: NodeColors{
		Fill: "#81C784",
		Text: "black",
		Edge: "#2E7D32",
	},
	Library: NodeColors{
		Fill: "#BA68C8",
		Text: "white",
		Edge: "#6A1B9A",
	},
}

type Direction string

const (
	DirectionHorizontal Direction = "LR"
	DirectionVertical   Direction = "TB"
)

func GenerateOutput(projectDeps, libDeps map[string][]string, rootModule, outputFile string, direction Direction, duplicateNodes bool) error {
	dotContent := createDotContent(projectDeps, libDeps, rootModule, direction, duplicateNodes)
	return writeSVG(dotContent, outputFile)
}

func createDotContent(projectDeps, libDeps map[string][]string, rootModule string, direction Direction, duplicateNodes bool) []string {
	dotContent := []string{
		"digraph Dependencies {",
		fmt.Sprintf("  rankdir=%s;", direction),
		"  node [shape=box, style=filled, width=2, height=0.5, fontname=\"Arial\"];",
		"  edge [penwidth=1.5, fontname=\"Arial\"];",
	}

	if duplicateNodes {
		// Track processed nodes and their parents
		processedNodes := make(map[string]string) // node -> parent
		addedNodes := make(map[string]bool)
		nodeCounters := make(map[string]int)
		queuedNodes := make(map[string]bool)    // Track nodes that are in the queue
		processedEdges := make(map[string]bool) // Track processed edges to ensure consistency

		// Add root node
		dotContent = append(dotContent,
			fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
				rootModule, DefaultColors.Root.Fill, rootModule, DefaultColors.Root.Text))
		addedNodes[rootModule] = true

		// Create a queue for BFS processing
		queue := []string{rootModule}
		queuedNodes[rootModule] = true

		for len(queue) > 0 {
			currentModule := queue[0]
			queue = queue[1:]
			delete(queuedNodes, currentModule)

			// Process project dependencies
			if deps, exists := projectDeps[currentModule]; exists {
				for _, dep := range deps {
					// If this dependency has already been processed, create a new instance
					if _, exists := processedNodes[dep]; exists {
						nodeCounters[dep]++
						nodeID := fmt.Sprintf("%s_%d", dep, nodeCounters[dep])
						dotContent = append(dotContent,
							fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
								nodeID, DefaultColors.Project.Fill, dep, DefaultColors.Project.Text))
						dotContent = append(dotContent,
							fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\"];",
								currentModule, nodeID, DefaultColors.Project.Edge))

						// Process dependencies for this duplicate node
						if childDeps, exists := projectDeps[dep]; exists {
							for _, childDep := range childDeps {
								edgeKey := fmt.Sprintf("%s->%s", nodeID, childDep)
								if !processedEdges[edgeKey] {
									nodeCounters[childDep]++
									childNodeID := fmt.Sprintf("%s_%d", childDep, nodeCounters[childDep])
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
											childNodeID, DefaultColors.Project.Fill, childDep, DefaultColors.Project.Text))
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\"];",
											nodeID, childNodeID, DefaultColors.Project.Edge))
									processedEdges[edgeKey] = true

									if !queuedNodes[childNodeID] {
										queue = append(queue, childNodeID)
										queuedNodes[childNodeID] = true
									}
								}
							}
						}
						if childDeps, exists := libDeps[dep]; exists {
							for _, childDep := range childDeps {
								edgeKey := fmt.Sprintf("%s->%s", nodeID, childDep)
								if !processedEdges[edgeKey] {
									nodeCounters[childDep]++
									childNodeID := fmt.Sprintf("%s_%d", childDep, nodeCounters[childDep])
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
											childNodeID, DefaultColors.Library.Fill, childDep, DefaultColors.Library.Text))
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" -> \"%s\" [style=dashed, color=\"%s\"];",
											nodeID, childNodeID, DefaultColors.Library.Edge))
									processedEdges[edgeKey] = true

									if !queuedNodes[childNodeID] {
										queue = append(queue, childNodeID)
										queuedNodes[childNodeID] = true
									}
								}
							}
						}
					} else {
						// Add the node if it hasn't been added yet
						if !addedNodes[dep] {
							dotContent = append(dotContent,
								fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
									dep, DefaultColors.Project.Fill, dep, DefaultColors.Project.Text))
							addedNodes[dep] = true
						}

						// Add the edge
						edgeKey := fmt.Sprintf("%s->%s", currentModule, dep)
						if !processedEdges[edgeKey] {
							dotContent = append(dotContent,
								fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\"];",
									currentModule, dep, DefaultColors.Project.Edge))
							processedEdges[edgeKey] = true
						}

						// Mark as processed and add to queue if not already queued
						processedNodes[dep] = currentModule
						if !queuedNodes[dep] {
							queue = append(queue, dep)
							queuedNodes[dep] = true
						}
					}
				}
			}

			// Process library dependencies
			if deps, exists := libDeps[currentModule]; exists {
				for _, dep := range deps {
					// If this dependency has already been processed, create a new instance
					if _, exists := processedNodes[dep]; exists {
						nodeCounters[dep]++
						nodeID := fmt.Sprintf("%s_%d", dep, nodeCounters[dep])
						dotContent = append(dotContent,
							fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
								nodeID, DefaultColors.Library.Fill, dep, DefaultColors.Library.Text))
						dotContent = append(dotContent,
							fmt.Sprintf("  \"%s\" -> \"%s\" [style=dashed, color=\"%s\"];",
								currentModule, nodeID, DefaultColors.Library.Edge))

						// Process dependencies for this duplicate node
						if childDeps, exists := projectDeps[dep]; exists {
							for _, childDep := range childDeps {
								edgeKey := fmt.Sprintf("%s->%s", nodeID, childDep)
								if !processedEdges[edgeKey] {
									nodeCounters[childDep]++
									childNodeID := fmt.Sprintf("%s_%d", childDep, nodeCounters[childDep])
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
											childNodeID, DefaultColors.Project.Fill, childDep, DefaultColors.Project.Text))
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\"];",
											nodeID, childNodeID, DefaultColors.Project.Edge))
									processedEdges[edgeKey] = true

									if !queuedNodes[childNodeID] {
										queue = append(queue, childNodeID)
										queuedNodes[childNodeID] = true
									}
								}
							}
						}
						if childDeps, exists := libDeps[dep]; exists {
							for _, childDep := range childDeps {
								edgeKey := fmt.Sprintf("%s->%s", nodeID, childDep)
								if !processedEdges[edgeKey] {
									nodeCounters[childDep]++
									childNodeID := fmt.Sprintf("%s_%d", childDep, nodeCounters[childDep])
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
											childNodeID, DefaultColors.Library.Fill, childDep, DefaultColors.Library.Text))
									dotContent = append(dotContent,
										fmt.Sprintf("  \"%s\" -> \"%s\" [style=dashed, color=\"%s\"];",
											nodeID, childNodeID, DefaultColors.Library.Edge))
									processedEdges[edgeKey] = true

									if !queuedNodes[childNodeID] {
										queue = append(queue, childNodeID)
										queuedNodes[childNodeID] = true
									}
								}
							}
						}
					} else {
						// Add the node if it hasn't been added yet
						if !addedNodes[dep] {
							dotContent = append(dotContent,
								fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
									dep, DefaultColors.Library.Fill, dep, DefaultColors.Library.Text))
							addedNodes[dep] = true
						}

						// Add the edge
						edgeKey := fmt.Sprintf("%s->%s", currentModule, dep)
						if !processedEdges[edgeKey] {
							dotContent = append(dotContent,
								fmt.Sprintf("  \"%s\" -> \"%s\" [style=dashed, color=\"%s\"];",
									currentModule, dep, DefaultColors.Library.Edge))
							processedEdges[edgeKey] = true
						}

						// Mark as processed and add to queue if not already queued
						processedNodes[dep] = currentModule
						if !queuedNodes[dep] {
							queue = append(queue, dep)
							queuedNodes[dep] = true
						}
					}
				}
			}
		}
	} else {
		// Original non-unique mode
		// Add root node
		dotContent = append(dotContent,
			fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
				rootModule, DefaultColors.Root.Fill, rootModule, DefaultColors.Root.Text))

		// Add project dependencies
		for module, deps := range projectDeps {
			for _, dep := range deps {
				dotContent = append(dotContent,
					fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
						dep, DefaultColors.Project.Fill, dep, DefaultColors.Project.Text))
				dotContent = append(dotContent,
					fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\"];",
						module, dep, DefaultColors.Project.Edge))
			}
		}

		// Add library dependencies
		for module, deps := range libDeps {
			for _, dep := range deps {
				dotContent = append(dotContent,
					fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\", fontcolor=\"%s\"];",
						dep, DefaultColors.Library.Fill, dep, DefaultColors.Library.Text))
				dotContent = append(dotContent,
					fmt.Sprintf("  \"%s\" -> \"%s\" [style=dashed, color=\"%s\"];",
						module, dep, DefaultColors.Library.Edge))
			}
		}
	}

	dotContent = append(dotContent, "}")
	return dotContent
}

func writeSVG(dotContent []string, outputFile string) error {
	cmd := exec.Command("dot", "-Tsvg", "-o", outputFile)
	cmd.Stdin = strings.NewReader(strings.Join(dotContent, "\n"))
	return cmd.Run()
}
