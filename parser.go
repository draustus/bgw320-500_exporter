package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type BroadbandStats struct {
	ConnectionSource string
	Connection       string
	NetworkType      string
	IPv4Address      string
	GatewayAddress   string
	MACAddress       string
	PrimaryDNS       string
	SecondaryDNS     string
	MTU              int

	EthernetLineState string
	EthernetSpeed     float64 // in Mbps
	EthernetDuplex    string

	IPv6Status         string
	IPv6ServiceType    string
	IPv6Address        string
	IPv6LinkLocal      string
	IPv6GatewayAddress string
	IPv6MTU            int

	IPv4RxPackets   float64
	IPv4TxPackets   float64
	IPv4RxBytes     float64
	IPv4TxBytes     float64
	IPv4RxUnicast   float64
	IPv4TxUnicast   float64
	IPv4RxMulticast float64
	IPv4TxMulticast float64
	IPv4RxDrops     float64
	IPv4TxDrops     float64
	IPv4RxErrors    float64
	IPv4TxErrors    float64
	IPv4Collisions  float64

	IPv6TxPackets  float64
	IPv6TxErrors   float64
	IPv6TxDiscards float64
}

type LANInterfaceStats struct {
	Name            string
	Enabled         bool
	ActiveDevices   float64
	InactiveDevices float64
}

type LANPortStats struct {
	Port              string // "Port 1", etc.
	State             string // "up", "down"
	TransmitSpeed     float64
	TransmitPackets   float64
	TransmitBytes     float64
	TransmitUnicast   float64
	TransmitMulticast float64
	TransmitDropped   float64
	TransmitErrors    float64
	ReceivePackets    float64
	ReceiveBytes      float64
	ReceiveUnicast    float64
	ReceiveMulticast  float64
	ReceiveDropped    float64
	ReceiveErrors     float64
}

type WiFiRadioStats struct {
	Band              string // "2.4 GHz" or "5 GHz"
	RadioStatus       string // "Enabled", "Disabled"
	PowerLevelPercent float64
	TransmitBytes     float64
	ReceiveBytes      float64
	TransmitPackets   float64
	ReceivePackets    float64
	TransmitErrors    float64
	ReceiveErrors     float64
	TransmitDiscards  float64
	ReceiveDiscards   float64
}

type WiFiClientStats struct {
	MACAddress      string
	IPAddress       string
	Band            string // "2.4 GHz" or "5 GHz"
	APName          string
	TransmitPackets float64
	ReceivePackets  float64
	TransmitBytes   float64
	ReceiveBytes    float64
	TransmitErrors  float64
	SignalStrength  float64 // dBm
	DisassocCount   float64
	DeauthCount     float64
}

type LANStats struct {
	IPv4Address          string
	Netmask              string
	DHCPServer           string
	SecondarySubnet      string
	PublicSubnet         string
	CascadedRouterStatus string
	IPPassthroughStatus  string
	IPv6Status           string

	Interfaces  []LANInterfaceStats
	Ports       []LANPortStats
	WiFiRadios  []WiFiRadioStats
	WiFiClients []WiFiClientStats

	IPv4TxPackets  float64
	IPv4TxErrors   float64
	IPv4TxDiscards float64
	IPv4RxPackets  float64
	IPv4RxErrors   float64
	IPv4RxDiscards float64
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func textOf(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textOf(c))
	}
	return sb.String()
}

func getCellText(n *html.Node) string {
	return cleanText(textOf(n))
}

func findAttr(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func extractRows(n *html.Node) [][]string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "tr" {
			var row []string
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					row = append(row, getCellText(c))
				}
			}
			rows = append(rows, row)
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "table" {
				continue
			}
			walk(c)
		}
	}
	walk(n)
	return rows
}

