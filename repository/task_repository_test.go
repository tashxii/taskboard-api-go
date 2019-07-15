package repository

import (
	"fmt"
	"taskboard-api-go/common"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

////
/// Model specific functions (Only replace model name, take care names are casesencitive!!)
//
func newTxAndTaskRepository() (tx *gorm.DB, repo *TaskRepository) {
	tx = orm.GetDB().Begin()
	repo = NewTaskRepository(tx)
	return
}

func createTaskTestData(tx *gorm.DB, idFormat string, findIdentify string, count int) []*model.Task {
	result := make([]*model.Task, 0, count)
	for i := 0; i < count; i++ {
		task := model.NewTask(
			"name"+common.GenerateID(),
			findIdentify,
			false,
			time.Now().UTC(),
		)
		task.ID = fmt.Sprintf("%s-%03d", idFormat, i)
		task.DispOrder = i + 1
		task.BoardID = "fixBoardID"
		result = append(result, task)
	}
	return result
}

func insertTaskTestData(tx *gorm.DB, tasks []*model.Task) (err error) {
	for _, task := range tasks {
		err = tx.Create(task).Error
		if err != nil {
			return
		}
	}
	return
}

////
/// Common repository functions' test
//
func TestTaskRepository_FindFirstTask(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	firstTasks := createTaskTestData(tx, "taskID-find", "findDescription", 5)
	secondTasks := createTaskTestData(tx, "taskID-not-find", "notFindDescription", 4)
	insertTasks := append(firstTasks, secondTasks...)
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}
	// Sort by id
	task, err := repo.FindFirstTask(&model.Task{Description: "findDescription"}, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find first: %+v", err)
	}
	expected := "taskID-find-000"
	if task.ID != expected {
		t.Errorf("expected task ID is %s, but got %s", expected, task.ID)
	}
}

func TestTaskRepository_FindTasks(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	firstTasks := createTaskTestData(tx, "taskID-find", "findDescription", 5)
	secondTasks := createTaskTestData(tx, "taskID-not-find", "notFindDescription", 4)
	insertTasks := append(firstTasks, secondTasks...)
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	// Get 3 data from 2nd(index=1) position
	const (
		offset = 1
		limit  = 3
	)
	tasks, err := repo.FindTasks(&model.Task{Description: "findDescription"}, offset, limit, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find: %+v", err)
	}

	// Length of data must be limit
	if len(tasks) != limit {
		t.Errorf("Expected result size = %d, but got %d", limit, len(tasks))
		return
	}
	// Head must be 001
	head := tasks[0]
	headExpected := "taskID-find-001"
	if head.ID != headExpected {
		t.Errorf("head ID must be %s, but got %s", headExpected, head.ID)
	}
	// Tail must be 003
	tail := tasks[len(tasks)-1]
	tailExpected := "taskID-find-003"
	if tail.ID != tailExpected {
		t.Errorf("tail ID must be %s, but got %s", tailExpected, tail.ID)
	}
}

func TestAddressRepository_CountTasks(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	expected := 5
	firstTasks := createTaskTestData(tx, "taskID-find", "findDescription", 5)
	secondTasks := createTaskTestData(tx, "taskID-not-find", "notFindDescription", 4)
	insertTasks := append(firstTasks, secondTasks...)
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	count, err := repo.CountTasks(&model.Task{Description: "findDescription"})
	if err != nil {
		t.Fatalf("failed to count Task: %+v", err)
	}
	if expected != count {
		t.Errorf("expected %d records, but got %d record", expected, count)
	}
}

