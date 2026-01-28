package service

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// ConfigService service layer
type ConfigService struct {
	logger           lib.Logger
	cache            lib.Cache
	configRepository repository.ConfigRepository
}

// NewConfigService creates a new config service
func NewConfigService(
	logger lib.Logger,
	cache lib.Cache,
	configRepository repository.ConfigRepository,
) ConfigService {
	return ConfigService{
		logger:           logger,
		cache:            cache,
		configRepository: configRepository,
	}
}

// WithTrx delegates transaction to repository database
func (a ConfigService) WithTrx(trxHandle *gorm.DB) ConfigService {
	a.configRepository = a.configRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询系统配置
func (a ConfigService) Query(param *system.ConfigQueryParam) (*system.ConfigQueryResult, error) {
	return a.configRepository.Query(param)
}

// Get 获取系统配置
func (a ConfigService) Get(id uint64) (*system.Config, error) {
	return a.configRepository.Get(id)
}

// GetByKey 根据Key获取配置
func (a ConfigService) GetByKey(key string) (*system.Config, error) {
	return a.configRepository.GetByKey(key)
}

// GetForm 获取系统配置表单数据
func (a ConfigService) GetForm(id uint64) (*system.ConfigForm, error) {
	config, err := a.configRepository.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.ConfigForm{
		ID:          config.ID,
		ConfigName:  config.ConfigName,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		Remark:      config.Remark,
	}, nil
}

// Create 创建系统配置
func (a ConfigService) Create(form *system.ConfigForm, createdBy uint64) error {
	// 检查配置key是否已存在
	exists, err := a.configRepository.ExistsByKey(form.ConfigKey, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.ConfigKeyAlreadyExists
	}

	config := &system.Config{
		ConfigName:  form.ConfigName,
		ConfigKey:   form.ConfigKey,
		ConfigValue: form.ConfigValue,
		Remark:      form.Remark,
		CreateBy:    createdBy,
		IsDeleted:   0,
	}

	return a.configRepository.Create(config)
}

// Update 更新系统配置
func (a ConfigService) Update(id uint64, form *system.ConfigForm, updatedBy uint64) error {
	// 检查配置是否存在
	_, err := a.configRepository.Get(id)
	if err != nil {
		return err
	}

	// 检查配置key是否已存在（排除自身）
	exists, err := a.configRepository.ExistsByKey(form.ConfigKey, id)
	if err != nil {
		return err
	}
	if exists {
		return errors.ConfigKeyAlreadyExists
	}

	config := &system.Config{
		ID:          id,
		ConfigName:  form.ConfigName,
		ConfigKey:   form.ConfigKey,
		ConfigValue: form.ConfigValue,
		Remark:      form.Remark,
		UpdateBy:    updatedBy,
	}

	return a.configRepository.Update(id, config)
}

// Delete 删除系统配置
func (a ConfigService) Delete(id uint64, deletedBy uint64) error {
	// 检查配置是否存在
	_, err := a.configRepository.Get(id)
	if err != nil {
		return err
	}

	return a.configRepository.Delete(id, deletedBy)
}

// RefreshCache 刷新配置缓存
func (a ConfigService) RefreshCache() error {
	// 删除旧缓存
	cacheKey := "sys:config"
	_, err := a.cache.Delete(cacheKey)
	if err != nil {
		a.logger.Zap.Warnf("Failed to delete config cache: %v", err)
	}

	// 获取所有配置
	configs, err := a.configRepository.GetAll()
	if err != nil {
		return err
	}

	// 写入缓存
	configMap := make(map[string]interface{})
	for _, config := range configs {
		configMap[config.ConfigKey] = config.ConfigValue
	}

	if len(configMap) > 0 {
		if err := a.cache.HMSet(cacheKey, configMap); err != nil {
			a.logger.Zap.Warnf("Failed to set config cache: %v", err)
		}
	}

	return nil
}

// GetSystemConfig 从缓存获取系统配置
func (a ConfigService) GetSystemConfig(key string) (string, error) {
	cacheKey := "sys:config"
	var value string
	err := a.cache.HGet(cacheKey, key, &value)
	if err != nil {
		// 如果缓存中没有，从数据库获取
		config, dbErr := a.configRepository.GetByKey(key)
		if dbErr != nil {
			return "", dbErr
		}
		return config.ConfigValue, nil
	}
	return value, nil
}
