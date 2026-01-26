package rpc

import (
	"log"
)

// Event represents an aria2 notification event
type Event struct {
	Gid string `json:"gid"` // GID of the download
}

// websocketResponse represents a WebSocket response from aria2
type websocketResponse struct {
	ClientResponse
	Method string  `json:"method"`
	Params []Event `json:"params"`
}

// Notifier handles RPC notifications from aria2 server
type Notifier interface {
	// OnDownloadStart is called when a download starts
	OnDownloadStart([]Event)
	// OnDownloadPause is called when a download is paused
	OnDownloadPause([]Event)
	// OnDownloadStop is called when a download is stopped by user
	OnDownloadStop([]Event)
	// OnDownloadComplete is called when a download completes
	OnDownloadComplete([]Event)
	// OnDownloadError is called when a download errors
	OnDownloadError([]Event)
	// OnBtDownloadComplete is called when a BitTorrent download completes but seeding continues
	OnBtDownloadComplete([]Event)
}

// DummyNotifier is a no-op notifier implementation
type DummyNotifier struct{}

func (DummyNotifier) OnDownloadStart(events []Event)      { log.Printf("%s started.", events) }
func (DummyNotifier) OnDownloadPause(events []Event)      { log.Printf("%s paused.", events) }
func (DummyNotifier) OnDownloadStop(events []Event)       { log.Printf("%s stopped.", events) }
func (DummyNotifier) OnDownloadComplete(events []Event)   { log.Printf("%s completed.", events) }
func (DummyNotifier) OnDownloadError(events []Event)      { log.Printf("%s error.", events) }
func (DummyNotifier) OnBtDownloadComplete(events []Event) { log.Printf("bt %s completed.", events) }

// LogNotifier logs all notifications
type LogNotifier struct {
	Logger func(format string, args ...interface{})
}

func (n LogNotifier) OnDownloadStart(events []Event) {
	if n.Logger != nil {
		n.Logger("Download started: %v", events)
	}
}

func (n LogNotifier) OnDownloadPause(events []Event) {
	if n.Logger != nil {
		n.Logger("Download paused: %v", events)
	}
}

func (n LogNotifier) OnDownloadStop(events []Event) {
	if n.Logger != nil {
		n.Logger("Download stopped: %v", events)
	}
}

func (n LogNotifier) OnDownloadComplete(events []Event) {
	if n.Logger != nil {
		n.Logger("Download completed: %v", events)
	}
}

func (n LogNotifier) OnDownloadError(events []Event) {
	if n.Logger != nil {
		n.Logger("Download error: %v", events)
	}
}

func (n LogNotifier) OnBtDownloadComplete(events []Event) {
	if n.Logger != nil {
		n.Logger("BT download completed: %v", events)
	}
}
