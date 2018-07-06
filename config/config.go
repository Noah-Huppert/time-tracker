package config

import "fmt"
import "os"
import "strconv"

// Config holds application configuration
type Config struct {
	// GRPCPort is the port the GRPC server serves requests on
	GRPCPort int

	// HTTPPort is the port static HTTP content will be served on
	HTTPPort int
}

// GRPCPortKey is the env variable key Config.GRPCPort is provided by
var GRPCPortKey string = "GRPC_PORT"

// HTTPPortKey is the env variable key Config.HTTPPort is provided by
var HTTPPortKey string = "HTTP_PORT"

// NewConfig loads configuration from the environment
func NewConfig() (*Config, error) {
	// Get GRPCPort
	grpcPortStr := os.Getenv(GRPCPortKey)
	if len(grpcPortStr) == 0 {
		return nil, fmt.Errorf("%s env variable must be set", GRPCPortKey)
	}

	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing GRPCPort into int: %s", err.Error())
	}

	// Get HTTPPortKey
	httpPortStr := os.Getenv(HTTPPortKey)
	if len(httpPortStr) == 0 {
		return nil, fmt.Errorf("%s env variable must be set", HTTPPortKey)
	}

	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTTPPort into int: %s", err.Error())
	}

	// Create
	return &Config{
		GRPCPort: grpcPort,
		HTTPPort: httpPort,
	}, nil
}
