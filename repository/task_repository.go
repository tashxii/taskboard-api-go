package repository

import (
	"database/sql"
	"sync"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"

	"github.com/jinzhu/gorm"
)

var lockTask = &sync.Mutex{}

// TaskRepository is repository of task table
type TaskRepository struct {
	tx *gorm.DB
}

// NewTaskRepository returns new instance of TaskRepository
func NewTaskRepository(tx *gorm.DB) *TaskRepository {
	if tx == nil {
		// Programing error!!
		panic("tx must be set")
	}
	return &TaskRepository{
		tx: tx,
	}
}

// FindFirstTask returns first Task matching with specified condition
func (repo *TaskRepository) FindFirstTask(condition interface{}, sortOrders []string) (result model.Task, err error) {
	query := repo.tx.Where(condition)
	if sortOrders == nil {
		sortOrders = []string{}
	}

	for _, sortOrder := range sortOrders {
		query = query.Order(sortOrder)
	}
	err = query.First(&result).Error
	return
}

// FindTasks returns Tasks matching with specified condition
func (repo *TaskRepository) FindTasks(condition interface{}, offset int, limit int, sortOrders []string) (result []model.Task, err error) {
	query := repo.tx.Where(condition)
	if offset >= 0 {
		query = query.Offset(offset)
	}
	if limit >= 0 {
		query = query.Limit(limit)
	}

	if sortOrders == nil {
		sortOrders = []string{}
	}
	for _, task := range sortOrders {
		query = query.Order(task)
	}

	err = query.Find(&result).Error
	return
}

// CountTasks returns the number of Tasks matching specfied condition
func (repo *TaskRepository) CountTasks(condition interface{}) (count int, err error) {
	var tasks []model.Task
	err = repo.tx.Where(condition).Find(&tasks).Count(&count).Error
	return
}

// CreateTask inserts new Task record
func (repo *TaskRepository) CreateTask(task *model.Task) error {
	return repo.CreateTasks([]*model.Task{task})
}

// UpdateTask updates Task record
func (repo *TaskRepository) UpdateTask(task *model.Task) error {
	return repo.UpdateTasks([]*model.Task{task})
}

// DeleteTask deletes Task record
func (repo *TaskRepository) DeleteTask(task *model.Task) error {
	return repo.DeleteTasks([]*model.Task{task})
}

// CreateTasks inserts new Task records.
func (repo *TaskRepository) CreateTasks(tasks []*model.Task) (err error) {
	lockTask.Lock()
	defer lockTask.Unlock()

	for _, task := range tasks {
		max := 0
		if task.BoardID != "" {
			max, err = repo.MaxTaskDispOrder(&model.Task{BoardID: task.BoardID})
			if err != nil {
				return
			}
		}
		task.DispOrder = max + 1
		err = repo.tx.Create(task).Error
		if err != nil {
			return
		}
	}
	return
}

// UpdateTasks updates task records
func (repo *TaskRepository) UpdateTasks(tasks []*model.Task) (err error) {
	lockTask.Lock()
	defer lockTask.Unlock()

	for _, task := range tasks {
		oldVersion := task.Version
		task.Version++
		db := repo.tx.Model(&model.Task{}).Where("version = ?", oldVersion).Save(task)
		count := db.RowsAffected
		err = db.Error
		// return ErrorRecordNotFoud as optimistic lock error
		if err == nil && count == 0 {
			return orm.ErrorRecordNotFound
		}
		if err != nil {
			return
		}
	}
	return
}

// DeleteTasks deletes Task records
func (repo *TaskRepository) DeleteTasks(tasks []*model.Task) (err error) {
	for _, task := range tasks {
		if task.ID == "" {
			continue // To avoid deleting all due to gorm warning, continue here.
		}
		err = repo.tx.Delete(task).Error
		if err != nil {
			return
		}
	}
	return
}

// MaxTaskDispOrder return max of disp order matching specified condition
func (repo *TaskRepository) MaxTaskDispOrder(condition interface{}) (max int, err error) {
	var out sql.NullInt64
	err = repo.tx.Model(&model.Task{}).Select("max(disp_order)").
		Where(condition).Row().Scan(&out)
	if err != nil {
		return
	}
	if !out.Valid {
		// no row selected -> returns 0
		return
	}
	return int(out.Int64), nil
}

// MoveToIceboxBoard move tasks to icebox board which matches specified board id
func (repo *TaskRepository) MoveToIceboxBoard(boardID string) (err error) {
	lockTask.Lock()
	defer lockTask.Unlock()
	max, err := repo.MaxTaskDispOrder(&model.Task{BoardID: model.SystemBoardIcebox.ID})
	err = repo.tx.Model(&model.Task{}).Where("board_id = ?", boardID).
		Update("disp_order", gorm.Expr("disp_order + ?", max)).Error
	if err != nil {
		return
	}
	return repo.tx.Model(&model.Task{}).Where("board_id = ?", boardID).
		Update(&model.Task{BoardID: model.SystemBoardIcebox.ID}).Error
}

// MoveTaskDispOrders changes task order position.
func (repo *TaskRepository) MoveTaskDispOrders(
	taskID, fromBoardID string, fromDispOrder int,
	toBoardID string, toDispOrder int,
) (err error) {
	if fromBoardID == toBoardID {
		low := fromDispOrder     // ex) 1
		high := toDispOrder      // ex) 3
		expr := "disp_order - 1" // a1 b2 c3 d4 => b1 c2 a3 d4
		if fromDispOrder > toDispOrder {
			high = fromDispOrder    // ex) 3
			low = toDispOrder       // ex) 1
			expr = "disp_order + 1" // a1 b2 c3 d4 => c1 a2 b3 d4
		}
		err = repo.tx.Model(&model.Task{}).
			Where("board_id = ? and disp_order >= ? and disp_order <= ?", fromBoardID, low, high).
			Update("disp_order", gorm.Expr(expr)).Error
	} else {
		// ex) from=3 to=2
		// a1 b2 c3 d4 e5  => a1 b2 d3 e4
		// x1 y2 z3        => x1 c2 y3 z4
		// shift - 1 (remove form source board order)
		err = repo.tx.Model(&model.Task{}).Where("board_id = ? and disp_order >= ?", fromBoardID, fromDispOrder).
			Update("disp_order", gorm.Expr("disp_order - 1")).Error
		if err != nil {
			return
		}
		// shift + 1 (insert to destination board order)
		err = repo.tx.Model(&model.Task{}).Where("board_id = ? and disp_order >= ?", toBoardID, toDispOrder).
			Update("disp_order", gorm.Expr("disp_order + 1")).Error
		if err != nil {
			return
		}
	}
	// move
	return repo.tx.Model(&model.Task{}).Where("id = ?", taskID).
		Update(&model.Task{DispOrder: toDispOrder, BoardID: toBoardID}).Error
}
