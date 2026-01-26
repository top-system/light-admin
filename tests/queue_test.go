package tests

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/top-system/light-admin/pkg/queue"
)

// SimpleTask 简单的内存任务
type SimpleTask struct {
	*queue.InMemoryTask
	Name      string
	executed  bool
	execCount int32
	mu        sync.Mutex
}

func NewSimpleTask(name string) *SimpleTask {
	return &SimpleTask{
		InMemoryTask: &queue.InMemoryTask{
			DBTask: &queue.DBTask{
				TaskModel: &queue.TaskModel{
					Type: "simple_task",
				},
			},
		},
		Name: name,
	}
}

func (t *SimpleTask) Do(ctx context.Context) (queue.Status, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executed = true
	atomic.AddInt32(&t.execCount, 1)
	return queue.StatusCompleted, nil
}

func (t *SimpleTask) IsExecuted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.executed
}

func (t *SimpleTask) ExecCount() int32 {
	return atomic.LoadInt32(&t.execCount)
}

// FailingTask 会失败的任务，用于测试重试
type FailingTask struct {
	*queue.InMemoryTask
	failCount    int32
	maxFails     int32
	successCount int32
}

func NewFailingTask(maxFails int32) *FailingTask {
	return &FailingTask{
		InMemoryTask: &queue.InMemoryTask{
			DBTask: &queue.DBTask{
				TaskModel: &queue.TaskModel{
					Type: "failing_task",
				},
			},
		},
		maxFails: maxFails,
	}
}

func (t *FailingTask) Do(ctx context.Context) (queue.Status, error) {
	count := atomic.AddInt32(&t.failCount, 1)
	if count <= t.maxFails {
		return queue.StatusError, errors.New("intentional failure")
	}
	atomic.AddInt32(&t.successCount, 1)
	return queue.StatusCompleted, nil
}

// SlowTask 慢任务，用于测试并发
type SlowTask struct {
	*queue.InMemoryTask
	Duration time.Duration
	done     chan struct{}
}

func NewSlowTask(duration time.Duration) *SlowTask {
	return &SlowTask{
		InMemoryTask: &queue.InMemoryTask{
			DBTask: &queue.DBTask{
				TaskModel: &queue.TaskModel{
					Type: "slow_task",
				},
			},
		},
		Duration: duration,
		done:     make(chan struct{}),
	}
}

func (t *SlowTask) Do(ctx context.Context) (queue.Status, error) {
	defer close(t.done)
	select {
	case <-time.After(t.Duration):
		return queue.StatusCompleted, nil
	case <-ctx.Done():
		return queue.StatusError, ctx.Err()
	}
}

func (t *SlowTask) Wait() {
	<-t.done
}

// TestQueueBasic 测试队列基本功能
func TestQueueBasic(t *testing.T) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(2),
		queue.WithName("test-queue"),
	)

	q.Start()
	defer q.Shutdown()

	task := NewSimpleTask("test-task-1")
	err := q.QueueTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to queue task: %v", err)
	}

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	if !task.IsExecuted() {
		t.Error("Task should have been executed")
	}

	if q.SubmittedTasks() != 1 {
		t.Errorf("Expected 1 submitted task, got %d", q.SubmittedTasks())
	}

	if q.SuccessTasks() != 1 {
		t.Errorf("Expected 1 success task, got %d", q.SuccessTasks())
	}
}

// TestQueueMultipleTasks 测试多个任务
func TestQueueMultipleTasks(t *testing.T) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(4),
		queue.WithName("multi-task-queue"),
	)

	q.Start()
	defer q.Shutdown()

	taskCount := 10
	tasks := make([]*SimpleTask, taskCount)

	for i := 0; i < taskCount; i++ {
		tasks[i] = NewSimpleTask("task-" + string(rune('A'+i)))
		err := q.QueueTask(context.Background(), tasks[i])
		if err != nil {
			t.Fatalf("Failed to queue task %d: %v", i, err)
		}
	}

	// 等待所有任务完成
	time.Sleep(2 * time.Second)

	executedCount := 0
	for _, task := range tasks {
		if task.IsExecuted() {
			executedCount++
		}
	}

	if executedCount != taskCount {
		t.Errorf("Expected %d tasks executed, got %d", taskCount, executedCount)
	}

	if q.SuccessTasks() != taskCount {
		t.Errorf("Expected %d success tasks, got %d", taskCount, q.SuccessTasks())
	}
}

