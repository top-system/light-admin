package queue

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// TaskModel represents the task model in database
type TaskModel struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Type          string         `gorm:"size:100;not null;index" json:"type"`
	Status        Status         `gorm:"size:50;not null;index" json:"status"`
	CorrelationID uuid.UUID      `gorm:"type:char(36);index" json:"correlationId"`
	OwnerID       uint64         `gorm:"index" json:"ownerId"`
	PrivateState  string         `gorm:"type:text" json:"privateState"`
	PublicState   TaskPublicState `gorm:"embedded;embeddedPrefix:public_" json:"publicState"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for TaskModel
func (TaskModel) TableName() string {
	return "sys_tasks"
}

// TaskPublicState represents the public state of a task
type TaskPublicState struct {
	RetryCount       int           `gorm:"default:0" json:"retryCount"`
	ExecutedDuration time.Duration `gorm:"default:0" json:"executedDuration"`
	Error            string        `gorm:"type:text" json:"error"`
	ErrorHistory     StringSlice   `gorm:"type:text" json:"errorHistory"`
	ResumeTime       int64         `gorm:"default:0" json:"resumeTime"`
}

// TaskOwner represents the owner of a task (simplified user interface)
type TaskOwner struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// StringSlice is a custom type for storing string slices in database
type StringSlice []string

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return s.unmarshal(v)
	case string:
		return s.unmarshal([]byte(v))
	}

	*s = []string{}
	return nil
}

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (interface{}, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return s.marshal()
}

func (s *StringSlice) unmarshal(data []byte) error {
	if len(data) == 0 || string(data) == "[]" {
		*s = []string{}
		return nil
	}

	// Simple JSON-like parsing
	str := string(data)
	if str[0] == '[' && str[len(str)-1] == ']' {
		str = str[1 : len(str)-1]
	}
	if str == "" {
		*s = []string{}
		return nil
	}

	// Split by comma and trim quotes
	var result []string
	current := ""
	inQuote := false
	for _, c := range str {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	*s = result
	return nil
}

func (s StringSlice) marshal() (string, error) {
	if len(s) == 0 {
		return "[]", nil
	}

	result := "["
	for i, str := range s {
		if i > 0 {
			result += ","
		}
		result += `"` + str + `"`
	}
	result += "]"
	return result, nil
}
