package queue

import (
	"context"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// TaskRepository interface for task persistence
type TaskRepository interface {
	// Create creates a new task in database
	Create(ctx context.Context, task *TaskModel) error
	// Update updates an existing task in database
	Update(ctx context.Context, task *TaskModel) error
	// GetByID gets a task by ID
	GetByID(ctx context.Context, id uint64) (*TaskModel, error)
	// GetPendingTasks gets all pending tasks by types
	GetPendingTasks(ctx context.Context, types ...string) ([]*TaskModel, error)
	// Delete deletes a task by ID
	Delete(ctx context.Context, id uint64) error
}

// GormTaskRepository implements TaskRepository using GORM
type GormTaskRepository struct {
	db *gorm.DB
}

// NewGormTaskRepository creates a new GORM task repository
func NewGormTaskRepository(db *gorm.DB) TaskRepository {
	return &GormTaskRepository{db: db}
}

func (r *GormTaskRepository) Create(ctx context.Context, task *TaskModel) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *GormTaskRepository) Update(ctx context.Context, task *TaskModel) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *GormTaskRepository) GetByID(ctx context.Context, id uint64) (*TaskModel, error) {
	var task TaskModel
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *GormTaskRepository) GetPendingTasks(ctx context.Context, types ...string) ([]*TaskModel, error) {
	var tasks []*TaskModel
	query := r.db.WithContext(ctx).
		Where("status IN ?", []Status{StatusQueued, StatusProcessing, StatusSuspending})

	if len(types) > 0 {
		query = query.Where("type IN ?", types)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *GormTaskRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&TaskModel{}, id).Error
}

// TaskArgs represents arguments for creating or updating a task
type TaskArgs struct {
	Status        Status
	Type          string
	PublicState   *TaskPublicState
	PrivateState  string
	OwnerID       uint64
	CorrelationID uuid.UUID
}

// InMemoryTaskRepository implements TaskRepository using in-memory storage
type InMemoryTaskRepository struct {
	tasks  map[uint64]*TaskModel
	nextID uint64
}

// NewInMemoryTaskRepository creates a new in-memory task repository
func NewInMemoryTaskRepository() TaskRepository {
	return &InMemoryTaskRepository{
		tasks:  make(map[uint64]*TaskModel),
		nextID: 1,
	}
}

func (r *InMemoryTaskRepository) Create(ctx context.Context, task *TaskModel) error {
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) Update(ctx context.Context, task *TaskModel) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) GetByID(ctx context.Context, id uint64) (*TaskModel, error) {
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *InMemoryTaskRepository) GetPendingTasks(ctx context.Context, types ...string) ([]*TaskModel, error) {
	var result []*TaskModel
	for _, task := range r.tasks {
		if task.Status == StatusQueued || task.Status == StatusProcessing || task.Status == StatusSuspending {
			if len(types) == 0 {
				result = append(result, task)
			} else {
				for _, t := range types {
					if task.Type == t {
						result = append(result, task)
						break
					}
				}
			}
		}
	}
	return result, nil
}

func (r *InMemoryTaskRepository) Delete(ctx context.Context, id uint64) error {
	delete(r.tasks, id)
	return nil
}
