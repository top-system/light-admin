package service

import (
	"fmt"
	"time"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
)

const (
	// 缓存过期时间
	permCacheExpiration = 30 * time.Minute

	// 缓存键前缀
	permCacheKeyUserRoles = "perm:user:%d:roles" // 用户角色ID列表
	permCacheKeyUserPerms = "perm:user:%d:perms" // 用户权限标识列表
)

// PermissionCache 权限缓存服务
type PermissionCache struct {
	logger             lib.Logger
	cache              lib.Cache
	userRoleRepository repository.UserRoleRepository
}

// NewPermissionCache creates a new permission cache service
func NewPermissionCache(
	logger lib.Logger,
	cache lib.Cache,
	userRoleRepository repository.UserRoleRepository,
) PermissionCache {
	return PermissionCache{
		logger:             logger,
		cache:              cache,
		userRoleRepository: userRoleRepository,
	}
}

// GetUserRoleIDs 获取用户角色ID列表（带缓存）
func (a PermissionCache) GetUserRoleIDs(userID uint64) ([]uint64, error) {
	cacheKey := fmt.Sprintf(permCacheKeyUserRoles, userID)

	// 尝试从缓存获取
	var roleIDs []uint64
	if err := a.cache.Get(cacheKey, &roleIDs); err == nil {
		return roleIDs, nil
	}

	// 从数据库获取
	roleIDs, err := a.userRoleRepository.GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if err := a.cache.Set(cacheKey, roleIDs, permCacheExpiration); err != nil {
		a.logger.Zap.Warn("Failed to cache user roles: " + err.Error())
	}

	return roleIDs, nil
}

// SetUserPerms 缓存用户权限
func (a PermissionCache) SetUserPerms(userID uint64, perms []string) {
	cacheKey := fmt.Sprintf(permCacheKeyUserPerms, userID)
	if err := a.cache.Set(cacheKey, perms, permCacheExpiration); err != nil {
		a.logger.Zap.Warn("Failed to cache user perms: " + err.Error())
	}
}

// GetUserPerms 从缓存获取用户权限
func (a PermissionCache) GetUserPerms(userID uint64) ([]string, bool) {
	cacheKey := fmt.Sprintf(permCacheKeyUserPerms, userID)
	var perms []string
	if err := a.cache.Get(cacheKey, &perms); err == nil {
		return perms, true
	}
	return nil, false
}

// InvalidateUserCache 清除用户权限缓存
func (a PermissionCache) InvalidateUserCache(userID uint64) {
	rolesKey := fmt.Sprintf(permCacheKeyUserRoles, userID)
	permsKey := fmt.Sprintf(permCacheKeyUserPerms, userID)

	if _, err := a.cache.Delete(rolesKey, permsKey); err != nil {
		a.logger.Zap.Warn("Failed to invalidate user cache: " + err.Error())
	}
}

// InvalidateRoleCache 清除角色权限缓存（角色权限变更时调用）
func (a PermissionCache) InvalidateRoleCache(roleID uint64) {
	// 角色权限变更时，需要清除所有拥有该角色的用户的权限缓存
	userIDs, err := a.userRoleRepository.GetUserIDsByRoleID(roleID)
	if err != nil {
		a.logger.Zap.Warn("Failed to get users for role: " + err.Error())
		return
	}

	// 清除这些用户的权限缓存
	for _, userID := range userIDs {
		a.InvalidateUserCache(userID)
	}
}
