package system

import (
	"time"

	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/queue"
)

// Task 队列任务模型（用于查询展示）
type Task struct {
	ID               uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Type             string       `gorm:"column:type;size:100;not null;index" json:"type"`
	Status           queue.Status `gorm:"column:status;size:50;not null;index" json:"status"`
	CorrelationID    string       `gorm:"column:correlation_id;type:char(36);index" json:"correlationId"`
	OwnerID          uint64       `gorm:"column:owner_id;index" json:"ownerId"`
	PrivateState     string       `gorm:"column:private_state;type:text" json:"privateState"`
	RetryCount       int          `gorm:"column:public_retry_count;default:0" json:"retryCount"`
	ExecutedDuration int64        `gorm:"column:public_executed_duration;default:0" json:"executedDuration"`
	Error            string       `gorm:"column:public_error;type:text" json:"error"`
	ErrorHistory     string       `gorm:"column:public_error_history;type:text" json:"errorHistory"`
	ResumeTime       int64        `gorm:"column:public_resume_time;default:0" json:"resumeTime"`
	CreatedAt        time.Time    `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        time.Time    `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "sys_tasks"
}

type Tasks []*Task

// TaskQueryParam 任务查询参数
type TaskQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	Keywords       string `query:"keywords"`
	Type           string `query:"type"`
	Status         string `query:"status"`
	CorrelationID  string `query:"correlationId"`
	CreateTimeFrom string `query:"createdAt[0]"`
	CreateTimeTo   string `query:"createdAt[1]"`
}

// TaskQueryResult 任务查询结果
type TaskQueryResult struct {
	List       Tasks           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

// TaskPageVO 任务分页视图对象
type TaskPageVO struct {
	ID               uint64 `json:"id"`
	Type             string `json:"type"`
	Status           string `json:"status"`
	CorrelationID    string `json:"correlationId"`
	OwnerID          uint64 `json:"ownerId"`
	RetryCount       int    `json:"retryCount"`
	ExecutedDuration int64  `json:"executedDuration"`
	Error            string `json:"error"`
	ErrorHistory     string `json:"errorHistory"`
	ResumeTime       int64  `json:"resumeTime"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// ToPageVOList 转换为分页视图对象列表
func (list Tasks) ToPageVOList() []*TaskPageVO {
	result := make([]*TaskPageVO, 0, len(list))
	for _, item := range list {
		result = append(result, &TaskPageVO{
			ID:               item.ID,
			Type:             item.Type,
			Status:           string(item.Status),
			CorrelationID:    item.CorrelationID,
			OwnerID:          item.OwnerID,
			RetryCount:       item.RetryCount,
			ExecutedDuration: item.ExecutedDuration,
			Error:            item.Error,
			ErrorHistory:     item.ErrorHistory,
			ResumeTime:       item.ResumeTime,
			CreatedAt:        item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}

// TaskStatsVO 任务统计视图对象
type TaskStatsVO struct {
	BusyWorkers     int   `json:"busyWorkers"`
	SuccessTasks    int   `json:"successTasks"`
	FailureTasks    int   `json:"failureTasks"`
	SubmittedTasks  int   `json:"submittedTasks"`
	SuspendingTasks int   `json:"suspendingTasks"`
	QueuedCount     int64 `json:"queuedCount"`
	ProcessingCount int64 `json:"processingCount"`
	CompletedCount  int64 `json:"completedCount"`
	ErrorCount      int64 `json:"errorCount"`
	CanceledCount   int64 `json:"canceledCount"`
}

// TaskTypeVO 任务类型视图对象
type TaskTypeVO struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
