package service

import (
	"sort"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/hash"
)

// UserService service layer
type UserService struct {
	logger             lib.Logger
	config             lib.Config
	userRepository     repository.UserRepository
	userRoleRepository repository.UserRoleRepository
	menuRepository     repository.MenuRepository
	roleRepository     repository.RoleRepository
	roleMenuRepository repository.RoleMenuRepository
	deptRepository     repository.DeptRepository
	permissionCache    PermissionCache
}

// NewUserService creates a new user service
func NewUserService(
	logger lib.Logger,
	config lib.Config,
	userRepository repository.UserRepository,
	userRoleRepository repository.UserRoleRepository,
	roleRepository repository.RoleRepository,
	roleMenuRepository repository.RoleMenuRepository,
	menuRepository repository.MenuRepository,
	deptRepository repository.DeptRepository,
	permissionCache PermissionCache,
) UserService {
	return UserService{
		logger:             logger,
		config:             config,
		userRepository:     userRepository,
		userRoleRepository: userRoleRepository,
		roleRepository:     roleRepository,
		roleMenuRepository: roleMenuRepository,
		menuRepository:     menuRepository,
		deptRepository:     deptRepository,
		permissionCache:    permissionCache,
	}
}

func (a UserService) GetSuperAdmin() *system.User {
	admin := a.config.SuperAdmin
	return &system.User{
		Username: admin.Username,
		Nickname: admin.Realname,
		Password: admin.Password,
	}
}

func (a UserService) IsSuperAdmin(username string) bool {
	return a.config.SuperAdmin.Username == username
}

// GetUserRoleIDs 获取用户角色ID列表
func (a UserService) GetUserRoleIDs(userID uint64) ([]uint64, error) {
	return a.userRoleRepository.GetRoleIDsByUserID(userID)
}

// WithTrx delegates transaction to repository database
func (a UserService) WithTrx(trxHandle *gorm.DB) UserService {
	a.userRepository = a.userRepository.WithTrx(trxHandle)
	a.userRoleRepository = a.userRoleRepository.WithTrx(trxHandle)

	return a
}

func (a UserService) Query(param *system.UserQueryParam) (userQR *system.UserQueryResult, err error) {
	if userQR, err = a.userRepository.Query(param); err != nil {
		return
	}

	// Get user roles
	uRoleQR, err := a.userRoleRepository.Query(
		&system.UserRoleQueryParam{UserIDs: userQR.List.ToIDs()},
	)
	if err != nil {
		return
	}

	m := uRoleQR.List.ToUserIDMap()
	for _, user := range userQR.List {
		if uRoles, ok := m[user.ID]; ok {
			user.RoleIds = uRoles.ToRoleIDs()
		}
	}

	// Get dept names
	deptIDs := make([]uint64, 0)
	deptIDMap := make(map[uint64]struct{})
	for _, user := range userQR.List {
		if user.DeptID > 0 {
			if _, ok := deptIDMap[user.DeptID]; !ok {
				deptIDs = append(deptIDs, user.DeptID)
				deptIDMap[user.DeptID] = struct{}{}
			}
		}
	}

	if len(deptIDs) > 0 {
		deptMap, err := a.deptRepository.GetByIDs(deptIDs)
		if err == nil {
			for _, user := range userQR.List {
				if dept, ok := deptMap[user.DeptID]; ok {
					user.DeptName = dept.Name
				}
			}
		}
	}

	return
}

func (a UserService) Verify(username, password string) (*system.User, error) {
	// super admin user
	admin := a.GetSuperAdmin()
	if admin.Username == username && admin.Password == password {
		return admin, nil
	}

	user, err := a.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	if user.Password != hash.SHA256(password) {
		return nil, errors.UserInvalidPassword
	} else if user.Status != 1 {
		return nil, errors.UserIsDisable
	}

	return user, nil
}

func (a UserService) Check(user *system.User) error {
	if user.Username == a.GetSuperAdmin().Username {
		return errors.UserInvalidUsername
	}

	if qr, err := a.Query(&system.UserQueryParam{Username: user.Username}); err != nil {
		return err
	} else if len(qr.List) > 0 {
		return errors.UserAlreadyExists
	}

	return nil
}

