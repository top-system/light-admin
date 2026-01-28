package service

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// RoleService service layer
type RoleService struct {
	logger             lib.Logger
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	roleMenuRepository repository.RoleMenuRepository
	menuRepository     repository.MenuRepository
	permissionCache    PermissionCache
}

// NewRoleService creates a new role service
func NewRoleService(
	logger lib.Logger,
	userRepository repository.UserRepository,
	roleRepository repository.RoleRepository,
	roleMenuRepository repository.RoleMenuRepository,
	menuRepository repository.MenuRepository,
	permissionCache PermissionCache,
) RoleService {
	return RoleService{
		logger:             logger,
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		roleMenuRepository: roleMenuRepository,
		menuRepository:     menuRepository,
		permissionCache:    permissionCache,
	}
}

// WithTrx delegates transaction to repository database
func (a RoleService) WithTrx(trxHandle *gorm.DB) RoleService {
	a.roleRepository = a.roleRepository.WithTrx(trxHandle)
	a.userRepository = a.userRepository.WithTrx(trxHandle)
	a.roleMenuRepository = a.roleMenuRepository.WithTrx(trxHandle)

	return a
}

func (a RoleService) Query(param *system.RoleQueryParam) (roleQR *system.RoleQueryResult, err error) {
	return a.roleRepository.Query(param)
}

func (a RoleService) Get(id uint64) (*system.Role, error) {
	role, err := a.roleRepository.Get(id)
	if err != nil {
		return nil, err
	}

	// Get role menu IDs
	menuIDs, err := a.roleMenuRepository.GetMenuIDsByRoleID(id)
	if err != nil {
		return nil, err
	}
	role.MenuIds = menuIDs

	return role, nil
}

func (a RoleService) GetByCode(code string) (*system.Role, error) {
	return a.roleRepository.GetByCode(code)
}

func (a RoleService) CheckName(item *system.Role) error {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{Name: item.Name})
	if err != nil {
		return err
	}

	for _, role := range qr.List {
		if role.ID != item.ID {
			return errors.RoleAlreadyExists
		}
	}

	return nil
}

func (a RoleService) CheckCode(item *system.Role) error {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{Code: item.Code})
	if err != nil {
		return err
	}

	for _, role := range qr.List {
		if role.ID != item.ID {
			return errors.RoleCodeAlreadyExists
		}
	}

	return nil
}

func (a RoleService) Create(role *system.Role) (uint64, error) {
	if err := a.CheckName(role); err != nil {
		return 0, err
	}

	if err := a.CheckCode(role); err != nil {
		return 0, err
	}

	if err := a.roleRepository.Create(role); err != nil {
		return 0, err
	}

	// Create role menu associations
	if len(role.MenuIds) > 0 {
		if err := a.assignMenusToRole(role.ID, role.MenuIds); err != nil {
			return 0, err
		}
	}

	return role.ID, nil
}

func (a RoleService) Update(id uint64, role *system.Role) error {
	oRole, err := a.Get(id)
	if err != nil {
		return err
	}

	role.ID = id

	if role.Name != oRole.Name {
		if err = a.CheckName(role); err != nil {
			return err
		}
	}

	if role.Code != oRole.Code {
		if err = a.CheckCode(role); err != nil {
			return err
		}
	}

	if err := a.roleRepository.Update(id, role); err != nil {
		return err
	}

	// 清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(id)

	return nil
}

func (a RoleService) Delete(id uint64) error {
	_, err := a.roleRepository.Get(id)
	if err != nil {
		return err
	}

	userQR, err := a.userRepository.Query(&system.UserQueryParam{
		RoleIDs: []uint64{id},
	})

	if err != nil {
		return err
	} else if userQR.Pagination.Total > 0 {
		return errors.RoleNotAllowDeleteWithUser
	}

	// 先清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(id)

	// Delete role menu associations
	if err := a.roleMenuRepository.DeleteByRoleID(id); err != nil {
		return err
	}

	if err := a.roleRepository.Delete(id); err != nil {
		return err
	}

	return nil
}

func (a RoleService) UpdateStatus(id uint64, status int) error {
	_, err := a.roleRepository.Get(id)
	if err != nil {
		return err
	}

	return a.roleRepository.UpdateStatus(id, status)
}

// GetRoleMenuIds 获取角色的菜单ID列表
func (a RoleService) GetRoleMenuIds(roleID uint64) ([]uint64, error) {
	return a.roleMenuRepository.GetMenuIDsByRoleID(roleID)
}

// AssignMenusToRole 为角色分配菜单
func (a RoleService) AssignMenusToRole(roleID uint64, menuIDs []uint64) error {
	_, err := a.roleRepository.Get(roleID)
	if err != nil {
		return err
	}

	// Delete existing associations
	if err := a.roleMenuRepository.DeleteByRoleID(roleID); err != nil {
		return err
	}

	if err := a.assignMenusToRole(roleID, menuIDs); err != nil {
		return err
	}

	// 清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(roleID)

	return nil
}

func (a RoleService) assignMenusToRole(roleID uint64, menuIDs []uint64) error {
	if len(menuIDs) == 0 {
		return nil
	}

	roleMenus := make([]*system.RoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		roleMenus = append(roleMenus, &system.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		})
	}

	return a.roleMenuRepository.BatchCreate(roleMenus)
}

// ListRoleOptions 获取角色下拉选项
func (a RoleService) ListRoleOptions() ([]system.RoleOption, error) {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{
		Status:          1,
		PaginationParam: dto.PaginationParam{PageSize: 999, PageNum: 1},
	})

	if err != nil {
		return nil, err
	}

	return qr.List.ToOptions(), nil
}
