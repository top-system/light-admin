package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// UserRole 用户角色关联模型
type UserRole struct {
	UserID uint64 `gorm:"column:user_id;primaryKey" json:"userId"`
	RoleID uint64 `gorm:"column:role_id;primaryKey" json:"roleId"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "t_user_role"
}

type UserRoles []*UserRole

type UserRoleQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	UserID  uint64
	UserIDs []uint64
}

type UserRoleQueryResult struct {
	List       UserRoles       `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a UserRoles) ToMap() map[uint64]*UserRole {
	m := make(map[uint64]*UserRole)
	for _, item := range a {
		m[item.RoleID] = item
	}
	return m
}

func (a UserRoles) ToRoleIDs() []uint64 {
	list := make([]uint64, len(a))
	for i, item := range a {
		list[i] = item.RoleID
	}
	return list
}

func (a UserRoles) ToUserIDMap() map[uint64]UserRoles {
	m := make(map[uint64]UserRoles)
	for _, item := range a {
		m[item.UserID] = append(m[item.UserID], item)
	}
	return m
}
