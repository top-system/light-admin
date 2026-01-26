package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// Log 系统操作日志模型
type Log struct {
	ID              uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Module          string       `gorm:"column:module;size:50;not null" json:"module"`
	RequestMethod   string       `gorm:"column:request_method;size:64;not null" json:"requestMethod"`
	RequestParams   string       `gorm:"column:request_params;type:text" json:"requestParams"`
	ResponseContent string       `gorm:"column:response_content;type:mediumtext" json:"responseContent"`
	Content         string       `gorm:"column:content;size:255;not null" json:"content"`
	RequestURI      string       `gorm:"column:request_uri;size:255" json:"requestUri"`
	Method          string       `gorm:"column:method;size:255" json:"method"`
	IP              string       `gorm:"column:ip;size:45" json:"ip"`
	Province        string       `gorm:"column:province;size:100" json:"province"`
	City            string       `gorm:"column:city;size:100" json:"city"`
	ExecutionTime   int64        `gorm:"column:execution_time" json:"executionTime"`
	Browser         string       `gorm:"column:browser;size:100" json:"browser"`
	BrowserVersion  string       `gorm:"column:browser_version;size:100" json:"browserVersion"`
	OS              string       `gorm:"column:os;size:100" json:"os"`
	CreateBy        uint64       `gorm:"column:create_by" json:"createBy"`
	CreateTime      dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

// TableName 指定表名
func (Log) TableName() string {
	return "sys_log"
}

type Logs []*Log

type LogQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Keywords       string `query:"keywords"`
	Module         string `query:"module"`
	CreateTimeFrom string `query:"createTime[0]"`
	CreateTimeTo   string `query:"createTime[1]"`
}

type LogQueryResult struct {
	List       Logs            `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// LogPageVO 日志分页视图对象
type LogPageVO struct {
	ID             uint64       `json:"id"`
	Module         string       `json:"module"`
	RequestMethod  string       `json:"requestMethod"`
	Content        string       `json:"content"`
	RequestURI     string       `json:"requestUri"`
	IP             string       `json:"ip"`
	Province       string       `json:"province"`
	City           string       `json:"city"`
	ExecutionTime  int64        `json:"executionTime"`
	Browser        string       `json:"browser"`
	BrowserVersion string       `json:"browserVersion"`
	OS             string       `json:"os"`
	CreateBy       uint64       `json:"createBy"`
	CreateTime     dto.DateTime `json:"createTime"`
}

// ToPageVOList 转换为分页视图对象列表
func (list Logs) ToPageVOList() []*LogPageVO {
	result := make([]*LogPageVO, 0, len(list))
	for _, item := range list {
		result = append(result, &LogPageVO{
			ID:             item.ID,
			Module:         item.Module,
			RequestMethod:  item.RequestMethod,
			Content:        item.Content,
			RequestURI:     item.RequestURI,
			IP:             item.IP,
			Province:       item.Province,
			City:           item.City,
			ExecutionTime:  item.ExecutionTime,
			Browser:        item.Browser,
			BrowserVersion: item.BrowserVersion,
			OS:             item.OS,
			CreateBy:       item.CreateBy,
			CreateTime:     item.CreateTime,
		})
	}
	return result
}
