package main

// Resource which can be acted on via the command line
type Resource interface {
	// Actions which can be executed on resource. Keys are the action names
	Actions() map[string]Action

	// Name of resource when references in the command line
	Name() string
}
