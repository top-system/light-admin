package downloader

import (
	"context"
	"encoding/gob"
	"fmt"
)

var (
	// ErrTaskNotFound is returned when task is not found
	ErrTaskNotFound = fmt.Errorf("task not found")
)

type (
	// Downloader interface defines the download client behavior
	Downloader interface {
		// CreateTask creates a task with the given URL and options, returns a task handle for future operations
		CreateTask(ctx context.Context, url string, options map[string]interface{}) (*TaskHandle, error)
		// Info returns the status of the task with the given handle
		Info(ctx context.Context, handle *TaskHandle) (*TaskStatus, error)
		// Cancel cancels the task with the given handle
		Cancel(ctx context.Context, handle *TaskHandle) error
		// SetFilesToDownload sets the files to download for the task with the given handle
		SetFilesToDownload(ctx context.Context, handle *TaskHandle, args ...*SetFileToDownloadArgs) error
		// Test tests the connection to the downloader
		Test(ctx context.Context) (string, error)
	}

	// TaskHandle represents a task handle for future operations
	TaskHandle struct {
		ID   string `json:"id"`
		Hash string `json:"hash"`
	}

	// Status represents the download status
	Status string

	// TaskStatus represents the status of a download task
	TaskStatus struct {
		FollowedBy    *TaskHandle `json:"-"` // Indicate if the task handle is changed
		SavePath      string      `json:"save_path,omitempty"`
		Name          string      `json:"name"`
		State         Status      `json:"state"`
		Total         int64       `json:"total"`
		Downloaded    int64       `json:"downloaded"`
		DownloadSpeed int64       `json:"download_speed"`
		Uploaded      int64       `json:"uploaded"`
		UploadSpeed   int64       `json:"upload_speed"`
		Hash          string      `json:"hash,omitempty"`
		Files         []TaskFile  `json:"files,omitempty"`
		Pieces        []byte      `json:"pieces,omitempty"` // Hexadecimal representation of the download progress
		NumPieces     int         `json:"num_pieces,omitempty"`
		ErrorMessage  string      `json:"error_message,omitempty"`
	}

	// TaskFile represents a file in a download task
	TaskFile struct {
		Index    int     `json:"index"`
		Name     string  `json:"name"`
		Size     int64   `json:"size"`
		Progress float64 `json:"progress"`
		Selected bool    `json:"selected"`
	}

	// SetFileToDownloadArgs represents arguments for setting files to download
	SetFileToDownloadArgs struct {
		Index    int  `json:"index"`
		Download bool `json:"download"`
	}
)

// Download status constants
const (
	StatusDownloading Status = "downloading"
	StatusSeeding     Status = "seeding"
	StatusCompleted   Status = "completed"
	StatusError       Status = "error"
	StatusUnknown     Status = "unknown"

	DownloaderCtxKey = "downloader"
)

func init() {
	gob.Register(TaskHandle{})
	gob.Register(TaskStatus{})
}

// Progress returns the download progress as a percentage
func (s *TaskStatus) Progress() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Downloaded) / float64(s.Total) * 100
}

// IsComplete returns true if the download is complete
func (s *TaskStatus) IsComplete() bool {
	return s.State == StatusCompleted
}

// IsError returns true if the download has an error
func (s *TaskStatus) IsError() bool {
	return s.State == StatusError
}

// IsActive returns true if the download is active
func (s *TaskStatus) IsActive() bool {
	return s.State == StatusDownloading || s.State == StatusSeeding
}
