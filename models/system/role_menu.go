package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// RoleMenu 角色菜单关联模型
type RoleMenu struct {
	RoleID uint64 `gorm:"column:role_id;not null;uniqueIndex:uk_roleid_menuid" json:"roleId"`
	MenuID uint64 `gorm:"column:menu_id;not null;uniqueIndex:uk_roleid_menuid" json:"menuId"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "t_role_menu"
}

type RoleMenus []*RoleMenu

type RoleMenuQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	RoleID  uint64
	RoleIDs []uint64
}

type RoleMenuQueryResult struct {
	List       RoleMenus       `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a RoleMenus) ToMap() map[uint64]*RoleMenu {
	m := make(map[uint64]*RoleMenu)
	for _, item := range a {
		m[item.MenuID] = item
	}
	return m
}

func (a RoleMenus) ToRoleIDMap() map[uint64]RoleMenus {
	m := make(map[uint64]RoleMenus)
	for _, item := range a {
		m[item.RoleID] = append(m[item.RoleID], item)
	}
	return m
}

func (a RoleMenus) ToMenuIDs() []uint64 {
	var idList []uint64
	m := make(map[uint64]struct{})

	for _, item := range a {
		if _, ok := m[item.MenuID]; ok {
			continue
		}
		idList = append(idList, item.MenuID)
		m[item.MenuID] = struct{}{}
	}

	return idList
}
