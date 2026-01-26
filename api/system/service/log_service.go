package service

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// LogService service layer
type LogService struct {
	logger        lib.Logger
	logRepository repository.LogRepository
}

// NewLogService creates a new log service
func NewLogService(
	logger lib.Logger,
	logRepository repository.LogRepository,
) LogService {
	return LogService{
		logger:        logger,
		logRepository: logRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a LogService) WithTrx(trxHandle *gorm.DB) LogService {
	a.logRepository = a.logRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询日志
func (a LogService) Query(param *system.LogQueryParam) (*system.LogQueryResult, error) {
	return a.logRepository.Query(param)
}

// Get 获取日志详情
func (a LogService) Get(id uint64) (*system.Log, error) {
	return a.logRepository.Get(id)
}

// Create 创建日志
func (a LogService) Create(log *system.Log) error {
	return a.logRepository.Create(log)
}

// Delete 删除日志
func (a LogService) Delete(id uint64) error {
	return a.logRepository.Delete(id)
}

// BatchDelete 批量删除日志
func (a LogService) BatchDelete(ids []uint64) error {
	return a.logRepository.BatchDelete(ids)
}
