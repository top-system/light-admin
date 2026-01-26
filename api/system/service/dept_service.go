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

// DeptService service layer
type DeptService struct {
	logger         lib.Logger
	deptRepository repository.DeptRepository
}

// NewDeptService creates a new dept service
func NewDeptService(
	logger lib.Logger,
	deptRepository repository.DeptRepository,
) DeptService {
	return DeptService{
		logger:         logger,
		deptRepository: deptRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a DeptService) WithTrx(trxHandle *gorm.DB) DeptService {
	a.deptRepository = a.deptRepository.WithTrx(trxHandle)
	return a
}

// GetDeptList 获取部门列表（树形）
func (a DeptService) GetDeptList(param *system.DeptQueryParam) ([]*system.DeptVO, error) {
	deptList, err := a.deptRepository.Query(param)
	if err != nil {
		return nil, err
	}

	if len(deptList) == 0 {
		return []*system.DeptVO{}, nil
	}

	// 获取所有部门ID
	deptIds := make(map[uint64]bool)
	for _, dept := range deptList {
		deptIds[dept.ID] = true
	}

	// 获取父节点ID
	parentIds := make(map[uint64]bool)
	for _, dept := range deptList {
		parentIds[dept.ParentID] = true
	}

	// 获取根节点ID（父节点ID中不包含在部门ID中的节点）
	var rootIds []uint64
	for parentId := range parentIds {
		if !deptIds[parentId] {
			rootIds = append(rootIds, parentId)
		}
	}

	// 递归生成部门树形列表
	var result []*system.DeptVO
	for _, rootId := range rootIds {
		children := a.recurDeptList(rootId, deptList)
		result = append(result, children...)
	}

	return result, nil
}

// recurDeptList 递归生成部门树形列表
func (a DeptService) recurDeptList(parentId uint64, deptList system.Depts) []*system.DeptVO {
	var result []*system.DeptVO

	for _, dept := range deptList {
		if dept.ParentID == parentId {
			deptVO := &system.DeptVO{
				ID:         dept.ID,
				Name:       dept.Name,
				Code:       dept.Code,
				ParentID:   dept.ParentID,
				Sort:       dept.Sort,
				Status:     dept.Status,
				CreateTime: dept.CreateTime,
				UpdateTime: dept.UpdateTime,
			}
			children := a.recurDeptList(dept.ID, deptList)
			if len(children) > 0 {
				deptVO.Children = children
			}
			result = append(result, deptVO)
		}
	}

	return result
}

// ListDeptOptions 部门下拉选项
func (a DeptService) ListDeptOptions() ([]*system.DeptOption, error) {
	deptList, err := a.deptRepository.GetAllEnabled()
	if err != nil {
		return nil, err
	}

	if len(deptList) == 0 {
		return []*system.DeptOption{}, nil
	}

	// 获取所有部门ID
	deptIds := make(map[uint64]bool)
	for _, dept := range deptList {
		deptIds[dept.ID] = true
	}

	// 获取父节点ID
	parentIds := make(map[uint64]bool)
	for _, dept := range deptList {
		parentIds[dept.ParentID] = true
	}

	// 获取根节点ID
	var rootIds []uint64
	for parentId := range parentIds {
		if !deptIds[parentId] {
			rootIds = append(rootIds, parentId)
		}
	}

	// 递归生成部门下拉选项
	var result []*system.DeptOption
	for _, rootId := range rootIds {
		children := a.recurDeptOptions(rootId, deptList)
		result = append(result, children...)
	}

	return result, nil
}

// recurDeptOptions 递归生成部门下拉选项
func (a DeptService) recurDeptOptions(parentId uint64, deptList system.Depts) []*system.DeptOption {
	var result []*system.DeptOption

	for _, dept := range deptList {
		if dept.ParentID == parentId {
			option := &system.DeptOption{
				Value: dept.ID,
				Label: dept.Name,
			}
			children := a.recurDeptOptions(dept.ID, deptList)
			if len(children) > 0 {
				option.Children = children
			}
			result = append(result, option)
		}
	}

	return result
}

// SaveDept 新增部门
func (a DeptService) SaveDept(form *system.DeptForm, createdBy uint64) (uint64, error) {
	// 校验部门编号是否存在
	existDept, err := a.deptRepository.GetByCode(form.Code)
	if err != nil {
		return 0, err
	}
	if existDept != nil {
		return 0, errors.New("部门编号已存在")
	}

	// 生成部门路径
	parentID := form.ParentID.Value()
	treePath, err := a.generateDeptTreePath(parentID)
	if err != nil {
		return 0, err
	}

	dept := &system.Dept{
		Name:     form.Name,
		Code:     form.Code,
		ParentID: parentID,
		TreePath: treePath,
		Sort:     form.Sort,
		Status:   form.Status,
		CreateBy: createdBy,
	}

	if err := a.deptRepository.Create(dept); err != nil {
		return 0, err
	}

	return dept.ID, nil
}

// GetDeptForm 获取部门表单数据
func (a DeptService) GetDeptForm(id uint64) (*system.DeptForm, error) {
	dept, err := a.deptRepository.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.DeptForm{
		ID:       dept.ID,
		Name:     dept.Name,
		Code:     dept.Code,
		ParentID: dto.FlexUint64(dept.ParentID),
		Sort:     dept.Sort,
		Status:   dept.Status,
	}, nil
}

// UpdateDept 更新部门
func (a DeptService) UpdateDept(id uint64, form *system.DeptForm, updatedBy uint64) (uint64, error) {
	// 检查部门是否存在
	_, err := a.deptRepository.Get(id)
	if err != nil {
		return 0, err
	}

	// 校验部门编号是否存在（排除自身）
	existDept, err := a.deptRepository.GetByCode(form.Code, id)
	if err != nil {
		return 0, err
	}
	if existDept != nil {
		return 0, errors.New("部门编号已存在")
	}

	// 生成部门路径
	parentID := form.ParentID.Value()
	treePath, err := a.generateDeptTreePath(parentID)
	if err != nil {
		return 0, err
	}

	dept := &system.Dept{
		ID:       id,
		Name:     form.Name,
		Code:     form.Code,
		ParentID: parentID,
		TreePath: treePath,
		Sort:     form.Sort,
		Status:   form.Status,
		UpdateBy: updatedBy,
	}

	if err := a.deptRepository.Update(id, dept); err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteByIds 删除部门
func (a DeptService) DeleteByIds(ids string, deletedBy uint64) error {
	if ids == "" {
		return errors.New("删除的部门数据为空")
	}

	idStrs := strings.Split(ids, ",")
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}

		// 删除部门及子部门
		if err := a.deptRepository.DeleteByTreePath(id, deletedBy); err != nil {
			return err
		}
	}

	return nil
}

// generateDeptTreePath 生成部门路径
func (a DeptService) generateDeptTreePath(parentId uint64) (string, error) {
	if parentId == 0 {
		return "0", nil
	}

	parent, err := a.deptRepository.Get(parentId)
	if err != nil {
		return "", err
	}

	return parent.TreePath + "," + strconv.FormatUint(parent.ID, 10), nil
}
