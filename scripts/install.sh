#!/bin/bash

# vex installation script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing vex...${NC}"

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    echo "Please install Go 1.22 or later from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]); then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION detected. vex requires Go 1.22 or later.${NC}"
fi

# Install vex
echo "Building vex..."
go install ./cmd/vex

# Verify installation
if command -v vex &> /dev/null; then
    echo -e "${GREEN}vex installed successfully!${NC}"
    echo ""
    echo "Run 'vex --help' to get started."
else
    echo -e "${YELLOW}vex was built but may not be in your PATH.${NC}"
    echo "Make sure \$GOPATH/bin (usually ~/go/bin) is in your PATH."
    echo ""
    echo "Add this to your shell profile:"
    echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
fi
