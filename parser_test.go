package main

import (
	"os"
	"testing"
)

func TestParseBroadbandStats(t *testing.T) {
	file, err := os.Open("testdata/broadbandstatistics.html")
	if err != nil {
		t.Fatalf("Failed to open broadband test data: %v", err)
	}
	defer func() { _ = file.Close() }()

	stats, err := ParseBroadbandStats(file)
	if err != nil {
		t.Fatalf("Failed to parse broadband stats: %v", err)
	}

	// Assertions based on broadbandstatistics.html contents
	if stats.ConnectionSource != "ETHERNET" {
		t.Errorf("Expected ConnectionSource to be 'ETHERNET', got '%s'", stats.ConnectionSource)
	}
	if stats.Connection != "Up" {
		t.Errorf("Expected Connection to be 'Up', got '%s'", stats.Connection)
	}
	if stats.NetworkType != "Lightspeed" {
		t.Errorf("Expected NetworkType to be 'Lightspeed', got '%s'", stats.NetworkType)
	}
	if stats.IPv4Address != "107.136.2.157" {
		t.Errorf("Expected IPv4Address to be '107.136.2.157', got '%s'", stats.IPv4Address)
	}
	if stats.GatewayAddress != "107.136.2.1" {
		t.Errorf("Expected GatewayAddress to be '107.136.2.1', got '%s'", stats.GatewayAddress)
	}
	if stats.MACAddress != "10:c4:ca:5c:ac:c0" {
		t.Errorf("Expected MACAddress to be '10:c4:ca:5c:ac:c0', got '%s'", stats.MACAddress)
	}
	if stats.PrimaryDNS != "68.94.156.8" {
		t.Errorf("Expected PrimaryDNS to be '68.94.156.8', got '%s'", stats.PrimaryDNS)
	}
	if stats.SecondaryDNS != "68.94.157.8" {
		t.Errorf("Expected SecondaryDNS to be '68.94.157.8', got '%s'", stats.SecondaryDNS)
	}
	if stats.MTU != 1500 {
		t.Errorf("Expected MTU to be 1500, got %d", stats.MTU)
	}

	// Ethernet
	if stats.EthernetLineState != "Up" {
		t.Errorf("Expected EthernetLineState to be 'Up', got '%s'", stats.EthernetLineState)
	}
	if stats.EthernetSpeed != 1000 {
		t.Errorf("Expected EthernetSpeed to be 1000, got %f", stats.EthernetSpeed)
	}
	if stats.EthernetDuplex != "full" {
		t.Errorf("Expected EthernetDuplex to be 'full', got '%s'", stats.EthernetDuplex)
	}

	// IPv6
	if stats.IPv6Status != "Available" {
		t.Errorf("Expected IPv6Status to be 'Available', got '%s'", stats.IPv6Status)
	}
	if stats.IPv6Address != "2001:506:4846:f35::1" {
		t.Errorf("Expected IPv6Address to be '2001:506:4846:f35::1', got '%s'", stats.IPv6Address)
	}

	// IPv4 Statistics
	if stats.IPv4RxPackets != 1554021474 {
		t.Errorf("Expected IPv4RxPackets to be 1554021474, got %f", stats.IPv4RxPackets)
	}
	if stats.IPv4TxPackets != 730879952 {
		t.Errorf("Expected IPv4TxPackets to be 730879952, got %f", stats.IPv4TxPackets)
	}
	if stats.IPv4RxBytes != 4066849231 {
		t.Errorf("Expected IPv4RxBytes to be 4066849231, got %f", stats.IPv4RxBytes)
	}
	if stats.IPv4TxBytes != 3415517179 {
		t.Errorf("Expected IPv4TxBytes to be 3415517179, got %f", stats.IPv4TxBytes)
	}
}

