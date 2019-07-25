package websocket

import (
	"fmt"
	"strings"
	"sync"

	"gopkg.in/olahol/melody.v1"
)

// WsManager manages session of websocket
type WsManager struct {
	lock     *sync.Mutex
	sessions map[string]*melody.Session
	mrouter  *melody.Melody
}

// NewWsManager creates new instance of WsManager(Websocket Manager)
func NewWsManager(mrouter *melody.Melody) *WsManager {
	ws := &WsManager{
		lock:     new(sync.Mutex),
		sessions: make(map[string]*melody.Session, 0),
		mrouter:  mrouter,
	}
	mrouter.HandleConnect(ws.Connect)
	mrouter.HandleDisconnect(ws.Disconnect)
	return ws
}

const (
	updateTasksMessage      = "UPDATE_TASKS"
	updateBoardsMessage     = "UPDATE_BOARDS"
	updateTaskBoardsMessage = "UPDATE_TASKBOARDS"
	updateUsersMessage      = "UPDATE_USERS"
	queryFromUserIDKey      = "from"
)

// SendUpdateTaskMessage sends a message to update tasks for other clients
func (w *WsManager) SendUpdateTaskMessage(fromUserID string, taskIDs ...string) {
	w.sendMessage(fromUserID, fmt.Sprintf("%s %s", updateTasksMessage, strings.Join(taskIDs, " ")))
}

// SendUpdateTaskBoardMessage sends a message to update taskboards for other clients
func (w *WsManager) SendUpdateTaskBoardMessage(fromUserID string, boardIDs ...string) {
	w.sendMessage(fromUserID, fmt.Sprintf("%s %s", updateTaskBoardsMessage, strings.Join(boardIDs, " ")))
}

// SendUpdateBoardMessage sends a message to update boards for other clients
func (w *WsManager) SendUpdateBoardMessage(fromUserID string, boardIDs ...string) {
	w.sendMessage(fromUserID, fmt.Sprintf("%s %s", updateBoardsMessage, strings.Join(boardIDs, " ")))
}

// SendUpdateUserMessage sends a message to update users for other clients
func (w *WsManager) SendUpdateUserMessage(fromUserID string, userIDs ...string) {
	w.sendMessage(fromUserID, fmt.Sprintf("%s %s", updateUsersMessage, strings.Join(userIDs, " ")))
}

func (w *WsManager) sendMessage(fromUserID string, message string) {
	w.lock.Lock()
	defer w.lock.Unlock()
	s, exists := w.sessions[fromUserID]
	if !exists {
		w.mrouter.Broadcast([]byte(message))
	} else {
		w.mrouter.BroadcastOthers([]byte(message), s)
	}
}

// Connect put a session to session's map
func (w *WsManager) Connect(s *melody.Session) {
	w.lock.Lock()
	defer w.lock.Unlock()
	fromUserID := s.Request.URL.Query().Get(queryFromUserIDKey)
	w.sessions[fromUserID] = s
}

// Disconnect remove a session from session's map
func (w *WsManager) Disconnect(s *melody.Session) {
	w.lock.Lock()
	defer w.lock.Unlock()
	for id, session := range w.sessions {
		if s == session {
			delete(w.sessions, id)
			return
		}
	}
}
