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
func newTxAndBoardRepository() (tx *gorm.DB, repo *BoardRepository) {
	tx = orm.GetDB().Begin()
	repo = NewBoardRepository(tx)
	return
}

func createBoardTestData(tx *gorm.DB, idFormat string, findIdentify bool, count int) []*model.Board {
	result := make([]*model.Board, 0, count)
	for i := 0; i < count; i++ {
		board := model.NewBoard(
			"name"+common.GenerateID(),
			findIdentify,
			false,
			time.Now().UTC(),
		)
		board.ID = fmt.Sprintf("%s-%03d", idFormat, i)
		result = append(result, board)
	}
	return result
}

func insertBoardTestData(tx *gorm.DB, boards []*model.Board) (err error) {
	for _, board := range boards {
		err = tx.Create(board).Error
		if err != nil {
			return
		}
	}
	return
}

////
/// Common repository functions' test
//
func TestBoardRepository_FindFirstBoard(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	firstBoards := createBoardTestData(tx, "boardID-find", true, 5)
	secondBoards := createBoardTestData(tx, "boardID-not-find", false, 4)
	insertBoards := append(firstBoards, secondBoards...)
	err := insertBoardTestData(tx, insertBoards)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}
	// Sort by id
	board, err := repo.FindFirstBoard(&model.Board{IsSystem: true}, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find first: %+v", err)
	}
	expected := "boardID-find-000"
	if board.ID != expected {
		t.Errorf("expected board ID is %s, but got %s", expected, board.ID)
	}
}

func TestBoardRepository_FindBoards(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	firstBoards := createBoardTestData(tx, "boardID-find", true, 5)
	secondBoards := createBoardTestData(tx, "boardID-not-find", false, 4)
	insertBoards := append(firstBoards, secondBoards...)
	err := insertBoardTestData(tx, insertBoards)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	// Get 3 data from 2nd(index=1) position
	const (
		offset = 1
		limit  = 3
	)
	boards, err := repo.FindBoards(&model.Board{IsSystem: true}, offset, limit, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find: %+v", err)
	}

	// Length of data must be limit
	if len(boards) != limit {
		t.Errorf("Expected result size = %d, but got %d", limit, len(boards))
		return
	}
	// Head must be 001
	head := boards[0]
	headExpected := "boardID-find-001"
	if head.ID != headExpected {
		t.Errorf("head ID must be %s, but got %s", headExpected, head.ID)
	}
	// Tail must be 003
	tail := boards[len(boards)-1]
	tailExpected := "boardID-find-003"
	if tail.ID != tailExpected {
		t.Errorf("tail ID must be %s, but got %s", tailExpected, tail.ID)
	}
}

func TestAddressRepository_CountBoards(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	expected := 5
	firstBoards := createBoardTestData(tx, "boardID-find", true, 5)
	secondBoards := createBoardTestData(tx, "boardID-not-find", false, 4)
	insertBoards := append(firstBoards, secondBoards...)
	err := insertBoardTestData(tx, insertBoards)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	count, err := repo.CountBoards(&model.Board{IsSystem: true})
	if err != nil {
		t.Fatalf("failed to count Board: %+v", err)
	}
	if expected != count {
		t.Errorf("expected %d records, but got %d record", expected, count)
	}
}

