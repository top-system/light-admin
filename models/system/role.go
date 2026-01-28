package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Role 角色模型
// Status: 1-正常 0-停用
// DataScope: 数据权限范围
type Role struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string       `gorm:"column:name;size:64;not null;uniqueIndex:uk_role_name" json:"name"`
	Code       string       `gorm:"column:code;size:32;not null;uniqueIndex:uk_role_code" json:"code"`
	Sort       int          `gorm:"column:sort" json:"sort"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	DataScope  int          `gorm:"column:data_scope" json:"dataScope"`
	CreateBy   uint64       `gorm:"column:create_by" json:"createBy"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   uint64       `gorm:"column:update_by" json:"updateBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`

	// 非数据库字段
	MenuIds []uint64 `gorm:"-" json:"menuIds,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "t_role"
}

type Roles []*Role

type RoleQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	IDs        []uint64 `query:"ids"`
	Name       string   `query:"name"`
	Code       string   `query:"code"`
	QueryValue string   `query:"query_value"`
	UserID     uint64   `query:"user_id"`
	Status     int      `query:"status"`
}

type RoleQueryResult struct {
	List       Roles           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a Roles) ToNames() []string {
	names := make([]string, len(a))
	for i, item := range a {
		names[i] = item.Name
	}
	return names
}

func (a Roles) ToCodes() []string {
	codes := make([]string, len(a))
	for i, item := range a {
		codes[i] = item.Code
	}
	return codes
}

func (a Roles) ToMap() map[uint64]*Role {
	m := make(map[uint64]*Role)
	for _, item := range a {
		m[item.ID] = item
	}
	return m
}

func (a Roles) ToIDs() []uint64 {
	ids := make([]uint64, len(a))
	for i, item := range a {
		ids[i] = item.ID
	}
	return ids
}

// RoleOption 角色下拉选项
type RoleOption struct {
	Value uint64 `json:"value"`
	Label string `json:"label"`
}

func (a Roles) ToOptions() []RoleOption {
	options := make([]RoleOption, len(a))
	for i, item := range a {
		options[i] = RoleOption{
			Value: item.ID,
			Label: item.Name,
		}
	}
	return options
}
