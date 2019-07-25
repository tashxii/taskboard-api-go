package boards

import (
	"net/http"
	"taskboard-api-go/controller/api"
	"taskboard-api-go/controller/websocket"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/service"

	"github.com/gin-gonic/gin"
)

type endPoint struct {
	boards          string
	boardid         string
	boardtasks      string
	boardorders     string
	taskorders      string
	taskboardFromID string
	ws              *websocket.WsManager
}

// EndPoint presents boards endpoint
var EndPoint = endPoint{
	boards:          "/boards",
	boardorders:     "/boardorders",
	boardid:         "boardid",
	taskboardFromID: "taskboard-from-id",
}

// SetWsManager sets websocket manager to EndPoint
func SetWsManager(ws *websocket.WsManager) {
	EndPoint.ws = ws
}

// RegisterRoute registers API endpoints for boards
func (p *endPoint) RegisterRoute(route *gin.RouterGroup) (err error) {
	route.GET(p.boards, list)
	route.POST(p.boards, create)
	route.GET(p.boards+"/:"+p.boardid, get)
	route.PUT(p.boards+"/:"+p.boardid, update)
	route.DELETE(p.boards+"/:"+p.boardid, delete)
	route.PUT(p.boardorders, updateBoardOrders)
	return
}

// find all boards
func list(c *gin.Context) {
	tx := orm.GetDB() // No transction
	srvc := service.NewBoardService(tx)
	boards, serr := srvc.FindBoards(&model.Board{}, []string{"disp_order, created_date"})
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	res := convertListBoardResponse(boards)
	c.IndentedJSON(http.StatusOK, res)
}

func create(c *gin.Context) {
	board, serr := getBoardByCreateRequest(c)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}

	// create board
	tx := orm.GetDB().Begin()
	srvc := service.NewBoardService(tx)
	serr = srvc.CreateBoard(board)
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

	res := convertBoardResponse(board)
	c.IndentedJSON(http.StatusOK, res)

	// websocket send message
	EndPoint.ws.SendUpdateBoardMessage(c.GetHeader(EndPoint.taskboardFromID))
}

// get a board
func get(c *gin.Context) {
	tx := orm.GetDB() // No transaction
	srvc := service.NewBoardService(tx)
	find, err := findBoardByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	res := convertBoardResponse(find)
	c.IndentedJSON(http.StatusOK, res)
}

func findBoardByPathParameter(c *gin.Context, srvc *service.BoardService) (find *model.Board, serr error) {
	boardID, serr := api.GetPathParameter(c, EndPoint.boardid)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return nil, serr
	}
	find, serr = srvc.FindBoard(&model.Board{ID: boardID})
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return nil, serr
	}
	return
}

// update board
func update(c *gin.Context) {
	tx := orm.GetDB().Begin()
	srvc := service.NewBoardService(tx)
	find, err := findBoardByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	board, serr := getBoardByUpdateRequest(c, find)
	if serr != nil {
		api.Rollback(tx)
		api.SetErrorStatus(c, serr)
		return
	}

	// update board
	serr = srvc.UpdateBoard(board)
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

	res := convertBoardResponse(board)
	c.IndentedJSON(http.StatusOK, res)

	// websocket send message
	EndPoint.ws.SendUpdateBoardMessage(c.GetHeader(EndPoint.taskboardFromID), board.ID)
}

// delete board
func delete(c *gin.Context) {
	tx := orm.GetDB().Begin()
	srvc := service.NewBoardService(tx)
	find, err := findBoardByPathParameter(c, srvc)
	if err != nil {
		api.Rollback(tx)
		return
	}
	// delete board
	serr := srvc.DeleteBoard(find)
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

	// websocket send message
	EndPoint.ws.SendUpdateBoardMessage(c.GetHeader(EndPoint.taskboardFromID), find.ID, model.SystemBoardIcebox.ID)
}

// update order of all boards
func updateBoardOrders(c *gin.Context) {
	req, serr := getUpdateBoardOrdersRequest(c)
	if serr != nil {
		api.SetErrorStatus(c, serr)
		return
	}
	tx := orm.GetDB().Begin()
	srvc := service.NewBoardService(tx)
	serr = srvc.UpdateBoardOrders(req.BoardIDs)
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

	// websocket send message
	EndPoint.ws.SendUpdateBoardMessage(c.GetHeader(EndPoint.taskboardFromID))
	EndPoint.ws.SendUpdateTaskBoardMessage(c.GetHeader(EndPoint.taskboardFromID))
}
