package main

import (
	"fmt"
	"log"

	"github.com/gen2brain/beeep"
	"github.com/google/gopacket/pcap"
)

func selectInterface(ifaceName string) {
	// Stop the current monitor
	if monitor != nil {
		monitor.Stop()
	}
	// Start a new monitor with the selected interface
	monitor = NewTrafficMonitor(ifaceName, true)
	err := monitor.Start()
	if err != nil {
		log.Println("Error starting monitor:", err)
		return
	}
	// Update the status with the new interface name
	status.InterfaceName = ifaceName
	err = status.Save()
	if err != nil {
		log.Println("Error saving status:", err)
	}
	log.Printf("Monitoring interface: %s\n", ifaceName)
	totalBytesStr := formatBytes(status.TotalBytes)
	threasholdBytesStr := formatBytes(status.ThresholdBytes)
	beeep.Notify("Start monitoring", fmt.Sprintf("Current usage: %s / %s", totalBytesStr, threasholdBytesStr), *iconPath)
	// Get output from the channel in a separate goroutine
	go processPackets()
}

func processPackets() {
	for packetInfo := range monitor.Output {
		status.TotalBytes += packetInfo.SizeBytes
		//log.Printf("%s -> %s (%dB / %s)\n", packetInfo.SrcIP, packetInfo.DstIP, packetInfo.SizeBytes, formatBytes(status.TotalBytes))
	}
}

func defaultInterface() string {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Println("Error finding devices:", err)
		return ""
	}
	for _, dev := range devices {
		if len(dev.Addresses) == 0 {
			continue
		}
		return dev.Name
	}
	return ""
}
