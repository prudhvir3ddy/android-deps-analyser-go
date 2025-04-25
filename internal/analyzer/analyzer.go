package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DependencyAnalyzer struct {
	projectPath string
}

func NewDependencyAnalyzer(projectPath string) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		projectPath: projectPath,
	}
}

func (a *DependencyAnalyzer) Analyze(rootModule string, maxDepth int) (map[string][]string, map[string][]string) {
	projectDeps := make(map[string][]string)
	libDeps := make(map[string][]string)
	queue := make([]string, 0)
	depth := make(map[string]int)

	queue = append(queue, rootModule)
	depth[rootModule] = 0

	fmt.Printf("\nAnalyzing dependencies for module: %s\n", rootModule)

	for len(queue) > 0 {
		currentModule := queue[0]
		queue = queue[1:]

		if maxDepth > 0 && depth[currentModule] >= maxDepth {
			continue
		}

		buildFile := a.findBuildFile(currentModule)
		fmt.Printf("Current build file %s:\n", buildFile)
		if buildFile != "" {
			projDeps, libs := a.parseDependencies(buildFile)
			fmt.Printf("Found dependencies for %s:\n", currentModule)
			fmt.Printf("  Project deps: %v\n", projDeps)
			fmt.Printf("  Library deps: %v\n", libs)

			// Only update if we found dependencies
			if len(projDeps) > 0 || len(libs) > 0 {
				projectDeps[currentModule] = projDeps
				libDeps[currentModule] = libs
			}

			for _, dep := range projDeps {
				queue = append(queue, dep)
				depth[dep] = depth[currentModule] + 1
			}
		}
	}

	return projectDeps, libDeps
}

func (a *DependencyAnalyzer) findBuildFile(moduleName string) string {
	modulePath := a.toPath(moduleName)
	buildFile := filepath.Join(a.projectPath, modulePath, "build.gradle.kts")
	if _, err := os.Stat(buildFile); err == nil {
		return buildFile
	}
	return ""
}

func (a *DependencyAnalyzer) parseDependencies(buildFile string) ([]string, []string) {
	projectDeps := make([]string, 0)
	libDeps := make([]string, 0)

	file, err := os.Open(buildFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", buildFile, err)
		return projectDeps, libDeps
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDepsBlock := false
	braceCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "dependencies {") {
			inDepsBlock = true
			braceCount = 1
			continue
		}

		if !inDepsBlock {
			continue
		}

		if strings.Contains(line, "{") {
			braceCount++
		}
		if strings.Contains(line, "}") {
			braceCount--
			if braceCount == 0 {
				inDepsBlock = false
			}
		}

		if dep := a.parseProjectDependency(line); dep != "" {
			projectDeps = append(projectDeps, dep)
		}
		if dep := a.parseLibraryDependency(line); dep != "" {
			libDeps = append(libDeps, dep)
		}
	}

	return projectDeps, libDeps
}

func (a *DependencyAnalyzer) parseProjectDependency(line string) string {
	if !strings.Contains(line, "projects.") {
		return ""
	}

	// Look for implementation, api, or compileOnly with projects.
	if !strings.Contains(line, "implementation(projects.") &&
		!strings.Contains(line, "api(projects.") &&
		!strings.Contains(line, "compileOnly(projects.") {
		return ""
	}

	start := strings.Index(line, "projects.")
	end := strings.Index(line[start:], ")")
	if end == -1 {
		return ""
	}

	dotNotation := line[start : start+end]
	module := a.dotToModule(dotNotation)
	return module
}

func (a *DependencyAnalyzer) parseLibraryDependency(line string) string {
	// Check for both libs. and deliverooLibs. prefixes
	if !strings.Contains(line, "libs.") && !strings.Contains(line, "deliverooLibs.") {
		return ""
	}

	// Look for implementation, api, or compileOnly with libs.
	if !strings.Contains(line, "implementation(") &&
		!strings.Contains(line, "api(") &&
		!strings.Contains(line, "compileOnly(") {
		return ""
	}

	// Find the start of the library name
	var start int
	if strings.Contains(line, "libs.") {
		start = strings.Index(line, "libs.")
	} else if strings.Contains(line, "deliverooLibs.") {
		start = strings.Index(line, "deliverooLibs.")
	} else {
		return ""
	}

	end := strings.Index(line[start:], ")")
	if end == -1 {
		return ""
	}

	// Extract the library name with the prefix
	libName := line[start : start+end]
	fmt.Printf("Found library dependency: %s from line: %s\n", libName, line)
	return libName
}

func (a *DependencyAnalyzer) toPath(moduleName string) string {
	if strings.HasPrefix(moduleName, ":") {
		moduleName = moduleName[1:]
	}
	parts := strings.Split(moduleName, ":")
	pathParts := make([]string, len(parts))
	for i, part := range parts {
		pathParts[i] = a.toKebabCase(part)
	}
	path := filepath.Join(pathParts...)
	return path
}

func (a *DependencyAnalyzer) dotToModule(dotNotation string) string {
	if strings.HasPrefix(dotNotation, "projects.") {
		dotNotation = dotNotation[8:]
	}
	parts := strings.Split(dotNotation, ".")
	moduleParts := make([]string, len(parts))
	for i, part := range parts {
		moduleParts[i] = a.toKebabCase(part)
	}
	module := strings.Join(moduleParts, ":")
	if !strings.HasPrefix(module, ":") {
		module = ":" + module
	}
	return module
}

func (a *DependencyAnalyzer) toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