// TestQueueConcurrency 测试并发处理
func TestQueueConcurrency(t *testing.T) {
	logger := queue.NewDefaultLogger()
	workerCount := 3
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(workerCount),
		queue.WithName("concurrent-queue"),
	)

	q.Start()
	defer q.Shutdown()

	// 提交多个慢任务
	taskCount := 5
	tasks := make([]*SlowTask, taskCount)
	for i := 0; i < taskCount; i++ {
		tasks[i] = NewSlowTask(200 * time.Millisecond)
		err := q.QueueTask(context.Background(), tasks[i])
		if err != nil {
			t.Fatalf("Failed to queue task %d: %v", i, err)
		}
	}

	// 给一些时间让任务开始
	time.Sleep(100 * time.Millisecond)

	// 检查并发 worker 数量
	busyWorkers := q.BusyWorkers()
	if busyWorkers > workerCount {
		t.Errorf("Busy workers (%d) should not exceed worker count (%d)", busyWorkers, workerCount)
	}

	// 等待所有任务完成
	for _, task := range tasks {
		task.Wait()
	}

	time.Sleep(100 * time.Millisecond)

	if q.SuccessTasks() != taskCount {
		t.Errorf("Expected %d success tasks, got %d", taskCount, q.SuccessTasks())
	}
}

// TestQueueWithRetry 测试重试机制
func TestQueueWithRetry(t *testing.T) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(1),
		queue.WithMaxRetry(3),
		queue.WithRetryDelay(100*time.Millisecond),
		queue.WithName("retry-queue"),
	)

	q.Start()
	defer q.Shutdown()

	// 创建一个会失败2次然后成功的任务
	task := NewFailingTask(2)
	err := q.QueueTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to queue task: %v", err)
	}

	// 等待任务完成（包括重试时间）
	time.Sleep(2 * time.Second)

	if atomic.LoadInt32(&task.successCount) != 1 {
		t.Error("Task should have succeeded after retries")
	}

	if atomic.LoadInt32(&task.failCount) != 3 {
		t.Errorf("Expected 3 attempts (2 fails + 1 success), got %d", task.failCount)
	}
}

// TestQueueShutdown 测试队列关闭
func TestQueueShutdown(t *testing.T) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(2),
		queue.WithName("shutdown-queue"),
	)

	q.Start()

	// 提交任务
	task := NewSimpleTask("shutdown-test")
	err := q.QueueTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to queue task: %v", err)
	}

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 关闭队列
	q.Shutdown()

	// 关闭后不能提交新任务
	err = q.QueueTask(context.Background(), NewSimpleTask("after-shutdown"))
	if err != queue.ErrQueueShutdown {
		t.Errorf("Expected ErrQueueShutdown, got %v", err)
	}
}

// TestQueueMetrics 测试指标统计
func TestQueueMetrics(t *testing.T) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		queue.NewTaskRegistry(),
		queue.WithWorkerCount(2),
		queue.WithName("metrics-queue"),
	)

	q.Start()
	defer q.Shutdown()

	// 初始状态
	if q.SubmittedTasks() != 0 {
		t.Errorf("Initial submitted should be 0, got %d", q.SubmittedTasks())
	}

	// 提交成功的任务
	for i := 0; i < 3; i++ {
		task := NewSimpleTask("success-task")
		q.QueueTask(context.Background(), task)
	}

	// 等待完成
	time.Sleep(500 * time.Millisecond)

	if q.SubmittedTasks() != 3 {
		t.Errorf("Expected 3 submitted, got %d", q.SubmittedTasks())
	}

	if q.SuccessTasks() != 3 {
		t.Errorf("Expected 3 success, got %d", q.SuccessTasks())
	}

	if q.FailureTasks() != 0 {
		t.Errorf("Expected 0 failures, got %d", q.FailureTasks())
	}
}

// TestTaskRegistry 测试任务注册表
func TestTaskRegistry(t *testing.T) {
	registry := queue.NewTaskRegistry()

	// 获取 ID
	id1 := registry.NextID()
	id2 := registry.NextID()

	if id1 >= id2 {
		t.Error("IDs should be incremental")
	}

	// 设置和获取任务
	task := NewSimpleTask("registry-test")
	registry.Set(id1, task)

	retrieved, ok := registry.Get(id1)
	if !ok {
		t.Error("Should find the task")
	}

	if retrieved != task {
		t.Error("Retrieved task should match")
	}

	// 删除任务
	registry.Delete(id1)
	_, ok = registry.Get(id1)
	if ok {
		t.Error("Should not find deleted task")
	}
}

