package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// UserNotice 用户通知公告模型
// IsRead: 读取状态（0: 未读, 1: 已读）
type UserNotice struct {
	ID         uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	NoticeID   uint64           `gorm:"column:notice_id;not null" json:"noticeId"`
	UserID     uint64           `gorm:"column:user_id;not null" json:"userId"`
	IsRead     int              `gorm:"column:is_read;default:0" json:"isRead"`
	ReadTime   dto.NullDateTime `gorm:"column:read_time" json:"readTime"`
	CreateTime dto.DateTime     `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime dto.DateTime     `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted  int              `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

// TableName 指定表名
func (UserNotice) TableName() string {
	return "t_user_notice"
}

type UserNotices []*UserNotice

type UserNoticeQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	NoticeID uint64 `query:"noticeId"`
	UserID   uint64 `query:"userId"`
	IsRead   *int   `query:"isRead"`
}

type UserNoticeQueryResult struct {
	List       UserNotices     `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// UserNoticePageVO 我的通知公告分页视图对象
type UserNoticePageVO struct {
	ID          uint64           `json:"id"`
	NoticeID    uint64           `json:"noticeId"`
	Title       string           `json:"title"`
	Type        int              `json:"type"`
	Level       string           `json:"level"`
	PublishTime dto.NullDateTime `json:"publishTime"`
	IsRead      int              `json:"isRead"`
}
