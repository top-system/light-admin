package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Notice 通知公告模型
// Type: 通知类型（关联字典编码：notice_type）
// Level: 通知等级（字典code：notice_level）L-低 M-中 H-高
// TargetType: 目标类型（1: 全体, 2: 指定）
// PublishStatus: 发布状态（0: 未发布, 1: 已发布, -1: 已撤回）
type Notice struct {
	ID            uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title         string           `gorm:"column:title;size:50" json:"title"`
	Content       string           `gorm:"column:content;type:text" json:"content"`
	Type          int              `gorm:"column:type;not null" json:"type"`
	Level         string           `gorm:"column:level;size:5;not null" json:"level"`
	TargetType    int              `gorm:"column:target_type;not null" json:"targetType"`
	TargetUserIds string           `gorm:"column:target_user_ids;size:255" json:"targetUserIds"`
	PublisherId   uint64           `gorm:"column:publisher_id" json:"publisherId"`
	PublishStatus int              `gorm:"column:publish_status;default:0;index:idx_publish_status" json:"publishStatus"`
	PublishTime   dto.NullDateTime `gorm:"column:publish_time" json:"publishTime"`
	RevokeTime    dto.NullDateTime `gorm:"column:revoke_time" json:"revokeTime"`
	CreateBy      uint64           `gorm:"column:create_by;not null" json:"createBy"`
	CreateTime    dto.DateTime     `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy      uint64           `gorm:"column:update_by" json:"updateBy"`
	UpdateTime    dto.DateTime     `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted     int              `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (Notice) TableName() string {
	return "t_notice"
}

type Notices []*Notice

type NoticeQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Title         string `query:"title"`
	Type          int    `query:"type"`
	PublishStatus *int   `query:"publishStatus"`
	UserID        uint64 `query:"-"` // 用于查询我的通知
}

type NoticeQueryResult struct {
	List       Notices         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// NoticeForm 通知公告表单
type NoticeForm struct {
	ID            uint64      `json:"id"`
	Title         string      `json:"title" validate:"required,max=50"`
	Content       string      `json:"content"`
	Type          dto.FlexInt `json:"type" validate:"required"`
	Level         string      `json:"level" validate:"required"`
	TargetType    int         `json:"targetType" validate:"required"`
	TargetUserIds []string    `json:"targetUserIds"`
}

// NoticePageVO 通知公告分页视图对象
type NoticePageVO struct {
	ID            uint64           `json:"id"`
	Title         string           `json:"title"`
	Type          int              `json:"type"`
	Level         string           `json:"level"`
	TargetType    int              `json:"targetType"`
	PublishStatus int              `json:"publishStatus"`
	PublishTime   dto.NullDateTime `json:"publishTime"`
	PublisherName string           `json:"publisherName"`
	CreateTime    dto.DateTime     `json:"createTime"`
}

// NoticeDetailVO 通知公告详情视图对象
type NoticeDetailVO struct {
	ID            uint64           `json:"id"`
	Title         string           `json:"title"`
	Content       string           `json:"content"`
	Type          int              `json:"type"`
	Level         string           `json:"level"`
	PublisherId   uint64           `json:"publisherId"`
	PublisherName string           `json:"publisherName"`
	PublishTime   dto.NullDateTime `json:"publishTime"`
}
