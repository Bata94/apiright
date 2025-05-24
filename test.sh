#!/usr/bin/env bash

echo "Testing APIRight Framework in Nix Environment..."

# Enter nix development shell and run tests
nix develop --command bash -c "
    echo 'Inside nix development shell'
    go version
    echo 'Building test framework...'
    go build -o test_framework test_framework.go
    echo 'Running test framework...'
    timeout 10s ./test_framework
    echo 'Test completed'
"

echo "Test script finished"