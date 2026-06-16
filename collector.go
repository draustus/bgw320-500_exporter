package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type BGWCollector struct {
	gatewayURL string
	client     *http.Client

	// Descriptors
	upDesc             *prometheus.Desc
	scrapeDurationDesc *prometheus.Desc

	// Broadband Info & Connection
	broadbandStatusDesc *prometheus.Desc
	broadbandInfoDesc   *prometheus.Desc
	broadbandMtuDesc    *prometheus.Desc

	// Broadband Ethernet
	ethernetLineStateDesc *prometheus.Desc
	ethernetSpeedDesc     *prometheus.Desc

	// Broadband IPv4 Statistics
	ipv4RxPacketsDesc   *prometheus.Desc
	ipv4TxPacketsDesc   *prometheus.Desc
	ipv4RxBytesDesc     *prometheus.Desc
	ipv4TxBytesDesc     *prometheus.Desc
	ipv4RxUnicastDesc   *prometheus.Desc
	ipv4TxUnicastDesc   *prometheus.Desc
	ipv4RxMulticastDesc *prometheus.Desc
	ipv4TxMulticastDesc *prometheus.Desc
	ipv4RxDropsDesc     *prometheus.Desc
	ipv4TxDropsDesc     *prometheus.Desc
	ipv4RxErrorsDesc    *prometheus.Desc
	ipv4TxErrorsDesc    *prometheus.Desc
	ipv4CollisionsDesc  *prometheus.Desc

	// Broadband IPv6 Statistics
	ipv6TxPacketsDesc  *prometheus.Desc
	ipv6TxErrorsDesc   *prometheus.Desc
	ipv6TxDiscardsDesc *prometheus.Desc

	// LAN Info
	lanInfoDesc *prometheus.Desc

	// LAN Interfaces
	lanInterfaceActiveDesc   *prometheus.Desc
	lanInterfaceInactiveDesc *prometheus.Desc
	lanInterfaceEnabledDesc  *prometheus.Desc

	// LAN IPv4 Statistics
	lanIpv4TxPacketsDesc  *prometheus.Desc
	lanIpv4TxErrorsDesc   *prometheus.Desc
	lanIpv4TxDiscardsDesc *prometheus.Desc
	lanIpv4RxPacketsDesc  *prometheus.Desc
	lanIpv4RxErrorsDesc   *prometheus.Desc
	lanIpv4RxDiscardsDesc *prometheus.Desc

	// LAN Wi-Fi
	wifiRadioEnabledDesc *prometheus.Desc
	wifiPowerLevelDesc   *prometheus.Desc
	wifiTxBytesDesc      *prometheus.Desc
	wifiRxBytesDesc      *prometheus.Desc
	wifiTxPacketsDesc    *prometheus.Desc
	wifiRxPacketsDesc    *prometheus.Desc
	wifiTxErrorsDesc     *prometheus.Desc
	wifiRxErrorsDesc     *prometheus.Desc
	wifiTxDiscardsDesc   *prometheus.Desc
	wifiRxDiscardsDesc   *prometheus.Desc

	// LAN Wi-Fi Clients
	wifiClientTxPacketsDesc *prometheus.Desc
	wifiClientRxPacketsDesc *prometheus.Desc
	wifiClientTxBytesDesc   *prometheus.Desc
	wifiClientRxBytesDesc   *prometheus.Desc
	wifiClientTxErrorsDesc  *prometheus.Desc
	wifiClientSignalDesc    *prometheus.Desc
	wifiClientDisassocDesc  *prometheus.Desc
	wifiClientDeauthDesc    *prometheus.Desc

	// LAN Ports
	lanPortSpeedDesc       *prometheus.Desc
	lanPortEnabledDesc     *prometheus.Desc
	lanPortTxPacketsDesc   *prometheus.Desc
	lanPortTxBytesDesc     *prometheus.Desc
	lanPortTxUnicastDesc   *prometheus.Desc
	lanPortTxMulticastDesc *prometheus.Desc
	lanPortTxDroppedDesc   *prometheus.Desc
	lanPortTxErrorsDesc    *prometheus.Desc
	lanPortRxPacketsDesc   *prometheus.Desc
	lanPortRxBytesDesc     *prometheus.Desc
	lanPortRxUnicastDesc   *prometheus.Desc
	lanPortRxMulticastDesc *prometheus.Desc
	lanPortRxDroppedDesc   *prometheus.Desc
	lanPortRxErrorsDesc    *prometheus.Desc
}