func TestTaskRepository_CreateTask(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	// Create 1 record
	insertTasks := createTaskTestData(tx, "taskID-create", "createDescription", 1)
	created := insertTasks[0]
	if err := repo.CreateTask(created); err != nil {
		t.Fatalf("Failed to create task: %+v", err)
	}

	// Find by ID
	var find = model.Task{}
	if err := tx.Where(&model.Task{ID: created.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find task: %+v", err)
	}
	// Verify created equals find
	if !assert.Equal(t, find, *created) {
		t.Errorf("expected: %+v, but got %+v", created.ID, find.ID)
	}
}

func TestTaskRepository_UpdateTask(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	// Create 1 record
	insertTasks := createTaskTestData(tx, "taskID-create", "createDescription", 1)
	created := insertTasks[0]
	if err := repo.CreateTask(created); err != nil {
		t.Fatalf("Failed to create task: %+v", err)
	}

	// Update the record
	updated := insertTasks[0]
	updated.Description = "updatedDescription"
	if err := repo.UpdateTask(updated); err != nil {
		t.Fatalf("failed to update: %+v", err)
	}

	// Find by ID
	var find = model.Task{}
	if err := tx.Where(&model.Task{ID: updated.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find task: %+v", err)
	}

	// Verify updated equals find
	if !assert.Equal(t, find, *updated) {
		t.Errorf("expected %v, but got %v", find, updated)
	}
}

func TestTaskRepository_DeleteTask(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	// Create 1 record
	insertTasks := createTaskTestData(tx, "taskId-delete", "deleteDescription", 1)
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to create Task: %+v", err)
	}
	deleted := insertTasks[0]
	t.Run("Record will not be deleted if ID is empty", func(t *testing.T) {
		deletedID := deleted.ID
		deleted.ID = "" // Clear ID
		if err := repo.DeleteTask(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}
		// record must exist
		result := model.Task{}
		if err := tx.Where(&model.Task{ID: deletedID}).Find(&result).Error; err != nil {
			t.Errorf("Failed to find: %+v", err)
		}
		result.ID = deletedID
		deleted.ID = deletedID // Restore ID

		// Verify not deleted
		if !assert.Equal(t, *deleted, result) {
			t.Error("Record is not same")
		}
	})

	t.Run("Record will be deleted if ID is set", func(t *testing.T) {
		if err := repo.DeleteTask(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}

		// Check record will not be found
		result := model.Task{}
		err := tx.Where(&model.Task{ID: deleted.ID}).First(&result).Error
		if !orm.IsRecordNotFoundError(err) {
			t.Errorf("Record must be deleted, but got: %+v", err)
		}
	})
}

// Test for CreateTasks, UpdateTasks, DeleteTasks are ommitted,
// because that they are called internally in each single version

////
/// Optimistic lock test (if version lock supported)
//
func TestTaskRepository_UpdateTaskOptimisticCheck(t *testing.T) {
	tx1, _ := newTxAndTaskRepository()
	tx2, repo2 := newTxAndTaskRepository()
	tx3, repo3 := newTxAndTaskRepository()
	defer tx1.Rollback()
	defer tx2.Rollback()
	defer tx3.Rollback()

	// Create and commit 1 record in tx1
	insertTasks := createTaskTestData(tx1, "taskID-optimistic", "", 1)
	err := insertTaskTestData(tx1, insertTasks)
	if err != nil {
		t.Fatalf("Failed to create task: %+v", err)
	}
	if err := tx1.Commit().Error; err != nil {
		t.Fatalf("Failed to commit data in tx1")
	}
	// Get commited data in tx2
	data := insertTasks[0]
	find, err := repo2.FindFirstTask(model.Task{ID: data.ID}, []string{})
	if err != nil {
		deleteCommitedTaskData(t, data)
	}
	find.Description = "UpdateInTx2"
	data.Description = "NotUpdateInTx3"
	// Find updated in tx2 (Version number incremented!!)
	err = repo2.UpdateTask(&find)
	if err != nil {
		deleteCommitedTaskData(t, data)
		t.Fatalf("Failed to update in tx2.")
	}
	err = tx2.Commit().Error
	if err != nil {
		deleteCommitedTaskData(t, data)
		t.Fatalf("Failed to commit data in tx2")
	}
	// Check to not update due to optimistic error
	if !assert.Error(t, repo3.UpdateTask(data)) {
		deleteCommitedTaskData(t, data)
		t.Fatalf("Not failed to update, No error occurred")
	}
	err = tx3.Commit().Error
	if err != nil {
		deleteCommitedTaskData(t, data)
		t.Fatalf("Failed to Commit tx3, no affected row")
	}

	tx4, _ := newTxAndTaskRepository()
	defer tx4.Rollback()
	var result = model.Task{}
	// Can be found as same as find(tx2)
	if err := tx4.Where(model.Task{ID: find.ID}).Find(&result).Error; err != nil {
		t.Fatalf("failed to retrieve Task: %+v", err)
	}
	deleteCommitedTaskData(t, data)
	if !assert.Equal(t, find, result) {
		t.Errorf("expected %v, but got %v", find, result)
	}
}

func deleteCommitedTaskData(t *testing.T, data *model.Task) {
	// Try to delete data in another transaction
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()
	err := repo.DeleteTask(data)
	if err != nil {
		t.Log("*** Failed to delete test data, please restart test. ***")
	} else {
		err := tx.Commit().Error
		if err != nil {
			t.Log("*** Failed to delete test data, please restart test. ***")
		}
	}
}

////
/// Other fuctions' test should be written in below
//
func TestTaskRepository_MaxTaskDispOrder(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	// Create 10 + 20 records
	firstTasks := createTaskTestData(tx, "taskID-order1st", "orderDescription", 10)
	for _, task := range firstTasks {
		task.BoardID = "firstBoardID"
	}
	secondTasks := createTaskTestData(tx, "taskID-order2nd", "orderDescription", 20)
	for _, task := range secondTasks {
		task.BoardID = "secondBoardID"
	}
	insertTasks := append(firstTasks, secondTasks...)
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to create tasks: %+v", err)
	}
	max, err := repo.MaxTaskDispOrder(&model.Task{BoardID: "firstBoardID"})
	if err != nil {
		t.Fatalf("Failed to get max of dispOrder of tasks: %+v", err)
	}
	// Verify created equals find
	if !assert.Equal(t, 10, max) {
		t.Errorf("expected: %d, but got %d", 10, max)
	}
}

func TestTaskRepository_MoveToIceboxBoard(t *testing.T) {
	tx, repo := newTxAndTaskRepository()
	defer tx.Rollback()

	// Create 2 records
	insertTasks := createTaskTestData(tx, "taskID-move", "moveToIceboxDescription", 3)
	insertTasks[0].BoardID = "firstBoardID"
	insertTasks[1].BoardID = "secondBoardID"
	insertTasks[2].BoardID = "firstBoardID"
	err := insertTaskTestData(tx, insertTasks)
	if err != nil {
		t.Fatalf("Failed to create tasks: %+v", err)
	}
	err = repo.MoveToIceboxBoard("firstBoardID")
	if err != nil {
		t.Fatalf("Failed to move tasks to IcebboxBoard: %+v", err)
	}
	findTasks, err := repo.FindTasks(&model.Task{Description: "moveToIceboxDescription"},
		0, orm.NoLimit, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to find tasks: %+v", err)
	}
	// Verify created equals find
	if len(findTasks) != 3 {
		t.Fatalf("")
	}
	// 0 and 2 will be changed.
	insertTasks[0].BoardID = model.SystemBoardIcebox.ID
	insertTasks[2].BoardID = model.SystemBoardIcebox.ID
	assert.Equal(t, *insertTasks[0], findTasks[0])
	assert.Equal(t, *insertTasks[1], findTasks[1])
	assert.Equal(t, *insertTasks[2], findTasks[2])
}
