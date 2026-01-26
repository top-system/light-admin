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

// DictService service layer
type DictService struct {
	logger             lib.Logger
	dictRepository     repository.DictRepository
	dictItemRepository repository.DictItemRepository
}

// NewDictService creates a new dict service
func NewDictService(
	logger lib.Logger,
	dictRepository repository.DictRepository,
	dictItemRepository repository.DictItemRepository,
) DictService {
	return DictService{
		logger:             logger,
		dictRepository:     dictRepository,
		dictItemRepository: dictItemRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a DictService) WithTrx(trxHandle *gorm.DB) DictService {
	a.dictRepository = a.dictRepository.WithTrx(trxHandle)
	a.dictItemRepository = a.dictItemRepository.WithTrx(trxHandle)
	return a
}

// GetDictPage 获取字典分页列表
func (a DictService) GetDictPage(param *system.DictQueryParam) (*system.DictQueryResult, error) {
	return a.dictRepository.Query(param)
}

// GetDictList 获取字典列表（下拉选项）
func (a DictService) GetDictList() ([]*system.DictOption, error) {
	list, err := a.dictRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return list.ToOptions(), nil
}

// GetDictForm 获取字典表单数据
func (a DictService) GetDictForm(id uint64) (*system.DictForm, error) {
	dict, err := a.dictRepository.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.DictForm{
		ID:       dict.ID,
		DictCode: dict.DictCode,
		Name:     dict.Name,
		Status:   dict.Status,
		Remark:   dict.Remark,
	}, nil
}

// SaveDict 新增字典
func (a DictService) SaveDict(form *system.DictForm, createdBy uint64) error {
	// 校验字典编码是否存在
	existDict, err := a.dictRepository.GetByCode(form.DictCode)
	if err != nil {
		return err
	}
	if existDict != nil {
		return errors.New("字典编码已存在")
	}

	dict := &system.Dict{
		DictCode: form.DictCode,
		Name:     form.Name,
		Status:   form.Status,
		Remark:   form.Remark,
		CreateBy: createdBy,
	}

	return a.dictRepository.Create(dict)
}

// UpdateDict 更新字典
func (a DictService) UpdateDict(id uint64, form *system.DictForm, updatedBy uint64) error {
	// 检查字典是否存在
	existDict, err := a.dictRepository.Get(id)
	if err != nil {
		return err
	}

	// 校验字典编码是否存在（排除自身）
	if existDict.DictCode != form.DictCode {
		dupDict, err := a.dictRepository.GetByCode(form.DictCode, id)
		if err != nil {
			return err
		}
		if dupDict != nil {
			return errors.New("字典编码已存在")
		}

		// 如果字典编码变更，需要更新字典项的字典编码
		if err := a.dictRepository.UpdateDictItemsCode(existDict.DictCode, form.DictCode); err != nil {
			return err
		}
	}

	dict := &system.Dict{
		ID:       id,
		DictCode: form.DictCode,
		Name:     form.Name,
		Status:   form.Status,
		Remark:   form.Remark,
		UpdateBy: updatedBy,
	}

	return a.dictRepository.Update(id, dict)
}

// DeleteDictByIds 删除字典
func (a DictService) DeleteDictByIds(ids string, deletedBy uint64) error {
	if ids == "" {
		return errors.New("删除的字典数据为空")
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
		return errors.New("删除的字典数据为空")
	}

	// 获取要删除的字典列表
	dictList, err := a.dictRepository.GetByIDs(idList)
	if err != nil {
		return err
	}

	// 删除字典
	if err := a.dictRepository.DeleteByIDs(idList, deletedBy); err != nil {
		return err
	}

	// 删除字典项
	if len(dictList) > 0 {
		dictCodes := make([]string, 0, len(dictList))
		for _, dict := range dictList {
			dictCodes = append(dictCodes, dict.DictCode)
		}
		if err := a.dictItemRepository.DeleteByDictCodes(dictCodes, deletedBy); err != nil {
			return err
		}
	}

	return nil
}

// GetDictCodesByIds 根据字典ID列表获取字典编码列表
func (a DictService) GetDictCodesByIds(ids string) ([]string, error) {
	if ids == "" {
		return []string{}, nil
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

	dictList, err := a.dictRepository.GetByIDs(idList)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(dictList))
	for _, dict := range dictList {
		codes = append(codes, dict.DictCode)
	}

	return codes, nil
}
