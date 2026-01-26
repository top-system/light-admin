package queue

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

type (
	// Task interface defines the task behavior
	Task interface {
		// Do executes the task and returns the next status
		Do(ctx context.Context) (Status, error)

		// ID returns the task ID
		ID() int
		// Type returns the task type
		Type() string
		// Status returns the task status
		Status() Status
		// Owner returns the task owner
		Owner() *TaskOwner
		// State returns the internal task state
		State() string
		// ShouldPersist returns true if the task should be persisted into DB
		ShouldPersist() bool
		// Persisted returns true if the task is persisted in DB
		Persisted() bool
		// Executed returns the duration of the task execution
		Executed() time.Duration
		// Retried returns the number of times the task has been retried
		Retried() int
		// Error returns the error of the task
		Error() error
		// ErrorHistory returns the error history of the task
		ErrorHistory() []error
		// Model returns the model of the task
		Model() *TaskModel
		// CorrelationID returns the correlation ID of the task
		CorrelationID() uuid.UUID
		// ResumeTime returns the time when the task is resumed
		ResumeTime() int64
		// ResumeAfter sets the time when the task should be resumed
		ResumeAfter(next time.Duration)
		// Progress returns the task progress
		Progress(ctx context.Context) Progresses
		// Summarize returns the task summary for UI display
		Summarize() *Summary
		// OnSuspend is called when queue decides to suspend the task
		OnSuspend(time int64)
		// OnPersisted is called when the task is persisted or updated in DB
		OnPersisted(task *TaskModel)
		// OnError is called when the task encounters an error
		OnError(err error, d time.Duration)
		// OnRetry is called when the iteration returns error and before retry
		OnRetry(err error)
		// OnIterationComplete is called when one iteration is completed
		OnIterationComplete(executed time.Duration)
		// OnStatusTransition is called when the task status is changed
		OnStatusTransition(newStatus Status)

		// Cleanup is called when the task is done or error
		Cleanup(ctx context.Context) error

		Lock()
		Unlock()
	}

	// ResumableTaskFactory creates a task from model
	ResumableTaskFactory func(model *TaskModel) Task

	// Progress represents task progress
	Progress struct {
		Total      int64  `json:"total"`
		Current    int64  `json:"current"`
		Identifier string `json:"identifier"`
	}

	// Progresses is a map of progress by name
	Progresses map[string]*Progress

	// Summary represents task summary for UI display
	Summary struct {
		NodeID int            `json:"-"`
		Phase  string         `json:"phase,omitempty"`
		Props  map[string]any `json:"props,omitempty"`
	}

	stateTransition func(ctx context.Context, task Task, newStatus Status, q *queue) error
)

var (
	taskFactories sync.Map
)

func init() {
	gob.Register(Progresses{})
}

// RegisterResumableTaskFactory registers a resumable task factory
func RegisterResumableTaskFactory(taskType string, factory ResumableTaskFactory) {
	taskFactories.Store(taskType, factory)
}

// NewTaskFromModel creates a task from TaskModel
func NewTaskFromModel(model *TaskModel) (Task, error) {
	if factory, ok := taskFactories.Load(model.Type); ok {
		return factory.(ResumableTaskFactory)(model), nil
	}

	return nil, fmt.Errorf("unknown task type: %s", model.Type)
}

// InMemoryTask implements part Task interface using in-memory data
type InMemoryTask struct {
	*DBTask
}

func (i *InMemoryTask) ShouldPersist() bool {
	return false
}

func (t *InMemoryTask) OnStatusTransition(newStatus Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		t.TaskModel.Status = newStatus
	}
}

// DBTask implements Task interface related to DB schema
type DBTask struct {
	DirectOwner *TaskOwner
	TaskModel   *TaskModel

	mu sync.Mutex
}

func (t *DBTask) ID() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return int(t.TaskModel.ID)
	}
	return 0
}

func (t *DBTask) Status() Status {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.Status
	}
	return ""
}

func (t *DBTask) Type() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TaskModel.Type
}

func (t *DBTask) Owner() *TaskOwner {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.DirectOwner != nil {
		return t.DirectOwner
	}
	return nil
}

func (t *DBTask) State() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.PrivateState
	}
	return ""
}

func (t *DBTask) Persisted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.TaskModel != nil && t.TaskModel.ID != 0
}

func (t *DBTask) Executed() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.PublicState.ExecutedDuration
	}
	return 0
}

func (t *DBTask) Retried() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.PublicState.RetryCount
	}
	return 0
}

func (t *DBTask) Error() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil && t.TaskModel.PublicState.Error != "" {
		return errors.New(t.TaskModel.PublicState.Error)
	}

	return nil
}

func (t *DBTask) ErrorHistory() []error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		result := make([]error, len(t.TaskModel.PublicState.ErrorHistory))
		for i, errStr := range t.TaskModel.PublicState.ErrorHistory {
			result[i] = errors.New(errStr)
		}
		return result
	}

	return nil
}

func (t *DBTask) Model() *TaskModel {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TaskModel
}

func (t *DBTask) Cleanup(ctx context.Context) error {
	return nil
}

func (t *DBTask) CorrelationID() uuid.UUID {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.CorrelationID
	}
	return uuid.Nil
}

func (t *DBTask) ShouldPersist() bool {
	return true
}

func (t *DBTask) OnPersisted(task *TaskModel) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.TaskModel = task
}

func (t *DBTask) OnError(err error, d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		t.TaskModel.PublicState.Error = err.Error()
		t.TaskModel.PublicState.ExecutedDuration += d
	}
}

