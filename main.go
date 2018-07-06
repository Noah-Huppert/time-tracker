package main

import "fmt"
import "net"
import "log"

import "github.com/Noah-Huppert/time-tracker/config"
import "github.com/Noah-Huppert/time-tracker/users"

import "google.golang.org/grpc"

func main() {
	// Get config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error loading configuration: %s\n", err.Error())
	}

	// Initialize RPC
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("error listening on GRPC port: %s\n", err.Error())
	}

	grpcServer := grpc.NewServer()
	users.RegisterUsersServer(grpcServer, &users.DefaultUsersServer{})
	grpcServer.Serve(listener)
}