func parsePercent(s string) float64 {
	s = strings.ReplaceAll(s, "%", "")
	s = strings.TrimSpace(s)
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func ParseBroadbandStats(r io.Reader) (*BroadbandStats, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse broadband HTML: %w", err)
	}

	stats := &BroadbandStats{}

	var findTables func(*html.Node)
	findTables = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			summary := findAttr(n, "summary")
			rows := extractRows(n)
			mapBroadbandTable(summary, rows, stats)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTables(c)
		}
	}
	findTables(doc)

	return stats, nil
}

func mapBroadbandTable(summary string, rows [][]string, stats *BroadbandStats) {
	switch summary {
	case "Summary of the most important WAN information":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			switch key {
			case "Broadband Connection Source":
				stats.ConnectionSource = val
			case "Broadband Connection":
				stats.Connection = val
			case "Broadband Network Type":
				stats.NetworkType = val
			case "Broadband IPv4 Address":
				stats.IPv4Address = val
			case "Gateway IPv4 Address":
				stats.GatewayAddress = val
			case "MAC Address":
				stats.MACAddress = val
			case "Primary DNS":
				stats.PrimaryDNS = val
			case "Secondary DNS":
				stats.SecondaryDNS = val
			case "MTU":
				stats.MTU, _ = strconv.Atoi(val)
			}
		}
	case "Ethernet Statistics Table":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			switch key {
			case "Line State":
				stats.EthernetLineState = val
			case "Current Speed (Mbps)":
				stats.EthernetSpeed, _ = strconv.ParseFloat(val, 64)
			case "Current Duplex":
				stats.EthernetDuplex = val
			}
		}
	case "IPv6 Table":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			switch key {
			case "Status":
				stats.IPv6Status = val
			case "Service Type":
				stats.IPv6ServiceType = val
			case "Global Unicast IPv6 Address":
				stats.IPv6Address = val
			case "Link Local Address":
				stats.IPv6LinkLocal = val
			case "Default IPv6 Gateway Address":
				stats.IPv6GatewayAddress = val
			case "MTU":
				stats.IPv6MTU, _ = strconv.Atoi(val)
			}
		}
	case "Ethernet IPv4 Statistics Table":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			v, _ := strconv.ParseFloat(val, 64)
			switch key {
			case "Receive Packets":
				stats.IPv4RxPackets = v
			case "Transmit Packets":
				stats.IPv4TxPackets = v
			case "Receive Bytes":
				stats.IPv4RxBytes = v
			case "Transmit Bytes":
				stats.IPv4TxBytes = v
			case "Receive Unicast":
				stats.IPv4RxUnicast = v
			case "Transmit Unicast":
				stats.IPv4TxUnicast = v
			case "Receive Multicast":
				stats.IPv4RxMulticast = v
			case "Transmit Multicast":
				stats.IPv4TxMulticast = v
			case "Receive Drops":
				stats.IPv4RxDrops = v
			case "Transmit Drops":
				stats.IPv4TxDrops = v
			case "Receive Errors":
				stats.IPv4RxErrors = v
			case "Transmit Errors":
				stats.IPv4TxErrors = v
			case "Collisions":
				stats.IPv4Collisions = v
			}
		}
	case "IPv6 Statistics Table":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			v, _ := strconv.ParseFloat(val, 64)
			switch key {
			case "Transmit Packets":
				stats.IPv6TxPackets = v
			case "Transmit Errors":
				stats.IPv6TxErrors = v
			case "Transmit Discards":
				stats.IPv6TxDiscards = v
			}
		}
	}
}

func ParseLANStats(r io.Reader) (*LANStats, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse LAN HTML: %w", err)
	}

	stats := &LANStats{}
	stats.WiFiRadios = []WiFiRadioStats{
		{Band: "2.4 GHz"},
		{Band: "5 GHz"},
	}
	stats.Ports = []LANPortStats{
		{Port: "Port 1"},
		{Port: "Port 2"},
		{Port: "Port 3"},
		{Port: "Port 4"},
	}

	var findTables func(*html.Node)
	findTables = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			summary := findAttr(n, "summary")
			rows := extractRows(n)
			mapLANTable(summary, rows, stats)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTables(c)
		}
	}
	findTables(doc)

	return stats, nil
}

