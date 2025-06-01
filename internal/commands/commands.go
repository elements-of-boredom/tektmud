package commands

import (
	"sync"
	"tektmud/internal/logger"
	"time"
)

var (
	mu = sync.Mutex{}
)

type Command interface {
	Name() string
}

type PrioritizedCommand interface {
	Command
	Priority() int // Lower number == higher priority, <0 = instant
}

type DelayedCommand interface {
	Command
	Delay() time.Duration
}

type CommandResult int

const (
	Continue      CommandResult = iota // Command succeeded, continue processing
	Cancel                             // Cancel further listener processing
	CancelRequeue                      // Cancel and requeue for next round
)

// CommandContext contains information about the command being executed
type CommandContext struct {
	UserId    uint64                 // Who initiated the command (0 for system)
	Command   Command                // The command being executed
	Data      map[string]interface{} // Additional context data
	Round     uint64                 // Current processing round
	Timestamp time.Time              // When command was queued
}

// DelayedCommandWrapper wraps commands that should execute after a delay
type DelayedCommandWrapper struct {
	wrappedCommand Command
	delay          time.Duration
	scheduledFor   time.Time
}

func (dcw *DelayedCommandWrapper) Name() string {
	return dcw.wrappedCommand.Name()
}

func (dcw *DelayedCommandWrapper) Delay() time.Duration {
	return dcw.delay
}

// Helper function to create delayed commands
func NewDelayedCommand(command Command, delay time.Duration) DelayedCommand {
	return &DelayedCommandWrapper{
		wrappedCommand: command,
		delay:          delay,
		scheduledFor:   time.Now().Add(delay),
	}
}

// CommandListener can react to commands
type CommandListener interface {
	// Handle processes the command and returns a result
	Handle(ctx *CommandContext) CommandResult

	// Priority returns the listener priority (lower = higher priority)
	Priority() int

	// Name returns the listener name for debugging
	Name() string
}

// RegisteredListener wraps a listener with its registration info
type RegisteredListenerWrapper struct {
	Listener CommandListener
	Priority int
}

var (
	cmdListenerLock = sync.RWMutex{}
	//listner management - organized by command name
	listenersByCommand map[string][]RegisteredListenerWrapper //commandName -> Sorted Listeners
	allListeners       map[string]CommandListener             //ListenerName -> listener (for removal)

	//Processing Controls
	currentRound uint64 //Might need to move this so its available outside here.
	tickRate     time.Duration
	minTickTime  time.Duration = 50 * time.Millisecond

	//Sync - The idea here is make adding to the queue able to be done w/out a reference
	// to the queue processor, but hide the queue processing itself from outsiders.
	gameQueueChan   chan *CommandContext = make(chan *CommandContext, 300)
	systemQueueChan chan *CommandContext = make(chan *CommandContext, 100)
	stopChan        chan struct{}        = make(chan struct{})
)

// TODO : Investigate a real Queue vs array for queues
type QueueProcessor struct {
	//Game queue for player/npc commands
	gameQueue   []*CommandContext
	gameRunning bool

	//System queue for immediate system commands
	systemQueue   []*CommandContext
	systemRunning bool

	//stats
	commandsProcessed  uint64
	listenersTriggered uint64
}

// New QueueProcessor creates a new queue processor
func NewQueueProcessor(tr time.Duration) *QueueProcessor {
	tickRate = tr
	return &QueueProcessor{
		gameQueue:   make([]*CommandContext, 0),
		systemQueue: make([]*CommandContext, 0),
	}
}

func RegisteredListener(listener CommandListener, commandNames ...string) {
	cmdListenerLock.Lock()
	defer cmdListenerLock.Unlock()

	listenerName := listener.Name()

	if allListeners == nil {
		allListeners = map[string]CommandListener{}
	}

	//store in alllisteners for removal purposes
	allListeners[listenerName] = listener

	//Register for each command name
	for _, commandName := range commandNames {
		registerListenerForCommand(listener, commandName)
	}
}

