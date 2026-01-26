package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Dict 字典模型
type Dict struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode   string       `gorm:"column:dict_code;size:100;not null;uniqueIndex:uk_dict_code" json:"dictCode"`
	Name       string       `gorm:"column:name;size:100;not null" json:"name"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	Remark     string       `gorm:"column:remark;size:255" json:"remark"`
	CreateBy   uint64       `gorm:"column:create_by" json:"createBy"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   uint64       `gorm:"column:update_by" json:"updateBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (Dict) TableName() string {
	return "t_dict"
}

type Dicts []*Dict

type DictQueryParam struct {
	dto.PaginationParam
	Keywords string `query:"keywords"`
}

type DictQueryResult struct {
	List       Dicts           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// DictForm 字典表单
type DictForm struct {
	ID       uint64 `json:"id"`
	DictCode string `json:"dictCode" validate:"required,max=100"`
	Name     string `json:"name" validate:"required,max=100"`
	Status   int    `json:"status"`
	Remark   string `json:"remark" validate:"max=255"`
}

// DictPageVO 字典分页视图对象
type DictPageVO struct {
	ID         uint64       `json:"id"`
	DictCode   string       `json:"dictCode"`
	Name       string       `json:"name"`
	Status     int          `json:"status"`
	Remark     string       `json:"remark"`
	CreateTime dto.DateTime `json:"createTime"`
}

// DictOption 字典下拉选项
type DictOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ToOptions 转换为下拉选项
func (list Dicts) ToOptions() []*DictOption {
	result := make([]*DictOption, 0, len(list))
	for _, item := range list {
		result = append(result, &DictOption{
			Value: item.DictCode,
			Label: item.Name,
		})
	}
	return result
}

// ToPageVOList 转换为分页视图对象列表
func (list Dicts) ToPageVOList() []*DictPageVO {
	result := make([]*DictPageVO, 0, len(list))
	for _, item := range list {
		result = append(result, &DictPageVO{
			ID:         item.ID,
			DictCode:   item.DictCode,
			Name:       item.Name,
			Status:     item.Status,
			Remark:     item.Remark,
			CreateTime: item.CreateTime,
		})
	}
	return result
}