func (a UserService) GetUserInfo(ID uint64) (*system.UserInfo, error) {
	user, err := a.Get(ID)
	if err != nil {
		return nil, err
	}

	userinfo := &system.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	}

	roleIDs, err := a.userRoleRepository.GetRoleIDsByUserID(ID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		roleQR, err := a.roleRepository.Query(&system.RoleQueryParam{
			IDs:    roleIDs,
			Status: 1,
		})
		if err != nil {
			return nil, err
		}
		userinfo.Roles = roleQR.List
	}

	return userinfo, nil
}

// GetCurrentUserInfo 获取当前登录用户的详细信息
func (a UserService) GetCurrentUserInfo(ID uint64, username string) (*dto.CurrentUserInfo, error) {
	// 超级管理员
	if a.IsSuperAdmin(username) {
		admin := a.GetSuperAdmin()
		return &dto.CurrentUserInfo{
			UserID:          0,
			Username:        admin.Username,
			Nickname:        admin.Nickname,
			Avatar:          "",
			CanSwitchTenant: false,
			Roles:           []string{"ROOT"},
			Perms:           []string{"*:*:*"},
		}, nil
	}

	user, err := a.Get(ID)
	if err != nil {
		return nil, err
	}

	info := &dto.CurrentUserInfo{
		UserID:          user.ID,
		Username:        user.Username,
		Nickname:        user.Nickname,
		Avatar:          user.Avatar,
		Gender:          user.Gender,
		Mobile:          user.Mobile,
		Email:           user.Email,
		CreateTime:      user.CreateTime,
		CanSwitchTenant: false,
		Roles:           []string{},
		Perms:           []string{},
	}

	// 获取部门名称
	if user.DeptID > 0 {
		deptMap, err := a.deptRepository.GetByIDs([]uint64{user.DeptID})
		if err == nil {
			if dept, ok := deptMap[user.DeptID]; ok {
				info.DeptName = dept.Name
			}
		}
	}

	// 获取用户角色
	roleIDs, err := a.userRoleRepository.GetRoleIDsByUserID(ID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		roleQR, err := a.roleRepository.Query(&system.RoleQueryParam{
			IDs:    roleIDs,
			Status: 1,
		})
		if err != nil {
			return nil, err
		}

		// 角色编码列表
		info.Roles = roleQR.List.ToCodes()

		// 获取角色关联的按钮权限
		perms, err := a.menuRepository.GetButtonPermsByRoleIDs(roleIDs)
		if err != nil {
			return nil, err
		}
		info.Perms = perms
	}

	return info, nil
}

func (a UserService) GetUserMenuTrees(ID uint64, username string) (system.MenuTrees, error) {
	if a.IsSuperAdmin(username) {
		menuQR, err := a.menuRepository.Query(&system.MenuQueryParam{
			Visible:    1,
			OrderParam: dto.OrderParam{Key: "sort", Direction: dto.OrderByASC},
		})

		if err != nil {
			return nil, err
		}

		return menuQR.List.ToMenuTrees(), nil
	}

	roleIDs, err := a.userRoleRepository.GetRoleIDsByUserID(ID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return nil, errors.UserNoPermission
	}

	menus, err := a.menuRepository.GetMenusByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	if len(menus) == 0 {
		return nil, errors.UserNoPermission
	}

	// 补充父级菜单
	menuMap := menus.ToMap()
	parentIDs := menus.SplitParentIDs()

	var missingIDs []uint64
	for _, parentID := range parentIDs {
		if _, ok := menuMap[parentID]; !ok {
			missingIDs = append(missingIDs, parentID)
		}
	}

	if len(missingIDs) > 0 {
		parentMenuQR, err := a.menuRepository.Query(&system.MenuQueryParam{
			IDs: missingIDs,
		})
		if err != nil {
			return nil, err
		}
		menus = append(menus, parentMenuQR.List...)
	}

	sort.Sort(menus)
	return menus.ToMenuTrees(), nil
}

func (a UserService) GetByUsername(username string) (*system.User, error) {
	userQR, err := a.Query(
		&system.UserQueryParam{Username: username, QueryPassword: true},
	)

	if err != nil {
		return nil, err
	} else if len(userQR.List) == 0 {
		return nil, errors.UserRecordNotFound
	}

	user := userQR.List[0]
	return user, nil
}

