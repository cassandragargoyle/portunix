package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/handlers"
	pb "go-plugin-template/proto"
)

func main() {
	var (
		port       = flag.Int("port", 50051, "gRPC server port")
		healthPort = flag.Int("health-port", 50052, "Health check port")
		metricsPort = flag.Int("metrics-port", 8080, "Metrics port")
		configPath = flag.String("config", "config.yaml", "Configuration file path")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure logging
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.WithFields(logrus.Fields{
		"plugin":  cfg.Plugin.Name,
		"version": cfg.Plugin.Version,
	}).Info("Starting plugin")

	// Start metrics server
	go startMetricsServer(*metricsPort)

	// Create gRPC server
	server := grpc.NewServer()

	// Register plugin service
	pluginHandler := handlers.NewPluginHandler(cfg)
	pb.RegisterPluginServiceServer(server, pluginHandler)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Start health check server
	go startHealthServer(*healthPort, healthServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logrus.Info("Shutting down gracefully...")
		pluginHandler.Shutdown()
		server.GracefulStop()
	}()

	logrus.WithField("port", *port).Info("Plugin server listening")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func startMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	logrus.WithField("port", port).Info("Metrics server listening")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func startHealthServer(port int, healthServer *health.Server) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen for health checks: %v", err)
	}

	healthGRPCServer := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(healthGRPCServer, healthServer)

	logrus.WithField("port", port).Info("Health check server listening")
	log.Fatal(healthGRPCServer.Serve(listener))
}