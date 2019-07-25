package api

import (
	"fmt"
	"taskboard-api-go/service"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Commit executes commit transaction
func Commit(tx *gorm.DB) error {
	err := tx.Commit().Error
	if err != nil {
		return service.NewDBCommitError(err)
	}
	return nil
}

// Rollback executes rollback transaction, only logging even if an error occurred
func Rollback(tx *gorm.DB) {
	err := tx.Rollback().Error
	if err != nil {
		err = errors.WithStack(err)
		fmt.Printf("Failed to rollback transaction Error:%+v\n", err)
	}
}