func TestParseLANStats(t *testing.T) {
	file, err := os.Open("testdata/lanstatistics.html")
	if err != nil {
		t.Fatalf("Failed to open lan test data: %v", err)
	}
	defer func() { _ = file.Close() }()

	stats, err := ParseLANStats(file)
	if err != nil {
		t.Fatalf("Failed to parse lan stats: %v", err)
	}

	// Critical LAN info
	if stats.IPv4Address != "192.168.1.254" {
		t.Errorf("Expected IPv4Address to be '192.168.1.254', got '%s'", stats.IPv4Address)
	}
	if stats.Netmask != "255.255.255.0" {
		t.Errorf("Expected Netmask to be '255.255.255.0', got '%s'", stats.Netmask)
	}
	if stats.DHCPServer != "Off" {
		t.Errorf("Expected DHCPServer to be 'Off', got '%s'", stats.DHCPServer)
	}

	// Interfaces
	if len(stats.Interfaces) != 5 {
		t.Fatalf("Expected 5 interfaces, got %d", len(stats.Interfaces))
	}
	// Interface 2 (Wi-Fi 2.4 GHz)
	wifi24 := stats.Interfaces[2]
	if wifi24.Name != "Wi-Fi 2.4 GHz" {
		t.Errorf("Expected interface 2 name to be 'Wi-Fi 2.4 GHz', got '%s'", wifi24.Name)
	}
	if !wifi24.Enabled {
		t.Errorf("Expected Wi-Fi 2.4 GHz enabled to be true")
	}
	if wifi24.ActiveDevices != 6 {
		t.Errorf("Expected ActiveDevices for Wi-Fi 2.4 GHz to be 6, got %f", wifi24.ActiveDevices)
	}

	// Wi-Fi Radios
	if len(stats.WiFiRadios) != 2 {
		t.Fatalf("Expected 2 Wi-Fi radios, got %d", len(stats.WiFiRadios))
	}
	radio24 := stats.WiFiRadios[0]
	if radio24.RadioStatus != "Enabled" {
		t.Errorf("Expected radio 2.4 GHz status to be 'Enabled', got '%s'", radio24.RadioStatus)
	}
	if radio24.PowerLevelPercent != 100 {
		t.Errorf("Expected radio 2.4 GHz power level to be 100, got %f", radio24.PowerLevelPercent)
	}
	if radio24.TransmitBytes != 289040566 {
		t.Errorf("Expected radio 2.4 GHz TransmitBytes to be 289040566, got %f", radio24.TransmitBytes)
	}

	radio50 := stats.WiFiRadios[1]
	if radio50.TransmitBytes != 7055713732 {
		t.Errorf("Expected radio 5 GHz TransmitBytes to be 7055713732, got %f", radio50.TransmitBytes)
	}

	// Wi-Fi Clients
	if len(stats.WiFiClients) != 12 {
		t.Fatalf("Expected 12 Wi-Fi clients, got %d", len(stats.WiFiClients))
	}
	client1 := stats.WiFiClients[0]
	if client1.MACAddress != "d8:3a:dd:90:74:48" {
		t.Errorf("Expected client 1 MACAddress to be 'd8:3a:dd:90:74:48', got '%s'", client1.MACAddress)
	}
	if client1.IPAddress != "192.168.1.53" {
		t.Errorf("Expected client 1 IPAddress to be '192.168.1.53', got '%s'", client1.IPAddress)
	}
	if client1.Band != "2.4 GHz" {
		t.Errorf("Expected client 1 Band to be '2.4 GHz', got '%s'", client1.Band)
	}
	if client1.APName != "cyrano" {
		t.Errorf("Expected client 1 APName to be 'cyrano', got '%s'", client1.APName)
	}
	if client1.SignalStrength != -32 {
		t.Errorf("Expected client 1 SignalStrength to be -32, got %f", client1.SignalStrength)
	}

	// Ports
	if len(stats.Ports) != 4 {
		t.Fatalf("Expected 4 ports, got %d", len(stats.Ports))
	}
	for i, port := range stats.Ports {
		if port.Port != "Port 1" && port.Port != "Port 2" && port.Port != "Port 3" && port.Port != "Port 4" {
			t.Errorf("Expected port name, got '%s'", port.Port)
		}
		if port.State != "down" {
			t.Errorf("Expected Port %d State to be 'down', got '%s'", i+1, port.State)
		}
	}
}
