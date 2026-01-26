package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jpillora/backoff"
)

type (
	// Queue interface defines the task queue behavior
	Queue interface {
		// Start resume tasks and starts all workers
		Start()
		// Shutdown stops all workers
		Shutdown()
		// QueueTask submits a task to the queue
		QueueTask(ctx context.Context, t Task) error
		// BusyWorkers returns the numbers of workers in the running process
		BusyWorkers() int
		// SuccessTasks returns the numbers of success tasks
		SuccessTasks() int
		// FailureTasks returns the numbers of failure tasks
		FailureTasks() int
		// SubmittedTasks returns the numbers of submitted tasks
		SubmittedTasks() int
		// SuspendingTasks returns the numbers of suspending tasks
		SuspendingTasks() int
	}

	queue struct {
		sync.Mutex
		routineGroup *routineGroup
		metric       *metric
		quit         chan struct{}
		ready        chan struct{}
		scheduler    Scheduler
		stopOnce     sync.Once
		stopFlag     int32
		rootCtx      context.Context
		cancel       context.CancelFunc

		// Dependencies
		logger         Logger
		taskRepository TaskRepository
		registry       TaskRegistry

		// Options
		*options
	}

	// Dep interface for dependency injection
	Dep interface {
		ForkWithLogger(ctx context.Context, l Logger) context.Context
	}
)

var (
	// CriticalErr is a non-retryable error
	CriticalErr = errors.New("non-retryable error")
)

