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
	gatewayURLFlag := flag.String("gateway-url", getEnv("GATEWAY_URL", "http://192.168.1.254"), "URL of the BGW320-500 gateway (including scheme, e.g. http://192.168.1.254).")
	timeoutFlag := flag.Duration("scrape-timeout", getEnvDuration("SCRAPE_TIMEOUT", 10*time.Second), "Scrape timeout duration.")

	flag.Parse()

	slog.Info("Starting bgw320-500_exporter",
		slog.String("listen_address", *listenAddrFlag),
		slog.String("gateway_url", *gatewayURLFlag),
		slog.Duration("scrape_timeout", *timeoutFlag),
	)

	// Create and register collector
	collector := NewBGWCollector(*gatewayURLFlag, *timeoutFlag)
	prometheus.MustRegister(collector)

	scrapeCtx, cancelScrape := context.WithCancel(context.Background())
	collector.Start(scrapeCtx)

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

	errChan := make(chan error, 1)
	go func() {
		slog.Info("Starting server", slog.String("listen_address", *listenAddrFlag))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case sig := <-stopChan:
		slog.Info("Shutting down gracefully...", slog.String("signal", sig.String()))
	case err := <-errChan:
		slog.Error("Failed to start server", "error", err)
		cancelScrape()
		os.Exit(1)
	}

	cancelScrape()

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

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
