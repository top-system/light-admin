package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/top-system/light-admin/lib"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSONB is a custom type for storing JSON data across different databases.
// On PostgreSQL it maps to the `jsonb` column type, on MySQL to `json`,
// and on SQLite to `text`.
type JSONB json.RawMessage

// GormDBDataType returns the appropriate column type for the current database.
func (JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "jsonb"
	case "mysql":
		return "json"
	default:
		return "text"
	}
}

// Scan implements the sql.Scanner interface.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB("null")
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONB.Scan: unsupported type %T", value)
	}

	if !json.Valid(bytes) {
		return errors.New("JSONB.Scan: invalid JSON")
	}

	*j = bytes
	return nil
}

// Value implements the driver.Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	if lib.IsPostgreSQL() {
		// Return []byte for PostgreSQL so the driver sends it as jsonb
		return []byte(j), nil
	}
	// MySQL and SQLite: return as string
	return string(j), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB.UnmarshalJSON: on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
