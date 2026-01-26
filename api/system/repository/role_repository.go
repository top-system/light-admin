package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// RoleRepository database structure
type RoleRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db lib.Database, logger lib.Logger) RoleRepository {
	return RoleRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a RoleRepository) WithTrx(trxHandle *gorm.DB) RoleRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context. ")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a RoleRepository) Query(param *system.RoleQueryParam) (*system.RoleQueryResult, error) {
	db := a.db.ORM.Model(&system.Role{}).Where("is_deleted = ?", 0)

	if v := param.IDs; len(v) > 0 {
		db = db.Where("id IN (?)", v)
	}

	if v := param.Name; v != "" {
		db = db.Where("name=?", v)
	}

	if v := param.Code; v != "" {
		db = db.Where("code=?", v)
	}

	if v := param.UserID; v != 0 {
		subQuery := a.db.ORM.Model(&system.UserRole{}).
			Where("user_id=?", v).
			Select("role_id")

		db = db.Where("id IN (?)", subQuery)
	}

	if v := param.QueryValue; v != "" {
		v = "%" + v + "%"
		db = db.Where("name LIKE ? OR code LIKE ?", v, v)
	}

	if v := param.Status; v != 0 {
		db = db.Where("status=?", v)
	}

	db = db.Order(param.OrderParam.ParseOrder())

	list := make(system.Roles, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.RoleQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a RoleRepository) Get(id uint64) (*system.Role, error) {
	role := new(system.Role)

	if ok, err := QueryOne(a.db.ORM.Model(role).Where("id=? AND is_deleted=?", id, 0), role); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return role, nil
}

func (a RoleRepository) GetByCode(code string) (*system.Role, error) {
	role := new(system.Role)

	if ok, err := QueryOne(a.db.ORM.Model(role).Where("code=? AND is_deleted=?", code, 0), role); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return role, nil
}

func (a RoleRepository) Create(role *system.Role) error {
	result := a.db.ORM.Model(role).Create(role)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleRepository) Update(id uint64, role *system.Role) error {
	result := a.db.ORM.Model(role).Where("id=?", id).Select(
		"name", "code", "sort", "status", "data_scope", "update_by",
	).Updates(role)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleRepository) Delete(id uint64) error {
	// 软删除
	result := a.db.ORM.Model(&system.Role{}).Where("id=?", id).Update("is_deleted", 1)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleRepository) UpdateStatus(id uint64, status int) error {
	result := a.db.ORM.Model(&system.Role{}).Where("id=?", id).Update("status", status)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
