package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DictItemRepository database structure
type DictItemRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewDictItemRepository creates a new dict item repository
func NewDictItemRepository(db lib.Database, logger lib.Logger) DictItemRepository {
	return DictItemRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a DictItemRepository) WithTrx(trxHandle *gorm.DB) DictItemRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询字典项分页列表
func (a DictItemRepository) Query(param *system.DictItemQueryParam) (*system.DictItemQueryResult, error) {
	db := a.db.ORM.Model(&system.DictItem{}).Where("is_deleted = ?", 0)

	if v := param.DictCode; v != "" {
		db = db.Where("dict_code = ?", v)
	}

	if v := param.Keywords; v != "" {
		db = db.Where("label LIKE ? OR value LIKE ?", "%"+v+"%", "%"+v+"%")
	}

	db = db.Order("sort ASC")

	var list system.DictItems
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return &system.DictItemQueryResult{
		List:       list,
		Pagination: pagination,
	}, nil
}

// GetByDictCode 根据字典编码获取字典项列表
func (a DictItemRepository) GetByDictCode(dictCode string) (system.DictItems, error) {
	var list system.DictItems
	if err := a.db.ORM.Model(&system.DictItem{}).
		Where("dict_code = ? AND is_deleted = ? AND status = ?", dictCode, 0, 1).
		Order("sort ASC").
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, nil
}

// Get 获取字典项
func (a DictItemRepository) Get(id uint64) (*system.DictItem, error) {
	item := new(system.DictItem)

	if ok, err := QueryOne(a.db.ORM.Model(item).Where("id=? AND is_deleted=?", id, 0), item); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return item, nil
}

// Create 创建字典项
func (a DictItemRepository) Create(item *system.DictItem) error {
	result := a.db.ORM.Create(item)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Update 更新字典项
func (a DictItemRepository) Update(id uint64, item *system.DictItem) error {
	result := a.db.ORM.Model(&system.DictItem{}).Where("id=?", id).Updates(map[string]interface{}{
		"dict_code": item.DictCode,
		"label":     item.Label,
		"value":     item.Value,
		"tag_type":  item.TagType,
		"sort":      item.Sort,
		"status":    item.Status,
		"remark":    item.Remark,
		"update_by": item.UpdateBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Delete 删除字典项（软删除）
func (a DictItemRepository) Delete(id uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.DictItem{}).Where("id=?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// DeleteByIDs 批量删除字典项
func (a DictItemRepository) DeleteByIDs(ids []uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.DictItem{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// DeleteByDictCodes 根据字典编码删除字典项
func (a DictItemRepository) DeleteByDictCodes(dictCodes []string, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.DictItem{}).
		Where("dict_code IN ?", dictCodes).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"update_by":  deletedBy,
		})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

