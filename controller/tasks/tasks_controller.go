package tasks

import (
	"net/http"
	"taskboard-api-go/controller/api"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/service"

	"github.com/gin-gonic/gin"
)

type endPoint struct {
	tasks      string
	taskorders string
	taskid     string
	boardid    string
}

// EndPoint presents boards endpoint
var EndPoint = endPoint{
	tasks:      "/tasks",
	taskorders: "/taskorders",
	taskid:     "taskid",
	boardid:    "boardid",
}

// RegisterRoute registers API endpoints for tasks
func (p *endPoint) RegisterRoute(route *gin.RouterGroup) (err error) {
	route.GET(p.tasks, list)
	route.POST(p.tasks, create)
	route.GET(p.tasks+"/:"+p.taskid, get)
	route.PUT(p.tasks+"/:"+p.taskid, update)
	route.DELETE(p.tasks+"/:"+p.taskid, delete)
	route.PUT(p.taskorders, updateTaskOrders)
	return
}

func list(c *gin.Context) {
	tx := orm.GetDB() // No transction
	srvc := service.NewTaskService(tx)
	boardID := c.Query(EndPoint.boardid)
	condition := &model.Task{}
	if boardID != "" {
		condition = &model.Task{BoardID: boardID}
	}
	tasks, serr := srvc.FindTasks(condition, []string{"disp_order, created_date, name"})
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	res := convertListTaskResponse(tasks)
	c.IndentedJSON(http.StatusOK, res)
}

func create(c *gin.Context) {
	task, serr := getTaskByCreateRequest(c)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}

	// create task
	tx := orm.GetDB().Begin()
	srvc := service.NewTaskService(tx)
	serr = srvc.CreateTask(task)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}
	serr = api.Commit(tx)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}

	res := convertTaskResponse(task)
	c.IndentedJSON(http.StatusOK, res)
}

func get(c *gin.Context) {
	tx := orm.GetDB() // No transaction
	srvc := service.NewTaskService(tx)
	find, err := findTaskByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	res := convertTaskResponse(find)
	c.IndentedJSON(http.StatusOK, res)
}

func findTaskByPathParameter(c *gin.Context, srvc *service.TaskService) (find *model.Task, serr error) {
	taskID, serr := api.GetPathParameter(c, EndPoint.taskid)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return nil, serr
	}
	find, serr = srvc.FindTask(&model.Task{ID: taskID})
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return nil, serr
	}
	return
}

func update(c *gin.Context) {
	tx := orm.GetDB().Begin()
	srvc := service.NewTaskService(tx)
	find, err := findTaskByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	task, serr := getTaskByUpdateRequest(c, find)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}

	// update task
	serr = srvc.UpdateTask(find, task)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}
	serr = api.Commit(tx)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}

	res := convertTaskResponse(task)
	c.IndentedJSON(http.StatusOK, res)
}

func delete(c *gin.Context) {
	tx := orm.GetDB().Begin()
	srvc := service.NewTaskService(tx)
	find, err := findTaskByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	// delete task
	serr := srvc.DeleteTask(find)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}
	serr = api.Commit(tx)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	c.Status(http.StatusOK)
}

// update order of tasks
func updateTaskOrders(c *gin.Context) {
	req, serr := getUpdateTaskOrdersRequest(c)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	tx := orm.GetDB().Begin()
	srvc := service.NewTaskService(tx)
	serr = srvc.UpdateTaskOrders(
		req.TaskID, req.FromBoardID, req.FromDispOrder, req.ToBoardID, req.ToDispOrder,
	)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}
	serr = api.Commit(tx)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	c.Status(http.StatusOK)
}
