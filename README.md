# Dependency Analyzer

A Go-based tool for analyzing Android module dependencies in a Gradle project. This tool generates a visual representation of module dependencies using Graphviz.

## Prerequisites

- Go 1.16 or later
- Graphviz (for SVG generation)
  - macOS: `brew install graphviz`
  - Ubuntu/Debian: `sudo apt-get install graphviz`
  - Windows: `choco install graphviz`

## Building

To build the project, run:

```bash
./build.sh
```

This will create a binary named `deps-analyzer` in the current directory.

## Usage

The tool analyzes dependencies by reading `build.gradle.kts` files in your Android project. It supports both project dependencies and library dependencies.

Basic usage:

```bash
./deps-analyzer -module <module_name> [-depth <max_depth>] [-output <output_file>]
```



### Parameters

- `-module`: The module to analyze (e.g., `:account:account-domain`)
- `-depth`: Maximum depth to analyze (optional, no limit if not specified)
- `-output`: Output SVG file path (optional, defaults to `module_dependencies.svg`)

### Example

```bash
./deps-analyzer -module :account:account-domain -depth 2 -output dependencies.svg
```

## Output

The tool generates two types of output:

1. Console output showing the dependency tree in text format
2. An SVG file visualizing the dependencies using Graphviz

The SVG visualization uses different colors to distinguish between:
- Root module (green)
- Project dependencies (light green)
- Library dependencies (purple)

## Features

- Analyzes both project and library dependencies
- Generates visual dependency graphs
- Supports depth-limited analysis
- Handles Gradle Kotlin DSL build files
- Converts between different module naming conventions

## License

MIT License 
