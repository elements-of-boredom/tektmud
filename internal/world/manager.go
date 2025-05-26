package world

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"tektmud/internal/character"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/rooms"
	"tektmud/internal/users"
	"time"
)

type WorldConfig struct {
	TickRate    time.Duration //How often our game loop runs
	DefaultArea string        //Default area for new characters
	DefaultRoom string        //Default room for new characters
}

// QueuedInput represents player input waiting to be processed
type QueuedInput struct {
	CharacterId uint64
	Input       string
	Timestamp   time.Time
}

type WorldManager struct {
	Config        *WorldConfig
	areaManager   *rooms.AreaManager
	tickManager   TickManager
	userManager   *users.UserManager
	characters    map[uint64]*character.Character            //CharacterId => Character
	connections   map[uint64]*connections.PlayerConnection   //CharacterId => PlayerConnection
	roomOccupants map[string]map[uint64]*character.Character //areaId:roomId => characters
	inputHandlers map[string]InputHandler                    //InputHandler.Id => InputHandler
	//Game loop
	ticker   *time.Ticker
	stopChan chan struct{}
	running  bool

	//Input throttling
	inputQueue    chan *QueuedInput
	maxInputQueue int

	//Sync
	mu sync.RWMutex
}

func NewWorldManager(um *users.UserManager) *WorldManager {
	c := configs.GetConfig()

	wc := &WorldConfig{
		TickRate:    time.Millisecond * time.Duration(c.Core.TickRate),
		DefaultArea: c.Core.DefaultArea,
		DefaultRoom: c.Core.DefaultRoom,
	}

	return &WorldManager{
		Config:        wc,
		areaManager:   rooms.NewAreaManager(),
		tickManager:   *NewTickManager(),
		userManager:   um,
		characters:    make(map[uint64]*character.Character),
		connections:   make(map[uint64]*connections.PlayerConnection),
		roomOccupants: make(map[string]map[uint64]*character.Character),
		inputHandlers: make(map[string]InputHandler),
		stopChan:      make(chan struct{}),
		inputQueue:    make(chan *QueuedInput, 1000), //Buffer for up to 1000 inputs
		maxInputQueue: 1,                             //Max inputs per character in queue
	}
}

func (wm *WorldManager) Initialize() error {
	slog.Info("Initializing world engine...")

	//Load all areas
	if err := wm.areaManager.LoadAllAreas(); err != nil {
		return fmt.Errorf("failed to load areas: %w", err)
	}

	//Validate the room connections
	if errors := wm.areaManager.ValidateRoomConnections(); len(errors) > 0 {
		slog.Warn("Warning: Found rooms with connection errors:", "count", len(errors))
		for _, err := range errors {
			log.Printf(" - %v", err)
		}
	}

	//Somehow get all our registered handlers
	wm.registerHandlers()

	slog.Info("Loaded world.", "areas", len(wm.areaManager.GetAreaList()), "rooms", wm.areaManager.GetRoomCount())

	return nil
}

// Start begins the game loop
func (wm *WorldManager) Start() {
	wm.mu.Lock()
	if wm.running {
		wm.mu.Unlock()
		return
	}
	wm.running = true
	wm.ticker = time.NewTicker(wm.Config.TickRate)
	wm.mu.Unlock()

	log.Printf("Starting world engine with tick rate: %v", wm.Config.TickRate)

	//Start input processing goroutine
	go wm.processInputQueue()

	//Queue initial heartbeat
	wm.tickManager.QueueDelayedAction(ActionHeartbeat, 30*time.Second, "", nil, HeartbeatCallback)

	go wm.gameLoop()
}

// Stop shuts down the game loop
func (wm *WorldManager) Stop() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if !wm.running {
		return
	}

	log.Println("Stopping world engine...")
	wm.running = false
	close(wm.stopChan)
	wm.ticker.Stop()
}

// handles queued player input to prevent spam.
func (wm *WorldManager) processInputQueue() {
	inputCounts := make(map[uint64]int)
	ticker := time.NewTicker((1 * time.Second))
	defer ticker.Stop()

	for {
		select {
		case input := <-wm.inputQueue:
			//check if char is still connected
			wm.mu.RLock()
			_, exists := wm.characters[input.CharacterId]
			wm.mu.RUnlock()

			if !exists {
				continue //Character dc'd
			}

			//Throttle to max N commands per second per char
			if inputCounts[input.CharacterId] >= wm.maxInputQueue {
				continue //We just pitch the extras for now. Probably need to send something later
			}

			inputCounts[input.CharacterId]++

			//queue the command to be processed on next tick
			wm.tickManager.QueueDelayedAction(
				ActionPlayerCommand,
				0, //Execute immediately on next tick
				strconv.FormatUint(input.CharacterId, 10),
				&PlayerCommandData{
					Command: strings.Fields(input.Input)[0],
					Args:    strings.Fields(input.Input)[1:],
				},
				PlayerCommandCallback,
			)

		case <-ticker.C:
			//Reset input counts every second
			inputCounts = make(map[uint64]int)

		case <-wm.stopChan:
			return
		}
	}
}