func mapLANTable(summary string, rows [][]string, stats *LANStats) {
	switch summary {
	case "This table displays the critical LAN status of the device.":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			switch key {
			case "Device IPv4 Address":
				stats.IPv4Address = val
			case "DHCPv4 Netmask":
				stats.Netmask = val
			case "DHCP Server":
				stats.DHCPServer = val
			case "Secondary Subnet":
				stats.SecondarySubnet = val
			case "Public Subnet":
				stats.PublicSubnet = val
			case "Cascaded Router Status":
				stats.CascadedRouterStatus = val
			case "IP Passthrough Status":
				stats.IPPassthroughStatus = val
			}
		}
	case "This table displays the LAN Interfaces of the device.":
		for _, row := range rows {
			if len(row) < 4 || row[0] == "Interface" {
				continue
			}
			enabled := strings.TrimSpace(strings.ToLower(row[1])) == "enabled"
			active, _ := strconv.ParseFloat(row[2], 64)
			inactive, _ := strconv.ParseFloat(row[3], 64)
			stats.Interfaces = append(stats.Interfaces, LANInterfaceStats{
				Name:            row[0],
				Enabled:         enabled,
				ActiveDevices:   active,
				InactiveDevices: inactive,
			})
		}
	case "This table displays IPv6 LAN information.":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			if key == "Status" {
				stats.IPv6Status = val
			}
		}
	case "This table displays IPv4 statistics.":
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			key, val := row[0], row[1]
			v, _ := strconv.ParseFloat(val, 64)
			switch key {
			case "Transmit Packets":
				stats.IPv4TxPackets = v
			case "Transmit Errors":
				stats.IPv4TxErrors = v
			case "Transmit Discards":
				stats.IPv4TxDiscards = v
			case "Receive Packets":
				stats.IPv4RxPackets = v
			case "Receive Errors":
				stats.IPv4RxErrors = v
			case "Receive Discards":
				stats.IPv4RxDiscards = v
			}
		}
	case "This table displays Wi-Fi status.":
		for _, row := range rows {
			if len(row) < 3 {
				continue
			}
			key := row[0]
			if key == "" || key == "2.4 GHz" || key == "5 GHz" {
				continue
			}
			switch key {
			case "Wi-Fi Radio Status":
				stats.WiFiRadios[0].RadioStatus = strings.TrimSpace(row[1])
				stats.WiFiRadios[1].RadioStatus = strings.TrimSpace(row[2])
			case "Power Level":
				stats.WiFiRadios[0].PowerLevelPercent = parsePercent(row[1])
				stats.WiFiRadios[1].PowerLevelPercent = parsePercent(row[2])
			}
		}
	case "This table displays Wi-Fi Statistics.":
		for _, row := range rows {
			if len(row) < 3 {
				continue
			}
			key := row[0]
			if key == "" || key == "2.4 GHz" || key == "5 GHz" {
				continue
			}
			val24, _ := strconv.ParseFloat(row[1], 64)
			val50, _ := strconv.ParseFloat(row[2], 64)
			switch key {
			case "Transmit Bytes":
				stats.WiFiRadios[0].TransmitBytes = val24
				stats.WiFiRadios[1].TransmitBytes = val50
			case "Receive Bytes":
				stats.WiFiRadios[0].ReceiveBytes = val24
				stats.WiFiRadios[1].ReceiveBytes = val50
			case "Transmit Packets":
				stats.WiFiRadios[0].TransmitPackets = val24
				stats.WiFiRadios[1].TransmitPackets = val50
			case "Receive Packets":
				stats.WiFiRadios[0].ReceivePackets = val24
				stats.WiFiRadios[1].ReceivePackets = val50
			case "Transmit Error Packets":
				stats.WiFiRadios[0].TransmitErrors = val24
				stats.WiFiRadios[1].TransmitErrors = val50
			case "Receive Error Packets":
				stats.WiFiRadios[0].ReceiveErrors = val24
				stats.WiFiRadios[1].ReceiveErrors = val50
			case "Transmit Discard Packets":
				stats.WiFiRadios[0].TransmitDiscards = val24
				stats.WiFiRadios[1].TransmitDiscards = val50
			case "Receive Discard Packets":
				stats.WiFiRadios[0].ReceiveDiscards = val24
				stats.WiFiRadios[1].ReceiveDiscards = val50
			}
		}
	case "Wi-Fi Client Connection Statistics Table":
		for _, row := range rows {
			if len(row) < 12 || row[0] == "MAC Address" {
				continue
			}
			mac := row[0]
			authState := row[1]
			_ = authState
			ip := row[2]
			ap := row[3]
			txPackets, _ := strconv.ParseFloat(row[4], 64)
			rxPackets, _ := strconv.ParseFloat(row[5], 64)
			txBytes, _ := strconv.ParseFloat(row[6], 64)
			rxBytes, _ := strconv.ParseFloat(row[7], 64)
			txErrors, _ := strconv.ParseFloat(row[8], 64)

			sigParts := strings.Fields(row[9])
			var sigVal float64
			if len(sigParts) > 0 {
				sigVal, _ = strconv.ParseFloat(sigParts[0], 64)
			}

			disassocCount, _ := strconv.ParseFloat(row[10], 64)
			deauthCount, _ := strconv.ParseFloat(row[11], 64)

			band := "unknown"
			apName := ap
			if strings.HasPrefix(ap, "2.4 GHz") {
				band = "2.4 GHz"
				apName = strings.TrimSpace(strings.TrimPrefix(ap, "2.4 GHz"))
			} else if strings.HasPrefix(ap, "5 GHz") {
				band = "5 GHz"
				apName = strings.TrimSpace(strings.TrimPrefix(ap, "5 GHz"))
			}

			stats.WiFiClients = append(stats.WiFiClients, WiFiClientStats{
				MACAddress:      mac,
				IPAddress:       ip,
				Band:            band,
				APName:          apName,
				TransmitPackets: txPackets,
				ReceivePackets:  rxPackets,
				TransmitBytes:   txBytes,
				ReceiveBytes:    rxBytes,
				TransmitErrors:  txErrors,
				SignalStrength:  sigVal,
				DisassocCount:   disassocCount,
				DeauthCount:     deauthCount,
			})
		}
	case "LAN Ethernet Statistics Table":
		for _, row := range rows {
			if len(row) < 5 {
				continue
			}
			key := row[0]
			if key == "" || key == "Port 1" || key == "Port 2" || key == "Port 3" || key == "Port 4" {
				continue
			}
			for i := range 4 {
				valStr := row[i+1]
				if key == "State" {
					stats.Ports[i].State = valStr
				} else {
					val, _ := strconv.ParseFloat(valStr, 64)
					switch key {
					case "Transmit Speed":
						stats.Ports[i].TransmitSpeed = val
					case "Transmit Packets":
						stats.Ports[i].TransmitPackets = val
					case "Transmit Bytes":
						stats.Ports[i].TransmitBytes = val
					case "Transmit Unicast":
						stats.Ports[i].TransmitUnicast = val
					case "Transmit Multicast":
						stats.Ports[i].TransmitMulticast = val
					case "Transmit Dropped":
						stats.Ports[i].TransmitDropped = val
					case "Transmit Errors":
						stats.Ports[i].TransmitErrors = val
					case "Receive Packets":
						stats.Ports[i].ReceivePackets = val
					case "Receive Bytes":
						stats.Ports[i].ReceiveBytes = val
					case "Receive Unicast":
						stats.Ports[i].ReceiveUnicast = val
					case "Receive Multicast":
						stats.Ports[i].ReceiveMulticast = val
					case "Receive Dropped":
						stats.Ports[i].ReceiveDropped = val
					case "Receive Errors":
						stats.Ports[i].ReceiveErrors = val
					}
				}
			}
		}
	}
}
