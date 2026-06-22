package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectorCollect(t *testing.T) {
	// Start a mock BGW gateway server serving the test data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var f *os.File
		var err error
		switch r.URL.Path {
		case "/cgi-bin/broadbandstatistics.ha":
			f, err = os.Open("testdata/broadbandstatistics.html")
		case "/cgi-bin/lanstatistics.ha":
			f, err = os.Open("testdata/lanstatistics.html")
		default:
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = f.Close() }()

		_, _ = io.Copy(w, f)
	}))
	defer server.Close()

	// Create a new collector pointing to the mock server
	collector := NewBGWCollector(server.URL, 2*time.Second)

	ctx := t.Context()
	collector.Start(ctx)

	// Wait for the first scrape to complete
	select {
	case <-collector.firstBroadbandDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for broadband scrape")
	}
	select {
	case <-collector.firstLANDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for LAN scrape")
	}

	// Registry
	reg := prometheus.NewRegistry()
	if err := reg.Register(collector); err != nil {
		t.Fatalf("Failed to register collector: %v", err)
	}

	// Gather metrics
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics are gathered and have correct values
	foundUp := false
	foundBytes := false
	foundWifiClients := 0

	for _, m := range metrics {
		name := m.GetName()
		switch name {
		case "bgw_up":
			foundUp = true
			if len(m.Metric) != 1 {
				t.Errorf("Expected 1 bgw_up metric, got %d", len(m.Metric))
			}
			val := m.Metric[0].Gauge.GetValue()
			if val != 1 {
				t.Errorf("Expected bgw_up to be 1, got %f", val)
			}
		case "bgw_broadband_ipv4_rx_bytes_total":
			foundBytes = true
			if len(m.Metric) != 1 {
				t.Errorf("Expected 1 rx_bytes metric, got %d", len(m.Metric))
			}
			val := m.Metric[0].Counter.GetValue()
			if val != 4066849231 {
				t.Errorf("Expected bgw_broadband_ipv4_rx_bytes_total to be 4066849231, got %f", val)
			}
		case "bgw_wifi_client_signal_strength_dbm":
			foundWifiClients = len(m.Metric)
		}
	}

	if !foundUp {
		t.Errorf("bgw_up metric not found")
	}
	if !foundBytes {
		t.Errorf("bgw_broadband_ipv4_rx_bytes_total metric not found")
	}
	if foundWifiClients != 12 {
		t.Errorf("Expected 12 wifi clients metrics, got %d", foundWifiClients)
	}
}
