package crontab

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/robfig/cron/v3"
)

type (
	// CronTaskFunc is the function type for cron tasks
	CronTaskFunc func(ctx context.Context)

	// CronType represents the type of cron task
	CronType string

	// cronRegistration represents a cron task registration
	cronRegistration struct {
		name   string
		spec   string
		fn     CronTaskFunc
		enable bool
	}

	// Logger interface for crontab logging
	Logger interface {
		Debug(format string, args ...interface{})
		Info(format string, args ...interface{})
		Warning(format string, args ...interface{})
		Error(format string, args ...interface{})
		CopyWithPrefix(prefix string) Logger
	}

	// Crontab represents the cron scheduler
	Crontab struct {
		cron          *cron.Cron
		logger        Logger
		registrations []cronRegistration
		entryIDs      map[string]cron.EntryID
		mu            sync.RWMutex
		started       bool
		contextData   map[string]interface{}
	}

	// Option configures a Crontab
	Option func(*Crontab)

	// TaskInfo represents information about a scheduled task
	TaskInfo struct {
		Name     string        `json:"name"`
		Spec     string        `json:"spec"`
		Enable   bool          `json:"enable"`
		EntryID  cron.EntryID  `json:"entryId"`
		Next     time.Time     `json:"next"`
		Prev     time.Time     `json:"prev"`
	}
)

// Context keys
type (
	CorrelationIDCtx struct{}
	LoggerCtx        struct{}
	UserCtx          struct{}
)

// Common cron types
const (
	CronTypeCleanup     CronType = "cleanup"
	CronTypeBackup      CronType = "backup"
	CronTypeReport      CronType = "report"
	CronTypeSync        CronType = "sync"
	CronTypeHealthCheck CronType = "health_check"
)

var (
	globalRegistrations []cronRegistration
	globalMu            sync.Mutex
)

// Register registers a global cron task that will be added to all new Crontab instances
func Register(name string, spec string, fn CronTaskFunc) {
	globalMu.Lock()
	defer globalMu.Unlock()

	globalRegistrations = append(globalRegistrations, cronRegistration{
		name:   name,
		spec:   spec,
		fn:     fn,
		enable: true,
	})
}

// RegisterWithType registers a global cron task with a CronType
func RegisterWithType(t CronType, spec string, fn CronTaskFunc) {
	Register(string(t), spec, fn)
}

