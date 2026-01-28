package migrate

import (
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/queue"
	"github.com/spf13/cobra"
)

var configFile string

func init() {
	pf := StartCmd.PersistentFlags()
	pf.StringVarP(&configFile, "config", "c",
		"config/config.yaml", "this parameter is used to start the service application")
}

var StartCmd = &cobra.Command{
	Use:          "migrate",
	Short:        "Migrate database",
	Example:      "{execfile} migrate -c config/config.yaml",
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		lib.SetConfigPath(configFile)
	},
	Run: func(cmd *cobra.Command, args []string) {
		config := lib.NewConfig()
		logger := lib.NewLogger(config)
		db := lib.NewDatabase(config, logger)

		if err := db.ORM.AutoMigrate(
			&system.User{},
			&system.UserRole{},
			&system.Role{},
			&system.RoleMenu{},
			&system.Menu{},
			&system.Config{},
			&system.Notice{},
			&system.UserNotice{},
			&system.Dept{},
			&system.Dict{},
			&system.DictItem{},
			&system.Log{},

			// 扩展功能模型 (可选)
			&queue.TaskModel{},    // 任务队列
			&system.DownloadTask{}, // 下载任务
		); err != nil {
			logger.Zap.Fatalf("Error to migrate database: %v", err)
		}

		logger.Zap.Info("Database migration completed successfully")
	},
}
