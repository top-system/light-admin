package system

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/top-system/light-admin/models/dto"
)

// Menu 菜单模型
// Type: 1-菜单 2-目录 3-外链 4-按钮
// Visible: 1-显示 0-隐藏
type Menu struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID   uint64       `gorm:"column:parent_id;not null;default:0;index" json:"parentId"`
	TreePath   string       `gorm:"column:tree_path;size:255;index:idx_tree_path" json:"treePath"`
	Name       string       `gorm:"column:name;size:64;not null" json:"name"`
	Type       int          `gorm:"column:type;not null;index:idx_type" json:"type"`
	RouteName  string       `gorm:"column:route_name;size:255" json:"routeName"`
	RoutePath  string       `gorm:"column:route_path;size:128" json:"routePath"`
	Component  string       `gorm:"column:component;size:128" json:"component"`
	Perm       string       `gorm:"column:perm;size:128" json:"perm"`
	AlwaysShow int          `gorm:"column:always_show;default:0" json:"alwaysShow"`
	KeepAlive  int          `gorm:"column:keep_alive;default:0" json:"keepAlive"`
	Visible    int          `gorm:"column:visible;default:1" json:"visible"`
	Sort       int          `gorm:"column:sort;default:0" json:"sort"`
	Icon       string       `gorm:"column:icon;size:64" json:"icon"`
	Redirect   string       `gorm:"column:redirect;size:128" json:"redirect"`
	Params     string       `gorm:"column:params;size:255" json:"params"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "t_menu"
}

// MenuTree 菜单树结构(用于展示和YAML解析)
type MenuTree struct {
	ID         uint64       `yaml:"-" json:"id"`
	ParentID   uint64       `yaml:"-" json:"parentId"`
	TreePath   string       `yaml:"-" json:"treePath,omitempty"`
	Name       string       `yaml:"name" json:"name"`
	Type       dto.MenuType `yaml:"type" json:"type"`
	RouteName  string       `yaml:"route_name,omitempty" json:"routeName,omitempty"`
	RoutePath  string       `yaml:"route_path,omitempty" json:"routePath,omitempty"`
	Component  string       `yaml:"component,omitempty" json:"component,omitempty"`
	Perm       string       `yaml:"perm,omitempty" json:"perm,omitempty"`
	AlwaysShow int          `yaml:"always_show,omitempty" json:"alwaysShow,omitempty"`
	KeepAlive  int          `yaml:"keep_alive,omitempty" json:"keepAlive,omitempty"`
	Visible    *int         `yaml:"visible" json:"visible"`
	Sort       int          `yaml:"sort" json:"sort"`
	Icon       string       `yaml:"icon,omitempty" json:"icon,omitempty"`
	Redirect   string       `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	Params     string       `yaml:"params,omitempty" json:"params,omitempty"`
	Children   MenuTrees    `yaml:"children,omitempty" json:"children,omitempty"`
}

type Menus []*Menu
type MenuTrees []*MenuTree

type MenuQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	IDs            []uint64 `query:"ids"`
	Name           string   `query:"name"`
	PrefixTreePath string   `query:"prefix_tree_path"`
	Keywords       string   `query:"keywords"`
	ParentID       *uint64  `query:"parent_id"`
	Type           int      `query:"type"`
	Visible        int      `query:"visible"`
	Tree           bool     `query:"tree"`
}

type MenuQueryResult struct {
	List       Menus           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// MenuForm 菜单表单（用于创建和更新）
type MenuForm struct {
	ID         uint64         `json:"id"`
	ParentID   dto.FlexUint64 `json:"parentId"`
	Name       string         `json:"name"`
	Type       dto.MenuType   `json:"type"`
	RouteName  string         `json:"routeName"`
	RoutePath  string         `json:"routePath"`
	Component  string         `json:"component"`
	Perm       string         `json:"perm"`
	AlwaysShow dto.FlexInt    `json:"alwaysShow"`
	KeepAlive  dto.FlexInt    `json:"keepAlive"`
	Visible    dto.FlexInt    `json:"visible"`
	Sort       dto.FlexInt    `json:"sort"`
	Icon       string         `json:"icon"`
	Redirect   string         `json:"redirect"`
	Params     string         `json:"params"`
}

// ToMenu 将 MenuForm 转换为 Menu 模型
func (f *MenuForm) ToMenu() *Menu {
	return &Menu{
		ID:         f.ID,
		ParentID:   f.ParentID.Value(),
		Name:       f.Name,
		Type:       f.Type.Value(),
		RouteName:  f.RouteName,
		RoutePath:  f.RoutePath,
		Component:  f.Component,
		Perm:       f.Perm,
		AlwaysShow: f.AlwaysShow.Value(),
		KeepAlive:  f.KeepAlive.Value(),
		Visible:    f.Visible.Value(),
		Sort:       f.Sort.Value(),
		Icon:       f.Icon,
		Redirect:   f.Redirect,
		Params:     f.Params,
	}
}

// ToForm 将 Menu 模型转换为 MenuForm
func (m *Menu) ToForm() *MenuForm {
	return &MenuForm{
		ID:         m.ID,
		ParentID:   dto.FlexUint64(m.ParentID),
		Name:       m.Name,
		Type:       dto.MenuType(m.Type),
		RouteName:  m.RouteName,
		RoutePath:  m.RoutePath,
		Component:  m.Component,
		Perm:       m.Perm,
		AlwaysShow: dto.FlexInt(m.AlwaysShow),
		KeepAlive:  dto.FlexInt(m.KeepAlive),
		Visible:    dto.FlexInt(m.Visible),
		Sort:       dto.FlexInt(m.Sort),
		Icon:       m.Icon,
		Redirect:   m.Redirect,
		Params:     m.Params,
	}
}

func (a Menus) Len() int {
	return len(a)
}

func (a Menus) Less(i, j int) bool {
	return a[i].Sort < a[j].Sort
}

func (a Menus) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Menus) ToMap() map[uint64]*Menu {
	m := make(map[uint64]*Menu)
	for _, menu := range a {
		m[menu.ID] = menu
	}
	return m
}

