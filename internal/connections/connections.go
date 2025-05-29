package connections

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	configs "tektmud/internal/config"
	"time"
)

type ConnectionId string

type ConnectionManager struct {
	connections map[ConnectionId]*PlayerConnection
	mu          sync.RWMutex
	maxPlayers  int
}

// Represents the status of the player
type ConnectionState int

const (
	StateConnected ConnectionState = iota
	StateInitialPrompt
	StateUsername
	StatePassword
	StateNewPassword
	StateConfirmPassword
	StateCollectEmail
	StateAuthenticated
	StateRejectedAuthentication

	//If the connection user doens't have a character
	//make sure we track where they are in creating it
	//No connection should be passed to to the world
	//if its for a user with no character.
	StateCharacterEval
	StateCharacterGenderChoice
	StateCharacterRaceChoice
	StateCharacterClassChoice
	StateCharacterNameChoice
)

type PlayerConnection struct {
	Id         ConnectionId
	Conn       net.Conn
	Reader     *bufio.Reader
	Writer     *bufio.Writer
	Username   string
	LastActive time.Time
	Mu         sync.Mutex
	state      ConnectionState
}

////////////////////////////////////////////
///////      ConnectionManager     ////////
///////////////////////////////////////////

func NewConnectionManager(c *configs.Config) (*ConnectionManager, error) {
	return &ConnectionManager{
		connections: make(map[ConnectionId]*PlayerConnection),
		maxPlayers:  c.Server.MaxPlayers,
	}, nil
}

func GenerateConnectionId() ConnectionId {
	return ConnectionId(fmt.Sprintf("conn_%d", time.Now().Nanosecond()))
}

func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.connections)
}

func (cm *ConnectionManager) Add(pc *PlayerConnection) {
	cm.mu.Lock()
	cm.connections[pc.Id] = pc
	cm.mu.Unlock()
}

func (cm *ConnectionManager) Remove(cid ConnectionId) {
	cm.mu.Lock()
	delete(cm.connections, cid)
	cm.mu.Unlock()
}

func (cm *ConnectionManager) CloseAll(f func(c *PlayerConnection)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conn := range cm.connections {
		f(conn)
	}
}

// //////////////////////////////////////////
// /////      PlayerConnection      ////////
// /////////////////////////////////////////
func (pc *PlayerConnection) SetState(state ConnectionState) {
	pc.Mu.Lock()
	defer pc.Mu.Unlock()

	pc.state = state
}

func (pc *PlayerConnection) GetState() ConnectionState {
	pc.Mu.Lock()
	defer pc.Mu.Unlock()

	return pc.state
}