// gameLoop is the main game tick loop
func (wm *WorldManager) gameLoop() {
	for {
		select {
		case <-wm.ticker.C:
			wm.tick()
		case <-wm.stopChan:
			return
		}
	}
}

// tick processes one game tick
func (wm *WorldManager) tick() {
	// Future: Process NPC actions, spell effects, regeneration, etc.
	// For now, this is just a placeholder for the game loop structure
	wm.tickManager.ProcessTick(wm)

	//Additional per-tick processing can go here
	// Example:
	// - Update temporary effects
	// - Process combat rounds (for NPC, no player actions should be automated for combat)
	// - Update any world state
}

// HandleInput processes input from a character (now queues it for processing)
func (wm *WorldManager) HandleInput(characterId uint64, input string) error {
	wm.mu.RLock()
	_, exists := wm.characters[characterId]
	wm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("character %d not found", characterId)
	}

	// Queue the input for processing (with timeout to prevent blocking)
	queuedInput := &QueuedInput{
		CharacterId: characterId,
		Input:       input,
		Timestamp:   time.Now(),
	}

	select {
	case wm.inputQueue <- queuedInput:
		// Successfully queued
		return nil
	case <-time.After(100 * time.Millisecond):
		// Queue is full, drop the input
		return fmt.Errorf("input queue full for character %d", characterId)
	}
}

// HandleInputDirect processes input immediately (for admin commands or special cases)
func (wm *WorldManager) HandleInputDirect(characterId uint64, input string) error {
	wm.mu.RLock()
	character, exists := wm.characters[characterId]
	wm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("character %d not found", characterId)
	}

	return ProcessInput(character, input, wm)
}

// AddCharacter brings a character into the world
func (wm *WorldManager) AddCharacter(character *character.Character, conn *connections.PlayerConnection) error {
	// Set up default handlers if character has none
	if len(character.Handlers) == 0 {
		wm.setupDefaultHandlers(character)
	}

	// Determine spawn location
	areaID, roomID := wm.getSpawnLocation(character)
	character.SetLocation(areaID, roomID)

	// Add to world
	wm.mu.Lock()
	wm.characters[character.Id] = character
	wm.connections[character.Id] = conn
	wm.addToRoom(character, areaID, roomID)
	wm.mu.Unlock()
	// Start regeneration for this character
	//wm.startCharacterRegeneration(character.Id)

	// Show the room to the character
	wm.ShowRoom(character)

	// Announce arrival to room (except to the character themselves)
	wm.SendToRoom(areaID, roomID, character.Name+" has entered the game.", character.Id)

	log.Printf("Character %s entered the world at %s:%s", character.Name, areaID, roomID)
	return nil
}

// getSpawnLocation determines where a character should spawn
func (wm *WorldManager) getSpawnLocation(character *character.Character) (string, string) {
	// Try to restore last location
	if character.LastLocation != "" {
		parts := rooms.SplitDestination(character.LastLocation)
		if len(parts) == 2 {
			if _, exists := wm.areaManager.GetRoom(parts[0], parts[1]); exists {
				return parts[0], parts[1]
			}
		}
	}
	c := configs.GetConfig()
	// Fall back to default location
	return c.Core.DefaultArea, c.Core.DefaultRoom
}

// RemoveCharacter removes a character from the world
func (wm *WorldManager) RemoveCharacter(characterId uint64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	character, exists := wm.characters[characterId]
	if !exists {
		return
	}

	// Announce departure
	areaID, roomID := character.GetLocation()
	wm.SendToRoom(areaID, roomID, character.Name+" has left the game.", characterId)

	// Save character state (facade)
	//wm.userManager.SaveUser(character)

	// Remove from room and world
	wm.removeFromRoom(character, areaID, roomID)
	delete(wm.characters, characterId)

	// Close connection and clean up
	if conn, exists := wm.connections[characterId]; exists {
		conn.Conn.Close()
		delete(wm.connections, characterId)
	}

	log.Printf("Character %s left the world", character.Name)
}

