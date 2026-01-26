package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Config 系统配置模型
type Config struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigName  string       `gorm:"column:config_name;size:50;not null" json:"configName"`
	ConfigKey   string       `gorm:"column:config_key;size:50;not null;uniqueIndex:uk_config_key" json:"configKey"`
	ConfigValue string       `gorm:"column:config_value;size:100;not null" json:"configValue"`
	Remark      string       `gorm:"column:remark;size:255" json:"remark"`
	CreateTime  dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	CreateBy    uint64       `gorm:"column:create_by" json:"createBy"`
	UpdateTime  dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	UpdateBy    uint64       `gorm:"column:update_by" json:"updateBy"`
	IsDeleted   int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (Config) TableName() string {
	return "t_config"
}

type Configs []*Config

type ConfigQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Keywords string `query:"keywords"`
}

type ConfigQueryResult struct {
	List       Configs         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// ConfigForm 系统配置表单
type ConfigForm struct {
	ID          uint64 `json:"id"`
	ConfigName  string `json:"configName" validate:"required,max=50"`
	ConfigKey   string `json:"configKey" validate:"required,max=50"`
	ConfigValue string `json:"configValue" validate:"required,max=100"`
	Remark      string `json:"remark" validate:"max=255"`
}

// ConfigVO 系统配置视图对象
type ConfigVO struct {
	ID          uint64       `json:"id"`
	ConfigName  string       `json:"configName"`
	ConfigKey   string       `json:"configKey"`
	ConfigValue string       `json:"configValue"`
	Remark      string       `json:"remark"`
	CreateTime  dto.DateTime `json:"createTime"`
}
