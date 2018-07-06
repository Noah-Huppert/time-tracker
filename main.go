package main

import "fmt"
import "net"
import "log"
import "net/http"
import "os"
import "os/signal"
import "context"

import "github.com/Noah-Huppert/time-tracker/config"
import "github.com/Noah-Huppert/time-tracker/users"

import "google.golang.org/grpc"

func main() {
	// Base context
	ctx, cancelCtx := context.WithCancel(context.Background())

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

	go func() {
		grpcServer.Serve(listener)
	}()

	log.Printf("started RPC server on :%d\n", cfg.GRPCPort)

	// Serve HTTP
	httpSrv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.HTTPPort)}
	http.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	go func() {
		httpSrv.ListenAndServe()
	}()

	log.Printf("started HTTP server on :%d\n", cfg.HTTPPort)

	// Setup exit signal handler
	exitSignalChan := make(chan os.Signal, 1)
	signal.Notify(exitSignalChan, os.Interrupt)

	go func() {
		select {
		case <-exitSignalChan:
			log.Println("stopping")
			cancelCtx()
		}
	}()

	// Wait until exited
	select {
	case <-ctx.Done():
		// Stop GRPC
		grpcServer.GracefulStop()
		log.Println("stopped GRPC server")

		// Stop HTTP
		if err := httpSrv.Shutdown(nil); err != nil {
			log.Fatalf("error stopping HTTP server: %s\n", err.Error())
		}
		log.Println("stopped HTTP server")

		log.Println("stopped")
	}

	log.Println("exiting")
}
