#!/bin/bash
echo "Starting FeaturePlus test workflow in DEBUG mode..."
export DEBUG=1
go run test_manager_workflow.go