// Registers this listener for all commands. Mostly helpful for something like a logger
func RegisterListenerForAllCommands(listener CommandListener) {
	listenerName := listener.Name()
	allListeners[listenerName] = listener

	//Add to a special "+" key for global listeners
	registerListenerForCommand(listener, "+")
}

func registerListenerForCommand(listener CommandListener, name string) {
	registeredListener := RegisteredListenerWrapper{
		Listener: listener,
		Priority: listener.Priority(),
	}

	if listenersByCommand == nil {
		listenersByCommand = map[string][]RegisteredListenerWrapper{}
	}

	if _, ok := listenersByCommand[name]; !ok {
		listenersByCommand[name] = []RegisteredListenerWrapper{}
	}

	//Add to command's listener list
	listeners := listenersByCommand[name]
	listeners = append(listeners, registeredListener)

	//Sort by priority (lower = higher priority)
	for i := len(listeners) - 1; i > 0; i-- {
		//Swap positions if lower
		if listeners[i].Priority < listeners[i-1].Priority {
			listeners[i], listeners[i-1] = listeners[i-1], listeners[i]
		} else {
			break
		}
	}
	//set our command listeners now that they are sorted.
	listenersByCommand[name] = listeners
}

func UnregisterListener(listenerName string) bool {
	cmdListenerLock.Lock()
	defer cmdListenerLock.Unlock()

	_, exists := allListeners[listenerName]
	if !exists {
		return false
	}

	//remove from all listeners
	delete(allListeners, listenerName)

	//Remove from all command listener lists
	for commandName, listeners := range listenersByCommand {
		filtered := make([]RegisteredListenerWrapper, 0, len(listeners))
		for _, regListener := range listeners {
			if regListener.Listener.Name() != listenerName {
				filtered = append(filtered, regListener)
			}
		}

		if len(filtered) == 0 {
			delete(listenersByCommand, commandName)
		} else {
			listenersByCommand[commandName] = filtered
		}
	}
	return true
}

// QueueGameCommand adds a command to the gameQueue
func QueueGameCommand(userId uint64, command Command) {
	ctx := &CommandContext{
		UserId:    userId,
		Command:   command,
		Data:      make(map[string]any),
		Round:     currentRound,
		Timestamp: time.Now(),
	}

	select {
	case gameQueueChan <- ctx:
		//successfully queued
	default:
		//queue full, drop command
		logger.Warn("Command dropped due to size limits", "user", userId, "cmd", command.Name())
	}
}

// QueueDelayedCommand adds a command to be processed after a delay
func QueueDelayedCommand(userId uint64, command Command, delay time.Duration) {
	// Create a delayed command wrapper
	delayedCmd := &DelayedCommandWrapper{
		wrappedCommand: command,
		delay:          delay,
		scheduledFor:   time.Now().Add(delay),
	}

	ctx := &CommandContext{
		UserId:    userId,
		Command:   delayedCmd,
		Data:      make(map[string]any),
		Round:     currentRound,
		Timestamp: time.Now(),
	}

	// Add to game queue - it will be processed when the delay expires
	select {
	case gameQueueChan <- ctx:
		// Successfully queued
	default:
		// Queue full, drop command
		logger.Warn("Delayed Command dropped due to size limits", "user", userId, "cmd", command.Name())
	}
}

// Start begins processing queues
func (qp *QueueProcessor) Start() {
	qp.gameRunning = true
	qp.systemRunning = true

	// Start system queue processor (immediate)
	go qp.processSystemQueue()

	// Start game queue processor (timed)
	go qp.processGameQueue()
}

// Stop halts queue processing
func (qp *QueueProcessor) Stop() {
	qp.gameRunning = false
	qp.systemRunning = false
	close(stopChan)
}

// processSystemQueue handles system commands immediately
func (qp *QueueProcessor) processSystemQueue() {
	for qp.systemRunning {
		select {
		case ctx := <-systemQueueChan:
			qp.processCommand(ctx)
		case <-stopChan:
			return
		}
	}
}

