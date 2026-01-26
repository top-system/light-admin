package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// DictItem 字典项模型
type DictItem struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode   string       `gorm:"column:dict_code;size:100;not null;index:idx_dict_code" json:"dictCode"`
	Label      string       `gorm:"column:label;size:100;not null" json:"label"`
	Value      string       `gorm:"column:value;size:100;not null" json:"value"`
	TagType    string       `gorm:"column:tag_type;size:50" json:"tagType"`
	Sort       int          `gorm:"column:sort;default:0" json:"sort"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	Remark     string       `gorm:"column:remark;size:255" json:"remark"`
	CreateBy   uint64       `gorm:"column:create_by" json:"createBy"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   uint64       `gorm:"column:update_by" json:"updateBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (DictItem) TableName() string {
	return "t_dict_item"
}

type DictItems []*DictItem

type DictItemQueryParam struct {
	dto.PaginationParam
	DictCode string `query:"dictCode"`
	Keywords string `query:"keywords"`
}

type DictItemQueryResult struct {
	List       DictItems       `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// DictItemForm 字典项表单
type DictItemForm struct {
	ID       uint64 `json:"id"`
	DictCode string `json:"dictCode"`
	Label    string `json:"label" validate:"required,max=100"`
	Value    string `json:"value" validate:"required,max=100"`
	TagType  string `json:"tagType" validate:"max=50"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
	Remark   string `json:"remark" validate:"max=255"`
}

// DictItemPageVO 字典项分页视图对象
type DictItemPageVO struct {
	ID         uint64       `json:"id"`
	DictCode   string       `json:"dictCode"`
	Label      string       `json:"label"`
	Value      string       `json:"value"`
	TagType    string       `json:"tagType"`
	Sort       int          `json:"sort"`
	Status     int          `json:"status"`
	Remark     string       `json:"remark"`
	CreateTime dto.DateTime `json:"createTime"`
}

// DictItemOptionVO 字典项下拉选项
type DictItemOptionVO struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	TagType string `json:"tagType,omitempty"`
}

// ToPageVOList 转换为分页视图对象列表
func (list DictItems) ToPageVOList() []*DictItemPageVO {
	result := make([]*DictItemPageVO, 0, len(list))
	for _, item := range list {
		result = append(result, &DictItemPageVO{
			ID:         item.ID,
			DictCode:   item.DictCode,
			Label:      item.Label,
			Value:      item.Value,
			TagType:    item.TagType,
			Sort:       item.Sort,
			Status:     item.Status,
			Remark:     item.Remark,
			CreateTime: item.CreateTime,
		})
	}
	return result
}

// ToOptionList 转换为下拉选项列表
func (list DictItems) ToOptionList() []*DictItemOptionVO {
	result := make([]*DictItemOptionVO, 0, len(list))
	for _, item := range list {
		result = append(result, &DictItemOptionVO{
			Label:   item.Label,
			Value:   item.Value,
			TagType: item.TagType,
		})
	}
	return result
}
