package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/top-system/light-admin/pkg/downloader"
)

type (
	// RemoteDownloadTask represents a remote download task
	RemoteDownloadTask struct {
		*DBTask

		l        Logger
		state    *RemoteDownloadTaskState
		d        downloader.Downloader
		progress Progresses
	}

	// RemoteDownloadTaskPhase represents the phase of the download task
	RemoteDownloadTaskPhase string

	// RemoteDownloadTaskState represents the internal state of a download task
	RemoteDownloadTaskState struct {
		URL                string                  `json:"url"`
		Dst                string                  `json:"dst,omitempty"`
		Downloader         string                  `json:"downloader"`
		Handle             *downloader.TaskHandle  `json:"handle,omitempty"`
		Status             *downloader.TaskStatus  `json:"status,omitempty"`
		Phase              RemoteDownloadTaskPhase `json:"phase,omitempty"`
		GetTaskStatusTried int                     `json:"get_task_status_tried,omitempty"`
		Options            map[string]interface{}  `json:"options,omitempty"`
	}
)

const (
	RemoteDownloadTaskPhaseNotStarted RemoteDownloadTaskPhase = ""
	RemoteDownloadTaskPhaseMonitor    RemoteDownloadTaskPhase = "monitor"
	RemoteDownloadTaskPhaseSeeding    RemoteDownloadTaskPhase = "seeding"

	GetTaskStatusMaxTries = 5

	// Summary keys
	SummaryKeyDownloadStatus = "download"
	SummaryKeySrcURL         = "src_url"
	SummaryKeyDownloader     = "downloader"
)

func init() {
	RegisterResumableTaskFactory(RemoteDownloadTaskType, NewRemoteDownloadTaskFromModel)
}

// NewRemoteDownloadTask creates a new RemoteDownloadTask
func NewRemoteDownloadTask(ctx context.Context, url string, downloaderName string, options map[string]interface{}, owner *TaskOwner) (Task, error) {
	state := &RemoteDownloadTaskState{
		URL:        url,
		Downloader: downloaderName,
		Options:    options,
	}
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	correlationID := uuid.Must(uuid.NewV4())

	t := &RemoteDownloadTask{
		DBTask: &DBTask{
			TaskModel: &TaskModel{
				Type:          RemoteDownloadTaskType,
				CorrelationID: correlationID,
				PrivateState:  string(stateBytes),
				PublicState:   TaskPublicState{},
			},
			DirectOwner: owner,
		},
		progress: make(Progresses),
	}
	return t, nil
}

// NewRemoteDownloadTaskFromModel creates a RemoteDownloadTask from model
func NewRemoteDownloadTaskFromModel(model *TaskModel) Task {
	return &RemoteDownloadTask{
		DBTask: &DBTask{
			TaskModel: model,
		},
		progress: make(Progresses),
	}
}

// SetDownloader sets the downloader instance for the task
func (m *RemoteDownloadTask) SetDownloader(d downloader.Downloader) {
	m.d = d
}

