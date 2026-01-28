package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// UserRoleRepository database structure
type UserRoleRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db lib.Database, logger lib.Logger) UserRoleRepository {
	return UserRoleRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a UserRoleRepository) WithTrx(trxHandle *gorm.DB) UserRoleRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context. ")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a UserRoleRepository) Query(param *system.UserRoleQueryParam) (*system.UserRoleQueryResult, error) {
	db := a.db.ORM.Model(system.UserRole{})

	if v := param.UserID; v != 0 {
		db = db.Where("user_id=?", v)
	}
	if v := param.UserIDs; len(v) > 0 {
		db = db.Where("user_id IN (?)", v)
	}

	// UserRole table doesn't have id column, order by user_id instead
	if param.OrderParam.Key == "" || param.OrderParam.Key == "id" {
		db = db.Order("user_id DESC")
	} else {
		db = db.Order(param.OrderParam.ParseOrder())
	}

	list := make(system.UserRoles, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.UserRoleQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a UserRoleRepository) GetRoleIDsByUserID(userID uint64) ([]uint64, error) {
	var roleIDs []uint64
	result := a.db.ORM.Model(&system.UserRole{}).
		Where("user_id=?", userID).
		Pluck("role_id", &roleIDs)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return roleIDs, nil
}

// GetUserIDsByRoleID 根据角色ID获取用户ID列表
func (a UserRoleRepository) GetUserIDsByRoleID(roleID uint64) ([]uint64, error) {
	var userIDs []uint64
	result := a.db.ORM.Model(&system.UserRole{}).
		Where("role_id=?", roleID).
		Pluck("user_id", &userIDs)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return userIDs, nil
}

func (a UserRoleRepository) Create(userRole *system.UserRole) error {
	result := a.db.ORM.Model(userRole).Create(userRole)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRoleRepository) BatchCreate(userRoles []*system.UserRole) error {
	if len(userRoles) == 0 {
		return nil
	}
	result := a.db.ORM.Create(&userRoles)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRoleRepository) DeleteByUserID(userID uint64) error {
	result := a.db.ORM.Where("user_id=?", userID).Delete(&system.UserRole{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRoleRepository) DeleteByRoleID(roleID uint64) error {
	result := a.db.ORM.Where("role_id=?", roleID).Delete(&system.UserRole{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
