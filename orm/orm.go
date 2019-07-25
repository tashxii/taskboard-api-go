package orm

import (
	"errors"
	"math"
	"reflect"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" //to load dialect
)

var instance *gorm.DB

// NoLimit is literal of unlimited search
const NoLimit = math.MaxInt32

// ErrorRecordNotFound is an error when record not found
var ErrorRecordNotFound = gorm.ErrRecordNotFound

// Init opens database
func Init(databasePath string) (err error) {
	opened, err := gorm.Open("sqlite3", databasePath)
	if err != nil {
		return
	}
	instance = opened
	instance.DB().SetMaxOpenConns(5)
	instance.DB().SetMaxIdleConns(5)
	return
}

// GetDB returns opend database of gorm
func GetDB() *gorm.DB {
	return instance
}

// Migrate create tables of models
func Migrate(models ...interface{}) error {
	db := GetDB()
	if db == nil {
		return errors.New("Database is not initilazed. Must call orm.Init() before calling this method")
	}
	return db.AutoMigrate(models...).Error
}

// TableName rerturns table name of model
func TableName(model interface{}) string {
	return gorm.ToTableName(reflect.TypeOf(model).Name())
}

// ColumnNames returns column names of model
func ColumnNames(model interface{}) []string {
	result := make([]string, 0)
	t := reflect.TypeOf(model)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() != reflect.Struct {
			name := f.Name
			result = append(result, gorm.ToColumnName(name))
		}
	}
	return result
}

// IsRecordNotFoundError checks whether err is due to no record found
func IsRecordNotFoundError(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}
