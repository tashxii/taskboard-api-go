package repository

import (
	"fmt"
	"taskboard-api-go/common"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

////
/// Model specific functions (Only replace model name, take care names are casesencitive!!)
//
func newTxAndUserRepository() (tx *gorm.DB, repo *UserRepository) {
	tx = orm.GetDB().Begin()
	repo = NewUserRepository(tx)
	return
}

func createUserTestData(tx *gorm.DB, idFormat string, findIdentify string, count int) []*model.User {
	result := make([]*model.User, 0, count)
	for i := 0; i < count; i++ {
		user := model.NewUser(
			"name"+common.GenerateID(),
			fmt.Sprintf("password-%03d", i),
			findIdentify,
		)
		user.ID = fmt.Sprintf("%s-%03d", idFormat, i)
		result = append(result, user)
	}
	return result
}

func insertUserTestData(tx *gorm.DB, users []*model.User) (err error) {
	for _, user := range users {
		err = tx.Create(user).Error
		if err != nil {
			return
		}
	}
	return
}

////
/// Common repository functions' test
//
func TestUserRepository_FindFirstUser(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	firstUsers := createUserTestData(tx, "userID-find", "findAvatar", 5)
	secondUsers := createUserTestData(tx, "userID-not-find", "notFindAvatar", 4)
	insertUsers := append(firstUsers, secondUsers...)
	err := insertUserTestData(tx, insertUsers)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}
	// Sort by id
	user, err := repo.FindFirstUser(&model.User{Avatar: "findAvatar"}, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find first: %+v", err)
	}
	expected := "userID-find-000"
	if user.ID != expected {
		t.Errorf("expected user ID is %s, but got %s", expected, user.ID)
	}
}

func TestUserRepository_FindUsers(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	firstUsers := createUserTestData(tx, "userID-find", "findAvatar", 5)
	secondUsers := createUserTestData(tx, "userID-not-find", "notFindAvatar", 4)
	insertUsers := append(firstUsers, secondUsers...)
	err := insertUserTestData(tx, insertUsers)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	// Get 3 data from 2nd(index=1) position
	const (
		offset = 1
		limit  = 3
	)
	users, err := repo.FindUsers(&model.User{Avatar: "findAvatar"}, offset, limit, []string{"id"})
	if err != nil {
		t.Fatalf("Failed to execute find: %+v", err)
	}

	// Length of data must be limit
	if len(users) != limit {
		t.Errorf("Expected result size = %d, but got %d", limit, len(users))
		return
	}
	// Head must be 001
	head := users[0]
	headExpected := "userID-find-001"
	if head.ID != headExpected {
		t.Errorf("head ID must be %s, but got %s", headExpected, head.ID)
	}
	// Tail must be 003
	tail := users[len(users)-1]
	tailExpected := "userID-find-003"
	if tail.ID != tailExpected {
		t.Errorf("tail ID must be %s, but got %s", tailExpected, tail.ID)
	}
}

func TestAddressRepository_CountUsers(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	expected := 5
	firstUsers := createUserTestData(tx, "userID-find", "findAvatar", 5)
	secondUsers := createUserTestData(tx, "userID-not-find", "notFindAvatar", 4)
	insertUsers := append(firstUsers, secondUsers...)
	err := insertUserTestData(tx, insertUsers)
	if err != nil {
		t.Fatalf("Failed to insert test data: %+v", err)
	}

	count, err := repo.CountUsers(&model.User{Avatar: "findAvatar"})
	if err != nil {
		t.Fatalf("failed to count User: %+v", err)
	}
	if expected != count {
		t.Errorf("expected %d records, but got %d record", expected, count)
	}
}

