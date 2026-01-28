package repository

import (
	"strconv"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DeptRepository database structure
type DeptRepository struct {
	db       lib.Database
	logger   lib.Logger
	dbCompat lib.DBCompat
}

// NewDeptRepository creates a new dept repository
func NewDeptRepository(db lib.Database, logger lib.Logger, dbCompat lib.DBCompat) DeptRepository {
	return DeptRepository{
		db:       db,
		logger:   logger,
		dbCompat: dbCompat,
	}
}

// WithTrx enables repository with transaction
func (a DeptRepository) WithTrx(trxHandle *gorm.DB) DeptRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query 查询部门列表
func (a DeptRepository) Query(param *system.DeptQueryParam) (system.Depts, error) {
	db := a.db.ORM.Model(&system.Dept{}).Where("is_deleted = ?", 0)

	if v := param.Keywords; v != "" {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}

	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	db = db.Order("sort ASC")

	var list system.Depts
	if err := db.Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, nil
}

// Get 获取部门
func (a DeptRepository) Get(id uint64) (*system.Dept, error) {
	dept := new(system.Dept)

	if ok, err := QueryOne(a.db.ORM.Model(dept).Where("id=? AND is_deleted=?", id, 0), dept); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return dept, nil
}

// GetByCode 根据编码获取部门
func (a DeptRepository) GetByCode(code string, excludeID ...uint64) (*system.Dept, error) {
	dept := new(system.Dept)
	db := a.db.ORM.Model(dept).Where("code = ? AND is_deleted = ?", code, 0)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		db = db.Where("id != ?", excludeID[0])
	}

	if ok, err := QueryOne(db, dept); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, nil
	}

	return dept, nil
}

// Create 创建部门
func (a DeptRepository) Create(dept *system.Dept) error {
	result := a.db.ORM.Create(dept)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Update 更新部门
func (a DeptRepository) Update(id uint64, dept *system.Dept) error {
	result := a.db.ORM.Model(dept).Where("id=?", id).Select(
		"name", "code", "parent_id", "tree_path", "sort", "status", "update_by",
	).Updates(dept)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// Delete 删除部门（软删除）
func (a DeptRepository) Delete(id uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Dept{}).Where("id=?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// DeleteByTreePath 根据tree_path删除部门及子部门
func (a DeptRepository) DeleteByTreePath(deptId uint64, deletedBy uint64) error {
	// 删除部门本身和所有子部门（tree_path包含该部门ID的）
	treePathExpr := a.dbCompat.TreePathLike("tree_path")
	result := a.db.ORM.Model(&system.Dept{}).
		Where("id = ? OR "+treePathExpr+" LIKE ?", deptId, "%,"+strconv.FormatUint(deptId, 10)+",%").
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"update_by":  deletedBy,
		})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// GetAllEnabled 获取所有启用的部门
func (a DeptRepository) GetAllEnabled() (system.Depts, error) {
	var list system.Depts
	if err := a.db.ORM.Model(&system.Dept{}).
		Where("is_deleted = ? AND status = ?", 0, 1).
		Order("sort ASC").
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, nil
}

// GetByIDs 根据ID列表获取部门Map
func (a DeptRepository) GetByIDs(ids []uint64) (map[uint64]*system.Dept, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var list system.Depts
	if err := a.db.ORM.Model(&system.Dept{}).
		Where("id IN (?) AND is_deleted = ?", ids, 0).
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	result := make(map[uint64]*system.Dept)
	for _, dept := range list {
		result[dept.ID] = dept
	}

	return result, nil
}
