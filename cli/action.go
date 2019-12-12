package main

// Action executed on a resource via the command line
type Action interface {
	// Execute action. Args are command line arguments.
	Execute(args []string, client APIClient) error

	// Help text to display to users
	Help() string
}
