package model

import (
	"database/sql"
	"taskboard-api-go/common"
	"time"
)

// Task present task of the app.
type Task struct {
	ID             string         `gorm:"primary_key;size:32"`
	Name           string         `gorm:"not null;size:255"`
	Description    string         `gorm:"size:8000"`
	AssigneeUserID sql.NullString `gorm:"size:32"`          // Null or String
	BoardID        string         `gorm:"not null;size:32"` // Default is IceboxBoardID
	DispOrder      int            `gorm:"not null"`
	CreatedDate    time.Time      `gorm:"not null"`
	IsClosed       bool           `gorm:"not null"`
	Version        int            `gorm:"not null"` // Version for optimistic lock
	EstimateSize   int
}

// NewTask returns created new task
func NewTask(name, description string, isClosed bool, now time.Time) *Task {
	return &Task{
		ID:             "task_" + common.GenerateID(),
		Name:           name,
		Description:    description,
		IsClosed:       isClosed,
		BoardID:        SystemBoardIcebox.ID,
		AssigneeUserID: sql.NullString{Valid: false},
		DispOrder:      0,
		CreatedDate:    now,
		Version:        1,
	}
}

// SetAssigneeUserID updates assigneeUserID by specifed value if it is not empty
func (t *Task) SetAssigneeUserID(assigneeUserID string) {
	if assigneeUserID != "" {
		// Update only if not empty
		t.AssigneeUserID = sql.NullString{String: assigneeUserID, Valid: true}
	}
}

// SetBoardID updates boardID by specifed value if it is not empty
func (t *Task) SetBoardID(boardID string) {
	if boardID != "" {
		// Update only if not empty
		t.BoardID = boardID
	}
}