func NewBGWCollector(gatewayURL string, insecureSkipVerify bool, timeout time.Duration) *BGWCollector {
	tr := &http.Transport{
		// #nosec G402: this exporter targets a local router with a self-signed certificate.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return &BGWCollector{
		gatewayURL: gatewayURL,
		client:     client,

		upDesc: prometheus.NewDesc(
			"bgw_up",
			"Was the last scrape of BGW statistics pages successful.",
			nil, nil,
		),
		scrapeDurationDesc: prometheus.NewDesc(
			"bgw_scrape_duration_seconds",
			"Duration of the scrape.",
			nil, nil,
		),

		// Broadband Info
		broadbandStatusDesc: prometheus.NewDesc(
			"bgw_broadband_status",
			"Broadband connection status (1 = Up, 0 = Down).",
			nil, nil,
		),
		broadbandInfoDesc: prometheus.NewDesc(
			"bgw_broadband_info",
			"Metadata about the broadband connection.",
			[]string{"connection_source", "network_type", "ipv4_address", "gateway_ipv4_address", "mac_address", "primary_dns", "secondary_dns", "ipv6_status", "ipv6_service_type", "ipv6_address", "ipv6_gateway_address"}, nil,
		),
		broadbandMtuDesc: prometheus.NewDesc(
			"bgw_broadband_mtu_bytes",
			"Broadband Maximum Transmit Unit size in bytes.",
			[]string{"proto"}, nil,
		),

		// Broadband Ethernet
		ethernetLineStateDesc: prometheus.NewDesc(
			"bgw_broadband_ethernet_line_up",
			"State of the WAN Ethernet line (1 = Up, 0 = Down).",
			nil, nil,
		),
		ethernetSpeedDesc: prometheus.NewDesc(
			"bgw_broadband_ethernet_speed_mbps",
			"Current WAN Ethernet line speed in Mbps.",
			nil, nil,
		),

		// Broadband IPv4 Stats
		ipv4RxPacketsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_packets_total",
			"Total IPv4 packets received on WAN.",
			nil, nil,
		),
		ipv4TxPacketsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_packets_total",
			"Total IPv4 packets transmitted on WAN.",
			nil, nil,
		),
		ipv4RxBytesDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_bytes_total",
			"Total IPv4 bytes received on WAN.",
			nil, nil,
		),
		ipv4TxBytesDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_bytes_total",
			"Total IPv4 bytes transmitted on WAN.",
			nil, nil,
		),
		ipv4RxUnicastDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_unicast_packets_total",
			"Total IPv4 unicast packets received on WAN.",
			nil, nil,
		),
		ipv4TxUnicastDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_unicast_packets_total",
			"Total IPv4 unicast packets transmitted on WAN.",
			nil, nil,
		),
		ipv4RxMulticastDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_multicast_packets_total",
			"Total IPv4 multicast packets received on WAN.",
			nil, nil,
		),
		ipv4TxMulticastDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_multicast_packets_total",
			"Total IPv4 multicast packets transmitted on WAN.",
			nil, nil,
		),
		ipv4RxDropsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_drops_total",
			"Total IPv4 packets dropped on receive on WAN.",
			nil, nil,
		),
		ipv4TxDropsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_drops_total",
			"Total IPv4 packets dropped on transmit on WAN.",
			nil, nil,
		),
		ipv4RxErrorsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_rx_errors_total",
			"Total IPv4 receive errors on WAN.",
			nil, nil,
		),
		ipv4TxErrorsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_tx_errors_total",
			"Total IPv4 transmit errors on WAN.",
			nil, nil,
		),
		ipv4CollisionsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv4_collisions_total",
			"Total collisions on WAN.",
			nil, nil,
		),

		// Broadband IPv6 Stats
		ipv6TxPacketsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv6_tx_packets_total",
			"Total IPv6 packets transmitted on WAN.",
			nil, nil,
		),
		ipv6TxErrorsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv6_tx_errors_total",
			"Total IPv6 transmit errors on WAN.",
			nil, nil,
		),
		ipv6TxDiscardsDesc: prometheus.NewDesc(
			"bgw_broadband_ipv6_tx_discards_total",
			"Total IPv6 packets discarded on transmit on WAN.",
			nil, nil,
		),

		// LAN Info
		lanInfoDesc: prometheus.NewDesc(
			"bgw_lan_info",
			"Metadata about the LAN configuration.",
			[]string{"device_ipv4_address", "netmask", "dhcp_server", "secondary_subnet", "public_subnet", "cascaded_router_status", "ip_passthrough_status", "ipv6_status"}, nil,
		),

		// LAN Interfaces
		lanInterfaceActiveDesc: prometheus.NewDesc(
			"bgw_lan_interface_active_devices",
			"Number of active devices on the LAN interface.",
			[]string{"interface"}, nil,
		),
		lanInterfaceInactiveDesc: prometheus.NewDesc(
			"bgw_lan_interface_inactive_devices",
			"Number of inactive devices on the LAN interface.",
			[]string{"interface"}, nil,
		),
		lanInterfaceEnabledDesc: prometheus.NewDesc(
			"bgw_lan_interface_enabled",
			"Whether the LAN interface is enabled (1 = Enabled, 0 = Disabled).",
			[]string{"interface"}, nil,
		),

		// LAN IPv4 Stats
		lanIpv4TxPacketsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_tx_packets_total",
			"Total IPv4 packets transmitted on LAN.",
			nil, nil,
		),
		lanIpv4TxErrorsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_tx_errors_total",
			"Total IPv4 transmit errors on LAN.",
			nil, nil,
		),
		lanIpv4TxDiscardsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_tx_discards_total",
			"Total IPv4 packets discarded on transmit on LAN.",
			nil, nil,
		),
		lanIpv4RxPacketsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_rx_packets_total",
			"Total IPv4 packets received on LAN.",
			nil, nil,
		),
		lanIpv4RxErrorsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_rx_errors_total",
			"Total IPv4 receive errors on LAN.",
			nil, nil,
		),
		lanIpv4RxDiscardsDesc: prometheus.NewDesc(
			"bgw_lan_ipv4_rx_discards_total",
			"Total IPv4 packets discarded on receive on LAN.",
			nil, nil,
		),

		// Wi-Fi Radios
		wifiRadioEnabledDesc: prometheus.NewDesc(
			"bgw_wifi_radio_enabled",
			"Whether the Wi-Fi radio band is enabled (1 = Enabled, 0 = Disabled).",
			[]string{"band"}, nil,
		),
		wifiPowerLevelDesc: prometheus.NewDesc(
			"bgw_wifi_power_level_percent",
			"Wi-Fi radio power level percentage.",
			[]string{"band"}, nil,
		),
		wifiTxBytesDesc: prometheus.NewDesc(
			"bgw_wifi_tx_bytes_total",
			"Total bytes transmitted on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiRxBytesDesc: prometheus.NewDesc(
			"bgw_wifi_rx_bytes_total",
			"Total bytes received on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiTxPacketsDesc: prometheus.NewDesc(
			"bgw_wifi_tx_packets_total",
			"Total packets transmitted on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiRxPacketsDesc: prometheus.NewDesc(
			"bgw_wifi_rx_packets_total",
			"Total packets received on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiTxErrorsDesc: prometheus.NewDesc(
			"bgw_wifi_tx_errors_total",
			"Total transmit error packets on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiRxErrorsDesc: prometheus.NewDesc(
			"bgw_wifi_rx_errors_total",
			"Total receive error packets on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiTxDiscardsDesc: prometheus.NewDesc(
			"bgw_wifi_tx_discards_total",
			"Total transmit discard packets on Wi-Fi band.",
			[]string{"band"}, nil,
		),
		wifiRxDiscardsDesc: prometheus.NewDesc(
			"bgw_wifi_rx_discards_total",
			"Total receive discard packets on Wi-Fi band.",
			[]string{"band"}, nil,
		),

		// Wi-Fi Clients
		wifiClientTxPacketsDesc: prometheus.NewDesc(
			"bgw_wifi_client_tx_packets_total",
			"Total packets transmitted to Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientRxPacketsDesc: prometheus.NewDesc(
			"bgw_wifi_client_rx_packets_total",
			"Total packets received from Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientTxBytesDesc: prometheus.NewDesc(
			"bgw_wifi_client_tx_bytes_total",
			"Total bytes transmitted to Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientRxBytesDesc: prometheus.NewDesc(
			"bgw_wifi_client_rx_bytes_total",
			"Total bytes received from Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientTxErrorsDesc: prometheus.NewDesc(
			"bgw_wifi_client_tx_errors_total",
			"Total transmit errors for Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientSignalDesc: prometheus.NewDesc(
			"bgw_wifi_client_signal_strength_dbm",
			"Signal strength of Wi-Fi client in dBm.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientDisassocDesc: prometheus.NewDesc(
			"bgw_wifi_client_disassoc_count_total",
			"Total disassociation count for Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),
		wifiClientDeauthDesc: prometheus.NewDesc(
			"bgw_wifi_client_deauth_count_total",
			"Total deauthentication count for Wi-Fi client.",
			[]string{"mac", "ip", "band", "ap"}, nil,
		),

		// LAN Ports
		lanPortSpeedDesc: prometheus.NewDesc(
			"bgw_lan_port_speed_mbps",
			"LAN port speed in Mbps.",
			[]string{"port"}, nil,
		),
		lanPortEnabledDesc: prometheus.NewDesc(
			"bgw_lan_port_enabled",
			"Whether the LAN port state is up (1 = Up, 0 = Down).",
			[]string{"port"}, nil,
		),
		lanPortTxPacketsDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_packets_total",
			"Total packets transmitted on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortTxBytesDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_bytes_total",
			"Total bytes transmitted on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortTxUnicastDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_unicast_packets_total",
			"Total unicast packets transmitted on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortTxMulticastDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_multicast_packets_total",
			"Total multicast packets transmitted on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortTxDroppedDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_dropped_total",
			"Total dropped transmit packets on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortTxErrorsDesc: prometheus.NewDesc(
			"bgw_lan_port_tx_errors_total",
			"Total transmit errors on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxPacketsDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_packets_total",
			"Total packets received on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxBytesDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_bytes_total",
			"Total bytes received on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxUnicastDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_unicast_packets_total",
			"Total unicast packets received on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxMulticastDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_multicast_packets_total",
			"Total multicast packets received on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxDroppedDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_dropped_total",
			"Total dropped receive packets on LAN port.",
			[]string{"port"}, nil,
		),
		lanPortRxErrorsDesc: prometheus.NewDesc(
			"bgw_lan_port_rx_errors_total",
			"Total receive errors on LAN port.",
			[]string{"port"}, nil,
		),
	}
}

