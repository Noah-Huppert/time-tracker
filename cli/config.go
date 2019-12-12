package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// configFileLdr reads configuration files from disk
type configFileLdr struct {
	// cfgDir is the directory configuration files are located
	cfgDir string
}

// newConfigFileLdr creates a new configuration file loader
func newConfigFileLdr() (*configFileLdr, error) {
	// Get home env var
	home := os.Getenv("HOME")
	if len(home) == 0 {
		return nil, fmt.Errorf("HOME env var not found, required")
	}

	// Get xdg config file
	xdgCfgHome := os.Getenv("XDG_CONFIG_HOME")
	if len(xdgCfgHome) == 0 {
		xdgCfgHome = filepath.Join(home, ".config/")
	}

	// cfgDir
	cfgDir := filepath.Join(xdgCfgHome, "timetrk")

	ldr := configFileLdr{
		cfgDir: cfgDir,
	}
	return &ldr, nil
}

// Get option value from disk
func (ldr configFileLdr) load(key string) (string, error) {
	optFilePath := filepath.Join(ldr.cfgDir, key)

	_, err := os.Stat(optFilePath)
	if os.IsNotExist(err) {
		return "", nil
	}

	optValBytes, err := ioutil.ReadFile(optFilePath)
	if err != nil {
		return "", fmt.Errorf("error opening %s file: %s", key,
			err.Error())
	}

	return string(optValBytes), nil
}

// Config for time tracker command line client.
// Passed via command line options or files stored on disk.
type Config struct {
	// APIHost is the host of the time tracker API server
	APIHost string

	// AuthToken is a time tracker authentication token
	AuthToken string
}

// NewConfig loads configuration from command line options or stored on disk
func NewConfig() (*Config, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var apiHost string
	fs.StringVar(&apiHost, "api-host", "", "Host of time tracker API server")

	var authToken string
	fs.StringVar(&authToken, "auth-token", "", "Authentication token, "+
		"overrides tokens stored on disk")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err.Error())
	}

	if len(apiHost) == 0 || len(authToken) == 0 {
		cfgLdr, err := newConfigFileLdr()
		if err != nil {
			return nil, fmt.Errorf("could not create config file loader: %s",
				err.Error())
		}

		if len(apiHost) == 0 {
			apiHost, err = cfgLdr.load("api-host")
			if err != nil {
				return nil, fmt.Errorf("failed to load api-host "+
					"configuration file: %s", err.Error())
			}
		}

		if len(authToken) == 0 {
			authToken, err = cfgLdr.load("auth-token")
			if err != nil {
				return nil, fmt.Errorf("failed to load auth-token "+
					"configuration file: %s", err.Error())
			}
		}
	}

	missingOpts := []string{}
	if len(apiHost) == 0 {
		missingOpts = append(missingOpts, "api-host")
	}

	if len(authToken) == 0 {
		missingOpts = append(missingOpts, "auth-token")
	}

	if len(missingOpts) > 0 {
		return nil, fmt.Errorf("no %s configuration value(s), set "+
			"option(s) or set file(s)", strings.Join(missingOpts, ", "))
	}

	cfg := Config{
		APIHost:   apiHost,
		AuthToken: authToken,
	}

	return &cfg, nil
}
