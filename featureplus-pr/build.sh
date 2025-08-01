#!/bin/bash

echo "Building FeaturePlus PR CLI..."
echo

# Build for current platform
go build -o featureplus-pr .

if [ $? -eq 0 ]; then
    echo
    echo "Build successful! Executable created: featureplus-pr"
    echo
    echo "You can now run the tool from anywhere with:"
    echo "  ./featureplus-pr config"
    echo "  ./featureplus-pr upload --feature-id 1 --task-id 1"
    echo
else
    echo
    echo "Build failed! Please check for errors above."
    echo
fi 