package service

import (
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
)

// NoticeService service layer
type NoticeService struct {
	logger               lib.Logger
	noticeRepository     repository.NoticeRepository
	userNoticeRepository repository.UserNoticeRepository
	userRepository       repository.UserRepository
}

// NewNoticeService creates a new notice service
func NewNoticeService(
	logger lib.Logger,
	noticeRepository repository.NoticeRepository,
	userNoticeRepository repository.UserNoticeRepository,
	userRepository repository.UserRepository,
) NoticeService {
	return NoticeService{
		logger:               logger,
		noticeRepository:     noticeRepository,
		userNoticeRepository: userNoticeRepository,
		userRepository:       userRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a NoticeService) WithTrx(trxHandle *gorm.DB) NoticeService {
	a.noticeRepository = a.noticeRepository.WithTrx(trxHandle)
	a.userNoticeRepository = a.userNoticeRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询通知公告
func (a NoticeService) Query(param *system.NoticeQueryParam) (*system.NoticeQueryResult, error) {
	return a.noticeRepository.Query(param)
}

// Get 获取通知公告
func (a NoticeService) Get(id uint64) (*system.Notice, error) {
	return a.noticeRepository.Get(id)
}

// GetForm 获取通知公告表单数据
func (a NoticeService) GetForm(id uint64) (*system.NoticeForm, error) {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return nil, err
	}

	var targetUserIds []string
	if notice.TargetUserIds != "" {
		targetUserIds = strings.Split(notice.TargetUserIds, ",")
	}

	return &system.NoticeForm{
		ID:            notice.ID,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          dto.FlexInt(notice.Type),
		Level:         notice.Level,
		TargetType:    notice.TargetType,
		TargetUserIds: targetUserIds,
	}, nil
}

// GetDetail 获取通知公告详情并标记为已读
func (a NoticeService) GetDetail(id uint64, userID uint64) (*system.NoticeDetailVO, error) {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return nil, err
	}

	// 标记为已读
	_ = a.userNoticeRepository.MarkAsRead(id, userID)

	// 获取发布人信息
	var publisherName string
	if notice.PublisherId > 0 {
		publisher, err := a.userRepository.Get(notice.PublisherId)
		if err == nil && publisher != nil {
			publisherName = publisher.Nickname
		}
	}

	return &system.NoticeDetailVO{
		ID:            notice.ID,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          notice.Type,
		Level:         notice.Level,
		PublisherId:   notice.PublisherId,
		PublisherName: publisherName,
		PublishTime:   notice.PublishTime,
	}, nil
}

// Create 创建通知公告
func (a NoticeService) Create(form *system.NoticeForm, createdBy uint64) error {
	// 如果目标类型是指定用户，则必须填写目标用户
	if form.TargetType == 2 && len(form.TargetUserIds) == 0 {
		return errors.New("推送指定用户不能为空")
	}

	notice := &system.Notice{
		Title:         form.Title,
		Content:       form.Content,
		Type:          form.Type.Value(),
		Level:         form.Level,
		TargetType:    form.TargetType,
		TargetUserIds: strings.Join(form.TargetUserIds, ","),
		PublishStatus: 0, // 未发布
		CreateBy:      createdBy,
		IsDeleted:     0,
	}

	return a.noticeRepository.Create(notice)
}

// Update 更新通知公告
func (a NoticeService) Update(id uint64, form *system.NoticeForm, updatedBy uint64) error {
	// 检查通知是否存在
	_, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	// 如果目标类型是指定用户，则必须填写目标用户
	if form.TargetType == 2 && len(form.TargetUserIds) == 0 {
		return errors.New("推送指定用户不能为空")
	}

	notice := &system.Notice{
		ID:            id,
		Title:         form.Title,
		Content:       form.Content,
		Type:          form.Type.Value(),
		Level:         form.Level,
		TargetType:    form.TargetType,
		TargetUserIds: strings.Join(form.TargetUserIds, ","),
		UpdateBy:      updatedBy,
	}

	return a.noticeRepository.Update(id, notice)
}

// Delete 删除通知公告
func (a NoticeService) Delete(ids string, deletedBy uint64) error {
	if ids == "" {
		return errors.New("删除的通知公告数据为空")
	}

	idStrs := strings.Split(ids, ",")
	idList := make([]uint64, 0, len(idStrs))
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}
		idList = append(idList, id)
	}

	if len(idList) == 0 {
		return errors.New("删除的通知公告数据为空")
	}

	// 删除通知公告
	if err := a.noticeRepository.BatchDelete(idList, deletedBy); err != nil {
		return err
	}

	// 删除用户通知状态
	return a.userNoticeRepository.DeleteByNoticeIDs(idList)
}

// Publish 发布通知公告
func (a NoticeService) Publish(id uint64, publisherId uint64) error {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	if notice.PublishStatus == 1 {
		return errors.New("通知公告已发布")
	}

	if notice.TargetType == 2 && notice.TargetUserIds == "" {
		return errors.New("推送指定用户不能为空")
	}

	// 更新发布状态
	if err := a.noticeRepository.UpdateStatus(id, 1, publisherId); err != nil {
		return err
	}

	// 删除该通告之前的用户通知数据（可能是重新发布）
	_ = a.userNoticeRepository.DeleteByNoticeID(id)

	// 获取目标用户列表
	var targetUsers system.Users
	if notice.TargetType == 1 {
		// 全体用户
		userQR, err := a.userRepository.Query(&system.UserQueryParam{})
		if err != nil {
			return err
		}
		targetUsers = userQR.List
	} else {
		// 指定用户
		targetUserIds := strings.Split(notice.TargetUserIds, ",")
		userIds := make([]uint64, 0, len(targetUserIds))
		for _, idStr := range targetUserIds {
			id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				continue
			}
			userIds = append(userIds, id)
		}
		// 这里简化处理，实际应该根据userIds查询用户
		userQR, err := a.userRepository.Query(&system.UserQueryParam{})
		if err != nil {
			return err
		}
		for _, user := range userQR.List {
			for _, uid := range userIds {
				if user.ID == uid {
					targetUsers = append(targetUsers, user)
					break
				}
			}
		}
	}

	// 创建用户通知记录
	userNotices := make([]*system.UserNotice, 0, len(targetUsers))
	for _, user := range targetUsers {
		userNotices = append(userNotices, &system.UserNotice{
			NoticeID: id,
			UserID:   user.ID,
			IsRead:   0,
		})
	}

	if len(userNotices) > 0 {
		return a.userNoticeRepository.BatchCreate(userNotices)
	}

	return nil
}

// Revoke 撤回通知公告
func (a NoticeService) Revoke(id uint64, updatedBy uint64) error {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	if notice.PublishStatus != 1 {
		return errors.New("通知公告未发布或已撤回")
	}

	// 更新撤回状态
	if err := a.noticeRepository.UpdateStatus(id, -1, updatedBy); err != nil {
		return err
	}

	// 删除用户通知状态
	return a.userNoticeRepository.DeleteByNoticeID(id)
}

// GetMyNoticePage 获取我的通知公告分页列表
func (a NoticeService) GetMyNoticePage(param *system.NoticeQueryParam) ([]system.UserNoticePageVO, int64, error) {
	return a.userNoticeRepository.GetMyNoticePage(param)
}

// ReadAll 全部标记为已读
func (a NoticeService) ReadAll(userID uint64) error {
	return a.userNoticeRepository.MarkAllAsRead(userID)
}
