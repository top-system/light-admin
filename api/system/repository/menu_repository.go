package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// MenuRepository database structure
type MenuRepository struct {
	db     lib.Database
	logger lib.Logger
}

// NewMenuRepository creates a new menu repository
func NewMenuRepository(db lib.Database, logger lib.Logger) MenuRepository {
	return MenuRepository{
		db:     db,
		logger: logger,
	}
}

// WithTrx enables repository with transaction
func (a MenuRepository) WithTrx(trxHandle *gorm.DB) MenuRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context. ")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a MenuRepository) Query(param *system.MenuQueryParam) (*system.MenuQueryResult, error) {
	db := a.db.ORM.Model(&system.Menu{})

	if v := param.IDs; len(v) > 0 {
		db = db.Where("id IN (?)", v)
	}

	if v := param.Name; v != "" {
		db = db.Where("name=?", v)
	}

	if v := param.ParentID; v != nil {
		db = db.Where("parent_id=?", *v)
	}

	if v := param.PrefixTreePath; v != "" {
		db = db.Where("tree_path LIKE ?", v+"%")
	}

	if v := param.Type; v != 0 {
		db = db.Where("type=?", v)
	}

	if v := param.Visible; v != 0 {
		db = db.Where("visible=?", v)
	}

	if v := param.Keywords; v != "" {
		v = "%" + v + "%"
		db = db.Where("name LIKE ?", v)
	}

	db = db.Order(param.OrderParam.ParseOrder())

	list := make(system.Menus, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.MenuQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a MenuRepository) Get(id uint64) (*system.Menu, error) {
	menu := new(system.Menu)

	if ok, err := QueryOne(a.db.ORM.Model(menu).Where("id=?", id), menu); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return menu, nil
}

func (a MenuRepository) Create(menu *system.Menu) error {
	result := a.db.ORM.Model(menu).Create(menu)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a MenuRepository) Update(id uint64, menu *system.Menu) error {
	result := a.db.ORM.Model(menu).Where("id=?", id).Select(
		"parent_id", "tree_path", "name", "type", "route_name", "route_path",
		"component", "perm", "always_show", "keep_alive", "visible",
		"sort", "icon", "redirect", "params",
	).Updates(menu)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a MenuRepository) Delete(id uint64) error {
	menu := new(system.Menu)

	result := a.db.ORM.Model(menu).Where("id=?", id).Delete(menu)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a MenuRepository) UpdateVisible(id uint64, visible int) error {
	menu := new(system.Menu)

	result := a.db.ORM.Model(menu).Where("id=?", id).Update("visible", visible)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a MenuRepository) UpdateTreePath(id uint64, treePath string) error {
	menu := new(system.Menu)

	result := a.db.ORM.Model(menu).Where("id=?", id).Update("tree_path", treePath)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

// GetMenusByRoleIDs 根据角色ID列表获取菜单
func (a MenuRepository) GetMenusByRoleIDs(roleIDs []uint64) (system.Menus, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	subQuery := a.db.ORM.Model(&system.RoleMenu{}).
		Where("role_id IN (?)", roleIDs).
		Select("DISTINCT menu_id")

	list := make(system.Menus, 0)
	result := a.db.ORM.Model(&system.Menu{}).
		Where("id IN (?)", subQuery).
		Order("sort ASC").
		Find(&list)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return list, nil
}

// GetButtonPermsByRoleIDs 获取角色关联的按钮权限标识列表
func (a MenuRepository) GetButtonPermsByRoleIDs(roleIDs []uint64) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	subQuery := a.db.ORM.Model(&system.RoleMenu{}).
		Where("role_id IN (?)", roleIDs).
		Select("DISTINCT menu_id")

	var perms []string
	result := a.db.ORM.Model(&system.Menu{}).
		Where("id IN (?)", subQuery).
		Where("type = ?", 4). // 按钮类型
		Where("perm != ''").
		Pluck("perm", &perms)

	if result.Error != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return perms, nil
}