func (c *BGWCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upDesc
	ch <- c.scrapeDurationDesc
	ch <- c.broadbandStatusDesc
	ch <- c.broadbandInfoDesc
	ch <- c.broadbandMtuDesc
	ch <- c.ethernetLineStateDesc
	ch <- c.ethernetSpeedDesc
	ch <- c.ipv4RxPacketsDesc
	ch <- c.ipv4TxPacketsDesc
	ch <- c.ipv4RxBytesDesc
	ch <- c.ipv4TxBytesDesc
	ch <- c.ipv4RxUnicastDesc
	ch <- c.ipv4TxUnicastDesc
	ch <- c.ipv4RxMulticastDesc
	ch <- c.ipv4TxMulticastDesc
	ch <- c.ipv4RxDropsDesc
	ch <- c.ipv4TxDropsDesc
	ch <- c.ipv4RxErrorsDesc
	ch <- c.ipv4TxErrorsDesc
	ch <- c.ipv4CollisionsDesc
	ch <- c.ipv6TxPacketsDesc
	ch <- c.ipv6TxErrorsDesc
	ch <- c.ipv6TxDiscardsDesc

	ch <- c.lanInfoDesc
	ch <- c.lanInterfaceActiveDesc
	ch <- c.lanInterfaceInactiveDesc
	ch <- c.lanInterfaceEnabledDesc
	ch <- c.lanIpv4TxPacketsDesc
	ch <- c.lanIpv4TxErrorsDesc
	ch <- c.lanIpv4TxDiscardsDesc
	ch <- c.lanIpv4RxPacketsDesc
	ch <- c.lanIpv4RxErrorsDesc
	ch <- c.lanIpv4RxDiscardsDesc

	ch <- c.wifiRadioEnabledDesc
	ch <- c.wifiPowerLevelDesc
	ch <- c.wifiTxBytesDesc
	ch <- c.wifiRxBytesDesc
	ch <- c.wifiTxPacketsDesc
	ch <- c.wifiRxPacketsDesc
	ch <- c.wifiTxErrorsDesc
	ch <- c.wifiRxErrorsDesc
	ch <- c.wifiTxDiscardsDesc
	ch <- c.wifiRxDiscardsDesc

	ch <- c.wifiClientTxPacketsDesc
	ch <- c.wifiClientRxPacketsDesc
	ch <- c.wifiClientTxBytesDesc
	ch <- c.wifiClientRxBytesDesc
	ch <- c.wifiClientTxErrorsDesc
	ch <- c.wifiClientSignalDesc
	ch <- c.wifiClientDisassocDesc
	ch <- c.wifiClientDeauthDesc

	ch <- c.lanPortSpeedDesc
	ch <- c.lanPortEnabledDesc
	ch <- c.lanPortTxPacketsDesc
	ch <- c.lanPortTxBytesDesc
	ch <- c.lanPortTxUnicastDesc
	ch <- c.lanPortTxMulticastDesc
	ch <- c.lanPortTxDroppedDesc
	ch <- c.lanPortTxErrorsDesc
	ch <- c.lanPortRxPacketsDesc
	ch <- c.lanPortRxBytesDesc
	ch <- c.lanPortRxUnicastDesc
	ch <- c.lanPortRxMulticastDesc
	ch <- c.lanPortRxDroppedDesc
	ch <- c.lanPortRxErrorsDesc
}

