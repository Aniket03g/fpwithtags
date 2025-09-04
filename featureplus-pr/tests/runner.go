package main

import (
	"fmt"
	"os"
	"strings"
	
	"featureplus-pr/tests"
)

// Import test functions from the tests package
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tests/runner.go [test-name]")
		fmt.Println("Available tests:")
		fmt.Println("  debug-auth     - Debug authentication issues")
		fmt.Println("  debug-cli      - Debug CLI configuration")
		fmt.Println("  test-token     - Test token handling")
		fmt.Println("  test-auth      - Test authentication")
		fmt.Println("  test-client    - Test client functionality")
		os.Exit(1)
	}

	testName := strings.ToLower(os.Args[1])
	
	switch testName {
	case "debug-auth":
		fmt.Println("Running debug-auth test...")
		tests.DebugAuth()
	case "debug-cli":
		fmt.Println("Running debug-cli test...")
		tests.DebugCLI()
	case "test-token":
		fmt.Println("Running test-token test...")
		tests.TestToken()
	case "test-auth":
		fmt.Println("Running test-auth test...")
		tests.TestAuth()
	case "test-client":
		fmt.Println("Running test-client test...")
		tests.TestClient()
	case "test-login":
		fmt.Println("Running test-login test...")
		tests.TestLogin()
	case "standalone":
		fmt.Println("Running standalone test...")
		tests.StandaloneTest()
	default:
		fmt.Printf("Unknown test: %s\n", testName)
		os.Exit(1)
	}
}
