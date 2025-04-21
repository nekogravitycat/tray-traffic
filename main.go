package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/google/gopacket/pcap"
)

var monitor *TrafficMonitor
var status *Status

func main() {
	systray.Run(onTrayReady, onTrayExit)
}

func onTrayReady() {
	populateTray()

	status = LoadOrDefaultStatus()
	startStatusMonitor(status)

	if status.InterfaceName == "" {
		status.InterfaceName = defaultInterface()
	}
	if status.InterfaceName == "" {
		log.Fatal("No network interface found")
		return
	}
	selectInterface(status.InterfaceName)
	ifaceItems[status.InterfaceName].Check() // Check the selected interface item
}

func onTrayExit() {
	fmt.Println("Exiting...")
	if monitor != nil {
		monitor.Stop()
	}
	systray.Quit()
}

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
	saveStatus()
	log.Printf("Monitoring interface: %s\n", ifaceName)
	// Get output from the channel in a separate goroutine
	threasholdBytesStr := formatBytes(status.ThresholdBytes)
	go func() {
		for packetInfo := range monitor.Output {
			status.TotalBytes += packetInfo.SizeBytes
			packetBytesStr := formatBytes(packetInfo.SizeBytes)
			totalBytesStr := formatBytes(status.TotalBytes)
			fmt.Printf("%s -> %s (%s / %s)\n", packetInfo.SrcIP, packetInfo.DstIP, packetBytesStr, totalBytesStr)
			systray.SetTooltip(fmt.Sprintf("%s / %s", totalBytesStr, threasholdBytesStr))
		}
	}()
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

func startStatusMonitor(status *Status) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Notify if the threshold has been exceeded
			if status.TotalBytes > status.ThresholdBytes && !status.Notified {
				fmt.Printf("Threshold exceeded: %s\n", formatBytes(status.TotalBytes))
				err := beeep.Notify("Traffic Threshold Exceeded", fmt.Sprintf("You have exceeded your threshold: %s", formatBytes(status.TotalBytes)), "assets/warning.png")
				if err != nil {
					log.Println("Error sending notification:", err)
				}
				status.Notified = true
			}
			// Reset status if the date has changed
			if status.Date.Day() != time.Now().Day() {
				status.TotalBytes = 0
				status.Notified = false
				status.Date = time.Now()
			}
			// Save the status to the file
			saveStatus()
		}
	}()
}

func saveStatus() {
	err := status.Save()
	if err != nil {
		log.Println("Error saving status:", err)
	}
}
