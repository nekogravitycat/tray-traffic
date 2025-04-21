package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Status struct {
	InterfaceName  string    `json:"interface_name"`
	TotalBytes     int       `json:"total_bytes"`
	ThresholdBytes int       `json:"threshold_bytes"`
	Notified       bool      `json:"notified"`
	Date           time.Time `json:"date"`
}

const statusFilePath = "status.json"

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

func NewStatus(interfaceName string, thresholdBytes int) *Status {
	return &Status{
		InterfaceName:  interfaceName,
		TotalBytes:     0,
		ThresholdBytes: thresholdBytes,
		Notified:       false,
		Date:           time.Now(),
	}
}
