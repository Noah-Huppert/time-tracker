package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// APIClient is a client for the time tracker API
type APIClient struct {
	// Cfg for command line client
	Cfg Config
}

// NotOKRespErr indicates an API response returned with an non-OK status code.
// A response body with an error will also be returned if this error occurs.
type NotOKRespErr struct{}

// Error message
func (e NotOKRespErr) Error() string {
	return "not-OK API response"
}

// IsNotOKRespErr tests if an error is a NotOKRespErr
func IsNotOKRespErr(err error) bool {
	_, ok := err.(NotOKRespErr)
	return ok
}

// APIResp is an API response
type APIResp struct {
	// Data in response body
	Data map[string]interface{}

	// Raw response
	Resp http.Response
}

// Error returns nil if there is no error in the response, and an error if there
// was an error. An error occurs when the Resp.StatusCode is not http.StatusOK.
// If an error cannot be found in the body a generic error will be returned.
func (resp APIResp) Error() error {
	if resp.Resp.StatusCode != http.StatusOK {
		if err, ok := resp.Data["error"]; ok {
			return fmt.Errorf("status code: %s (%s), error: %s",
				resp.Resp.Status, resp.Resp.StatusCode, err)
		} else {
			return fmt.Errorf("unknown server error")
		}
	}

	return nil
}

// Req makes a request to the time tracker API server. Response body is
// decoded as JSON and stored in respDat.
// Sets the req Host and authentication header.
func (client APIClient) Req(req http.Request) (*APIResp, error) {
	// Set request Host to api host
	req.URL.Host = client.Cfg.APIHost

	// Set authentication header
	req.Header.Set("Authorization", client.Cfg.AuthToken)

	// Make request
	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %s", err.Error())
	}

	// Decode response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %s", err.Error())
	}

	if len(respBytes) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	var respDat map[string]interface{}
	if err = json.Unmarshal(respBytes, &respDat); err != nil {
		return nil, fmt.Errorf("failed to decode API response as JSON: %s",
			err.Error())
	}

	apiResp := APIResp{
		Data: respDat,
		Resp: *resp,
	}

	if resp.StatusCode != http.StatusOK {
		return &apiResp, NotOKRespErr{}
	}

	return &apiResp, nil
}
