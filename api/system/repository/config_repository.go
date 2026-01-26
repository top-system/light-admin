package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// ConfigRepository database structure
type ConfigRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db lib.Database, logger lib.Logger) ConfigRepository {
	return ConfigRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a ConfigRepository) WithTrx(trxHandle *gorm.DB) ConfigRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a ConfigRepository) Query(param *system.ConfigQueryParam) (*system.ConfigQueryResult, error) {
	db := a.db.ORM.Model(&system.Config{}).Where("is_deleted = ?", 0)

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("config_name LIKE ? OR config_key LIKE ?", v, v)
	}

	db = db.Order(param.OrderParam.ParseOrder())

	list := make(system.Configs, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.ConfigQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a ConfigRepository) Get(id uint64) (*system.Config, error) {
	config := new(system.Config)

	if ok, err := QueryOne(a.db.ORM.Model(config).Where("id=? AND is_deleted=?", id, 0), config); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return config, nil
}

func (a ConfigRepository) GetByKey(key string) (*system.Config, error) {
	config := new(system.Config)

	if ok, err := QueryOne(a.db.ORM.Model(config).Where("config_key=? AND is_deleted=?", key, 0), config); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return config, nil
}

func (a ConfigRepository) GetAll() (system.Configs, error) {
	list := make(system.Configs, 0)
	result := a.db.ORM.Model(&system.Config{}).Where("is_deleted = ?", 0).Find(&list)
	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}
	return list, nil
}

func (a ConfigRepository) ExistsByKey(key string, excludeID uint64) (bool, error) {
	var count int64
	db := a.db.ORM.Model(&system.Config{}).Where("config_key=? AND is_deleted=?", key, 0)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	result := db.Count(&count)
	if result.Error != nil {
		return false, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}
	return count > 0, nil
}

func (a ConfigRepository) Create(config *system.Config) error {
	result := a.db.ORM.Model(config).Create(config)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a ConfigRepository) Update(id uint64, config *system.Config) error {
	result := a.db.ORM.Model(config).Where("id=?", id).Select(
		"config_name", "config_key", "config_value", "remark", "update_by",
	).Updates(config)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a ConfigRepository) Delete(id uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Config{}).Where("id=?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
