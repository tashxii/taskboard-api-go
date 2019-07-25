package tasks

import (
	"database/sql"
	"taskboard-api-go/model"
	"taskboard-api-go/service"
	"time"

	"github.com/gin-gonic/gin"
)

// ID             string         `gorm:"primary_key;size:32"`
// Name           string         `gorm:"not null;size:255"`
// Description    string         `gorm:"size:8000"`
// AssigneeUserID sql.NullString `gorm:"size:32"`           // Null or String
// BoardID        string         `gorm:"not null; size:32"` // Default is IceboxBoardID
// DispOrder      int            `gorm:"not null"`
// CreatedDate    time.Time      `gorm:"not null"`
// IsClosed       bool           `gorm:"not null"`
// Version        int            `gorm:"not null"` // Version for optimistic lock
// EstimateSize   int

type taskResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AssigneeUserID string `json:"assigneeUserId"`
	BoardID        string `json:"boardId"`
	DispOrder      int    `json:"dispOrder"`
	CreatedDate    string `json:"createDate"`
	IsClosed       bool   `json:"isClosed"`
	Version        int    `json:"version"`
	EstimateSize   int    `json:"estimateSize"`
}

type createRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	AssigneeUserID string `json:"assigneeUserId"`
	BoardID        string `json:"boardId"`
	IsClosed       bool   `json:"isClosed"`
	EstimateSize   int    `json:"estimateSize"`
}

type updateRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AssigneeUserID string `json:"assigneeUserId"`
	BoardID        string `json:"boardId"`
	IsClosed       bool   `json:"isClosed"`
	Version        int    `json:"version"`
	EstimateSize   int    `json:"estimateSize"`
}

type updateTaskOrdersRequest struct {
	TaskID        string `json:"taskId"`
	FromBoardID   string `json:"fromBoardId"`
	FromDispOrder int    `json:"fromDispOrder"`
	ToBoardID     string `json:"toBoardId"`
	ToDispOrder   int    `json:"toDispOrder"`
}

func convertTaskResponse(task *model.Task) *taskResponse {
	return &taskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		AssigneeUserID: task.AssigneeUserID.String,
		BoardID:        task.BoardID,
		DispOrder:      task.DispOrder,
		CreatedDate:    task.CreatedDate.Format(time.RFC3339),
		IsClosed:       task.IsClosed,
		Version:        task.Version,
		EstimateSize:   task.EstimateSize,
	}
}

func convertListTaskResponse(tasks []model.Task) (res []*taskResponse) {
	res = make([]*taskResponse, 0, len(tasks))
	for _, task := range tasks {
		res = append(res, convertTaskResponse(&task))
	}
	return
}

func getTaskByCreateRequest(c *gin.Context) (*model.Task, error) {
	var req createRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return nil, service.NewBadRequestError(err)
	}
	task := model.NewTask(
		req.Name,
		req.Description,
		req.IsClosed,
		time.Now().UTC(),
	)
	task.SetAssigneeUserID(req.AssigneeUserID)
	task.SetBoardID(req.BoardID)
	task.EstimateSize = req.EstimateSize
	return task, nil
}

func getTaskByUpdateRequest(c *gin.Context, find *model.Task) (*model.Task, error) {
	var req updateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return nil, service.NewBadRequestError(err)
	}
	newAssigneeUserID := sql.NullString{}
	if req.AssigneeUserID != "" {
		// Set only if not empty
		newAssigneeUserID.String = req.AssigneeUserID
		newAssigneeUserID.Valid = true
	}
	task := &model.Task{
		ID:             find.ID,
		Name:           req.Name,
		Description:    req.Description,
		AssigneeUserID: newAssigneeUserID,
		BoardID:        req.BoardID,
		DispOrder:      find.DispOrder,
		CreatedDate:    find.CreatedDate,
		IsClosed:       req.IsClosed,
		Version:        req.Version,
		EstimateSize:   req.EstimateSize,
	}
	task.SetAssigneeUserID(req.AssigneeUserID)
	return task, nil
}

func getUpdateTaskOrdersRequest(c *gin.Context) (*updateTaskOrdersRequest, error) {
	var req updateTaskOrdersRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return nil, service.NewBadRequestError(err)
	}
	return &req, nil
}
