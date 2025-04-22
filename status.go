package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

const statusFilePath = "status.json"

type Status struct {
	InterfaceName  string    `json:"interface_name"`
	TotalBytes     int       `json:"total_bytes"`
	ThresholdBytes int       `json:"threshold_bytes"`
	Notified       bool      `json:"notified"`
	Date           time.Time `json:"date"`
}

func NewStatus(interfaceName string, thresholdBytes int) *Status {
	return &Status{
		InterfaceName:  interfaceName,
		TotalBytes:     0,
		ThresholdBytes: thresholdBytes,
		Notified:       false,
		Date:           time.Now(),
	}
}

func LoadOrDefaultStatus() *Status {
	// Check if the status file exists
	if _, err := os.Stat(statusFilePath); os.IsNotExist(err) {
		// If it doesn't exist, return default values
		return NewStatus("", 1000000)
	}
	f, err := os.Open(statusFilePath)
	if err != nil {
		log.Fatalf("Error opening status file: %v\n", err)
	}
	defer f.Close()

	var s Status
	err = json.NewDecoder(f).Decode(&s)
	if err != nil {
		log.Fatalf("Error decoding status file: %v\n", err)
	}
	return &s
}

func (s *Status) Save() error {
	f, err := os.Create(statusFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(s)
}

func startStatusMonitor(status *Status) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Notify if the threshold has been exceeded
			if status.TotalBytes > status.ThresholdBytes && !status.Notified {
				totalBytesStr := formatBytes(status.TotalBytes)
				thresholdBytesStr := formatBytes(status.ThresholdBytes)
				fmt.Printf("Threshold exceeded: %s / %s\n", totalBytesStr, thresholdBytesStr)
				err := beeep.Notify(
					"Traffic Threshold Exceeded",
					fmt.Sprintf("You have exceeded your threshold: %s / %s", totalBytesStr, thresholdBytesStr),
					*iconPath,
				)
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
			err := status.Save()
			if err != nil {
				log.Println("Error saving status:", err)
			}
		}
	}()
}
