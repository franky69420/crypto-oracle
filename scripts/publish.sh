#!/bin/bash

# Script to publish updates to GitHub
# Usage: ./scripts/publish.sh "Commit message"

# Check if a commit message was provided
if [ $# -eq 0 ]; then
    echo "Error: No commit message provided"
    echo "Usage: ./scripts/publish.sh \"Commit message\""
    exit 1
fi

COMMIT_MESSAGE="$1"

# Clean up any binaries
echo "Cleaning up build artifacts..."
make clean

# Build to ensure code compiles
echo "Building to validate code..."
make build
make build-token-scan

# Run tests
echo "Running tests..."
make test

# Add all changes
echo "Adding changes to git..."
git add .

# Commit changes
echo "Committing changes with message: $COMMIT_MESSAGE"
git commit -m "$COMMIT_MESSAGE"

# Push to GitHub
echo "Pushing to GitHub..."
git push

echo "Done! Changes have been published to GitHub." 