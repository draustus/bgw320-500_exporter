package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Setup command line flags and environment variables fallback
	listenAddrFlag := flag.String("listen-address", getEnv("LISTEN_ADDRESS", ":9191"), "Address to listen on for web interface and telemetry.")
	gatewayURLFlag := flag.String("gateway-url", getEnv("GATEWAY_URL", "https://192.168.1.254"), "URL of the BGW320-500 gateway (including scheme, e.g. https://192.168.1.254).")
	insecureTLSFlag := flag.Bool("insecure-skip-verify", getEnvBool("INSECURE_SKIP_VERIFY", true), "Skip TLS verification for the gateway certificate.")
	timeoutFlag := flag.Duration("scrape-timeout", getEnvDuration("SCRAPE_TIMEOUT", 10*time.Second), "Scrape timeout duration.")

	flag.Parse()

	slog.Info("Starting bgw320-500_exporter",
		slog.String("listen_address", *listenAddrFlag),
		slog.String("gateway_url", *gatewayURLFlag),
		slog.Bool("insecure_skip_verify", *insecureTLSFlag),
		slog.Duration("scrape_timeout", *timeoutFlag),
	)

	// Create and register collector
	collector := NewBGWCollector(*gatewayURLFlag, *insecureTLSFlag, *timeoutFlag)
	prometheus.MustRegister(collector)

	// Setup HTTP Handlers
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html>
			<head><title>BGW320-500 Exporter</title></head>
			<body>
			<h1>BGW320-500 Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	server := &http.Server{
		Addr:              *listenAddrFlag,
		Handler:           nil, // uses http.DefaultServeMux
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting server", slog.String("listen_address", *listenAddrFlag))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-stopChan
	slog.Info("Shutting down gracefully...", slog.String("signal", sig.String()))

	// Dynamically calculate shutdown timeout based on scrape timeout + buffer
	shutdownTimeout := *timeoutFlag + 5*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	if err := server.Shutdown(ctx); err != nil {
		cancel()
		slog.Error("Graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	cancel()
	slog.Info("Server stopped successfully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return value == "true" || value == "1"
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
