package service

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// TaskService service layer
type TaskService struct {
	logger         lib.Logger
	taskRepository repository.TaskRepository
}

// NewTaskService creates a new task service
func NewTaskService(
	logger lib.Logger,
	taskRepository repository.TaskRepository,
) TaskService {
	return TaskService{
		logger:         logger,
		taskRepository: taskRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a TaskService) WithTrx(trxHandle *gorm.DB) TaskService {
	a.taskRepository = a.taskRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询任务
func (a TaskService) Query(param *system.TaskQueryParam) (*system.TaskQueryResult, error) {
	return a.taskRepository.Query(param)
}

// Get 获取任务详情
func (a TaskService) Get(id uint64) (*system.Task, error) {
	return a.taskRepository.Get(id)
}

// Delete 删除任务
func (a TaskService) Delete(id uint64) error {
	return a.taskRepository.Delete(id)
}

// BatchDelete 批量删除任务
func (a TaskService) BatchDelete(ids []uint64) error {
	return a.taskRepository.BatchDelete(ids)
}

// GetTaskTypes 获取所有任务类型
func (a TaskService) GetTaskTypes() ([]system.TaskTypeVO, error) {
	return a.taskRepository.GetTaskTypes()
}

// GetStats 获取任务统计信息
func (a TaskService) GetStats() (*system.TaskStatsVO, error) {
	return a.taskRepository.GetStatusCounts()
}
