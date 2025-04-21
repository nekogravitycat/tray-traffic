package main

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketInfo struct {
	SrcIP     string
	DstIP     string
	SizeBytes int
}

type TrafficMonitor struct {
	InterfaceName string
	ExcludeLocal  bool
	Output        chan PacketInfo
	stopCh        chan struct{}
}

func NewTrafficMonitor(interfaceName string, excludeLocal bool) *TrafficMonitor {
	return &TrafficMonitor{
		InterfaceName: interfaceName,
		ExcludeLocal:  excludeLocal,
		Output:        make(chan PacketInfo),
		stopCh:        make(chan struct{}),
	}
}

func (monitor *TrafficMonitor) Start() error {
	// Open the network interface for packet capturing
	handle, err := pcap.OpenLive(monitor.InterfaceName, 128, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	// Set a BPF filter to capture only IP and IPv6 packets
	err = handle.SetBPFFilter("ip or ip6")
	if err != nil {
		return err
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// Start the monitoring loop in a goroutine
	go func() {
		defer handle.Close()

		for {
			select {
			case <-monitor.stopCh:
				return
			case packet := <-packetSource.Packets():
				if packet == nil {
					continue
				}

				// Extract IP layer (either IPv4 or IPv6)
				ipLayer := packet.Layer(layers.LayerTypeIPv4)
				if ipLayer == nil {
					ipLayer = packet.Layer(layers.LayerTypeIPv6)
				}
				if ipLayer == nil {
					continue
				}

				var srcIP, dstIP string
				switch layer := ipLayer.(type) {
				case *layers.IPv4:
					srcIP = layer.SrcIP.String()
					dstIP = layer.DstIP.String()
				case *layers.IPv6:
					srcIP = layer.SrcIP.String()
					dstIP = layer.DstIP.String()
				}

				// Skip local traffic if the flag is enabled
				if monitor.ExcludeLocal && isLocalIP(srcIP) && isLocalIP(dstIP) {
					continue
				}

				// Send the packet info through the output channel
				monitor.Output <- PacketInfo{
					SrcIP:     srcIP,
					DstIP:     dstIP,
					SizeBytes: packet.Metadata().Length,
				}
			}
		}
	}()
	return nil
}

// Stop halts the monitoring process
func (monitor *TrafficMonitor) Stop() {
	select {
	case <-monitor.stopCh:
		// already closed or used
	default:
		close(monitor.stopCh) // Safely close to notify goroutine to exit
	}
}

// isLocalIP checks whether an IP address is a private (local) address.
func isLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.IsPrivate()
}

// Format the byte size into a human-readable string (KB, MB, GB, etc.)
func formatBytes(bytes int) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
