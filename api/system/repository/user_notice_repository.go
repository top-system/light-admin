package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// UserNoticeRepository database structure
type UserNoticeRepository struct {
	db       lib.Database
	logger   lib.Logger
	dbCompat lib.DBCompat
}

// NewUserNoticeRepository creates a new user notice repository
func NewUserNoticeRepository(db lib.Database, logger lib.Logger, dbCompat lib.DBCompat) UserNoticeRepository {
	return UserNoticeRepository{
		db:       db,
		logger:   logger,
		dbCompat: dbCompat,
	}
}

// WithTrx enables repository with transaction
func (a UserNoticeRepository) WithTrx(trxHandle *gorm.DB) UserNoticeRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

// GetMyNoticePage 获取我的通知公告分页列表
func (a UserNoticeRepository) GetMyNoticePage(param *system.NoticeQueryParam) ([]system.UserNoticePageVO, int64, error) {
	var list []system.UserNoticePageVO
	var total int64

	db := a.db.ORM.Table("t_user_notice un").
		Select("un.id, un.notice_id, n.title, n.type, n.level, n.publish_time, un.is_read").
		Joins("LEFT JOIN t_notice n ON un.notice_id = n.id").
		Where("un.user_id = ?", param.UserID).
		Where("un.is_deleted = ?", 0).
		Where("n.is_deleted = ?", 0).
		Where("n.publish_status = ?", 1)

	if v := param.Title; v != "" {
		db = db.Where("n.title LIKE ?", "%"+v+"%")
	}

	if v := param.Type; v != 0 {
		db = db.Where("n.type = ?", v)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	// Get page data
	offset := (param.PageNum - 1) * param.PageSize
	if err := db.Order("n.publish_time DESC").Offset(offset).Limit(param.PageSize).Scan(&list).Error; err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	return list, total, nil
}

func (a UserNoticeRepository) Create(userNotice *system.UserNotice) error {
	result := a.db.ORM.Create(userNotice)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserNoticeRepository) BatchCreate(userNotices []*system.UserNotice) error {
	if len(userNotices) == 0 {
		return nil
	}
	result := a.db.ORM.Create(&userNotices)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserNoticeRepository) MarkAsRead(noticeID, userID uint64) error {
	result := a.db.ORM.Model(&system.UserNotice{}).
		Where("notice_id = ? AND user_id = ? AND is_read = ?", noticeID, userID, 0).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": a.dbCompat.Now(),
		})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserNoticeRepository) MarkAllAsRead(userID uint64) error {
	result := a.db.ORM.Model(&system.UserNotice{}).
		Where("user_id = ? AND is_read = ?", userID, 0).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": a.dbCompat.Now(),
		})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserNoticeRepository) DeleteByNoticeID(noticeID uint64) error {
	result := a.db.ORM.Where("notice_id = ?", noticeID).Delete(&system.UserNotice{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a UserNoticeRepository) DeleteByNoticeIDs(noticeIDs []uint64) error {
	result := a.db.ORM.Where("notice_id IN ?", noticeIDs).Delete(&system.UserNotice{})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
