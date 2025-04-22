# tray-traffic

A lightweight background network traffic monitor, featuring a tray icon interface, threshold notifications.

## Features

- Monitor real-time traffic of a selected network interface
- Automatically exclude local traffic (e.g., LAN, 127.0.0.1)
- Display status in system tray
- Custom usage threshold notifications (with desktop alerts)
- Persist usage data and settings in JSON

## Getting Started

### Prerequisites

- [Npcap](https://npcap.com/#download) installed

### Build

```bash
go build -ldflags="-H windowsgui"
