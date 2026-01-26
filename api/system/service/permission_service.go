package service

import (
	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// PermissionService 基于 perm 标识的权限服务
type PermissionService struct {
	logger             lib.Logger
	menuRepository     repository.MenuRepository
	roleMenuRepository repository.RoleMenuRepository
	userRoleRepository repository.UserRoleRepository
	roleRepository     repository.RoleRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	logger lib.Logger,
	menuRepository repository.MenuRepository,
	roleMenuRepository repository.RoleMenuRepository,
	userRoleRepository repository.UserRoleRepository,
	roleRepository repository.RoleRepository,
) PermissionService {
	return PermissionService{
		logger:             logger,
		menuRepository:     menuRepository,
		roleMenuRepository: roleMenuRepository,
		userRoleRepository: userRoleRepository,
		roleRepository:     roleRepository,
	}
}

// GetRolePerms 获取角色的权限标识列表
func (a PermissionService) GetRolePerms(roleIDs []uint64) ([]string, error) {
	return a.menuRepository.GetButtonPermsByRoleIDs(roleIDs)
}

// HasPerm 检查角色是否有指定权限
func (a PermissionService) HasPerm(roleIDs []uint64, perm string) (bool, error) {
	if perm == "" {
		return true, nil
	}

	perms, err := a.GetRolePerms(roleIDs)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == perm || p == "*:*:*" {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoleIDs 获取用户的角色ID列表
func (a PermissionService) GetUserRoleIDs(userID uint64) ([]uint64, error) {
	return a.userRoleRepository.GetRoleIDsByUserID(userID)
}

// GetUserPerms 获取用户的所有权限标识
func (a PermissionService) GetUserPerms(userID uint64) ([]string, error) {
	roleIDs, err := a.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}

	return a.GetRolePerms(roleIDs)
}

// GetUserRoleCodes 获取用户的角色编码列表
func (a PermissionService) GetUserRoleCodes(userID uint64) ([]string, error) {
	roleIDs, err := a.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return nil, nil
	}

	roleQR, err := a.roleRepository.Query(&system.RoleQueryParam{
		IDs:    roleIDs,
		Status: 1,
	})
	if err != nil {
		return nil, err
	}

	return roleQR.List.ToCodes(), nil
}
