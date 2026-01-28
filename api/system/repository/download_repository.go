package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DownloadRepository database structure
type DownloadRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewDownloadRepository creates a new download repository
func NewDownloadRepository(db lib.Database, logger lib.Logger) DownloadRepository {
	return DownloadRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a DownloadRepository) WithTrx(trxHandle *gorm.DB) DownloadRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询下载任务列表
func (a DownloadRepository) Query(param *system.DownloadTaskQueryParam) (*system.DownloadTaskQueryResult, error) {
	db := a.db.ORM.Model(&system.DownloadTask{})

	if v := param.Status; v != "" {
		db = db.Where("status = ?", v)
	}

	if v := param.Downloader; v != "" {
		db = db.Where("downloader = ?", v)
	}

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("name LIKE ? OR url LIKE ? OR save_path LIKE ? OR hash LIKE ?", v, v, v, v)
	}

	if v := param.CreateTimeFrom; v != "" {
		db = db.Where("created_at >= ?", v)
	}

	if v := param.CreateTimeTo; v != "" {
		db = db.Where("created_at <= ?", v+" 23:59:59")
	}

	db = db.Order("created_at DESC")

	list := make(system.DownloadTasks, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.DownloadTaskQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

// Get 获取下载任务详情
func (a DownloadRepository) Get(id uint64) (*system.DownloadTask, error) {
	task := new(system.DownloadTask)

	if ok, err := QueryOne(a.db.ORM.Model(task).Where("id=?", id), task); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return task, nil
}

// GetByTaskID 根据任务ID获取下载任务
func (a DownloadRepository) GetByTaskID(taskID string) (*system.DownloadTask, error) {
	task := new(system.DownloadTask)

	if ok, err := QueryOne(a.db.ORM.Model(task).Where("task_id=?", taskID), task); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return task, nil
}

// Create 创建下载任务
func (a DownloadRepository) Create(task *system.DownloadTask) error {
	result := a.db.ORM.Create(task)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Update 更新下载任务
func (a DownloadRepository) Update(task *system.DownloadTask) error {
	result := a.db.ORM.Save(task)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// UpdateStatus 更新下载任务状态
func (a DownloadRepository) UpdateStatus(id uint64, status string, downloaded, total, downloadSpeed, uploaded, uploadSpeed int64, errorMessage string) error {
	result := a.db.ORM.Model(&system.DownloadTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         status,
		"downloaded":     downloaded,
		"total":          total,
		"download_speed": downloadSpeed,
		"uploaded":       uploaded,
		"upload_speed":   uploadSpeed,
		"error_message":  errorMessage,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Delete 删除下载任务
func (a DownloadRepository) Delete(id uint64) error {
	result := a.db.ORM.Where("id=?", id).Delete(&system.DownloadTask{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// BatchDelete 批量删除下载任务
func (a DownloadRepository) BatchDelete(ids []uint64) error {
	result := a.db.ORM.Where("id IN ?", ids).Delete(&system.DownloadTask{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// GetStatusCounts 获取各状态的任务数量
func (a DownloadRepository) GetStatusCounts() (*system.DownloadTaskStatsVO, error) {
	stats := &system.DownloadTaskStatsVO{}

	// 查询各状态数量
	var counts []struct {
		Status string
		Count  int64
	}

	result := a.db.ORM.Model(&system.DownloadTask{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&counts)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	for _, c := range counts {
		stats.TotalCount += c.Count
		switch c.Status {
		case "downloading":
			stats.DownloadingCount = c.Count
		case "seeding":
			stats.SeedingCount = c.Count
		case "completed":
			stats.CompletedCount = c.Count
		case "error":
			stats.ErrorCount = c.Count
		}
	}

	return stats, nil
}

// GetActiveTaskIDs 获取活跃任务ID列表（用于状态同步）
func (a DownloadRepository) GetActiveTaskIDs() ([]system.DownloadTask, error) {
	var tasks []system.DownloadTask
	result := a.db.ORM.Model(&system.DownloadTask{}).
		Where("status IN ?", []string{"downloading", "seeding", "unknown", "queued"}).
		Select("id, queue_task_id, task_id, hash, downloader").
		Find(&tasks)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return tasks, nil
}

// GetByQueueTaskID 根据队列任务ID获取下载任务
func (a DownloadRepository) GetByQueueTaskID(queueTaskID uint64) (*system.DownloadTask, error) {
	task := new(system.DownloadTask)

	if ok, err := QueryOne(a.db.ORM.Model(task).Where("queue_task_id=?", queueTaskID), task); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return task, nil
}

// UpdateFromDownloader 从下载器同步更新任务完整信息
func (a DownloadRepository) UpdateFromDownloader(id uint64, taskID, hash, name, savePath, status string, downloaded, total, downloadSpeed, uploaded, uploadSpeed int64, errorMessage string) error {
	updates := map[string]interface{}{
		"status":         status,
		"downloaded":     downloaded,
		"total":          total,
		"download_speed": downloadSpeed,
		"uploaded":       uploaded,
		"upload_speed":   uploadSpeed,
		"error_message":  errorMessage,
	}

	// 只在有值时更新这些字段
	if taskID != "" {
		updates["task_id"] = taskID
	}
	if hash != "" {
		updates["hash"] = hash
	}
	if name != "" {
		updates["name"] = name
	}
	if savePath != "" {
		updates["save_path"] = savePath
	}

	result := a.db.ORM.Model(&system.DownloadTask{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
