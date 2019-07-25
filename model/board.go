package model

import (
	"taskboard-api-go/common"
	"time"
)

// Board presents a board which has plural tasks
type Board struct {
	ID          string    `gorm:"primary_key;size:32"`
	Name        string    `gorm:"unique;size:255"`
	DispOrder   int       `gorm:"not null"`
	IsSystem    bool      `gorm:"not null"`
	IsClosed    bool      `gorm:"not null"`
	CreatedDate time.Time `gorm:"not null"`
	Version     int       `gorm:"not null"` // Version for optimistic lock
}

// SystemBoardIcebox is a system board
var SystemBoardIcebox = &Board{
	ID:          "board_icebox",
	Name:        "Icebox",
	DispOrder:   0,
	IsSystem:    true,
	CreatedDate: time.Now().UTC(),
	Version:     1,
}

// SystemBoardTodo is a system board id
var SystemBoardTodo = &Board{
	ID:          "board_todo",
	Name:        "Todo",
	DispOrder:   1,
	IsSystem:    true,
	CreatedDate: time.Now().UTC(),
	Version:     1,
}

// SystemBoardDoing is a system board id
var SystemBoardDoing = &Board{
	ID:          "board_doing",
	Name:        "Doing",
	DispOrder:   2,
	IsSystem:    true,
	CreatedDate: time.Now().UTC(),
	Version:     1,
}

// SystemBoardDone is a system board id
var SystemBoardDone = &Board{
	ID:          "board_done",
	Name:        "Done",
	DispOrder:   3,
	IsSystem:    true,
	CreatedDate: time.Now().UTC(),
	Version:     1,
}

// NewBoard returns created new board
func NewBoard(name string, isSystem, isClosed bool, now time.Time) *Board {
	return &Board{
		ID:          "board_" + common.GenerateID(),
		Name:        name,
		DispOrder:   0,
		IsSystem:    isSystem,
		IsClosed:    isClosed,
		CreatedDate: now,
		Version:     1,
	}
}
