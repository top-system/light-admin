package queue

// Status represents the status of a task
type Status string

const (
	// StatusQueued indicates the task is queued and waiting to be processed
	StatusQueued Status = "queued"
	// StatusProcessing indicates the task is being processed
	StatusProcessing Status = "processing"
	// StatusCompleted indicates the task has completed successfully
	StatusCompleted Status = "completed"
	// StatusError indicates the task has failed with an error
	StatusError Status = "error"
	// StatusCanceled indicates the task has been canceled
	StatusCanceled Status = "canceled"
	// StatusSuspending indicates the task is suspended and waiting to be resumed
	StatusSuspending Status = "suspending"
)

// IsTerminal returns true if the status is a terminal state
func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusError || s == StatusCanceled
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}
