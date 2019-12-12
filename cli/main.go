/*
Time tracker CLI. Performs API actions on resources.

Usage: timetrk [GLOBAL_OPTS] ACTION RESOURCE [SUB_CMD_OPTS]

Global options:

-auth-token AUTH_TOK: Auth token, overrides any stored authentication
                      credentials on disk

Sub-command options change based on the action and resource.

Arguments:

- ACTION: Action to execute
- RESOURCE: Resource on which to execute action

Authentication credentials stored in $XDG_CONFIG_HOME/timetrk/auth-token.
XDG_CONFIG_HOME defaults to $HOME/.config.

Resources:

- employee
- client
- auth-token
- contract
- project
- task

See each resource's struct .Actions() method for the actions they implement.
*/
package main

import (
	"flag"
	"os"

	"github.com/Noah-Huppert/golog"
)

func main() {
	// Setup
	golog := golog.NewStdLogger("timetrk")

	cfg, err := NewConfig()
	if err != nil {
		golog.Fatalf("failed to global configuration: %s", err.Error())
	}

	apiClient := APIClient{
		Cfg: *cfg,
	}

	args := []string{}
	skipNext := false
	for _, arg := range os.Args[1:] {
		if skipNext {
			skipNext = false
			continue
		}

		if arg[0] == '-' {
			skipNext = true
			continue
		}

		args = append(args, arg)
	}

	argsProblem := ""
	if len(args) == 0 {
		argsProblem = "arguments required"
	} else if len(args) == 1 {
		argsProblem = "RESOURCE argument required"
	} else if len(args) > 2 {
		argsProblem = "too many arguments"
	}

	if len(argsProblem) > 0 {
		golog.Fatalf("usage: %#v, %s [GLOBAL_OPTS] ACTION RESOURCE "+
			"[SUB_CMD_OPTS]\nproblem: %s"+
			argsProblem, flag.Args(), os.Args[0])
	}

	resources := map[string]Resource{}
	employee := EmployeeResource{}
	resources[employee.Name()] = employee

	// Run action on resource
	action := args[0]
	resource := args[1]

	if resourceInstance, ok := resources[resource]; ok {
		if actionExec, ok := resourceInstance.Actions()[action]; ok {
			if err = actionExec.Execute(args[2:], apiClient); err != nil {
				golog.Infof("%s %s", action, resource)
			} else {
				golog.Fatalf("failed to %s on %s: %s", action, resource,
					err.Error())
			}
		} else {
			golog.Fatalf("unknown action %s on resource %s", action,
				resource)
		}
	} else {
		golog.Fatalf("unknown resource: %s", resource)
	}
}
