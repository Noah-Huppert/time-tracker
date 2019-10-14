package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONResponder writes a JSON response
type JSONResponder struct {
	// Status code written in response. Defaults to http.StatusOK
	Status int

	// Data to write as JSON
	Data interface{}
}

// Respond writes the JSON response
func (h JSONResponder) Respond(ectx EndpointContext, w http.ResponseWriter,
	r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	if h.Status == 0 {
		h.Status = http.StatusOK
	}

	w.WriteHeader(h.Status)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(h.Data)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %s", err.Error())
	}

	return nil
}
