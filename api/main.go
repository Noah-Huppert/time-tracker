/*
Time tracker HTTP API.

Requests pass data using URL parameters for all HTTP verbs.  
JSON encoded bodies may be used for the POST, PUT, PATCH, 
DELETE HTTP verbs.

Responses will always be JSON encoded.

RESTful API. With resources: users, projects, time entry.
*/
package main

import (
	"context"
	"net/http"

	"github.com/Noah-Huppert/golog"
	"github.com/gorilla/mux"
)

func main() {
	// {{{1 Initialization
	ctx := context.WithCancel(context.Background())
	logger := golog.NewStdLogger("time-tracker-api")

	handler := mux.NewRouter()
	handler.Handle("/api/v0/
}