// New creates a new Crontab instance
func New(logger Logger, opts ...Option) *Crontab {
	c := &Crontab{
		cron:          cron.New(cron.WithSeconds()),
		logger:        logger,
		registrations: make([]cronRegistration, 0),
		entryIDs:      make(map[string]cron.EntryID),
		contextData:   make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Copy global registrations
	globalMu.Lock()
	for _, r := range globalRegistrations {
		c.registrations = append(c.registrations, r)
	}
	globalMu.Unlock()

	return c
}

// NewWithStandardParser creates a new Crontab with standard 5-field cron parser (minute, hour, day, month, weekday)
func NewWithStandardParser(logger Logger, opts ...Option) *Crontab {
	c := &Crontab{
		cron:          cron.New(),
		logger:        logger,
		registrations: make([]cronRegistration, 0),
		entryIDs:      make(map[string]cron.EntryID),
		contextData:   make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Copy global registrations
	globalMu.Lock()
	for _, r := range globalRegistrations {
		c.registrations = append(c.registrations, r)
	}
	globalMu.Unlock()

	return c
}

// WithContextData sets context data that will be passed to all tasks
func WithContextData(key string, value interface{}) Option {
	return func(c *Crontab) {
		c.contextData[key] = value
	}
}

// AddTask adds a new cron task
func (c *Crontab) AddTask(name string, spec string, fn CronTaskFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if task already exists
	for _, r := range c.registrations {
		if r.name == name {
			return fmt.Errorf("crontab: task %q already exists", name)
		}
	}

	reg := cronRegistration{
		name:   name,
		spec:   spec,
		fn:     fn,
		enable: true,
	}

	c.registrations = append(c.registrations, reg)

	// If cron is already started, add the task immediately
	if c.started {
		return c.scheduleTask(reg)
	}

	return nil
}

// AddTaskWithType adds a new cron task with CronType
func (c *Crontab) AddTaskWithType(t CronType, spec string, fn CronTaskFunc) error {
	return c.AddTask(string(t), spec, fn)
}

// RemoveTask removes a cron task by name
func (c *Crontab) RemoveTask(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find and remove from registrations
	found := false
	newRegs := make([]cronRegistration, 0, len(c.registrations))
	for _, r := range c.registrations {
		if r.name == name {
			found = true
			continue
		}
		newRegs = append(newRegs, r)
	}

	if !found {
		return fmt.Errorf("crontab: task %q not found", name)
	}

	c.registrations = newRegs

	// Remove from cron if started
	if entryID, ok := c.entryIDs[name]; ok {
		c.cron.Remove(entryID)
		delete(c.entryIDs, name)
	}

	c.logger.Info("Cron task %q removed", name)
	return nil
}

// EnableTask enables a disabled task
func (c *Crontab) EnableTask(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, r := range c.registrations {
		if r.name == name {
			if r.enable {
				return nil // Already enabled
			}

			c.registrations[i].enable = true

			// Schedule if cron is started
			if c.started {
				if err := c.scheduleTask(c.registrations[i]); err != nil {
					return err
				}
			}

			c.logger.Info("Cron task %q enabled", name)
			return nil
		}
	}

	return fmt.Errorf("crontab: task %q not found", name)
}

// DisableTask disables a task
func (c *Crontab) DisableTask(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, r := range c.registrations {
		if r.name == name {
			if !r.enable {
				return nil // Already disabled
			}

			c.registrations[i].enable = false

			// Remove from cron if started
			if entryID, ok := c.entryIDs[name]; ok {
				c.cron.Remove(entryID)
				delete(c.entryIDs, name)
			}

			c.logger.Info("Cron task %q disabled", name)
			return nil
		}
	}

	return fmt.Errorf("crontab: task %q not found", name)
}

// UpdateTaskSpec updates the cron spec for a task
func (c *Crontab) UpdateTaskSpec(name string, newSpec string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, r := range c.registrations {
		if r.name == name {
			// Remove old entry
			if entryID, ok := c.entryIDs[name]; ok {
				c.cron.Remove(entryID)
				delete(c.entryIDs, name)
			}

			// Update spec
			c.registrations[i].spec = newSpec

			// Re-schedule if enabled and started
			if r.enable && c.started {
				if err := c.scheduleTask(c.registrations[i]); err != nil {
					return err
				}
			}

			c.logger.Info("Cron task %q spec updated to %q", name, newSpec)
			return nil
		}
	}

	return fmt.Errorf("crontab: task %q not found", name)
}

// Start starts the cron scheduler
func (c *Crontab) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("crontab: already started")
	}

	c.logger.Info("Initializing crontab jobs...")

	// Schedule all enabled tasks
	for _, r := range c.registrations {
		if r.enable {
			if err := c.scheduleTask(r); err != nil {
				c.logger.Warning("Failed to schedule cron task %q: %s", r.name, err)
			}
		}
	}

	c.cron.Start()
	c.started = true

	c.logger.Info("Crontab started with %d tasks", len(c.entryIDs))
	return nil
}

// Stop stops the cron scheduler
func (c *Crontab) Stop() context.Context {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return context.Background()
	}

	c.logger.Info("Stopping crontab...")
	ctx := c.cron.Stop()
	c.started = false
	c.logger.Info("Crontab stopped")

	return ctx
}

// RunTask runs a task immediately by name
func (c *Crontab) RunTask(name string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, r := range c.registrations {
		if r.name == name {
			go c.taskWrapper(r.name, r.spec, r.fn)()
			return nil
		}
	}

	return fmt.Errorf("crontab: task %q not found", name)
}

// GetTasks returns information about all registered tasks
func (c *Crontab) GetTasks() []TaskInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tasks := make([]TaskInfo, 0, len(c.registrations))
	entries := c.cron.Entries()
	entryMap := make(map[cron.EntryID]cron.Entry)
	for _, e := range entries {
		entryMap[e.ID] = e
	}

	for _, r := range c.registrations {
		info := TaskInfo{
			Name:   r.name,
			Spec:   r.spec,
			Enable: r.enable,
		}

		if entryID, ok := c.entryIDs[r.name]; ok {
			info.EntryID = entryID
			if entry, ok := entryMap[entryID]; ok {
				info.Next = entry.Next
				info.Prev = entry.Prev
			}
		}

		tasks = append(tasks, info)
	}

	return tasks
}