func (a Menus) SplitParentIDs() []uint64 {
	idList := make([]uint64, 0, len(a))
	mIDList := make(map[uint64]struct{})

	for _, item := range a {
		if _, ok := mIDList[item.ID]; ok || item.TreePath == "" {
			continue
		}

		for _, pp := range strings.Split(item.TreePath, ",") {
			if pp == "" {
				continue
			}
			id, err := strconv.ParseUint(pp, 10, 64)
			if err != nil {
				continue
			}
			if _, ok := mIDList[id]; ok {
				continue
			}

			idList = append(idList, id)
			mIDList[id] = struct{}{}
		}
	}

	return idList
}

func (a Menus) ToIDs() []uint64 {
	ids := make([]uint64, len(a))
	for i, item := range a {
		ids[i] = item.ID
	}
	return ids
}

func (a Menus) ToMenuTrees() MenuTrees {
	menuTrees := make(MenuTrees, len(a))
	for i, menu := range a {
		visible := menu.Visible
		menuTrees[i] = &MenuTree{
			ID:         menu.ID,
			ParentID:   menu.ParentID,
			TreePath:   menu.TreePath,
			Name:       menu.Name,
			Type:       dto.MenuType(menu.Type),
			RouteName:  menu.RouteName,
			RoutePath:  menu.RoutePath,
			Component:  menu.Component,
			Perm:       menu.Perm,
			AlwaysShow: menu.AlwaysShow,
			KeepAlive:  menu.KeepAlive,
			Visible:    &visible,
			Sort:       menu.Sort,
			Icon:       menu.Icon,
			Redirect:   menu.Redirect,
			Params:     menu.Params,
		}
	}

	return menuTrees.ToTree()
}

// ToFlatMenuTrees 转换为扁平的菜单树列表（用于搜索结果）
func (a Menus) ToFlatMenuTrees() MenuTrees {
	menuTrees := make(MenuTrees, len(a))
	for i, menu := range a {
		visible := menu.Visible
		menuTrees[i] = &MenuTree{
			ID:         menu.ID,
			ParentID:   menu.ParentID,
			TreePath:   menu.TreePath,
			Name:       menu.Name,
			Type:       dto.MenuType(menu.Type),
			RouteName:  menu.RouteName,
			RoutePath:  menu.RoutePath,
			Component:  menu.Component,
			Perm:       menu.Perm,
			AlwaysShow: menu.AlwaysShow,
			KeepAlive:  menu.KeepAlive,
			Visible:    &visible,
			Sort:       menu.Sort,
			Icon:       menu.Icon,
			Redirect:   menu.Redirect,
			Params:     menu.Params,
		}
	}
	return menuTrees
}

func (a MenuTrees) ToTree() MenuTrees {
	menuTreeMap := make(map[uint64]*MenuTree)
	for _, menuTree := range a {
		menuTreeMap[menuTree.ID] = menuTree
	}

	menuTrees := make(MenuTrees, 0)
	for _, menuTree := range a {
		if menuTree.ParentID == 0 {
			menuTrees = append(menuTrees, menuTree)
			continue
		}

		if parentMenuTree, ok := menuTreeMap[menuTree.ParentID]; ok {
			if parentMenuTree.Children == nil {
				children := MenuTrees{menuTree}
				parentMenuTree.Children = children
				continue
			}

			parentMenuTree.Children = append(parentMenuTree.Children, menuTree)
		}
	}

	return menuTrees
}

// GetParamsMap 获取路由参数Map
func (a *Menu) GetParamsMap() map[string]string {
	if a.Params == "" {
		return nil
	}
	var params map[string]string
	if err := json.Unmarshal([]byte(a.Params), &params); err != nil {
		return nil
	}
	return params
}
