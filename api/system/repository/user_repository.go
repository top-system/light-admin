package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// UserRepository database structure
type UserRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db lib.Database, logger lib.Logger) UserRepository {
	return UserRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a UserRepository) WithTrx(trxHandle *gorm.DB) UserRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context. ")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// Query gets all users
func (a UserRepository) Query(param *system.UserQueryParam) (*system.UserQueryResult, error) {
	db := a.db.ORM.Model(&system.User{}).Where("is_deleted = ?", 0)

	if v := param.QueryPassword; !v {
		db = db.Omit("password")
	}

	if v := param.Username; v != "" {
		db = db.Where("username = (?)", v)
	}

	if v := param.Nickname; v != "" {
		db = db.Where("nickname = (?)", v)
	}

	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	if v := param.DeptID; v > 0 {
		db = db.Where("dept_id = ?", v)
	}

	if v := param.RoleIDs; len(v) > 0 {
		subQuery := a.db.ORM.Model(&system.UserRole{}).
			Select("user_id").
			Where("role_id IN (?)", v)

		db = db.Where("id IN (?)", subQuery)
	}

	if v := param.QueryValue; v != "" {
		v = "%" + v + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ? OR mobile LIKE ? OR email LIKE ?", v, v, v, v)
	}

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ? OR mobile LIKE ? OR email LIKE ?", v, v, v, v)
	}

	if v := param.CreateTimeFrom; v != "" {
		db = db.Where("create_time >= ?", v)
	}

	if v := param.CreateTimeTo; v != "" {
		db = db.Where("create_time <= ?", v+" 23:59:59")
	}

	db = db.Order(param.OrderParam.ParseOrder())

	list := make(system.Users, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.UserQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a UserRepository) Get(id uint64) (*system.User, error) {
	user := new(system.User)

	if ok, err := QueryOne(a.db.ORM.Model(user).Where("id=? AND is_deleted=?", id, 0), user); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return user, nil
}

func (a UserRepository) Create(user *system.User) error {
	result := a.db.ORM.Model(user).Create(user)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRepository) Update(id uint64, user *system.User) error {
	result := a.db.ORM.Model(user).Where("id=?", id).Select(
		"username", "nickname", "gender", "dept_id", "avatar",
		"mobile", "status", "email", "update_by",
	).Updates(user)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRepository) Delete(id uint64) error {
	// 软删除
	result := a.db.ORM.Model(&system.User{}).Where("id=?", id).Update("is_deleted", 1)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRepository) UpdateStatus(id uint64, status int) error {
	result := a.db.ORM.Model(&system.User{}).Where("id=?", id).Update("status", status)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserRepository) UpdatePassword(id uint64, password string) error {
	result := a.db.ORM.Model(&system.User{}).Where("id=?", id).Update("password", password)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
