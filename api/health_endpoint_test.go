package main

import (
	"net/http"
	"testing"
)

// TestHealth ensures the health check endpoint works
func TestHealth(t *testing.T) {
	e := httpTester(t)

	e.GET("/api/v0/health").Expect().
		Status(http.StatusOK).
		JSON().Equal(map[string]interface{}{"ok": true})
}
