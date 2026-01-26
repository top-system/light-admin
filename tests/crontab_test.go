package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/top-system/light-admin/pkg/crontab"
)

// TestCrontabBasic 测试基本功能
func TestCrontabBasic(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	var executed int32

	// 添加每秒执行的任务
	err := c.AddTask("test-task", "* * * * * *", func(ctx context.Context) {
		atomic.AddInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 启动
	err = c.Start()
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// 等待几秒
	time.Sleep(2500 * time.Millisecond)

	// 停止
	c.Stop()

	count := atomic.LoadInt32(&executed)
	if count < 2 {
		t.Errorf("Expected at least 2 executions, got %d", count)
	}
}

// TestCrontabAddRemoveTask 测试添加和删除任务
func TestCrontabAddRemoveTask(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	// 添加任务
	err := c.AddTask("task1", crontab.EveryMinute, func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 添加重复名称任务应失败
	err = c.AddTask("task1", crontab.EveryMinute, func(ctx context.Context) {})
	if err == nil {
		t.Error("Expected error when adding duplicate task")
	}

	// 删除任务
	err = c.RemoveTask("task1")
	if err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	// 删除不存在的任务应失败
	err = c.RemoveTask("task1")
	if err == nil {
		t.Error("Expected error when removing non-existent task")
	}
}

// TestCrontabEnableDisable 测试启用禁用任务
func TestCrontabEnableDisable(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	var executed int32

	err := c.AddTask("toggle-task", "* * * * * *", func(ctx context.Context) {
		atomic.AddInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	c.Start()
	defer c.Stop()

	// 等待执行
	time.Sleep(1500 * time.Millisecond)
	count1 := atomic.LoadInt32(&executed)

	// 禁用任务
	err = c.DisableTask("toggle-task")
	if err != nil {
		t.Fatalf("Failed to disable task: %v", err)
	}

	// 等待
	time.Sleep(1500 * time.Millisecond)
	count2 := atomic.LoadInt32(&executed)

	// 禁用后不应继续执行
	if count2 > count1+1 {
		t.Errorf("Task should not execute after disable, count1=%d, count2=%d", count1, count2)
	}

	// 启用任务
	err = c.EnableTask("toggle-task")
	if err != nil {
		t.Fatalf("Failed to enable task: %v", err)
	}

	// 等待执行
	time.Sleep(1500 * time.Millisecond)
	count3 := atomic.LoadInt32(&executed)

	if count3 <= count2 {
		t.Error("Task should execute after enable")
	}
}

// TestCrontabUpdateSpec 测试更新执行时间
func TestCrontabUpdateSpec(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	err := c.AddTask("update-task", crontab.EveryMinute, func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	c.Start()
	defer c.Stop()

	// 获取任务信息
	task, err := c.GetTask("update-task")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	oldSpec := task.Spec

	// 更新执行时间
	err = c.UpdateTaskSpec("update-task", crontab.EveryHour)
	if err != nil {
		t.Fatalf("Failed to update spec: %v", err)
	}

	task, _ = c.GetTask("update-task")
	if task.Spec == oldSpec {
		t.Error("Spec should be updated")
	}
}

// TestCrontabRunTask 测试立即执行任务
func TestCrontabRunTask(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	var executed int32

	// 添加每小时执行的任务（正常情况下不会执行）
	err := c.AddTask("manual-task", crontab.EveryHour, func(ctx context.Context) {
		atomic.AddInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	c.Start()
	defer c.Stop()

	// 手动触发执行
	err = c.RunTask("manual-task")
	if err != nil {
		t.Fatalf("Failed to run task: %v", err)
	}

	// 等待执行完成
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&executed) != 1 {
		t.Error("Task should be executed once")
	}
}

// TestCrontabGetTasks 测试获取任务列表
func TestCrontabGetTasks(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	c.AddTask("task1", crontab.EveryMinute, func(ctx context.Context) {})
	c.AddTask("task2", crontab.EveryHour, func(ctx context.Context) {})
	c.AddTask("task3", crontab.EveryDay, func(ctx context.Context) {})

	tasks := c.GetTasks()
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	if c.TaskCount() != 3 {
		t.Errorf("Expected TaskCount=3, got %d", c.TaskCount())
	}

	if c.ActiveTaskCount() != 3 {
		t.Errorf("Expected ActiveTaskCount=3, got %d", c.ActiveTaskCount())
	}

	// 禁用一个任务
	c.DisableTask("task2")

	if c.ActiveTaskCount() != 2 {
		t.Errorf("Expected ActiveTaskCount=2, got %d", c.ActiveTaskCount())
	}
}

// TestCrontabContext 测试上下文传递
func TestCrontabContext(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger,
		crontab.WithContextData("test-key", "test-value"),
	)

	var gotValue string
	var gotCid bool

	err := c.AddTask("context-task", "* * * * * *", func(ctx context.Context) {
		if v, ok := ctx.Value("test-key").(string); ok {
			gotValue = v
		}
		cid := crontab.CorrelationIDFromContext(ctx)
		if cid.String() != "00000000-0000-0000-0000-000000000000" {
			gotCid = true
		}
	})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	c.Start()
	time.Sleep(1500 * time.Millisecond)
	c.Stop()

	if gotValue != "test-value" {
		t.Errorf("Expected context value 'test-value', got '%s'", gotValue)
	}

	if !gotCid {
		t.Error("Expected correlation ID in context")
	}
}

// TestCrontabPanicRecovery 测试 panic 恢复
func TestCrontabPanicRecovery(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	var normalExecuted int32

	// 添加会 panic 的任务
	c.AddTask("panic-task", "* * * * * *", func(ctx context.Context) {
		panic("intentional panic")
	})

	// 添加正常任务
	c.AddTask("normal-task", "* * * * * *", func(ctx context.Context) {
		atomic.AddInt32(&normalExecuted, 1)
	})

	c.Start()
	time.Sleep(2500 * time.Millisecond)
	c.Stop()

	// 正常任务应该继续执行
	if atomic.LoadInt32(&normalExecuted) < 2 {
		t.Error("Normal task should continue executing after panic in other task")
	}
}

// TestCrontabStartStop 测试启动停止
func TestCrontabStartStop(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	if c.IsRunning() {
		t.Error("Should not be running before Start")
	}

	c.Start()

	if !c.IsRunning() {
		t.Error("Should be running after Start")
	}

	// 重复启动应报错
	err := c.Start()
	if err == nil {
		t.Error("Expected error on double start")
	}

	c.Stop()

	if c.IsRunning() {
		t.Error("Should not be running after Stop")
	}
}

// TestCrontabStandardParser 测试标准解析器
func TestCrontabStandardParser(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.NewWithStandardParser(logger)

	var executed int32

	// 标准格式：每分钟
	err := c.AddTask("std-task", "* * * * *", func(ctx context.Context) {
		atomic.AddInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("Failed to add task with standard format: %v", err)
	}

	if c.TaskCount() != 1 {
		t.Error("Task should be added")
	}
}

// TestCrontabTaskInfo 测试任务信息
func TestCrontabTaskInfo(t *testing.T) {
	logger := crontab.NewDefaultLogger()
	c := crontab.New(logger)

	c.AddTask("info-task", crontab.EveryMinute, func(ctx context.Context) {})
	c.Start()
	defer c.Stop()

	task, err := c.GetTask("info-task")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if task.Name != "info-task" {
		t.Errorf("Expected name 'info-task', got '%s'", task.Name)
	}

	if task.Spec != crontab.EveryMinute {
		t.Errorf("Expected spec '%s', got '%s'", crontab.EveryMinute, task.Spec)
	}

	if !task.Enable {
		t.Error("Task should be enabled")
	}

	if task.Next.IsZero() {
		t.Error("Next run time should be set")
	}

	// 获取不存在的任务
	_, err = c.GetTask("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

// TestDefaultLogger 测试默认日志记录器
func TestDefaultLogger(t *testing.T) {
	logger := crontab.NewDefaultLogger()

	// 这些不应该 panic
	logger.Debug("debug %s", "test")
	logger.Info("info %s", "test")
	logger.Warning("warning %s", "test")
	logger.Error("error %s", "test")

	prefixed := logger.CopyWithPrefix("[TEST]")
	prefixed.Info("prefixed message")
}

// BenchmarkCrontabAddTask 基准测试：添加任务
func BenchmarkCrontabAddTask(b *testing.B) {
	logger := crontab.NewDefaultLogger()

	for i := 0; i < b.N; i++ {
		c := crontab.New(logger)
		for j := 0; j < 100; j++ {
			c.AddTask(
				"task-"+string(rune(j)),
				crontab.EveryMinute,
				func(ctx context.Context) {},
			)
		}
	}
}
