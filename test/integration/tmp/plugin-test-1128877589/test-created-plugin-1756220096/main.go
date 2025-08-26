package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	// Import your plugin protocol definitions
)

func main() {
	// Parse command line arguments
	port := "9001"
	if len(os.Args) > 2 && os.Args[1] == "--port" {
		port = os.Args[2]
	}

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	
	// Register your plugin service
	// pb.RegisterPluginServiceServer(server, &YourPluginService{})

	fmt.Printf("%s plugin starting on port %s\n", "/tmp/plugin-test-1128877589/test-created-plugin-1756220096", port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Shutting down...")
		server.GracefulStop()
	}()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