// TestScheduler 测试调度器
func TestScheduler(t *testing.T) {
	logger := queue.NewDefaultLogger()
	scheduler := queue.NewFifoScheduler(10, logger)

	// 添加任务
	task1 := NewSimpleTask("task1")
	task2 := NewSimpleTask("task2")

	err := scheduler.Queue(task1)
	if err != nil {
		t.Fatalf("Failed to queue task1: %v", err)
	}

	err = scheduler.Queue(task2)
	if err != nil {
		t.Fatalf("Failed to queue task2: %v", err)
	}

	// 请求任务（基于 ResumeTime 的堆排序，ResumeTime 相同时按入队顺序）
	retrieved, err := scheduler.Request()
	if err != nil {
		t.Fatalf("Failed to request task: %v", err)
	}

	// 验证获取到了一个任务
	if retrieved == nil {
		t.Error("Should get a task")
	}

	// 获取第二个任务
	retrieved2, err := scheduler.Request()
	if err != nil {
		t.Fatalf("Failed to request second task: %v", err)
	}

	if retrieved2 == nil {
		t.Error("Should get second task")
	}

	// 队列应该为空
	_, err = scheduler.Request()
	if err != queue.ErrNoTaskInQueue {
		t.Errorf("Expected ErrNoTaskInQueue, got %v", err)
	}

	// 关闭调度器
	err = scheduler.Shutdown()
	if err != nil {
		t.Fatalf("Failed to shutdown scheduler: %v", err)
	}

	// 关闭后不能添加任务
	err = scheduler.Queue(NewSimpleTask("after-shutdown"))
	if err != queue.ErrQueueShutdown {
		t.Errorf("Expected ErrQueueShutdown, got %v", err)
	}
}

// TestInMemoryRepository 测试内存任务仓库
func TestInMemoryRepository(t *testing.T) {
	repo := queue.NewInMemoryTaskRepository()
	ctx := context.Background()

	// 创建任务
	task := &queue.TaskModel{
		Type:   "test_task",
		Status: queue.StatusQueued,
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.ID == 0 {
		t.Error("Task should have an ID after creation")
	}

	// 获取任务
	retrieved, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.Type != task.Type {
		t.Error("Retrieved task type should match")
	}

	// 更新任务
	task.Status = queue.StatusProcessing
	err = repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, task.ID)
	if retrieved.Status != queue.StatusProcessing {
		t.Error("Task status should be updated")
	}

	// 获取待处理任务
	pending, err := repo.GetPendingTasks(ctx)
	if err != nil {
		t.Fatalf("Failed to get pending tasks: %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pending))
	}

	// 删除任务
	err = repo.Delete(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	_, err = repo.GetByID(ctx, task.ID)
	if err == nil {
		t.Error("Should not find deleted task")
	}
}

// TestTaskStatus 测试任务状态
func TestTaskStatus(t *testing.T) {
	tests := []struct {
		status     queue.Status
		isTerminal bool
	}{
		{queue.StatusQueued, false},
		{queue.StatusProcessing, false},
		{queue.StatusSuspending, false},
		{queue.StatusCompleted, true},
		{queue.StatusError, true},
		{queue.StatusCanceled, true},
	}

	for _, tt := range tests {
		if tt.status.IsTerminal() != tt.isTerminal {
			t.Errorf("Status %s: expected IsTerminal=%v, got %v",
				tt.status, tt.isTerminal, tt.status.IsTerminal())
		}
	}
}

// BenchmarkQueueThroughput 基准测试：队列吞吐量
func BenchmarkQueueThroughput(b *testing.B) {
	logger := queue.NewDefaultLogger()
	q := queue.New(
		logger,
		nil,
		nil,
		queue.WithWorkerCount(4),
		queue.WithName("benchmark-queue"),
	)

	q.Start()
	defer q.Shutdown()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := NewSimpleTask("bench-task")
		q.QueueTask(context.Background(), task)
	}

	// 等待所有任务完成
	for q.BusyWorkers() > 0 || q.SubmittedTasks() > q.SuccessTasks()+q.FailureTasks() {
		time.Sleep(10 * time.Millisecond)
	}
}