func TestUserRepository_CreateUser(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	// Create 1 record
	insertUsers := createUserTestData(tx, "userID-create", "createAvatar", 1)
	created := insertUsers[0]
	if err := repo.CreateUser(created); err != nil {
		t.Fatalf("Failed to create user: %+v", err)
	}

	// Find by ID
	var find = model.User{}
	if err := tx.Where(&model.User{ID: created.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find user: %+v", err)
	}
	// Verify created equals find
	if !assert.Equal(t, find, *created) {
		t.Errorf("expected: %+v, but got %+v", created.ID, find.ID)
	}
}

func TestUserRepository_UpdateUser(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	// Create 1 record
	insertUsers := createUserTestData(tx, "userID-create", "createAvatar", 1)
	created := insertUsers[0]
	if err := repo.CreateUser(created); err != nil {
		t.Fatalf("Failed to create user: %+v", err)
	}

	// Update the record
	updated := insertUsers[0]
	updated.Avatar = "updatedAvatar"
	if err := repo.UpdateUser(updated); err != nil {
		t.Fatalf("failed to update: %+v", err)
	}

	// Find by ID
	var find = model.User{}
	if err := tx.Where(&model.User{ID: updated.ID}).First(&find).Error; err != nil {
		t.Fatalf("Failed to find user: %+v", err)
	}

	// Verify updated equals find
	if !assert.Equal(t, find, *updated) {
		t.Errorf("expected %v, but got %v", find, updated)
	}
}

func TestUserRepository_DeleteUser(t *testing.T) {
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()

	// Create 1 record
	insertUsers := createUserTestData(tx, "userId-delete", "deleteAvatar", 1)
	err := insertUserTestData(tx, insertUsers)
	if err != nil {
		t.Fatalf("Failed to create User: %+v", err)
	}
	deleted := insertUsers[0]
	t.Run("Record will not be deleted if ID is empty", func(t *testing.T) {
		deletedID := deleted.ID
		deleted.ID = "" // Clear ID
		if err := repo.DeleteUser(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}
		// record must exist
		result := model.User{}
		if err := tx.Where(&model.User{ID: deletedID}).Find(&result).Error; err != nil {
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
		if err := repo.DeleteUser(deleted); err != nil {
			t.Errorf("Failed to delete: %+v", err)
		}

		// Check record will not be found
		result := model.User{}
		err := tx.Where(&model.User{ID: deleted.ID}).First(&result).Error
		if !orm.IsRecordNotFoundError(err) {
			t.Errorf("Record must be deleted, but got: %+v", err)
		}
	})
}

// Test for CreateUsers, UpdateUsers, DeleteUsers are ommitted,
// because that they are called internally in each single version

////
/// Optimistic lock test (if version lock supported)
//
func TestUserRepository_UpdateUserOptimisticCheck(t *testing.T) {
	tx1, _ := newTxAndUserRepository()
	tx2, repo2 := newTxAndUserRepository()
	tx3, repo3 := newTxAndUserRepository()
	defer tx1.Rollback()
	defer tx2.Rollback()
	defer tx3.Rollback()

	// Create and commit 1 record in tx1
	insertUsers := createUserTestData(tx1, "userID-optimistic", "", 1)
	err := insertUserTestData(tx1, insertUsers)
	if err != nil {
		t.Fatalf("Failed to create user: %+v", err)
	}
	if err := tx1.Commit().Error; err != nil {
		t.Fatalf("Failed to commit data in tx1")
	}
	// Get commited data in tx2
	data := insertUsers[0]
	find, err := repo2.FindFirstUser(model.User{ID: data.ID}, []string{})
	if err != nil {
		deleteCommitedUserData(t, data)
	}
	find.Avatar = "UpdateInTx2"
	data.Avatar = "NotUpdateInTx3"
	// Find updated in tx2 (Version number incremented!!)
	err = repo2.UpdateUser(&find)
	if err != nil {
		deleteCommitedUserData(t, data)
		t.Fatalf("Failed to update in tx2.")
	}
	err = tx2.Commit().Error
	if err != nil {
		deleteCommitedUserData(t, data)
		t.Fatalf("Failed to commit data in tx2")
	}
	// Check to not update due to optimistic error
	if !assert.Error(t, repo3.UpdateUser(data)) {
		deleteCommitedUserData(t, data)
		t.Fatalf("Not failed to update, No error occurred")
	}
	err = tx3.Commit().Error
	if err != nil {
		deleteCommitedUserData(t, data)
		t.Fatalf("Failed to Commit tx3, no affected row")
	}

	tx4, _ := newTxAndUserRepository()
	defer tx4.Rollback()
	var result = model.User{}
	// Can be found as same as find(tx2)
	if err := tx4.Where(model.User{ID: find.ID}).Find(&result).Error; err != nil {
		t.Fatalf("failed to retrieve User: %+v", err)
	}
	deleteCommitedUserData(t, data)
	if !assert.Equal(t, find, result) {
		t.Errorf("expected %v, but got %v", find, result)
	}
}

func deleteCommitedUserData(t *testing.T, data *model.User) {
	// Try to delete data in another transaction
	tx, repo := newTxAndUserRepository()
	defer tx.Rollback()
	err := repo.DeleteUser(data)
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
