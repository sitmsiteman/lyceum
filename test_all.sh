#!/bin/bash

# Stop on error
set -e

echo "1. Building all components..."

# Build tlgviewer (Updated to use tlgcore)
echo "   - Building tlgviewer..."
go build -o tlgviewer ./cmd/tlgviewer

# Build readauth (May need check if it compiles with old logic or needs update)
echo "   - Building readauth..."
go build -o readauth ./cmd/readauth

# Build search
echo "   - Building search..."
go build -o search ./cmd/search

# Build the new test suite
echo "   - Building test_full..."
go build -o test_full ./cmd/test_full

echo "Build Success!"
echo "---------------------------------------------------"

# Ask for directory if not provided
DIR="."
if [ "$1" != "" ]; then
    DIR="$1"
else
    echo "Usage: ./test_all.sh [directory_with_tlg_files]"
    echo "Defaulting to current directory..."
fi

echo "2. Running Feature Test Suite on: $DIR"
./test_full -d "$DIR"

rm tlgviewer readauth search test_full
