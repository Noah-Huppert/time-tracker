package main

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestMain(t *testing.T) {
	go func() {
		main()
	}()

	e := httpexpect.New(t, "http://localhost:8000")
	e.GET("/api/v0/health").Expect().
		Status(http.StatusOK).
		JSON().Equal(map[string]interface{}{"ok": true})
}
