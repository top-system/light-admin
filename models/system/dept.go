package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Dept 部门模型
// Status: 状态（1: 正常, 0: 禁用）
type Dept struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string       `gorm:"column:name;size:100;not null" json:"name"`
	Code       string       `gorm:"column:code;size:100;not null;uniqueIndex:uk_code" json:"code"`
	ParentID   uint64       `gorm:"column:parent_id;default:0" json:"parentId"`
	TreePath   string       `gorm:"column:tree_path;size:255;not null" json:"treePath"`
	Sort       int          `gorm:"column:sort;default:0" json:"sort"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	CreateBy   uint64       `gorm:"column:create_by" json:"createBy"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   uint64       `gorm:"column:update_by" json:"updateBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (Dept) TableName() string {
	return "t_dept"
}

type Depts []*Dept

type DeptQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Keywords string `query:"keywords"`
	Status   *int   `query:"status"`
}

type DeptQueryResult struct {
	List       Depts           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// DeptForm 部门表单
type DeptForm struct {
	ID       uint64         `json:"id"`
	Name     string         `json:"name" validate:"required,max=100"`
	Code     string         `json:"code" validate:"required,max=100"`
	ParentID dto.FlexUint64 `json:"parentId"`
	Sort     int            `json:"sort"`
	Status   int            `json:"status"`
}

// DeptVO 部门视图对象
type DeptVO struct {
	ID         uint64       `json:"id"`
	Name       string       `json:"name"`
	Code       string       `json:"code"`
	ParentID   uint64       `json:"parentId"`
	Sort       int          `json:"sort"`
	Status     int          `json:"status"`
	CreateTime dto.DateTime `json:"createTime"`
	UpdateTime dto.DateTime `json:"updateTime"`
	Children   []*DeptVO    `json:"children,omitempty"`
}

// DeptOption 部门下拉选项
type DeptOption struct {
	Value    uint64        `json:"value"`
	Label    string        `json:"label"`
	Children []*DeptOption `json:"children,omitempty"`
}
