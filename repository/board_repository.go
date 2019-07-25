package repository

import (
	"sync"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"

	"github.com/jinzhu/gorm"
)

var lockBoard = &sync.Mutex{}

// BoardRepository is repository of board table
type BoardRepository struct {
	tx *gorm.DB
}

// NewBoardRepository returns new instance of BoardRepository
func NewBoardRepository(tx *gorm.DB) *BoardRepository {
	if tx == nil {
		// Programing error!!
		panic("tx must be set")
	}
	return &BoardRepository{
		tx: tx,
	}
}

// FindFirstBoard returns first Board matching with specified condition
func (repo *BoardRepository) FindFirstBoard(condition interface{}, sortOrders []string) (result model.Board, err error) {
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

// FindBoards returns Boards matching with specified condition
func (repo *BoardRepository) FindBoards(condition interface{}, offset int, limit int, sortOrders []string) (result []model.Board, err error) {
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
	for _, board := range sortOrders {
		query = query.Order(board)
	}

	err = query.Find(&result).Error
	return
}

// CountBoards returns the number of Boards matching specfied condition
func (repo *BoardRepository) CountBoards(condition interface{}) (count int, err error) {
	var boards []model.Board
	err = repo.tx.Where(condition).Find(&boards).Count(&count).Error
	return
}

// CreateBoard inserts new Board record
func (repo *BoardRepository) CreateBoard(board *model.Board) error {
	return repo.CreateBoards([]*model.Board{board})
}

// UpdateBoard updates Board record
func (repo *BoardRepository) UpdateBoard(board *model.Board) error {
	return repo.UpdateBoards([]*model.Board{board})
}

// DeleteBoard deletes Board record
func (repo *BoardRepository) DeleteBoard(board *model.Board) error {
	return repo.DeleteBoards([]*model.Board{board})
}

// CreateBoards inserts new Board records.
func (repo *BoardRepository) CreateBoards(boards []*model.Board) (err error) {
	lockBoard.Lock()
	defer lockBoard.Unlock()

	count, err := repo.CountBoards(&model.Board{})
	if err != nil {
		return
	}
	for _, board := range boards {
		err = repo.tx.Create(board).Error
		count++
		board.DispOrder = count
		if err != nil {
			return
		}
	}
	return
}

// UpdateBoards updates board records
func (repo *BoardRepository) UpdateBoards(boards []*model.Board) (err error) {
	lockBoard.Lock()
	defer lockBoard.Unlock()

	for _, board := range boards {
		oldVersion := board.Version
		board.Version++
		db := repo.tx.Model(&model.Board{}).Where("version = ?", oldVersion).Updates(board)
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

// DeleteBoards deletes Board records
func (repo *BoardRepository) DeleteBoards(boards []*model.Board) (err error) {
	for _, board := range boards {
		if board.ID == "" {
			continue // To avoid deleting all due to gorm warning, continue here.
		}
		err = repo.tx.Delete(board).Error
		if err != nil {
			return
		}
	}
	return
}

// UpdateBoardOrders changes display orders of boards
func (repo *BoardRepository) UpdateBoardOrders(boardIDs []string) (err error) {
	for i, boardID := range boardIDs {
		err = repo.tx.Model(&model.Board{}).Where("id = ?", boardID).
			Update(model.Board{DispOrder: i}).Error
		if err != nil {
			return
		}
	}
	return nil
}
