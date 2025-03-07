#!/bin/bash

# Script to run Go test cases across all services and subdirectories.

echo "Running Go tests across all services..."

# Find all directories containing go.mod and run 'go test ./...'
find . -name "go.mod" | while read modfile; do
    # Get the directory of the current go.mod file
    service_dir=$(dirname "$modfile")

    echo "------------------------------------------"
    echo "Running tests in: $service_dir"
    echo "------------------------------------------"

    # Change into the service directory and run tests
    (cd "$service_dir" && go test ./... -v)

    # Check the test result
    if [ $? -ne 0 ]; then
        echo "❌ Tests failed in: $service_dir"
    else
        echo "✅ Tests passed in: $service_dir"
    fi
done

echo "------------------------------------------"
echo "All tests completed."
