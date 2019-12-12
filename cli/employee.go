package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// EmployeeModel resource with actions
type EmployeeModel struct {
	// ID is a unique identifier
	ID int

	// Name of employee
	Name string
}

// EmployeeResource can perform actions on employee resources
type EmployeeResource struct{}

// Name of employee resource
func (e EmployeeResource) Name() string {
	return "employee"
}

// Actions employee resource implements
func (e EmployeeResource) Actions() map[string]Action {
	return map[string]Action{
		"get": GetEmployee{},
	}
}

// GetEmployee action
type GetEmployee struct{}

// Help for get employee action
func (a GetEmployee) Help() string {
	return "get employee by id or name, cannot search by both"
}

// Execute get employee action
func (a GetEmployee) Execute(args []string, client APIClient) error {
	// Get action parameters
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var tgtID string
	flagSet.StringVar(&tgtID, "id", "", "id of employee to find")

	var tgtName string
	flagSet.StringVar(&tgtName, "name", "", "name of employee to find")
	flagSet.Parse(args)

	if len(tgtID) > 0 && len(tgtName) > 0 {
		return fmt.Errorf("cannot provide -id and -name options at the " +
			"same time")
	} else if len(tgtID) == 0 && len(tgtName) == 0 {
		return fmt.Errorf("either -id or -name option must be provided")
	}

	// Make api request
	query := ""
	if len(tgtID) > 0 {
		query += fmt.Sprintf("id=%s", tgtID)
	}
	if len(tgtName) > 0 {
		query += fmt.Sprintf("tgtName=%s", tgtName)
	}

	resp, err := client.Req(http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path:     "/api/v0/",
			RawQuery: query,
		},
	})
	fmt.Printf("employee.get resp=%#v\n", resp)
	if err != nil {
		return fmt.Errorf("failed to make API request: %s", err.Error())
	}
	return nil
}
