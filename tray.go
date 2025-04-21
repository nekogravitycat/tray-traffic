package main

import (
	"log"

	"github.com/getlantern/systray"
	"github.com/google/gopacket/pcap"
)

var ifaceItems = make(map[string]*systray.MenuItem)

func populateTray() {
	systray.SetTitle("Tray Traffic Monitor")
	systray.SetTooltip("Monitor network traffic and notify on threshold exceedance")

	// Create the menu for selecting the interface
	ifacesMenu := systray.AddMenuItem("Select Interface", "Select the network interface to monitor")
	// Get all network interfaces
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Error fetching network devices:", err)
	}
	// Create a sub-menu for each network interface
	for k := range ifaceItems {
		delete(ifaceItems, k)
	}
	for _, device := range devices {
		item := ifacesMenu.AddSubMenuItemCheckbox(device.Description, device.Name, false)
		ifaceItems[device.Name] = item
		// Check for click events on the interface items
		go func(clickedItem *systray.MenuItem, clickedIfaceName string) {
			for range clickedItem.ClickedCh {
				selectInterface(clickedIfaceName)
				// Update the check state of the items
				for _, i := range ifaceItems {
					if i == clickedItem {
						i.Check()
					} else {
						i.Uncheck()
					}
				}
			}
		}(item, device.Name)
	}
	// Create a menu item for exiting the application
	exitItem := systray.AddMenuItem("Exit", "Exit the application")
	go func() {
		<-exitItem.ClickedCh
		systray.Quit()
	}()
}
