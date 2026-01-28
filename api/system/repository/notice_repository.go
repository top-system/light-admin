package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// NoticeRepository database structure
type NoticeRepository struct {
	db       lib.Database
	logger   lib.Logger
	dbCompat lib.DBCompat
}

// NewNoticeRepository creates a new notice repository
func NewNoticeRepository(db lib.Database, logger lib.Logger, dbCompat lib.DBCompat) NoticeRepository {
	return NoticeRepository{
		db:       db,
		logger:   logger,
		dbCompat: dbCompat,
	}
}

// WithTrx enables repository with transaction
func (a NoticeRepository) WithTrx(trxHandle *gorm.DB) NoticeRepository {
	if trxHandle == nil {
		a.logger.Zap.Error("Transaction Database not found in echo context.")
		return a
	}

	a.db.ORM = trxHandle
	return a
}

func (a NoticeRepository) Query(param *system.NoticeQueryParam) (*system.NoticeQueryResult, error) {
	db := a.db.ORM.Model(&system.Notice{}).Where("is_deleted = ?", 0)

	if v := param.Title; v != "" {
		db = db.Where("title LIKE ?", "%"+v+"%")
	}

	if v := param.Type; v != 0 {
		db = db.Where("type = ?", v)
	}

	if v := param.PublishStatus; v != nil {
		db = db.Where("publish_status = ?", *v)
	}

	db = db.Order("create_time DESC")

	list := make(system.Notices, 0)
	pagination, err := QueryPagination(db, param.PaginationParam, &list)
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	qr := &system.NoticeQueryResult{
		Pagination: pagination,
		List:       list,
	}

	return qr, nil
}

func (a NoticeRepository) Get(id uint64) (*system.Notice, error) {
	notice := new(system.Notice)

	if ok, err := QueryOne(a.db.ORM.Model(notice).Where("id=? AND is_deleted=?", id, 0), notice); err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	} else if !ok {
		return nil, errors.DatabaseRecordNotFound
	}

	return notice, nil
}

func (a NoticeRepository) Create(notice *system.Notice) error {
	result := a.db.ORM.Model(notice).Create(notice)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a NoticeRepository) Update(id uint64, notice *system.Notice) error {
	result := a.db.ORM.Model(notice).Where("id=?", id).Select(
		"title", "content", "type", "level", "target_type",
		"target_user_ids", "update_by",
	).Updates(notice)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a NoticeRepository) UpdateStatus(id uint64, status int, publisherId uint64) error {
	updates := map[string]interface{}{
		"publish_status": status,
		"publisher_id":   publisherId,
	}
	if status == 1 {
		updates["publish_time"] = a.dbCompat.Now()
	} else if status == -1 {
		updates["revoke_time"] = a.dbCompat.Now()
	}

	result := a.db.ORM.Model(&system.Notice{}).Where("id=?", id).Updates(updates)
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a NoticeRepository) Delete(id uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Notice{}).Where("id=?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}

func (a NoticeRepository) BatchDelete(ids []uint64, deletedBy uint64) error {
	result := a.db.ORM.Model(&system.Notice{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_deleted": 1,
		"update_by":  deletedBy,
	})
	if result.Error != nil {
		return errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}

	return nil
}
