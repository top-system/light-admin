package database

import (
	"gorm.io/gorm"
)

// Model base model
type Model struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement;" json:"id"`
	CreatedAt Datetime       `gorm:"column:created_at;autoCreateTime;" json:"created_at"`
	UpdatedAt Datetime       `gorm:"column:updated_at;autoUpdateTime;" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;" json:"-"`
}
