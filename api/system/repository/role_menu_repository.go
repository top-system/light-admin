package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// RoleMenuRepository database structure
type RoleMenuRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewRoleMenuRepository creates a new role menu repository
func NewRoleMenuRepository(db lib.Database, logger lib.Logger) RoleMenuRepository {
	return RoleMenuRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a RoleMenuRepository) WithTrx(trxHandle *gorm.DB) RoleMenuRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context. ")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a RoleMenuRepository) Query(param *system.RoleMenuQueryParam) (*system.RoleMenuQueryResult, error) {
	db := a.db.ORM.Model(&system.RoleMenu{})

	if v := param.RoleID; v != 0 {
		db = db.Where("role_id=?", v)
	}

	if v := param.RoleIDs; len(v) > 0 {
		db = db.Where("role_id IN (?)", v)
	}

	// RoleMenu table doesn't have id column, order by role_id instead
	if param.OrderParam.Key == "" || param.OrderParam.Key == "id" {
		db = db.Order("role_id DESC")
	} else {
		db = db.Order(param.OrderParam.ParseOrder())
	}

	list := make([]*system.RoleMenu, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.RoleMenuQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a RoleMenuRepository) GetMenuIDsByRoleID(roleID uint64) ([]uint64, error) {
	var menuIDs []uint64
	result := a.db.ORM.Model(&system.RoleMenu{}).
		Where("role_id=?", roleID).
		Pluck("menu_id", &menuIDs)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return menuIDs, nil
}

func (a RoleMenuRepository) Create(roleMenu *system.RoleMenu) error {
	result := a.db.ORM.Model(roleMenu).Create(roleMenu)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleMenuRepository) BatchCreate(roleMenus []*system.RoleMenu) error {
	if len(roleMenus) == 0 {
		return nil
	}
	result := a.db.ORM.Create(&roleMenus)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleMenuRepository) DeleteByRoleID(roleID uint64) error {
	result := a.db.ORM.Where("role_id=?", roleID).Delete(&system.RoleMenu{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a RoleMenuRepository) DeleteByMenuID(menuID uint64) error {
	result := a.db.ORM.Where("menu_id=?", menuID).Delete(&system.RoleMenu{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
