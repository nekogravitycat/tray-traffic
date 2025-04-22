package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
)

var iconPath *string
var monitor *TrafficMonitor
var status *Status

func main() {
	// Set the icon path for the tray icon
	iconPathStr, err := saveIconToTemp()
	iconPath = &iconPathStr
	if err != nil {
		log.Printf("Error saving icon to temp: %v", err)
	}

	// Ensure only one instance of the application is running
	f, err := ensureSingleton()
	if err != nil {
		beeep.Notify("Error starting application", err.Error(), *iconPath)
		log.Fatalf("Error ensuring singleton: %v", err)
		return
	}
	defer f.Unlock()

	systray.Run(onTrayReady, onTrayExit)
}

func ensureSingleton() (*flock.Flock, error) {
	const lockFileName = "tray-traffic.lock"
	lockPath := filepath.Join(os.TempDir(), lockFileName)

	f := flock.New(lockPath)
	locked, err := f.TryLock()

	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}

	if !locked {
		return nil, fmt.Errorf("another instance is already running")
	}

	return f, nil
}
