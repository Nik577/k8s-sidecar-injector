package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s-sidecar-injector/pkg/mutation"
	"k8s-sidecar-injector/pkg/webhook"
)

func main() {
	// Initialize structured logging (Go 1.21 slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Configuration via environment variables
	port := getEnv("WEBHOOK_PORT", "8443")
	certFile := getEnv("WEBHOOK_CERT_FILE", "/etc/webhook/certs/tls.crt")
	keyFile := getEnv("WEBHOOK_KEY_FILE", "/etc/webhook/certs/tls.key")
	
	sidecarImage := getEnv("SIDECAR_IMAGE", "falcosecurity/falco-no-driver:latest")
	sidecarName := getEnv("SIDECAR_NAME", "security-agent")
	
	slog.Info("Starting k8s-sidecar-injector",
		"port", port,
		"cert", certFile,
		"key", keyFile,
		"sidecar_image", sidecarImage,
	)

	server := &webhook.Server{
		SidecarConfig: mutation.SidecarConfig{
			Image: sidecarImage,
			Name:  sidecarName,
			Args: []string{
				"/usr/bin/falco",
				"-A",
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", server.HandleMutate)
	mux.HandleFunc("/healthz", server.HandleHealthz)
	mux.HandleFunc("/readyz", server.HandleReadyz)
	mux.Handle("/metrics", server.HandleMetrics())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Webhook server listening", "addr", srv.Addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServeTLS failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-stop
	slog.Info("Shutting down webhook server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited gracefully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
