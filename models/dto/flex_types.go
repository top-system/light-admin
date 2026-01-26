package dto

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"
)

// FlexUint64 可以同时接受字符串和数字格式的 uint64
type FlexUint64 uint64

// UnmarshalJSON 自定义 JSON 反序列化
func (f *FlexUint64) UnmarshalJSON(data []byte) error {
	// 尝试作为数字解析
	var num uint64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexUint64(num)
		return nil
	}

	// 尝试作为字符串解析
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*f = 0
			return nil
		}
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*f = FlexUint64(num)
		return nil
	}

	return nil
}

// MarshalJSON 自定义 JSON 序列化（输出为数字）
func (f FlexUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(f))
}

// Value 获取 uint64 值
func (f FlexUint64) Value() uint64 {
	return uint64(f)
}

// FlexInt 可以同时接受字符串和数字格式的 int
type FlexInt int

// UnmarshalJSON 自定义 JSON 反序列化
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	// 尝试作为数字解析
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexInt(num)
		return nil
	}

	// 尝试作为字符串解析
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*f = 0
			return nil
		}
		num, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*f = FlexInt(num)
		return nil
	}

	return nil
}

// MarshalJSON 自定义 JSON 序列化（输出为数字）
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// Value 获取 int 值
func (f FlexInt) Value() int {
	return int(f)
}

// MenuType 菜单类型，可以接受字符串("M","C","B")或数字(1,2,4)
// 数据库: 1-菜单 2-目录 4-按钮
// 前端:   M-菜单 C-目录 B-按钮
type MenuType int

// UnmarshalJSON 自定义 JSON 反序列化
func (m *MenuType) UnmarshalJSON(data []byte) error {
	// 尝试作为数字解析
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*m = MenuType(num)
		return nil
	}

	// 尝试作为字符串解析
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		switch str {
		case "M", "1":
			*m = 1 // 菜单
		case "C", "2":
			*m = 2 // 目录
		case "B", "4":
			*m = 4 // 按钮
		default:
			*m = 0
		}
		return nil
	}

	return nil
}

// MarshalJSON 自定义 JSON 序列化（输出为字符串）
func (m MenuType) MarshalJSON() ([]byte, error) {
	var str string
	switch int(m) {
	case 1:
		str = "M" // 菜单
	case 2:
		str = "C" // 目录
	case 4:
		str = "B" // 按钮
	default:
		str = ""
	}
	return json.Marshal(str)
}

// Value 获取 int 值
func (m MenuType) Value() int {
	return int(m)
}

// DateTime 自定义时间类型，JSON序列化时格式化为 "YYYY-MM-DD HH:mm:ss"
type DateTime time.Time

const DateTimeFormat = "2006-01-02 15:04:05"

// MarshalJSON 自定义 JSON 序列化
func (t DateTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(tt.Format(DateTimeFormat))
}

// UnmarshalJSON 自定义 JSON 反序列化
func (t *DateTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if str == "" {
		*t = DateTime(time.Time{})
		return nil
	}
	tt, err := time.ParseInLocation(DateTimeFormat, str, time.Local)
	if err != nil {
		return err
	}
	*t = DateTime(tt)
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (t DateTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (t *DateTime) Scan(value interface{}) error {
	if value == nil {
		*t = DateTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = DateTime(v)
	}
	return nil
}

// Time 获取 time.Time 值
func (t DateTime) Time() time.Time {
	return time.Time(t)
}

// NullDateTime 可空的时间类型，JSON序列化时格式化为 "YYYY-MM-DD HH:mm:ss" 或 null
type NullDateTime struct {
	Time  time.Time
	Valid bool
}

// MarshalJSON 自定义 JSON 序列化
func (t NullDateTime) MarshalJSON() ([]byte, error) {
	if !t.Valid || t.Time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(t.Time.Format(DateTimeFormat))
}

// UnmarshalJSON 自定义 JSON 反序列化
func (t *NullDateTime) UnmarshalJSON(data []byte) error {
	var str *string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if str == nil || *str == "" {
		t.Valid = false
		t.Time = time.Time{}
		return nil
	}
	tt, err := time.ParseInLocation(DateTimeFormat, *str, time.Local)
	if err != nil {
		return err
	}
	t.Time = tt
	t.Valid = true
	return nil
}

// Value 实现 driver.Valuer 接口
func (t NullDateTime) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Time, nil
}

// Scan 实现 sql.Scanner 接口
func (t *NullDateTime) Scan(value interface{}) error {
	if value == nil {
		t.Time, t.Valid = time.Time{}, false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time, t.Valid = v, true
	}
	return nil
}

// NewNullDateTime 从 *time.Time 创建 NullDateTime
func NewNullDateTime(t *time.Time) NullDateTime {
	if t == nil {
		return NullDateTime{Valid: false}
	}
	return NullDateTime{Time: *t, Valid: true}
}