func (a UserService) Get(id uint64) (*system.User, error) {
	user, err := a.userRepository.Get(id)
	if err != nil {
		return nil, err
	}

	roleIDs, err := a.userRoleRepository.GetRoleIDsByUserID(id)
	if err != nil {
		return nil, err
	}
	user.RoleIds = roleIDs

	return user, nil
}

func (a UserService) Create(user *system.User) (uint64, error) {
	if err := a.Check(user); err != nil {
		return 0, err
	}

	user.Password = hash.SHA256(user.Password)

	if err := a.userRepository.Create(user); err != nil {
		return 0, err
	}

	// Create user role associations
	if len(user.RoleIds) > 0 {
		if err := a.assignRolesToUser(user.ID, user.RoleIds); err != nil {
			return 0, err
		}
	}

	return user.ID, nil
}

func (a UserService) Update(id uint64, user *system.User) error {
	oUser, err := a.Get(id)
	if err != nil {
		return err
	}

	if user.Username != oUser.Username {
		if err := a.Check(user); err != nil {
			return err
		}
	}

	if user.Password != "" {
		user.Password = hash.SHA256(user.Password)
	} else {
		user.Password = oUser.Password
	}

	user.ID = oUser.ID
	user.CreateTime = oUser.CreateTime

	// Update user role associations if provided
	if user.RoleIds != nil {
		// Delete existing associations
		if err := a.userRoleRepository.DeleteByUserID(id); err != nil {
			return err
		}

		if err := a.assignRolesToUser(id, user.RoleIds); err != nil {
			return err
		}

		// 清除用户权限缓存
		a.permissionCache.InvalidateUserCache(id)
	}

	if err := a.userRepository.Update(id, user); err != nil {
		return err
	}

	return nil
}

func (a UserService) assignRolesToUser(userID uint64, roleIDs []uint64) error {
	if len(roleIDs) == 0 {
		return nil
	}

	userRoles := make([]*system.UserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, &system.UserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	return a.userRoleRepository.BatchCreate(userRoles)
}

func (a UserService) Delete(id uint64) error {
	_, err := a.userRepository.Get(id)
	if err != nil {
		return err
	}

	// 清除用户权限缓存
	a.permissionCache.InvalidateUserCache(id)

	if err := a.userRoleRepository.DeleteByUserID(id); err != nil {
		return err
	}

	return a.userRepository.Delete(id)
}

func (a UserService) UpdateStatus(id uint64, status int) error {
	_, err := a.userRepository.Get(id)
	if err != nil {
		return err
	}

	return a.userRepository.UpdateStatus(id, status)
}

// ResetPassword 重置用户密码
func (a UserService) ResetPassword(id uint64, password string) error {
	_, err := a.userRepository.Get(id)
	if err != nil {
		return err
	}

	return a.userRepository.UpdatePassword(id, hash.SHA256(password))
}

// GetUserForm 获取用户表单数据
func (a UserService) GetUserForm(id uint64) (*system.UserForm, error) {
	user, err := a.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.UserForm{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Status:   user.Status,
		DeptId:   user.DeptID,
		RoleIds:  user.RoleIds,
	}, nil
}

// ListUserOptions 获取用户下拉选项
func (a UserService) ListUserOptions() ([]*system.UserOption, error) {
	status := 1
	qr, err := a.Query(&system.UserQueryParam{
		Status:          &status,
		PaginationParam: dto.PaginationParam{PageSize: 999, PageNum: 1},
	})
	if err != nil {
		return nil, err
	}

	return qr.List.ToOptions(), nil
}

// UpdateProfile 更新用户个人资料
func (a UserService) UpdateProfile(id uint64, username string, profile *system.ProfileForm) error {
	// 超级管理员不支持更新资料（配置文件用户）
	if a.IsSuperAdmin(username) {
		return errors.UserCannotUpdate
	}

	_, err := a.userRepository.Get(id)
	if err != nil {
		return err
	}

	return a.userRepository.UpdateProfile(id, profile)
}
