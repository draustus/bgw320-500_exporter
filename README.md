# BGW320-500 Prometheus Exporter

A Prometheus metrics exporter written in Go for the AT&T BGW320-500 residential gateway. It scrapes the gateway's status and statistics pages, parses the HTML tables, and exposes standard Prometheus metrics via an HTTP endpoint.

## Features

- **Concurrent Scraping**: Fetches the Broadband and LAN statistics pages in parallel to minimize scrape duration.
- **Robust Parsing**: Parses HTML directly without external scripting dependencies. Tested against real BGW320-500 firmware output.
- **Comprehensive Metrics**:
  - Broadband connection source, network type, status, and IP details.
  - Broadband Ethernet WAN state, duplex, and speed.
  - Broadband IPv4 & IPv6 traffic statistics (bytes, packets, drops, errors, collisions).
  - LAN status (DHCP state, IP settings, interfaces).
  - LAN Ethernet port link state and traffic statistics.
  - LAN Wi-Fi status and traffic statistics per band (2.4 GHz & 5 GHz).
  - LAN Wi-Fi client stats (per client MAC, IP, RSSI, bytes/packets, connection drops).

## Configuration

The exporter can be configured via command-line flags or environment variables:

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--listen-address` | `LISTEN_ADDRESS` | `:9191` | The address the HTTP server listens on. |
| `--gateway-url` | `GATEWAY_URL` | `http://192.168.1.254` | Base URL of the BGW320-500 gateway. |
| `--scrape-timeout` | `SCRAPE_TIMEOUT` | `10s` | HTTP timeout when scraping the gateway. |

## How to Run

### Prerequisite

Make sure you have Go installed (Go 1.25+ is recommended).

### 1. Build

To compile the exporter binary:

```bash
go build -o bgw320-500_exporter
```

### 2. Run

Start the exporter locally:

```bash
./bgw320-500_exporter --gateway-url="http://192.168.1.254" --listen-address=":9191"
```

Once running, the metrics will be available at [http://localhost:9191/metrics](http://localhost:9191/metrics). A simple status page is served at the root `/`.

## Running Tests

To run the unit tests (which execute against captured HTML fixtures in `testdata/`):

```bash
go test -v ./...
```

## Exported Metrics

### Broadband Metrics
- `bgw_up`: Scrape success status (1 = Success, 0 = Failure).
- `bgw_scrape_duration_seconds`: Time taken to fetch and parse the gateway statistics.
- `bgw_broadband_status`: Broadband status (1 = Up, 0 = Down).
- `bgw_broadband_info`: Info metric exposing metadata (connection source, network type, IP address, MAC, DNS).
- `bgw_broadband_mtu_bytes`: MTU size (labeled by protocol: `ipv4`, `ipv6`).
- `bgw_broadband_ethernet_line_up`: WAN Ethernet physical line status.
- `bgw_broadband_ethernet_speed_mbps`: WAN Ethernet line speed.
- `bgw_broadband_ipv4_rx_bytes_total` / `bgw_broadband_ipv4_tx_bytes_total`: Total WAN IPv4 bytes.
- `bgw_broadband_ipv4_rx_packets_total` / `bgw_broadband_ipv4_tx_packets_total`: Total WAN IPv4 packets.
- `bgw_broadband_ipv4_rx_unicast_packets_total` / `bgw_broadband_ipv4_tx_unicast_packets_total`: WAN IPv4 unicast packets.
- `bgw_broadband_ipv4_rx_multicast_packets_total` / `bgw_broadband_ipv4_tx_multicast_packets_total`: WAN IPv4 multicast packets.
- `bgw_broadband_ipv4_rx_drops_total` / `bgw_broadband_ipv4_tx_drops_total`: WAN IPv4 dropped packets.
- `bgw_broadband_ipv4_rx_errors_total` / `bgw_broadband_ipv4_tx_errors_total`: WAN IPv4 packet errors.
- `bgw_broadband_ipv4_collisions_total`: WAN collision count.
- `bgw_broadband_ipv6_tx_packets_total`: WAN IPv6 packets transmitted.
- `bgw_broadband_ipv6_tx_errors_total`: WAN IPv6 transmit errors.
- `bgw_broadband_ipv6_tx_discards_total`: WAN IPv6 transmit discards.

### LAN Metrics
- `bgw_lan_info`: Info metric with LAN subnet configuration and state.
- `bgw_lan_interface_enabled`: Enabled state of LAN interfaces (Ethernet, 5G Ethernet, Wi-Fi 2.4/5GHz, Mesh).
- `bgw_lan_interface_active_devices` / `bgw_lan_interface_inactive_devices`: Count of active/inactive devices per interface.
- `bgw_lan_ipv4_rx_bytes_total` / `bgw_lan_ipv4_tx_bytes_total`: LAN-wide traffic bytes.
- `bgw_lan_ipv4_rx_packets_total` / `bgw_lan_ipv4_tx_packets_total`: LAN-wide traffic packets.
- `bgw_lan_port_enabled`: LAN port state (1 = Up, 0 = Down) per port (`Port 1` to `Port 4`).
- `bgw_lan_port_speed_mbps`: LAN port speed per port.
- `bgw_lan_port_rx_bytes_total` / `bgw_lan_port_tx_bytes_total`: Traffic bytes per physical LAN port.
- `bgw_lan_port_rx_packets_total` / `bgw_lan_port_tx_packets_total`: Traffic packets per physical LAN port.

### Wi-Fi Metrics
- `bgw_wifi_radio_enabled`: Wi-Fi band state (2.4 GHz vs 5 GHz).
- `bgw_wifi_power_level_percent`: Radio power output percentage per band.
- `bgw_wifi_rx_bytes_total` / `bgw_wifi_tx_bytes_total`: Total Wi-Fi bytes transferred.
- `bgw_wifi_rx_packets_total` / `bgw_wifi_tx_packets_total`: Total Wi-Fi packets transferred.
- `bgw_wifi_rx_errors_total` / `bgw_wifi_tx_errors_total`: Error packets count per band.
- `bgw_wifi_client_signal_strength_dbm`: RSSI signal level (dBm) per client.
- `bgw_wifi_client_rx_bytes_total` / `bgw_wifi_client_tx_bytes_total`: Bytes transferred per client.
- `bgw_wifi_client_rx_packets_total` / `bgw_wifi_client_tx_packets_total`: Packets transferred per client.
- `bgw_wifi_client_disassoc_count_total` / `bgw_wifi_client_deauth_count_total`: Connection drop event counters per client.

## Grafana Dashboard

A sample Grafana dashboard is included in `grafana/bgw320-500_dashboard.json`.

### Import
1. Open Grafana and go to Dashboards > Import.
2. Upload `grafana/bgw320-500_dashboard.json` or paste its contents.
3. Choose your Prometheus datasource.
4. Import the dashboard.

## License

This project is open-source software licensed under the MIT License.
