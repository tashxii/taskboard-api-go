package repository_test

import (
	"fmt"
	"os"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"testing"
)

func TestMain(m *testing.M) {
	// Check test database file exits or not
	testDbFile := "./repository_test.sqlite3"
	_, err := os.Stat(testDbFile)
	if !os.IsNotExist(err) {
		// Try to remove
		err = os.Remove(testDbFile)
		if err != nil {
			fmt.Printf("Failed to remove [%s]\n", testDbFile)
		}
		_, err := os.Stat(testDbFile)
		if !os.IsNotExist(err) {
			// Still exits, fail...
			fmt.Printf("Test db file [%s] exists, please remove it before executing test\n", testDbFile)
			os.Exit(1)
		}
	}

	// Prepare test database file
	err = orm.Init(testDbFile)
	if err != nil {
		fmt.Printf("Failed to init test db file [%s]\n", testDbFile)
		os.Exit(1)
	}

	// Create tables
	err = orm.Migrate(
		&model.User{},
		&model.Task{},
		&model.Board{},
	)
	if err != nil {
		fmt.Printf("Failed to create tables: %+v\n", err)
		err := orm.GetDB().Close()
		if err != nil {
			fmt.Printf("Failed to close database: %+v", err)
		}
		err = os.Remove(testDbFile)
		if err != nil {
			fmt.Printf("Failed to remove [%s] err:%+v\n", testDbFile, err)
		}
		os.Exit(1)
	}

	// Execute test
	ret := m.Run()

	err = orm.GetDB().Close()
	if err != nil {
		fmt.Printf("Failed to close database: %+v\n", err)
	}
	if ret != 0 {
		fmt.Printf("Test failed, the database file [%s] is kept for investigation\n", testDbFile)
	} else {
		err = os.Remove(testDbFile)
		if err != nil {
			fmt.Printf("Failed to remove [%s]\n", testDbFile)
		}
	}
	os.Exit(ret)
}
