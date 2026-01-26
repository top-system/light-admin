package main

import (
	"github.com/top-system/light-admin/cmd"
)

// @title Echo Admin API
// @version 1.0
// @description Echo Admin 后台管理系统 API 文档
// @termsOfService https://github.com/top-system/light-admin

// @contact.name LiuSha
// @contact.email liusha@email.cn

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:2222
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	cmd.Execute()
}