func (t *DBTask) OnRetry(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		if t.TaskModel.PublicState.ErrorHistory == nil {
			t.TaskModel.PublicState.ErrorHistory = make([]string, 0)
		}

		t.TaskModel.PublicState.ErrorHistory = append(t.TaskModel.PublicState.ErrorHistory, err.Error())
		t.TaskModel.PublicState.RetryCount++
	}
}

func (t *DBTask) OnIterationComplete(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		t.TaskModel.PublicState.ExecutedDuration += d
	}
}

func (t *DBTask) ResumeTime() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		return t.TaskModel.PublicState.ResumeTime
	}
	return 0
}

func (t *DBTask) OnSuspend(time int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		t.TaskModel.PublicState.ResumeTime = time
	}
}

func (t *DBTask) Progress(ctx context.Context) Progresses {
	return nil
}

func (t *DBTask) OnStatusTransition(newStatus Status) {
	// Nop
}

func (t *DBTask) Lock() {
	t.mu.Lock()
}

func (t *DBTask) Unlock() {
	t.mu.Unlock()
}

func (t *DBTask) Summarize() *Summary {
	return &Summary{}
}

func (t *DBTask) ResumeAfter(next time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.TaskModel != nil {
		t.TaskModel.PublicState.ResumeTime = time.Now().Add(next).Unix()
	}
}

var stateTransitions map[Status]map[Status]stateTransition

func init() {
	stateTransitions = map[Status]map[Status]stateTransition{
		"": {
			StatusQueued: persistTask,
		},
		StatusQueued: {
			StatusProcessing: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				if err := persistTask(ctx, task, newStatus, q); err != nil {
					return err
				}
				return nil
			},
			StatusQueued: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				return nil
			},
			StatusError: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.metric.IncFailureTask()
				return persistTask(ctx, task, newStatus, q)
			},
		},
		StatusProcessing: {
			StatusQueued: persistTask,
			StatusCompleted: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.logger.Info("Execution completed in %s with %d retries, clean up...", task.Executed(), task.Retried())
				q.metric.IncSuccessTask()

				if err := task.Cleanup(ctx); err != nil {
					q.logger.Error("Task cleanup failed: %s", err.Error())
				}

				if q.registry != nil {
					q.registry.Delete(task.ID())
				}

				if err := persistTask(ctx, task, newStatus, q); err != nil {
					return err
				}
				return nil
			},
			StatusError: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.logger.Error("Execution failed with error in %s with %d retries, clean up...", task.Executed(), task.Retried())
				q.metric.IncFailureTask()

				if err := task.Cleanup(ctx); err != nil {
					q.logger.Error("Task cleanup failed: %s", err.Error())
				}

				if q.registry != nil {
					q.registry.Delete(task.ID())
				}

				if err := persistTask(ctx, task, newStatus, q); err != nil {
					return err
				}

				return nil
			},
			StatusCanceled: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.logger.Info("Execution canceled, clean up...")
				q.metric.IncFailureTask()

				if err := task.Cleanup(ctx); err != nil {
					q.logger.Error("Task cleanup failed: %s", err.Error())
				}

				if q.registry != nil {
					q.registry.Delete(task.ID())
				}

				if err := persistTask(ctx, task, newStatus, q); err != nil {
					return err
				}

				return nil
			},
			StatusProcessing: persistTask,
			StatusSuspending: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.metric.IncSuspendingTask()
				if err := persistTask(ctx, task, newStatus, q); err != nil {
					return err
				}
				q.logger.Info("Task %d suspended, resume time: %d", task.ID(), task.ResumeTime())
				return q.QueueTask(ctx, task)
			},
		},
		StatusSuspending: {
			StatusProcessing: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.metric.DecSuspendingTask()
				return persistTask(ctx, task, newStatus, q)
			},
			StatusError: func(ctx context.Context, task Task, newStatus Status, q *queue) error {
				q.metric.IncFailureTask()
				return persistTask(ctx, task, newStatus, q)
			},
		},
	}
}

func persistTask(ctx context.Context, task Task, newState Status, q *queue) error {
	// Persist task into repository
	if task.ShouldPersist() {
		if err := saveTaskToRepository(ctx, task, newState, q); err != nil {
			return err
		}
	} else {
		task.OnStatusTransition(newState)
	}

	return nil
}

func saveTaskToRepository(ctx context.Context, task Task, newStatus Status, q *queue) error {
	var (
		errStr     string
		errHistory []string
	)
	if err := task.Error(); err != nil {
		errStr = err.Error()
	}

	for _, err := range task.ErrorHistory() {
		errHistory = append(errHistory, err.Error())
	}

	model := task.Model()
	if model == nil {
		model = &TaskModel{}
	}

	model.Status = newStatus
	model.Type = task.Type()
	model.PublicState = TaskPublicState{
		RetryCount:       task.Retried(),
		ExecutedDuration: task.Executed(),
		ErrorHistory:     errHistory,
		Error:            errStr,
		ResumeTime:       task.ResumeTime(),
	}
	model.PrivateState = task.State()
	if task.Owner() != nil {
		model.OwnerID = task.Owner().ID
	}
	model.CorrelationID = correlationIDFromContext(ctx)

	var err error

	if q.taskRepository != nil {
		if !task.Persisted() {
			err = q.taskRepository.Create(ctx, model)
		} else {
			err = q.taskRepository.Update(ctx, model)
		}
		if err != nil {
			return fmt.Errorf("failed to persist task into DB: %w", err)
		}
	}

	task.OnPersisted(model)
	return nil
}

func correlationIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(CorrelationIDCtx{}).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