// MoveCharacter attempts to move a character in a direction
func (wm *WorldManager) MoveCharacter(character *character.Character, direction rooms.Direction) (bool, string) {
	areaID, roomID := character.GetLocation()

	// Find the exit
	exit, exists := wm.areaManager.GetRoomExit(areaID, roomID, direction)
	if !exists {
		return false, "You can't go that way."
	}

	// Parse destination
	destAreaID := areaID
	destRoomID := exit.Destination

	if len(exit.Destination) > 0 && exit.Destination != roomID {
		parts := rooms.SplitDestination(exit.Destination)
		if len(parts) == 2 {
			destAreaID = parts[0]
			destRoomID = parts[1]
		}
	}

	// Validate destination exists
	if _, exists := wm.areaManager.GetRoom(destAreaID, destRoomID); !exists {
		return false, "That exit leads nowhere."
	}

	// Perform the move
	wm.mu.Lock()
	wm.removeFromRoom(character, areaID, roomID)
	character.SetLocation(destAreaID, destRoomID)
	wm.addToRoom(character, destAreaID, destRoomID)
	wm.mu.Unlock()

	// Announce movement
	wm.SendToRoom(areaID, roomID, character.Name+" leaves "+string(direction)+".", character.Id)
	wm.SendToRoom(destAreaID, destRoomID, character.Name+" arrives.", character.Id)

	return true, ""
}

// ShowRoom displays a room description to a character
func (wm *WorldManager) ShowRoom(character *character.Character) {
	areaId, roomId := character.GetLocation()
	roomDesc := wm.areaManager.FormatRoom(areaId, roomId)

	// Add other characters in the room
	wm.mu.RLock()
	roomKey := areaId + ":" + roomId
	if occupants, exists := wm.roomOccupants[roomKey]; exists {
		var others []string
		for _, other := range occupants {
			if other.Id != character.Id {
				others = append(others, other.Name)
			}
		}
		if len(others) > 0 {
			roomDesc += "\n\nAlso here: " + strings.Join(others, ", ")
		}
	}
	wm.mu.RUnlock()

	wm.SendToCharacter(character, roomDesc)
}

// addToRoom adds a character to a room's occupant list
func (wm *WorldManager) addToRoom(c *character.Character, areaID, roomID string) {
	roomKey := areaID + ":" + roomID
	if wm.roomOccupants[roomKey] == nil {
		var charmap = make(map[uint64]*character.Character)
		wm.roomOccupants[roomKey] = charmap
	}
	wm.roomOccupants[roomKey][c.Id] = c
}

// removeFromRoom removes a character from a room's occupant list
func (wm *WorldManager) removeFromRoom(character *character.Character, areaID, roomID string) {
	roomKey := areaID + ":" + roomID
	if occupants, exists := wm.roomOccupants[roomKey]; exists {
		delete(occupants, character.Id)

		// Clean up empty room
		if len(occupants) == 0 {
			delete(wm.roomOccupants, roomKey)
		}
	}
}

// SendToCharacter sends a message to a specific character
func (wm *WorldManager) SendToCharacter(character *character.Character, message string) {
	wm.mu.RLock()
	conn, exists := wm.connections[character.Id]
	wm.mu.RUnlock()

	if exists {
		conn.Conn.Write([]byte(message + "\n"))
	}
}

// SendToRoom sends a message to all characters in a room except excluded ones
func (wm *WorldManager) SendToRoom(areaID, roomID, message string, excludeIDs ...uint64) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	roomKey := areaID + ":" + roomID
	occupants, exists := wm.roomOccupants[roomKey]
	if !exists {
		return
	}

	// Create exclusion map for efficiency
	exclude := make(map[uint64]bool)
	for _, id := range excludeIDs {
		exclude[id] = true
	}

	// Send to all non-excluded occupants
	for _, character := range occupants {
		if !exclude[character.Id] {
			if conn, exists := wm.connections[character.Id]; exists {
				conn.Conn.Write([]byte(message + "\n"))
			}
		}
	}
}

func (wm *WorldManager) registerHandlers() {
	quit := NewQuitHandler()
	movement := NewMovementHandler()
	look := NewLookHandler()
	say := NewSayHandler()
	defaultHandler := NewDefaultHandler()

	wm.inputHandlers = map[string]InputHandler{
		quit.Name():     quit,
		movement.Name(): movement,
		look.Name():     look,
		say.Name():      say,

		//ALWAYS LAST
		defaultHandler.Name(): defaultHandler,
	}

}

// setupDefaultHandlers adds standard handlers to a character
func (wm *WorldManager) setupDefaultHandlers(character *character.Character) {

	// Add admin handlers if character is an admin
	/*
		if character.IsAdmin() {
			character.AddHandler(NewAdminHandler())
			character.AddHandler(NewAdminHelpHandler())
			character.AddHandler(NewDebugHandler()) // Debug commands for admins
		}
	*/

	//character.AddHandler(NewSpellHandler())  // Spell casting
	//character.AddHandler(NewAttackHandler()) // Combat commands

	//TODO: Fix this ugly string crap
	character.AddHandler("quit")
	character.AddHandler("movement")
	character.AddHandler("look")
	character.AddHandler("say")
	character.AddHandler("default") // Always last
}
