package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DictRepository database structure
type DictRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewDictRepository creates a new dict repository
func NewDictRepository(db lib.Database, logger lib.Logger) DictRepository {
	return DictRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a DictRepository) WithTrx(trxHandle *gorm.DB) DictRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询字典分页列表
func (a DictRepository) Query(param *system.DictQueryParam) (*system.DictQueryResult, error) {
	db := a.db.ORM.Model(&system.Dict{}).Where("is_deleted = ?", 0)

	if v := param.Keywords; v != "" {
		db = db.Where("name LIKE ? OR dict_code LIKE ?", "%"+v+"%", "%"+v+"%")
	}

	db = db.Order("create_time DESC")

	var list system.Dicts
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return &system.DictQueryResult{
		List:       list,
		Pagination: pagination,
	}, nil
}

// GetAll 获取所有启用的字典
func (a DictRepository) GetAll() (system.Dicts, error) {
	var list system.Dicts
	if err := a.db.ORM.Model(&system.Dict{}).
		Where("is_deleted = ? AND status = ?", 0, 1).
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, nil
}

// Get 获取字典
func (a DictRepository) Get(id uint64) (*system.Dict, error) {
	dict := new(system.Dict)

	if ok, err := QueryOne(a.db.ORM.Model(dict).Where("id=? AND is_deleted=?", id, 0), dict); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return dict, nil
}

// GetByCode 根据编码获取字典
func (a DictRepository) GetByCode(dictCode string, excludeID ...uint64) (*system.Dict, error) {
	dict := new(system.Dict)
	db := a.db.ORM.Model(dict).Where("dict_code = ? AND is_deleted = ?", dictCode, 0)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		db = db.Where("id != ?", excludeID[0])
	}

	if ok, err := QueryOne(db, dict); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, nil
	}

	return dict, nil
}

// GetByIDs 根据ID列表获取字典列表
func (a DictRepository) GetByIDs(ids []uint64) (system.Dicts, error) {
	var list system.Dicts
	if err := a.db.ORM.Model(&system.Dict{}).
		Where("id IN ? AND is_deleted = ?", ids, 0).
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, nil
}

// Create 创建字典
func (a DictRepository) Create(dict *system.Dict) error {
	result := a.db.ORM.Create(dict)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Update 更新字典
func (a DictRepository) Update(id uint64, dict *system.Dict) error {
	result := a.db.ORM.Model(&system.Dict{}).Where("id=?", id).Updates(map[string]interface{}{
		"dict_code": dict.DictCode,
		"name":      dict.Name,
		"status":    dict.Status,
		"remark":    dict.Remark,
		"update_by": dict.UpdateBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Delete 删除字典（软删除）
func (a DictRepository) Delete(id uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Dict{}).Where("id=?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// DeleteByIDs 批量删除字典
func (a DictRepository) DeleteByIDs(ids []uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Dict{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// UpdateDictItemsCode 更新字典项的字典编码
func (a DictRepository) UpdateDictItemsCode(oldCode, newCode string) error {
	result := a.db.ORM.Model(&system.DictItem{}).
		Where("dict_code = ?", oldCode).
		Update("dict_code", newCode)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

