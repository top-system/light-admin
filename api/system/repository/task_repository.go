package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/queue"
)

// TaskRepository database structure
type TaskRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db lib.Database, logger lib.Logger) TaskRepository {
	return TaskRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a TaskRepository) WithTrx(trxHandle *gorm.DB) TaskRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询任务列表
func (a TaskRepository) Query(param *system.TaskQueryParam) (*system.TaskQueryResult, error) {
	db := a.db.ORM.Model(&system.Task{})

	if v := param.Type; v != "" {
		db = db.Where("type = ?", v)
	}

	if v := param.Status; v != "" {
		db = db.Where("status = ?", v)
	}

	if v := param.CorrelationID; v != "" {
		db = db.Where("correlation_id = ?", v)
	}

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("type LIKE ? OR correlation_id LIKE ? OR public_error LIKE ?", v, v, v)
	}

	if v := param.CreateTimeFrom; v != "" {
		db = db.Where("created_at >= ?", v)
	}

	if v := param.CreateTimeTo; v != "" {
		db = db.Where("created_at <= ?", v+" 23:59:59")
	}

	db = db.Order("created_at DESC")

	list := make(system.Tasks, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.TaskQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

// Get 获取任务详情
func (a TaskRepository) Get(id uint64) (*system.Task, error) {
	task := new(system.Task)

	if ok, err := QueryOne(a.db.ORM.Model(task).Where("id=?", id), task); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return task, nil
}

// Delete 删除任务
func (a TaskRepository) Delete(id uint64) error {
	result := a.db.ORM.Where("id=?", id).Delete(&system.Task{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// BatchDelete 批量删除任务
func (a TaskRepository) BatchDelete(ids []uint64) error {
	result := a.db.ORM.Where("id IN ?", ids).Delete(&system.Task{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// GetTaskTypes 获取所有任务类型
func (a TaskRepository) GetTaskTypes() ([]system.TaskTypeVO, error) {
	var types []string
	result := a.db.ORM.Model(&system.Task{}).Distinct("type").Pluck("type", &types)
	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	taskTypes := make([]system.TaskTypeVO, 0, len(types))
	for _, t := range types {
		taskTypes = append(taskTypes, system.TaskTypeVO{
			Label: t,
			Value: t,
		})
	}

	return taskTypes, nil
}

// GetStatusCounts 获取各状态的任务数量
func (a TaskRepository) GetStatusCounts() (*system.TaskStatsVO, error) {
	stats := &system.TaskStatsVO{}

	// 查询各状态数量
	var counts []struct {
		Status string
		Count  int64
	}

	result := a.db.ORM.Model(&system.Task{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&counts)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	for _, c := range counts {
		switch queue.Status(c.Status) {
		case queue.StatusQueued:
			stats.QueuedCount = c.Count
		case queue.StatusProcessing:
			stats.ProcessingCount = c.Count
		case queue.StatusCompleted:
			stats.CompletedCount = c.Count
		case queue.StatusError:
			stats.ErrorCount = c.Count
		case queue.StatusCanceled:
			stats.CanceledCount = c.Count
		}
	}

	return stats, nil
}