// New creates a new queue
func New(l Logger, taskRepository TaskRepository, registry TaskRegistry, opts ...Option) Queue {
	o := newDefaultOptions()
	for _, opt := range opts {
		opt.apply(o)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &queue{
		routineGroup:   newRoutineGroup(),
		scheduler:      NewFifoScheduler(0, l),
		quit:           make(chan struct{}),
		ready:          make(chan struct{}, 1),
		metric:         &metric{},
		options:        o,
		logger:         l,
		registry:       registry,
		taskRepository: taskRepository,
		rootCtx:        ctx,
		cancel:         cancel,
	}
}

// Start to enable all worker
func (q *queue) Start() {
	q.routineGroup.Run(func() {
		// Resume tasks in DB
		if len(q.options.resumeTaskType) > 0 && q.taskRepository != nil {
			ctx := context.TODO()
			tasks, err := q.taskRepository.GetPendingTasks(ctx, q.resumeTaskType...)
			if err != nil {
				q.logger.Warning("Failed to get pending tasks from DB for given type %v: %s", q.resumeTaskType, err)
			}

			resumed := 0
			for _, t := range tasks {
				resumedTask, err := NewTaskFromModel(t)
				if err != nil {
					q.logger.Warning("Failed to resume task %d: %s", t.ID, err)
					continue
				}

				if resumedTask.Status() == StatusSuspending {
					q.metric.IncSuspendingTask()
					q.metric.IncSubmittedTask()
				}

				if err := q.QueueTask(ctx, resumedTask); err != nil {
					q.logger.Warning("Failed to resume task %d: %s", t.ID, err)
				}
				resumed++
			}

			q.logger.Info("Resumed %d tasks from DB.", resumed)
		}

		q.start()
	})
	q.logger.Info("Queue %q started with %d workers.", q.name, q.workerCount)
}

// Shutdown stops all queues
func (q *queue) Shutdown() {
	q.logger.Info("Shutting down queue %q...", q.name)
	defer func() {
		q.routineGroup.Wait()
	}()

	if !atomic.CompareAndSwapInt32(&q.stopFlag, 0, 1) {
		return
	}

	q.stopOnce.Do(func() {
		q.cancel()
		if q.metric.BusyWorkers() > 0 {
			q.logger.Info("shutdown all tasks in queue %q: %d workers", q.name, q.metric.BusyWorkers())
		}

		if err := q.scheduler.Shutdown(); err != nil {
			q.logger.Error("failed to shutdown scheduler in queue %q: %s", q.name, err)
		}
		close(q.quit)
	})
}

// BusyWorkers returns the numbers of workers in the running process
func (q *queue) BusyWorkers() int {
	return int(q.metric.BusyWorkers())
}

// SuccessTasks returns the numbers of success tasks
func (q *queue) SuccessTasks() int {
	return int(q.metric.SuccessTasks())
}

// FailureTasks returns the numbers of failure tasks
func (q *queue) FailureTasks() int {
	return int(q.metric.FailureTasks())
}

// SubmittedTasks returns the numbers of submitted tasks
func (q *queue) SubmittedTasks() int {
	return int(q.metric.SubmittedTasks())
}

// SuspendingTasks returns the numbers of suspending tasks
func (q *queue) SuspendingTasks() int {
	return int(q.metric.SuspendingTasks())
}

// QueueTask to queue single task
func (q *queue) QueueTask(ctx context.Context, t Task) error {
	if atomic.LoadInt32(&q.stopFlag) == 1 {
		return ErrQueueShutdown
	}

	if t.Status() != StatusSuspending {
		q.metric.IncSubmittedTask()
		if err := q.transitStatus(ctx, t, StatusQueued); err != nil {
			return err
		}
	}

	if err := q.scheduler.Queue(t); err != nil {
		return err
	}
	owner := ""
	if t.Owner() != nil {
		owner = t.Owner().Email
	}
	q.logger.Info("New Task with type %q submitted to queue %q by %q", t.Type(), q.name, owner)
	if q.registry != nil {
		q.registry.Set(t.ID(), t)
	}

	return nil
}

// newContext creates a new context for a new task iteration
func (q *queue) newContext(t Task) context.Context {
	l := q.logger.CopyWithPrefix(fmt.Sprintf("[Cid: %s TaskID: %d Queue: %s]", t.CorrelationID(), t.ID(), q.name))
	ctx := context.WithValue(q.rootCtx, CorrelationIDCtx{}, t.CorrelationID())
	ctx = context.WithValue(ctx, LoggerCtx{}, l)
	ctx = context.WithValue(ctx, UserCtx{}, t.Owner())
	return ctx
}

func (q *queue) work(t Task) {
	ctx := q.newContext(t)
	l := loggerFromContext(ctx)
	timeIterationStart := time.Now()

	var err error
	// to handle panic cases from inside the worker
	// in such case, we start a new goroutine
	defer func() {
		q.metric.DecBusyWorker()
		e := recover()
		if e != nil {
			l.Error("Panic error in queue %q: %v", q.name, e)
			t.OnError(fmt.Errorf("panic error: %v", e), time.Since(timeIterationStart))

			_ = q.transitStatus(ctx, t, StatusError)
		}
		q.schedule()
	}()

	err = q.transitStatus(ctx, t, StatusProcessing)
	if err != nil {
		l.Error("failed to transit task %d to processing: %s", t.ID(), err.Error())
		panic(err)
	}

	for {
		timeIterationStart = time.Now()
		var next Status
		next, err = q.run(ctx, t)
		if err != nil {
			t.OnError(err, time.Since(timeIterationStart))
			l.Error("runtime error in queue %q: %s", q.name, err.Error())

			_ = q.transitStatus(ctx, t, StatusError)
			break
		}

		// iteration completes
		t.OnIterationComplete(time.Since(timeIterationStart))
		_ = q.transitStatus(ctx, t, next)
		if next != StatusProcessing {
			break
		}
	}
}

func (q *queue) run(ctx context.Context, t Task) (Status, error) {
	l := loggerFromContext(ctx)

	// create channel with buffer size 1 to avoid goroutine leak
	done := make(chan struct {
		err  error
		next Status
	}, 1)
	panicChan := make(chan interface{}, 1)
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(ctx, q.maxTaskExecution-t.Executed())
	defer func() {
		cancel()
	}()

	// run the job
	go func() {
		// handle panic issue
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()

		l.Debug("Iteration started.")
		next, err := t.Do(ctx)
		l.Debug("Iteration ended with err=%s", err)
		if err != nil && q.maxRetry-t.Retried() > 0 && !errors.Is(err, CriticalErr) && atomic.LoadInt32(&q.stopFlag) != 1 {
			// Retry needed
			t.OnRetry(err)
			b := &backoff.Backoff{
				Max:    q.backoffMaxDuration,
				Factor: q.backoffFactor,
			}
			delay := q.retryDelay
			if q.retryDelay == 0 {
				delay = b.ForAttempt(float64(t.Retried()))
			}

			// Resume after to retry
			l.Info("Will be retried in %s", delay)
			t.OnSuspend(time.Now().Add(delay).Unix())
			err = nil
			next = StatusSuspending
		}

		done <- struct {
			err  error
			next Status
		}{err: err, next: next}
	}()

	select {
	case p := <-panicChan:
		panic(p)
	case <-ctx.Done(): // timeout reached
		return StatusError, ctx.Err()
	case <-q.quit: // shutdown service
		// cancel job
		cancel()

		leftTime := q.maxTaskExecution - t.Executed() - time.Since(startTime)
		// wait job
		select {
		case <-time.After(leftTime):
			return StatusError, context.DeadlineExceeded
		case r := <-done: // job finish
			return r.next, r.err
		case p := <-panicChan:
			panic(p)
		}
	case r := <-done: // job finish
		return r.next, r.err
	}
}

// transitStatus updates task status
func (q *queue) transitStatus(ctx context.Context, task Task, to Status) (err error) {
	old := task.Status()
	transition, ok := stateTransitions[task.Status()][to]
	if !ok {
		err = fmt.Errorf("invalid state transition from %s to %s", old, to)
	} else {
		if innerErr := transition(ctx, task, to, q); innerErr != nil {
			err = fmt.Errorf("failed to transit task status from %s to %s: %w", old, to, innerErr)
		}
	}

	l := loggerFromContext(ctx)
	if err != nil {
		l.Error(err.Error())
	}

	l.Info("Task %d status changed from %q to %q.", task.ID(), old, to)
	return
}

// schedule to check worker number
func (q *queue) schedule() {
	q.Lock()
	defer q.Unlock()
	if q.BusyWorkers() >= q.workerCount {
		return
	}

	select {
	case q.ready <- struct{}{}:
	default:
	}
}

// start to start all worker
func (q *queue) start() {
	tasks := make(chan Task, 1)

	for {
		// check worker number
		q.schedule()

		select {
		// wait worker ready
		case <-q.ready:
		case <-q.quit:
			return
		}

		// request task from queue in background
		q.routineGroup.Run(func() {
			for {
				t, err := q.scheduler.Request()
				if t == nil || err != nil {
					if err != nil {
						select {
						case <-q.quit:
							if !errors.Is(err, ErrNoTaskInQueue) {
								close(tasks)
								return
							}
						case <-time.After(q.taskPullInterval):
							// sleep to fetch new task
						}
					}
				}
				if t != nil {
					tasks <- t
					return
				}

				select {
				case <-q.quit:
					if !errors.Is(err, ErrNoTaskInQueue) {
						close(tasks)
						return
					}
				default:
				}
			}
		})

		t, ok := <-tasks
		if !ok {
			return
		}

		// start new task
		q.metric.IncBusyWorker()
		q.routineGroup.Run(func() {
			q.work(t)
		})
	}
}

func loggerFromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(LoggerCtx{}).(Logger); ok {
		return l
	}
	return NewDefaultLogger()
}

// CorrelationID returns correlation ID from context
func CorrelationID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(CorrelationIDCtx{}).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
