package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// LogRepository database structure
type LogRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewLogRepository creates a new log repository
func NewLogRepository(db lib.Database, logger lib.Logger) LogRepository {
	return LogRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a LogRepository) WithTrx(trxHandle *gorm.DB) LogRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询日志列表
func (a LogRepository) Query(param *system.LogQueryParam) (*system.LogQueryResult, error) {
	db := a.db.ORM.Model(&system.Log{})

	if v := param.Module; v != "" {
		db = db.Where("module = ?", v)
	}

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("content LIKE ? OR request_uri LIKE ?", v, v)
	}

	if v := param.CreateTimeFrom; v != "" {
		db = db.Where("create_time >= ?", v)
	}

	if v := param.CreateTimeTo; v != "" {
		db = db.Where("create_time <= ?", v+" 23:59:59")
	}

	db = db.Order("create_time DESC")

	list := make(system.Logs, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.LogQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

// Get 获取日志详情
func (a LogRepository) Get(id uint64) (*system.Log, error) {
	log := new(system.Log)

	if ok, err := QueryOne(a.db.ORM.Model(log).Where("id=?", id), log); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return log, nil
}

// Create 创建日志
func (a LogRepository) Create(log *system.Log) error {
	result := a.db.ORM.Create(log)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Delete 删除日志
func (a LogRepository) Delete(id uint64) error {
	result := a.db.ORM.Where("id=?", id).Delete(&system.Log{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// BatchDelete 批量删除日志
func (a LogRepository) BatchDelete(ids []uint64) error {
	result := a.db.ORM.Where("id IN ?", ids).Delete(&system.Log{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