func (c *BGWCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	type bbRes struct {
		stats *BroadbandStats
		err   error
	}

	type lanRes struct {
		stats *LANStats
		err   error
	}

	bbChan := make(chan bbRes, 1)
	lanChan := make(chan lanRes, 1)

	go func() {
		stats, err := c.scrapeBroadband()
		bbChan <- bbRes{stats, err}
	}()

	go func() {
		stats, err := c.scrapeLAN()
		lanChan <- lanRes{stats, err}
	}()

	bRes := <-bbChan
	lRes := <-lanChan

	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue, duration)

	if bRes.err != nil || lRes.err != nil {
		ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 0)
		slog.Error("Error scraping BGW statistics",
			"broadband_error", bRes.err,
			"lan_error", lRes.err,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 1)

	c.collectBroadband(ch, bRes.stats)
	c.collectLAN(ch, lRes.stats)
}

func (c *BGWCollector) scrapeBroadband() (*BroadbandStats, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.gatewayURL+"/cgi-bin/broadbandstatistics.ha", nil)
	if err != nil {
		return nil, fmt.Errorf("create broadband request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch broadband statistics: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return ParseBroadbandStats(resp.Body)
}

func (c *BGWCollector) scrapeLAN() (*LANStats, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.gatewayURL+"/cgi-bin/lanstatistics.ha", nil)
	if err != nil {
		return nil, fmt.Errorf("create LAN request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch LAN statistics: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return ParseLANStats(resp.Body)
}

func (c *BGWCollector) collectBroadband(ch chan<- prometheus.Metric, stats *BroadbandStats) {
	// Broadband Status
	statusVal := 0.0
	if stats.Connection == "Up" {
		statusVal = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.broadbandStatusDesc, prometheus.GaugeValue, statusVal)

	// Broadband Info
	ch <- prometheus.MustNewConstMetric(c.broadbandInfoDesc, prometheus.GaugeValue, 1.0,
		stats.ConnectionSource,
		stats.NetworkType,
		stats.IPv4Address,
		stats.GatewayAddress,
		stats.MACAddress,
		stats.PrimaryDNS,
		stats.SecondaryDNS,
		stats.IPv6Status,
		stats.IPv6ServiceType,
		stats.IPv6Address,
		stats.IPv6GatewayAddress,
	)

	// MTU
	ch <- prometheus.MustNewConstMetric(c.broadbandMtuDesc, prometheus.GaugeValue, float64(stats.MTU), "ipv4")
	if stats.IPv6MTU > 0 {
		ch <- prometheus.MustNewConstMetric(c.broadbandMtuDesc, prometheus.GaugeValue, float64(stats.IPv6MTU), "ipv6")
	}

	// Ethernet line status
	ethLineVal := 0.0
	if stats.EthernetLineState == "Up" {
		ethLineVal = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.ethernetLineStateDesc, prometheus.GaugeValue, ethLineVal)
	ch <- prometheus.MustNewConstMetric(c.ethernetSpeedDesc, prometheus.GaugeValue, stats.EthernetSpeed)

	// Broadband IPv4 statistics
	ch <- prometheus.MustNewConstMetric(c.ipv4RxPacketsDesc, prometheus.CounterValue, stats.IPv4RxPackets)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxPacketsDesc, prometheus.CounterValue, stats.IPv4TxPackets)
	ch <- prometheus.MustNewConstMetric(c.ipv4RxBytesDesc, prometheus.CounterValue, stats.IPv4RxBytes)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxBytesDesc, prometheus.CounterValue, stats.IPv4TxBytes)
	ch <- prometheus.MustNewConstMetric(c.ipv4RxUnicastDesc, prometheus.CounterValue, stats.IPv4RxUnicast)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxUnicastDesc, prometheus.CounterValue, stats.IPv4TxUnicast)
	ch <- prometheus.MustNewConstMetric(c.ipv4RxMulticastDesc, prometheus.CounterValue, stats.IPv4RxMulticast)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxMulticastDesc, prometheus.CounterValue, stats.IPv4TxMulticast)
	ch <- prometheus.MustNewConstMetric(c.ipv4RxDropsDesc, prometheus.CounterValue, stats.IPv4RxDrops)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxDropsDesc, prometheus.CounterValue, stats.IPv4TxDrops)
	ch <- prometheus.MustNewConstMetric(c.ipv4RxErrorsDesc, prometheus.CounterValue, stats.IPv4RxErrors)
	ch <- prometheus.MustNewConstMetric(c.ipv4TxErrorsDesc, prometheus.CounterValue, stats.IPv4TxErrors)
	ch <- prometheus.MustNewConstMetric(c.ipv4CollisionsDesc, prometheus.CounterValue, stats.IPv4Collisions)

	// Broadband IPv6 statistics
	ch <- prometheus.MustNewConstMetric(c.ipv6TxPacketsDesc, prometheus.CounterValue, stats.IPv6TxPackets)
	ch <- prometheus.MustNewConstMetric(c.ipv6TxErrorsDesc, prometheus.CounterValue, stats.IPv6TxErrors)
	ch <- prometheus.MustNewConstMetric(c.ipv6TxDiscardsDesc, prometheus.CounterValue, stats.IPv6TxDiscards)
}

