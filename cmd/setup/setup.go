package setup

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/file"
	"github.com/top-system/light-admin/pkg/hash"
)

var configFile string
var menuFile string

func init() {
	pf := StartCmd.PersistentFlags()
	pf.StringVarP(&configFile, "config", "c",
		"config/config.yaml", "this parameter is used to start the service application")
	pf.StringVarP(&menuFile, "menu", "m",
		"config/menu.yaml", "this parameter is used to set the initialized menu data.")
}

var StartCmd = &cobra.Command{
	Use:          "setup",
	Short:        "Set up data for the application",
	Example:      "{execfile} setup -c config/config.yaml -m config/menu.yaml",
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		lib.SetConfigPath(configFile)
	},
	Run: func(cmd *cobra.Command, args []string) {
		config := lib.NewConfig()
		logger := lib.NewLogger(config)
		db := lib.NewDatabase(config, logger)

		// 初始化 repositories
		menuRepo := repository.NewMenuRepository(db, logger)
		roleMenuRepo := repository.NewRoleMenuRepository(db, logger)
		roleRepo := repository.NewRoleRepository(db, logger)
		userRepo := repository.NewUserRepository(db, logger)
		userRoleRepo := repository.NewUserRoleRepository(db, logger)
		dictRepo := repository.NewDictRepository(db, logger)
		dictItemRepo := repository.NewDictItemRepository(db, logger)

		// 初始化 services
		menuService := service.NewMenuService(
			logger,
			menuRepo,
			roleMenuRepo,
		)

		// Step 1: 导入菜单数据
		if !file.IsFile(menuFile) {
			logger.Zap.Fatal("menu file does not exist")
		}

		fs, err := os.Open(menuFile)
		if err != nil {
			logger.Zap.Fatalf("menu file could not be opened: %v", err)
		}
		defer fs.Close()

		var menuTrees system.MenuTrees
		yd := yaml.NewDecoder(fs)
		if err = yd.Decode(&menuTrees); err != nil {
			logger.Zap.Fatalf("menu file decode error: %v", err)
		}

		if err = menuService.CreateMenus(0, menuTrees); err != nil {
			logger.Zap.Fatalf("menu file init err: %v", err)
		}
		logger.Zap.Info("Step 1: Menu data imported successfully")

		// Step 2: 创建超级管理员角色
		var roleID uint64
		adminRole := &system.Role{
			Name:   "超级管理员",
			Code:   "ROOT",
			Sort:   1,
			Status: 1,
		}

		// 检查角色是否已存在
		existingRole, _ := roleRepo.Query(&system.RoleQueryParam{Code: "ROOT"})
		if existingRole != nil && len(existingRole.List) > 0 {
			roleID = existingRole.List[0].ID
			logger.Zap.Info("Step 2: ROOT role already exists, skipping creation")
		} else {
			if err := roleRepo.Create(adminRole); err != nil {
				logger.Zap.Fatalf("failed to create admin role: %v", err)
			}
			roleID = adminRole.ID
			logger.Zap.Info("Step 2: ROOT role created successfully")
		}

		// Step 3: 为角色分配所有菜单权限
		// 获取所有菜单
		menuQR, err := menuRepo.Query(&system.MenuQueryParam{
			PaginationParam: dto.PaginationParam{PageSize: 9999, PageNum: 1},
		})
		if err != nil {
			logger.Zap.Fatalf("failed to query menus: %v", err)
		}

		// 先删除该角色的所有权限，再重新分配
		if err := roleMenuRepo.DeleteByRoleID(roleID); err != nil {
			logger.Zap.Warnf("failed to delete existing role menus: %v", err)
		}

		// 为角色分配所有菜单权限
		roleMenus := make([]*system.RoleMenu, 0, len(menuQR.List))
		for _, menu := range menuQR.List {
			roleMenus = append(roleMenus, &system.RoleMenu{
				RoleID: roleID,
				MenuID: menu.ID,
			})
		}

		if err := roleMenuRepo.BatchCreate(roleMenus); err != nil {
			logger.Zap.Warnf("failed to create role menus: %v", err)
		}
		logger.Zap.Infof("Step 3: Assigned %d permissions to ROOT role", len(menuQR.List))

		// Step 4: 创建管理员用户
		adminUsername := "admin"
		adminPassword := "123456" // 默认密码

		// 检查用户是否已存在
		existingUser, _ := userRepo.Query(&system.UserQueryParam{Username: adminUsername})
		if existingUser != nil && len(existingUser.List) > 0 {
			logger.Zap.Info("Step 4: Admin user already exists, skipping creation")
		} else {
			adminUser := &system.User{
				Username: adminUsername,
				Nickname: "系统管理员",
				Password: hash.SHA256(adminPassword),
				Email:    "admin@example.com",
				Status:   1,
				Gender:   1,
			}

			if err := userRepo.Create(adminUser); err != nil {
				logger.Zap.Fatalf("failed to create admin user: %v", err)
			}

			// 关联用户和角色
			userRole := &system.UserRole{
				UserID: adminUser.ID,
				RoleID: roleID,
			}
			if err := userRoleRepo.Create(userRole); err != nil {
				logger.Zap.Fatalf("failed to create user role: %v", err)
			}

			logger.Zap.Info("Step 4: Admin user created successfully")
		}

		// Step 5: 初始化字典数据
		initDicts := []struct {
			DictCode string
			Name     string
		}{
			{"gender", "性别"},
			{"notice_type", "通知类型"},
			{"notice_level", "通知级别"},
		}

		for _, d := range initDicts {
			existingDict, _ := dictRepo.GetByCode(d.DictCode)
			if existingDict != nil {
				continue // 已存在则跳过
			}

			dict := &system.Dict{
				DictCode: d.DictCode,
				Name:     d.Name,
				Status:   1,
				CreateBy: 1,
			}
			if err := dictRepo.Create(dict); err != nil {
				logger.Zap.Warnf("failed to create dict %s: %v", d.DictCode, err)
			}
		}
		logger.Zap.Info("Step 5: Dict data initialized successfully")

		// Step 6: 初始化字典项数据
		initDictItems := []struct {
			DictCode string
			Value    string
			Label    string
			TagType  string
			Sort     int
		}{
			// 性别
			{"gender", "1", "男", "primary", 1},
			{"gender", "2", "女", "danger", 2},
			{"gender", "0", "保密", "info", 3},
			// 通知类型
			{"notice_type", "1", "系统升级", "success", 1},
			{"notice_type", "2", "系统维护", "primary", 2},
			{"notice_type", "3", "安全警告", "danger", 3},
			{"notice_type", "4", "假期通知", "success", 4},
			{"notice_type", "5", "公司新闻", "primary", 5},
			{"notice_type", "99", "其他", "info", 99},
			// 通知级别
			{"notice_level", "L", "低", "info", 1},
			{"notice_level", "M", "中", "warning", 2},
			{"notice_level", "H", "高", "danger", 3},
		}

		for _, item := range initDictItems {
			// 检查是否已存在
			existingItems, _ := dictItemRepo.GetByDictCode(item.DictCode)
			exists := false
			for _, existing := range existingItems {
				if existing.Value == item.Value {
					exists = true
					break
				}
			}
			if exists {
				continue
			}

			dictItem := &system.DictItem{
				DictCode: item.DictCode,
				Value:    item.Value,
				Label:    item.Label,
				TagType:  item.TagType,
				Sort:     item.Sort,
				Status:   1,
				CreateBy: 1,
			}
			if err := dictItemRepo.Create(dictItem); err != nil {
				logger.Zap.Warnf("failed to create dict item %s-%s: %v", item.DictCode, item.Value, err)
			}
		}
		logger.Zap.Info("Step 6: Dict item data initialized successfully")

		logger.Zap.Info("========================================")
		logger.Zap.Info("Setup completed!")
		logger.Zap.Info("Admin credentials:")
		logger.Zap.Infof("  Username: %s", adminUsername)
		logger.Zap.Infof("  Password: %s", adminPassword)
		logger.Zap.Info("========================================")
	},
}