func TestBoardRepository_CreateBoard(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	// Create 1 record
	insertBoards := createBoardTestData(tx, "boardID-create", true, 1)
	created := insertBoards[0]
	if err := repo.CreateBoard(created); err != nil {
		t.Fatalf("Failed to create board: %+v", err)
	}

	// Find by ID
	var find = model.Board{}
	if err := tx.Where(&model.Board{ID: created.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find board: %+v", err)
	}
	// Verify created equals find
	if !assert.Equal(t, find, *created) {
		t.Errorf("expected: %+v, but got %+v", created.ID, find.ID)
	}
}

func TestBoardRepository_UpdateBoard(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	// Create 1 record
	insertBoards := createBoardTestData(tx, "boardID-create", true, 1)
	created := insertBoards[0]
	if err := repo.CreateBoard(created); err != nil {
		t.Fatalf("Failed to create board: %+v", err)
	}

	// Update the record
	updated := insertBoards[0]
	updated.IsSystem = true
	if err := repo.UpdateBoard(updated); err != nil {
		t.Fatalf("failed to update: %+v", err)
	}

	// Find by ID
	var find = model.Board{}
	if err := tx.Where(&model.Board{ID: updated.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find board: %+v", err)
	}

	// Verify updated equals find
	if !assert.Equal(t, find, *updated) {
		t.Errorf("expected %v, but got %v", find, updated)
	}
}

func TestBoardRepository_DeleteBoard(t *testing.T) {
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()

	// Create 1 record
	insertBoards := createBoardTestData(tx, "boardId-delete", true, 1)
	err := insertBoardTestData(tx, insertBoards)
	if err != nil {
		t.Fatalf("Failed to create Board: %+v", err)
	}
	deleted := insertBoards[0]
	t.Run("Record will not be deleted if ID is empty", func(t *testing.T) {
		deletedID := deleted.ID
		deleted.ID = "" // Clear ID
		if err := repo.DeleteBoard(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}
		// record must exist
		result := model.Board{}
		if err := tx.Where(&model.Board{ID: deletedID}).Find(&result).Error; err != nil {
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
		if err := repo.DeleteBoard(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}

		// Check record will not be found
		result := model.Board{}
		err := tx.Where(&model.Board{ID: deleted.ID}).First(&result).Error
		if !orm.IsRecordNotFoundError(err) {
			t.Errorf("Record must be deleted, but got: %+v", err)
		}
	})
}

// Test for CreateBoards, UpdateBoards, DeleteBoards are ommitted,
// because that they are called internally in each single version

////
/// Optimistic lock test (if version lock supported)
//
func TestBoardRepository_UpdateBoardOptimisticCheck(t *testing.T) {
	tx1, _ := newTxAndBoardRepository()
	tx2, repo2 := newTxAndBoardRepository()
	tx3, repo3 := newTxAndBoardRepository()
	defer tx1.Rollback()
	defer tx2.Rollback()
	defer tx3.Rollback()

	// Create and commit 1 record in tx1
	insertBoards := createBoardTestData(tx1, "boardID-optimistic", true, 1)
	err := insertBoardTestData(tx1, insertBoards)
	if err != nil {
		t.Fatalf("Failed to create board: %+v", err)
	}
	if err := tx1.Commit().Error; err != nil {
		t.Fatalf("Failed to commit data in tx1")
	}
	// Get commited data in tx2
	data := insertBoards[0]
	find, err := repo2.FindFirstBoard(model.Board{ID: data.ID}, []string{})
	if err != nil {
		deleteCommitedBoardData(t, data)
	}
	find.Name = "updateInTx2"
	data.Name = "notUpdateInTx3"
	// Find updated in tx2 (Version number incremented!!)
	err = repo2.UpdateBoard(&find)
	if err != nil {
		deleteCommitedBoardData(t, data)
		t.Fatalf("Failed to update in tx2.")
	}
	err = tx2.Commit().Error
	if err != nil {
		deleteCommitedBoardData(t, data)
		t.Fatalf("Failed to commit data in tx2")
	}
	// Check to not update due to optimistic error
	if !assert.Error(t, repo3.UpdateBoard(data)) {
		deleteCommitedBoardData(t, data)
		t.Fatalf("Not failed to update, No error occurred")
	}
	err = tx3.Commit().Error
	if err != nil {
		deleteCommitedBoardData(t, data)
		t.Fatalf("Failed to Commit tx3, no affected row")
	}

	tx4, _ := newTxAndBoardRepository()
	defer tx4.Rollback()
	var result = model.Board{}
	// Can be found as same as find(tx2)
	if err := tx4.Where(model.Board{ID: find.ID}).Find(&result).Error; err != nil {
		t.Fatalf("failed to retrieve Board: %+v", err)
	}
	deleteCommitedBoardData(t, data)
	if !assert.Equal(t, find, result) {
		t.Errorf("expected %v, but got %v", find, result)
	}
}

func deleteCommitedBoardData(t *testing.T, data *model.Board) {
	// Try to delete data in another transaction
	tx, repo := newTxAndBoardRepository()
	defer tx.Rollback()
	err := repo.DeleteBoard(data)
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