// GetTask returns information about a specific task
func (c *Crontab) GetTask(name string) (*TaskInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, r := range c.registrations {
		if r.name == name {
			info := &TaskInfo{
				Name:   r.name,
				Spec:   r.spec,
				Enable: r.enable,
			}

			if entryID, ok := c.entryIDs[r.name]; ok {
				info.EntryID = entryID
				entry := c.cron.Entry(entryID)
				info.Next = entry.Next
				info.Prev = entry.Prev
			}

			return info, nil
		}
	}

	return nil, fmt.Errorf("crontab: task %q not found", name)
}

// IsRunning returns whether the crontab is running
func (c *Crontab) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}

// TaskCount returns the number of registered tasks
func (c *Crontab) TaskCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.registrations)
}

// ActiveTaskCount returns the number of active (enabled) tasks
func (c *Crontab) ActiveTaskCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, r := range c.registrations {
		if r.enable {
			count++
		}
	}
	return count
}

// scheduleTask schedules a single task (must be called with lock held)
func (c *Crontab) scheduleTask(r cronRegistration) error {
	wrappedFn := c.taskWrapper(r.name, r.spec, r.fn)
	entryID, err := c.cron.AddFunc(r.spec, wrappedFn)
	if err != nil {
		return fmt.Errorf("failed to add cron task %q with spec %q: %w", r.name, r.spec, err)
	}

	c.entryIDs[r.name] = entryID
	c.logger.Info("Cron task %q scheduled with spec %q", r.name, r.spec)
	return nil
}

// taskWrapper wraps a task function with logging and context
func (c *Crontab) taskWrapper(name, spec string, task CronTaskFunc) func() {
	return func() {
		cid := uuid.Must(uuid.NewV4())
		c.logger.Info("Executing cron task %q with Cid %q", name, cid)

		startTime := time.Now()
		ctx := context.Background()

		// Add correlation ID to context
		ctx = context.WithValue(ctx, CorrelationIDCtx{}, cid)

		// Add logger with prefix to context
		prefixedLogger := c.logger.CopyWithPrefix(fmt.Sprintf("[Cid: %s Cron: %s]", cid, name))
		ctx = context.WithValue(ctx, LoggerCtx{}, prefixedLogger)

		// Add custom context data
		for key, value := range c.contextData {
			ctx = context.WithValue(ctx, key, value)
		}

		// Execute task with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("Cron task %q panicked: %v", name, r)
				}
			}()
			task(ctx)
		}()

		duration := time.Since(startTime)
		c.logger.Info("Cron task %q completed in %s", name, duration)
	}
}

// DefaultLogger is a simple logger implementation
type DefaultLogger struct {
	prefix string
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	fmt.Printf("[WARN] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) CopyWithPrefix(prefix string) Logger {
	return &DefaultLogger{
		prefix: prefix + " ",
	}
}

// LoggerFromContext retrieves the logger from context
func LoggerFromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(LoggerCtx{}).(Logger); ok {
		return l
	}
	return NewDefaultLogger()
}

// CorrelationIDFromContext retrieves the correlation ID from context
func CorrelationIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(CorrelationIDCtx{}).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// Predefined cron specs
const (
	EveryMinute     = "0 * * * * *"     // Every minute (with seconds)
	EveryFiveMinute = "0 */5 * * * *"   // Every 5 minutes
	EveryTenMinute  = "0 */10 * * * *"  // Every 10 minutes
	EveryHour       = "0 0 * * * *"     // Every hour
	EveryDay        = "0 0 0 * * *"     // Every day at midnight
	EveryWeek       = "0 0 0 * * 0"     // Every week on Sunday
	EveryMonth      = "0 0 0 1 * *"     // Every month on the 1st

	// Standard format (without seconds)
	StandardEveryMinute     = "* * * * *"     // Every minute
	StandardEveryFiveMinute = "*/5 * * * *"   // Every 5 minutes
	StandardEveryTenMinute  = "*/10 * * * *"  // Every 10 minutes
	StandardEveryHour       = "0 * * * *"     // Every hour
	StandardEveryDay        = "0 0 * * *"     // Every day at midnight
	StandardEveryWeek       = "0 0 * * 0"     // Every week on Sunday
	StandardEveryMonth      = "0 0 1 * *"     // Every month on the 1st
)
