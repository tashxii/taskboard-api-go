package service

import (
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/repository"

	"github.com/jinzhu/gorm"
)

// BoardService provides apis for board management.
type BoardService struct {
	tx        *gorm.DB
	boardRepo *repository.BoardRepository
	taskRepo  *repository.TaskRepository
}

// NewBoardService return new instance of BoardService.
func NewBoardService(tx *gorm.DB) *BoardService {
	return &BoardService{
		tx:        tx,
		boardRepo: repository.NewBoardRepository(tx),
		taskRepo:  repository.NewTaskRepository(tx),
	}
}

// FindBoard returns board matching specified condition
func (s *BoardService) FindBoard(condition interface{}) (*model.Board, error) {
	find, err := s.boardRepo.FindFirstBoard(condition, []string{"id"})
	if err != nil {
		if err == orm.ErrorRecordNotFound {
			return nil, NewSvcErrorf(ErrorCodeNotFound, err, "Board not found")
		}
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find board")
	}
	return &find, nil
}

// FindBoards finds all boards
func (s *BoardService) FindBoards(condition interface{}, sortOrders []string) ([]model.Board, error) {
	boards, err := s.boardRepo.FindBoards(condition, 0, orm.NoLimit, sortOrders)
	if err != nil {
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find boards")
	}
	return boards, nil
}

// CreateBoard creates new board
func (s *BoardService) CreateBoard(board *model.Board) error {
	count, err := s.boardRepo.CountBoards(&model.Board{})
	if err != nil {
		return NewSvcError(ErrorCodeDB, err, "Failed to count boards")
	}
	board.DispOrder = count
	err = s.boardRepo.CreateBoard(board)
	if err != nil {
		return NewSvcError(ErrorCodeDB, err, "Failed to create board")
	}
	return nil
}

// UpdateBoard updates specifed board
func (s *BoardService) UpdateBoard(board *model.Board) error {
	err := s.boardRepo.UpdateBoard(board)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to update board. ID:%s", board.ID)
	}
	return nil
}

// DeleteBoard deletes specifed board
func (s *BoardService) DeleteBoard(board *model.Board) error {
	err := s.boardRepo.DeleteBoard(board)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to delete board. ID:%s", board.ID)
	}
	err = s.taskRepo.MoveToIceboxBoard(board.ID)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to move tasks to iceboax. BoardID:%s", board.ID)
	}
	return nil
}

// CreateSystemBoards creates all system boards if not exist
func (s *BoardService) CreateSystemBoards() error {
	boards, serr := s.FindBoards(&model.Board{IsSystem: true}, []string{"disp_order"})
	if serr != nil {
		return serr
	}
	isFindIcebox := false
	isFindTodo := false
	isFindDoing := false
	isFindDone := false
	for _, board := range boards {
		switch board.ID {
		case model.SystemBoardIcebox.ID:
			isFindIcebox = true
		case model.SystemBoardTodo.ID:
			isFindTodo = true
		case model.SystemBoardDoing.ID:
			isFindDoing = true
		case model.SystemBoardDone.ID:
			isFindDone = true
		}
	}
	if isFindIcebox != true {
		serr = s.CreateBoard(model.SystemBoardIcebox)
		if serr != nil {
			return serr
		}
	}
	if isFindTodo != true {
		serr = s.CreateBoard(model.SystemBoardTodo)
		if serr != nil {
			return serr
		}
	}
	if isFindDoing != true {
		serr = s.CreateBoard(model.SystemBoardDoing)
		if serr != nil {
			return serr
		}
	}
	if isFindDone != true {
		serr = s.CreateBoard(model.SystemBoardDone)
		if serr != nil {
			return serr
		}
	}
	return nil
}

// UpdateBoardOrders updates order of boards
func (s *BoardService) UpdateBoardOrders(boardIDs []string) error {
	err := s.boardRepo.UpdateBoardOrders(boardIDs)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to update board's order")
	}
	return nil
}