// Do executes the download task
func (m *RemoteDownloadTask) Do(ctx context.Context) (Status, error) {
	// Get logger from context
	if l, ok := ctx.Value(LoggerCtx{}).(Logger); ok {
		m.l = l
	} else {
		m.l = NewDefaultLogger()
	}

	// Unmarshal state
	state := &RemoteDownloadTaskState{}
	if err := json.Unmarshal([]byte(m.State()), state); err != nil {
		return StatusError, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	m.state = state

	// Check if downloader is set
	if m.d == nil {
		return StatusError, fmt.Errorf("downloader not set, please set downloader before executing task (%w)", CriticalErr)
	}

	var next Status
	var err error

	switch m.state.Phase {
	case RemoteDownloadTaskPhaseNotStarted:
		next, err = m.createDownloadTask(ctx)
	case RemoteDownloadTaskPhaseMonitor, RemoteDownloadTaskPhaseSeeding:
		next, err = m.monitor(ctx)
	}

	// Save state
	newStateStr, marshalErr := json.Marshal(m.state)
	if marshalErr != nil {
		return StatusError, fmt.Errorf("failed to marshal state: %w", marshalErr)
	}

	m.Lock()
	m.TaskModel.PrivateState = string(newStateStr)
	m.Unlock()

	return next, err
}

func (m *RemoteDownloadTask) createDownloadTask(ctx context.Context) (Status, error) {
	if m.state.Handle != nil {
		m.state.Phase = RemoteDownloadTaskPhaseMonitor
		return StatusSuspending, nil
	}

	m.l.Info("Creating download task for URL: %s", m.state.URL)

	// Create download task
	handle, err := m.d.CreateTask(ctx, m.state.URL, m.state.Options)
	if err != nil {
		return StatusError, fmt.Errorf("failed to create download task: %w", err)
	}

	m.state.Handle = handle
	m.state.Phase = RemoteDownloadTaskPhaseMonitor

	m.l.Info("Download task created with handle: %v", handle)
	return StatusSuspending, nil
}

func (m *RemoteDownloadTask) monitor(ctx context.Context) (Status, error) {
	resumeAfter := 10 * time.Second // Check every 10 seconds

	// Update task status
	status, err := m.d.Info(ctx, m.state.Handle)
	if err != nil {
		if errors.Is(err, downloader.ErrTaskNotFound) && m.state.Status != nil {
			// If task is not found, but it previously existed, consider it as canceled
			m.l.Warning("task not found, consider it as canceled")
			return StatusCanceled, nil
		}

		m.state.GetTaskStatusTried++
		if m.state.GetTaskStatusTried >= GetTaskStatusMaxTries {
			return StatusError, fmt.Errorf("failed to get task status after %d retry: %w", m.state.GetTaskStatusTried, err)
		}

		m.l.Warning("failed to get task info: %s, will retry.", err)
		m.ResumeAfter(resumeAfter)
		return StatusSuspending, nil
	}

	// Follow to new handle if needed
	if status.FollowedBy != nil {
		m.l.Info("Task handle updated to %v", status.FollowedBy)
		m.state.Handle = status.FollowedBy
		m.ResumeAfter(0)
		return StatusSuspending, nil
	}

	m.state.Status = status
	m.state.GetTaskStatusTried = 0

	// Update progress
	m.Lock()
	m.progress["download"] = &Progress{
		Total:      status.Total,
		Current:    status.Downloaded,
		Identifier: status.Name,
	}
	m.Unlock()

	m.l.Debug("Monitor %q task state: %s, progress: %.2f%%", status.Name, status.State, status.Progress())

	switch status.State {
	case downloader.StatusSeeding:
		m.l.Info("Download task seeding: %s", status.Name)
		if m.state.Phase == RemoteDownloadTaskPhaseMonitor {
			m.state.Phase = RemoteDownloadTaskPhaseSeeding
		}
		// Continue monitoring seeding
		m.ResumeAfter(resumeAfter)
		return StatusSuspending, nil

	case downloader.StatusCompleted:
		m.l.Info("Download task completed: %s", status.Name)
		return StatusCompleted, nil

	case downloader.StatusDownloading:
		m.ResumeAfter(resumeAfter)
		return StatusSuspending, nil

	case downloader.StatusUnknown, downloader.StatusError:
		return StatusError, fmt.Errorf("download task failed with state %q (%w), errorMsg: %s", status.State, CriticalErr, status.ErrorMessage)
	}

	m.ResumeAfter(resumeAfter)
	return StatusSuspending, nil
}

func (m *RemoteDownloadTask) Cleanup(ctx context.Context) error {
	if m.state != nil && m.state.Handle != nil && m.d != nil {
		// Optionally cancel the download task on error
		if m.Status() == StatusError || m.Status() == StatusCanceled {
			if err := m.d.Cancel(ctx, m.state.Handle); err != nil {
				m.l.Warning("failed to cancel download task: %s", err)
			}
		}
	}
	return nil
}

func (m *RemoteDownloadTask) Summarize() *Summary {
	if m.state == nil {
		if err := json.Unmarshal([]byte(m.State()), &m.state); err != nil {
			return nil
		}
	}

	var status *downloader.TaskStatus
	if m.state.Status != nil {
		statusCopy := *m.state.Status
		status = &statusCopy
		// Redact save path for security
		status.SavePath = ""
	}

	return &Summary{
		Phase: string(m.state.Phase),
		Props: map[string]any{
			SummaryKeySrcURL:         m.state.URL,
			SummaryKeyDownloader:     m.state.Downloader,
			SummaryKeyDownloadStatus: status,
		},
	}
}

func (m *RemoteDownloadTask) Progress(ctx context.Context) Progresses {
	m.Lock()
	defer m.Unlock()

	merged := make(Progresses)
	for k, v := range m.progress {
		merged[k] = v
	}
	return merged
}

// GetState returns the current task state
func (m *RemoteDownloadTask) GetState() *RemoteDownloadTaskState {
	if m.state == nil {
		m.state = &RemoteDownloadTaskState{}
		json.Unmarshal([]byte(m.State()), m.state)
	}
	return m.state
}

// GetHandle returns the download handle
func (m *RemoteDownloadTask) GetHandle() *downloader.TaskHandle {
	state := m.GetState()
	return state.Handle
}

// GetDownloadStatus returns the current download status
func (m *RemoteDownloadTask) GetDownloadStatus() *downloader.TaskStatus {
	state := m.GetState()
	return state.Status
}

// SetFilesToDownload sets the files to download for the task
func (m *RemoteDownloadTask) SetFilesToDownload(ctx context.Context, args ...*downloader.SetFileToDownloadArgs) error {
	handle := m.GetHandle()
	if handle == nil {
		return fmt.Errorf("download task not created")
	}

	if m.d == nil {
		return fmt.Errorf("downloader not set")
	}

	return m.d.SetFilesToDownload(ctx, handle, args...)
}

// CancelDownload cancels the download task
func (m *RemoteDownloadTask) CancelDownload(ctx context.Context) error {
	handle := m.GetHandle()
	if handle == nil {
		return nil
	}

	if m.d == nil {
		return fmt.Errorf("downloader not set")
	}

	return m.d.Cancel(ctx, handle)
}

// DownloaderRegistry is a thread-safe registry for downloaders
type DownloaderRegistry struct {
	mu          sync.RWMutex
	downloaders map[string]downloader.Downloader
}

// NewDownloaderRegistry creates a new downloader registry
func NewDownloaderRegistry() *DownloaderRegistry {
	return &DownloaderRegistry{
		downloaders: make(map[string]downloader.Downloader),
	}
}

// Register registers a downloader
func (r *DownloaderRegistry) Register(name string, d downloader.Downloader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downloaders[name] = d
}

// Get returns a downloader by name
func (r *DownloaderRegistry) Get(name string) (downloader.Downloader, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.downloaders[name]
	return d, ok
}

// List returns all registered downloader names
func (r *DownloaderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.downloaders))
	for name := range r.downloaders {
		names = append(names, name)
	}
	return names
}
