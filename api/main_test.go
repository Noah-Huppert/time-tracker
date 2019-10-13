package main

import (
	"os"
	"testing"

	"github.com/gavv/httpexpect"
)

// TestMain starts the API server and invokes the tests
func TestMain(m *testing.M) {
	go func() {
		main()
	}()

	os.Exit(m.Run())
}

// httpTester creates a new httpexpect.Expect instance
func httpTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.New(t, "http://localhost:8000")
}
