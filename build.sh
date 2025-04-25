#!/bin/bash

# Exit on error
set -e

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Check if Graphviz is installed
if ! command -v dot &> /dev/null; then
    echo "Error: Graphviz is not installed. Please install Graphviz first."
    echo "On macOS, you can install it using: brew install graphviz"
    echo "On Ubuntu/Debian: sudo apt-get install graphviz"
    echo "On Windows: choco install graphviz"
    exit 1
fi

# Build the project
echo "Building the project..."
go build -o deps-analyzer cmd/main.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful! Binary created as 'deps-analyzer'"
    echo ""
    echo "Usage:"
    echo "  ./deps-analyzer -module <module_name> [-depth <max_depth>] [-output <output_file>]"
    echo ""
    echo "Example:"
    echo "  ./deps-analyzer -module :account:account-domain -depth 2 -output dependencies.svg"
else
    echo "Build failed!"
    exit 1
fi 