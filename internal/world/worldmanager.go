package world

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"tektmud/internal/character"
	"tektmud/internal/commands"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/listeners"
	"tektmud/internal/logger"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
	"time"
)

type WorldConfig struct {
	TickRate    time.Duration //Default is 100ms or 10 ticks a second, which means we can support balances of N.N seconds
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
	tickManager   TickManager
	playerManager *players.PlayerManager
	areaManager   *rooms.AreaManager
	tmpl          *templates.TemplateManager
	characters    map[uint64]*character.Character          //CharacterId => Character
	connections   map[uint64]*connections.PlayerConnection //CharacterId => PlayerConnection

	inputHandlers    map[string]InputHandler //InputHandler.Id => InputHandler
	commandProcessor *commands.QueueProcessor
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

func NewWorldManager(pm *players.PlayerManager, tm *templates.TemplateManager) *WorldManager {
	c := configs.GetConfig()

	wc := &WorldConfig{
		TickRate:    time.Millisecond * time.Duration(c.Core.TickRate),
		DefaultArea: c.Core.DefaultArea,
		DefaultRoom: c.Core.DefaultRoom,
	}

	return &WorldManager{
		Config:           wc,
		tickManager:      *NewTickManager(),
		commandProcessor: commands.NewQueueProcessor(wc.TickRate),
		playerManager:    pm,
		tmpl:             tm,
		characters:       make(map[uint64]*character.Character),
		connections:      make(map[uint64]*connections.PlayerConnection),
		inputHandlers:    make(map[string]InputHandler),
		stopChan:         make(chan struct{}),
		inputQueue:       make(chan *QueuedInput, 1000), //Buffer for up to 1000 inputs
		maxInputQueue:    1,                             //Max inputs per character in queue
	}
}

func (wm *WorldManager) Initialize() error {
	logger.Info("Initializing world engine...")

	//Load all areas
	if am, err := rooms.Initialize(); err != nil {
		return err
	} else {
		wm.areaManager = am
	}

	//Somehow get all our registered handlers
	//wm.registerHandlers()
	//Register listeners
	wm.registerListeners()

	logger.Info("Loaded world.", "areas", len(wm.areaManager.GetAreaList()), "rooms", wm.areaManager.GetRoomCount())

	return nil
}

func (wm *WorldManager) registerListeners() {

	//Register our input listener
	var inputListener = listeners.NewInputListener(wm.areaManager, wm.playerManager)
	var promptListener = listeners.NewPromptListener(wm.playerManager)
	var messageListener = listeners.NewMessageListener(wm.areaManager, wm.playerManager, wm.connections, wm.tmpl)
	var displayRoomListener = listeners.NewDisplayRoomListener(wm.areaManager, wm.playerManager, wm.tmpl)
	var quitListener = listeners.NewQuitListener(wm)

	commands.RegisteredListener(inputListener, commands.Input{}.Name())
	commands.RegisteredListener(messageListener, commands.Message{}.Name())
	commands.RegisteredListener(displayRoomListener, commands.DisplayRoom{}.Name())
	commands.RegisteredListener(quitListener, commands.PlayerQuit{}.Name())
	commands.RegisteredListener(promptListener, commands.SendPrompt{}.Name())

}

// Start begins the game loop
func (wm *WorldManager) Start() {
	wm.mu.Lock()
	if wm.running {
		wm.mu.Unlock()
		return
	}
	wm.running = true
	//The ticker controls how frequently balances are restored
	wm.ticker = time.NewTicker(wm.Config.TickRate)
	wm.mu.Unlock()

	//start processing commands
	wm.commandProcessor.Start()

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
				logger.Debug("Throwing away command", "cmd", input.Input, "character", input.CharacterId)
				continue //We just pitch the extras for now. Probably need to send something later
			}

			inputCounts[input.CharacterId]++

			//queue the command to be processed on next tick
			logger.Debug("Length of input", "len", len(input.Input))
			var cmd string = ""
			if len(input.Input) > 0 {
				cmd = strings.Fields(input.Input)[0]
			}
			var args []string = []string{}
			if len(strings.Fields(input.Input)) > 1 {
				args = strings.Fields(input.Input)[1:]
			}
			wm.tickManager.QueueDelayedAction(
				ActionPlayerCommand,
				0, //Execute immediately on next tick
				strconv.FormatUint(input.CharacterId, 10),
				&PlayerCommandData{
					Command: cmd,
					Args:    args,
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

// HandleInput processes input from a character passed in from the server
func (wm *WorldManager) HandleInput(characterId uint64, input string) error {
	wm.mu.RLock()
	_, exists := wm.characters[characterId]
	wm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("character %d not found", characterId)
	}

	//Create a command type of Input
	//Even though we don't queue player's we still feed it through the queue
	//incase something else cares about the action as well.
	commands.QueueGameCommand(characterId, commands.Input{
		PlayerId: characterId,
		Text:     input,
	})

	/*
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
	*/
	return nil
}

// HandleInputDirect processes input immediately (for admin commands or special cases)
func (wm *WorldManager) HandleInputImmediate(characterId uint64, input string) error {
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

	// Determine spawn location
	areaId, roomId := wm.getSpawnLocation(character)
	character.SetLocation(areaId, roomId)

	// Add to world
	wm.mu.Lock()
	wm.characters[character.Id] = character
	wm.connections[character.Id] = conn
	rooms.AddToRoom(character.Id, areaId, roomId)
	wm.mu.Unlock()
	// Start regeneration for this character
	//wm.startCharacterRegeneration(character.Id)

	// Show the room to the character
	if r, exists := wm.areaManager.GetRoom(areaId, roomId); exists {
		r.ShowRoom(character.Id)

		// Announce arrival to room (except to the character themselves)
		r.SendText(character.Name+" has entered the game.", character.Id)
	}
	commands.QueueGameCommand(character.Id, commands.SendPrompt{PlayerId: character.Id})

	log.Printf("Character %s entered the world at %s:%s", character.Name, areaId, roomId)
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
	wm.mu.RLock()
	character, exists := wm.characters[characterId]
	wm.mu.RUnlock()
	if !exists {
		return
	}

	player, err := wm.playerManager.GetPlayerById(characterId)
	if err != nil {
		logger.Error("Unable to find player with character id", "name", character.Name, "id", characterId)
	} else {
		// Save character state (facade) TODO:
		wm.playerManager.UpdatePlayer(player)
	}

	// Announce departure
	areaID, roomID := character.GetLocation()
	if r, exists := wm.areaManager.GetRoom(areaID, roomID); exists {
		r.SendText(character.Name+" has left the game.", characterId)
	}

	// Remove from world
	wm.mu.Lock()
	delete(wm.characters, characterId)
	rooms.RemoveFromRoom(character.Id, areaID, roomID)
	wm.mu.Unlock()

	// Close connection and clean up
	if conn, exists := wm.connections[characterId]; exists {
		conn.Conn.Close()
		delete(wm.connections, characterId)
	}

	logger.Printf("Character %s left the world", character.Name)
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
