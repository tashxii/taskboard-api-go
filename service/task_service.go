package service

import (
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/repository"

	"github.com/jinzhu/gorm"
)

// TaskService provides apis for task management.
type TaskService struct {
	tx       *gorm.DB
	taskRepo *repository.TaskRepository
}

// NewTaskService return new instance of TaskService.
func NewTaskService(tx *gorm.DB) *TaskService {
	return &TaskService{
		tx:       tx,
		taskRepo: repository.NewTaskRepository(tx),
	}
}

// FindTask returns task matching specified condition
func (s *TaskService) FindTask(condition interface{}) (*model.Task, error) {
	find, err := s.taskRepo.FindFirstTask(condition, []string{"id"})
	if err != nil {
		if err == orm.ErrorRecordNotFound {
			return nil, NewSvcErrorf(ErrorCodeNotFound, err, "Task not found")
		}
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find task")
	}
	return &find, nil
}

// FindTasks finds all tasks
func (s *TaskService) FindTasks(condition interface{}, sortOrders []string) ([]model.Task, error) {
	tasks, err := s.taskRepo.FindTasks(condition, 0, orm.NoLimit, sortOrders)
	if err != nil {
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find tasks")
	}
	return tasks, nil
}

// CreateTask creates new task
func (s *TaskService) CreateTask(task *model.Task) error {
	max, err := s.taskRepo.MaxTaskDispOrder(&model.Task{BoardID: task.BoardID})
	if err != nil {
		return NewSvcError(ErrorCodeDB, err, "Failed to get max disp order")
	}
	task.DispOrder = max + 1
	err = s.taskRepo.CreateTask(task)
	if err != nil {
		return NewSvcError(ErrorCodeDB, err, "Failed to create task")
	}
	return nil
}

// UpdateTask updates specifed task
func (s *TaskService) UpdateTask(find *model.Task, task *model.Task) error {
	// Set dispOrder
	if find.BoardID != task.BoardID {
		dispOrder, err := s.taskRepo.MaxTaskDispOrder(&model.Task{BoardID: task.BoardID})
		if err != nil {
			return NewSvcErrorf(ErrorCodeDB, err, "Failed to get max disp order of task. ID:%s", task.ID)
		}
		task.DispOrder = dispOrder + 1
	}

	err := s.taskRepo.UpdateTask(task)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to update task. ID:%s", task.ID)
	}
	return nil
}

// DeleteTask deletes specifed task
func (s *TaskService) DeleteTask(task *model.Task) error {
	err := s.taskRepo.DeleteTask(task)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to delete task. ID:%s", task.ID)
	}
	return nil
}

// UpdateTaskOrders changes display order of tasks.
func (s *TaskService) UpdateTaskOrders(taskID, fromBoardID string, fromDispOrder int,
	toBoardID string, toDispOrder int,
) (err error) {
	err = s.taskRepo.MoveTaskDispOrders(taskID, fromBoardID, fromDispOrder, toBoardID, toDispOrder)
	if err != nil {
		return
	}
	return
}
