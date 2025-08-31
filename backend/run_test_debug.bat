@echo off
echo Starting FeaturePlus test workflow in DEBUG mode...
set DEBUG=1
go run test_manager_workflow.go