// processGameQueue handles game commands with timing control
func (qp *QueueProcessor) processGameQueue() {
	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	for qp.gameRunning {
		select {
		case <-ticker.C:
			qp.processGameRound()
		case <-stopChan:
			return
		}
	}
}

// processGameRound processes one round of game commands
func (qp *QueueProcessor) processGameRound() {
	roundStart := time.Now()
	currentRound++

	//collect all queued commands for this round
	var roundCommands []*CommandContext

	//Drain the queue channel into our round batch
	//TODO: Do i need locking here?
	collecting := true
	for collecting {
		select {
		case ctx := <-gameQueueChan:
			//Check if this is a delayed command that is ready
			if delayedCmd, ok := ctx.Command.(*DelayedCommandWrapper); ok {
				if time.Now().Before(delayedCmd.scheduledFor) {
					//Not ready, requeue it for later
					gameQueueChan <- ctx
					continue
				}

				//Ready to process, unwrap the command
				ctx.Command = delayedCmd.wrappedCommand
			}

			ctx.Round = currentRound
			roundCommands = append(roundCommands, ctx)
		default:
			collecting = false
		}
	}

	//Process all commands in this round
	for _, ctx := range roundCommands {
		result := qp.processCommand(ctx)

		//handle requeue requests
		if result == CancelRequeue {
			gameQueueChan <- ctx
		}
	}

	//Ensure minimum time between rounds
	procesingTime := time.Since(roundStart)
	if procesingTime < minTickTime {
		time.Sleep(minTickTime - procesingTime)
	}
}

// processCommand runs a command through its registered listeners (0(1) lookup)
func (qp *QueueProcessor) processCommand(ctx *CommandContext) CommandResult {
	qp.commandsProcessed++

	commandName := ctx.Command.Name()

	//Get listners for this command
	commandListeners := listenersByCommand[commandName]

	//Check for global listners
	globalListeners := listenersByCommand["+"]

	//Combine
	allListeners := make([]RegisteredListenerWrapper, 0, len(commandListeners)+len(globalListeners))
	allListeners = append(allListeners, commandListeners...)
	allListeners = append(allListeners, globalListeners...)

	//Sort by priority
	if len(allListeners) > 1 {
		qp.sortListenersByPriority(allListeners)
	}

	//Process through listeners in priority roder
	for _, regListener := range allListeners {
		qp.listenersTriggered++

		result := regListener.Listener.Handle(ctx)

		switch result {
		case Cancel:
			return result //stop processing further listeners
		case CancelRequeue:
			continue //TODO
		case Continue:
			//move to next
			continue
		}
	}
	return Continue
}

// sortListenersByPriority sorts listeners by priority (lower = higher priority)
func (qp *QueueProcessor) sortListenersByPriority(listeners []RegisteredListenerWrapper) {
	// Simple insertion sort since the lists are usually small and mostly sorted
	for i := 1; i < len(listeners); i++ {
		key := listeners[i]
		j := i - 1

		for j >= 0 && listeners[j].Priority > key.Priority {
			listeners[j+1] = listeners[j]
			j--
		}
		listeners[j+1] = key
	}
}

// GetStats returns processing statistics
func (qp *QueueProcessor) GetStats() map[string]any {
	commandCounts := make(map[string]int)
	for commandName, listeners := range listenersByCommand {
		commandCounts[commandName] = len(listeners)
	}

	return map[string]any{
		"current_round":        currentRound,
		"commands_processed":   qp.commandsProcessed,
		"listeners_triggered":  qp.listenersTriggered,
		"game_queue_size":      len(gameQueueChan),
		"system_queue_size":    len(systemQueueChan),
		"registered_listeners": len(allListeners),
		"command_listener_map": commandCounts,
		"tick_rate":            tickRate.String(),
	}
}

// GetListenersForCommand returns listener names for a specific command (debugging)
func (qp *QueueProcessor) GetListenersForCommand(commandName string) []string {
	listeners := listenersByCommand[commandName]
	names := make([]string, len(listeners))
	for i, regListener := range listeners {
		names[i] = regListener.Listener.Name()
	}
	return names
}
