package system

import (
	"time"

	"github.com/top-system/light-admin/models/dto"
)

// DownloadTask 下载任务模型
type DownloadTask struct {
	ID            uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID        string       `gorm:"column:task_id;size:100;not null;index" json:"taskId"`
	Hash          string       `gorm:"column:hash;size:100;index" json:"hash"`
	Name          string       `gorm:"column:name;size:500" json:"name"`
	URL           string       `gorm:"column:url;type:text" json:"url"`
	Downloader    string       `gorm:"column:downloader;size:50;not null" json:"downloader"` // aria2 / qbittorrent
	Status        string       `gorm:"column:status;size:50;not null;index" json:"status"`
	Total         int64        `gorm:"column:total;default:0" json:"total"`
	Downloaded    int64        `gorm:"column:downloaded;default:0" json:"downloaded"`
	DownloadSpeed int64        `gorm:"column:download_speed;default:0" json:"downloadSpeed"`
	Uploaded      int64        `gorm:"column:uploaded;default:0" json:"uploaded"`
	UploadSpeed   int64        `gorm:"column:upload_speed;default:0" json:"uploadSpeed"`
	SavePath      string       `gorm:"column:save_path;size:500" json:"savePath"`
	ErrorMessage  string       `gorm:"column:error_message;type:text" json:"errorMessage"`
	OwnerID       uint64       `gorm:"column:owner_id;index" json:"ownerId"`
	CreatedAt     time.Time    `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt     dto.DateTime `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (DownloadTask) TableName() string {
	return "sys_download_tasks"
}

type DownloadTasks []*DownloadTask

// DownloadTaskQueryParam 下载任务查询参数
type DownloadTaskQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Keywords       string `query:"keywords"`
	Status         string `query:"status"`
	Downloader     string `query:"downloader"`
	CreateTimeFrom string `query:"createdAt[0]"`
	CreateTimeTo   string `query:"createdAt[1]"`
}

// DownloadTaskQueryResult 下载任务查询结果
type DownloadTaskQueryResult struct {
	List       DownloadTasks   `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// DownloadTaskPageVO 下载任务分页视图对象
type DownloadTaskPageVO struct {
	ID            uint64  `json:"id"`
	TaskID        string  `json:"taskId"`
	Hash          string  `json:"hash"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Downloader    string  `json:"downloader"`
	Status        string  `json:"status"`
	Total         int64   `json:"total"`
	Downloaded    int64   `json:"downloaded"`
	DownloadSpeed int64   `json:"downloadSpeed"`
	Uploaded      int64   `json:"uploaded"`
	UploadSpeed   int64   `json:"uploadSpeed"`
	SavePath      string  `json:"savePath"`
	ErrorMessage  string  `json:"errorMessage"`
	Progress      float64 `json:"progress"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// ToPageVOList 转换为分页视图对象列表
func (list DownloadTasks) ToPageVOList() []*DownloadTaskPageVO {
	result := make([]*DownloadTaskPageVO, 0, len(list))
	for _, item := range list {
		var progress float64
		if item.Total > 0 {
			progress = float64(item.Downloaded) / float64(item.Total) * 100
		}
		result = append(result, &DownloadTaskPageVO{
			ID:            item.ID,
			TaskID:        item.TaskID,
			Hash:          item.Hash,
			Name:          item.Name,
			URL:           item.URL,
			Downloader:    item.Downloader,
			Status:        item.Status,
			Total:         item.Total,
			Downloaded:    item.Downloaded,
			DownloadSpeed: item.DownloadSpeed,
			Uploaded:      item.Uploaded,
			UploadSpeed:   item.UploadSpeed,
			SavePath:      item.SavePath,
			ErrorMessage:  item.ErrorMessage,
			Progress:      progress,
			CreatedAt:     item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}

// DownloadTaskStatsVO 下载任务统计视图对象
type DownloadTaskStatsVO struct {
	DownloadingCount int64 `json:"downloadingCount"`
	SeedingCount     int64 `json:"seedingCount"`
	CompletedCount   int64 `json:"completedCount"`
	ErrorCount       int64 `json:"errorCount"`
	TotalCount       int64 `json:"totalCount"`
}

// DownloadTaskCreateForm 创建下载任务表单
type DownloadTaskCreateForm struct {
	URL        string                 `json:"url" validate:"required"`
	Downloader string                 `json:"downloader" validate:"required"`
	Options    map[string]interface{} `json:"options"`
}

// DownloadTaskDetailVO 下载任务详情视图对象
type DownloadTaskDetailVO struct {
	DownloadTaskPageVO
	Files []DownloadTaskFileVO `json:"files"`
}

// DownloadTaskFileVO 下载任务文件视图对象
type DownloadTaskFileVO struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
	Selected bool    `json:"selected"`
}

// SetFileDownloadForm 设置文件下载表单
type SetFileDownloadForm struct {
	Files []SetFileDownloadItem `json:"files" validate:"required"`
}

// SetFileDownloadItem 设置文件下载项
type SetFileDownloadItem struct {
	Index    int  `json:"index"`
	Download bool `json:"download"`
}