func (c *BGWCollector) collectLAN(ch chan<- prometheus.Metric, stats *LANStats) {
	// LAN Info
	ch <- prometheus.MustNewConstMetric(c.lanInfoDesc, prometheus.GaugeValue, 1.0,
		stats.IPv4Address,
		stats.Netmask,
		stats.DHCPServer,
		stats.SecondarySubnet,
		stats.PublicSubnet,
		stats.CascadedRouterStatus,
		stats.IPPassthroughStatus,
		stats.IPv6Status,
	)

	// Interfaces
	for _, inf := range stats.Interfaces {
		infEnabledVal := 0.0
		if inf.Enabled {
			infEnabledVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.lanInterfaceEnabledDesc, prometheus.GaugeValue, infEnabledVal, inf.Name)
		ch <- prometheus.MustNewConstMetric(c.lanInterfaceActiveDesc, prometheus.GaugeValue, inf.ActiveDevices, inf.Name)
		ch <- prometheus.MustNewConstMetric(c.lanInterfaceInactiveDesc, prometheus.GaugeValue, inf.InactiveDevices, inf.Name)
	}

	// LAN IPv4 statistics
	ch <- prometheus.MustNewConstMetric(c.lanIpv4TxPacketsDesc, prometheus.CounterValue, stats.IPv4TxPackets)
	ch <- prometheus.MustNewConstMetric(c.lanIpv4TxErrorsDesc, prometheus.CounterValue, stats.IPv4TxErrors)
	ch <- prometheus.MustNewConstMetric(c.lanIpv4TxDiscardsDesc, prometheus.CounterValue, stats.IPv4TxDiscards)
	ch <- prometheus.MustNewConstMetric(c.lanIpv4RxPacketsDesc, prometheus.CounterValue, stats.IPv4RxPackets)
	ch <- prometheus.MustNewConstMetric(c.lanIpv4RxErrorsDesc, prometheus.CounterValue, stats.IPv4RxErrors)
	ch <- prometheus.MustNewConstMetric(c.lanIpv4RxDiscardsDesc, prometheus.CounterValue, stats.IPv4RxDiscards)

	// Wi-Fi Radios
	for _, r := range stats.WiFiRadios {
		radioEnabledVal := 0.0
		if r.RadioStatus == "Enabled" {
			radioEnabledVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.wifiRadioEnabledDesc, prometheus.GaugeValue, radioEnabledVal, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiPowerLevelDesc, prometheus.GaugeValue, r.PowerLevelPercent, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiTxBytesDesc, prometheus.CounterValue, r.TransmitBytes, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiRxBytesDesc, prometheus.CounterValue, r.ReceiveBytes, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiTxPacketsDesc, prometheus.CounterValue, r.TransmitPackets, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiRxPacketsDesc, prometheus.CounterValue, r.ReceivePackets, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiTxErrorsDesc, prometheus.CounterValue, r.TransmitErrors, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiRxErrorsDesc, prometheus.CounterValue, r.ReceiveErrors, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiTxDiscardsDesc, prometheus.CounterValue, r.TransmitDiscards, r.Band)
		ch <- prometheus.MustNewConstMetric(c.wifiRxDiscardsDesc, prometheus.CounterValue, r.ReceiveDiscards, r.Band)
	}

	// Wi-Fi Clients
	for _, cl := range stats.WiFiClients {
		ch <- prometheus.MustNewConstMetric(c.wifiClientTxPacketsDesc, prometheus.CounterValue, cl.TransmitPackets, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientRxPacketsDesc, prometheus.CounterValue, cl.ReceivePackets, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientTxBytesDesc, prometheus.CounterValue, cl.TransmitBytes, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientRxBytesDesc, prometheus.CounterValue, cl.ReceiveBytes, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientTxErrorsDesc, prometheus.CounterValue, cl.TransmitErrors, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientSignalDesc, prometheus.GaugeValue, cl.SignalStrength, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientDisassocDesc, prometheus.CounterValue, cl.DisassocCount, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
		ch <- prometheus.MustNewConstMetric(c.wifiClientDeauthDesc, prometheus.CounterValue, cl.DeauthCount, cl.MACAddress, cl.IPAddress, cl.Band, cl.APName)
	}

	// LAN Ports
	for _, p := range stats.Ports {
		portEnabledVal := 0.0
		if p.State == "up" {
			portEnabledVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.lanPortEnabledDesc, prometheus.GaugeValue, portEnabledVal, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortSpeedDesc, prometheus.GaugeValue, p.TransmitSpeed, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxPacketsDesc, prometheus.CounterValue, p.TransmitPackets, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxBytesDesc, prometheus.CounterValue, p.TransmitBytes, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxUnicastDesc, prometheus.CounterValue, p.TransmitUnicast, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxMulticastDesc, prometheus.CounterValue, p.TransmitMulticast, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxDroppedDesc, prometheus.CounterValue, p.TransmitDropped, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortTxErrorsDesc, prometheus.CounterValue, p.TransmitErrors, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxPacketsDesc, prometheus.CounterValue, p.ReceivePackets, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxBytesDesc, prometheus.CounterValue, p.ReceiveBytes, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxUnicastDesc, prometheus.CounterValue, p.ReceiveUnicast, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxMulticastDesc, prometheus.CounterValue, p.ReceiveMulticast, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxDroppedDesc, prometheus.CounterValue, p.ReceiveDropped, p.Port)
		ch <- prometheus.MustNewConstMetric(c.lanPortRxErrorsDesc, prometheus.CounterValue, p.ReceiveErrors, p.Port)
	}
}
