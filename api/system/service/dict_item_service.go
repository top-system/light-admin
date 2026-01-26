package service

import (
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DictItemService service layer
type DictItemService struct {
	logger             lib.Logger
	dictItemRepository repository.DictItemRepository
}

// NewDictItemService creates a new dict item service
func NewDictItemService(
	logger lib.Logger,
	dictItemRepository repository.DictItemRepository,
) DictItemService {
	return DictItemService{
		logger:             logger,
		dictItemRepository: dictItemRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a DictItemService) WithTrx(trxHandle *gorm.DB) DictItemService {
	a.dictItemRepository = a.dictItemRepository.WithTrx(trxHandle)
	return a
}

// GetDictItemPage 获取字典项分页列表
func (a DictItemService) GetDictItemPage(param *system.DictItemQueryParam) (*system.DictItemQueryResult, error) {
	return a.dictItemRepository.Query(param)
}

// GetDictItems 获取字典项列表（下拉选项）
func (a DictItemService) GetDictItems(dictCode string) ([]*system.DictItemOptionVO, error) {
	list, err := a.dictItemRepository.GetByDictCode(dictCode)
	if err != nil {
		return nil, err
	}

	return list.ToOptionList(), nil
}

// GetDictItemForm 获取字典项表单数据
func (a DictItemService) GetDictItemForm(id uint64) (*system.DictItemForm, error) {
	item, err := a.dictItemRepository.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.DictItemForm{
		ID:       item.ID,
		DictCode: item.DictCode,
		Label:    item.Label,
		Value:    item.Value,
		TagType:  item.TagType,
		Sort:     item.Sort,
		Status:   item.Status,
		Remark:   item.Remark,
	}, nil
}

// SaveDictItem 新增字典项
func (a DictItemService) SaveDictItem(form *system.DictItemForm, createdBy uint64) error {
	item := &system.DictItem{
		DictCode: form.DictCode,
		Label:    form.Label,
		Value:    form.Value,
		TagType:  form.TagType,
		Sort:     form.Sort,
		Status:   form.Status,
		Remark:   form.Remark,
		CreateBy: createdBy,
	}

	return a.dictItemRepository.Create(item)
}

// UpdateDictItem 更新字典项
func (a DictItemService) UpdateDictItem(id uint64, form *system.DictItemForm, updatedBy uint64) error {
	// 检查字典项是否存在
	_, err := a.dictItemRepository.Get(id)
	if err != nil {
		return err
	}

	item := &system.DictItem{
		ID:       id,
		DictCode: form.DictCode,
		Label:    form.Label,
		Value:    form.Value,
		TagType:  form.TagType,
		Sort:     form.Sort,
		Status:   form.Status,
		Remark:   form.Remark,
		UpdateBy: updatedBy,
	}

	return a.dictItemRepository.Update(id, item)
}

// DeleteDictItemByIds 删除字典项
func (a DictItemService) DeleteDictItemByIds(ids string, deletedBy uint64) error {
	if ids == "" {
		return errors.New("删除的字典项数据为空")
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
		return errors.New("删除的字典项数据为空")
	}

	return a.dictItemRepository.DeleteByIDs(idList, deletedBy)
}
