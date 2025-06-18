package world

import (
	"container/heap"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"tektmud/internal/logger"
	"time"
)

// ActionType represents different types of actions that can be queued
type ActionType string

const (
	ActionPlayerCommand  ActionType = "player_command"
	ActionNPCAction      ActionType = "npc_action"
	ActionSpellEffect    ActionType = "spell_effect"
	ActionRegeneration   ActionType = "regeneration"
	ActionBalanceRestore ActionType = "balance_restore"
	ActionHeartbeat      ActionType = "heartbeat"
)

// Action represents a queued action with timing information
type Action struct {
	Id          string                             //Unique ID
	Type        ActionType                         //Type of action
	ExecuteAt   time.Time                          //When to execute
	Priority    int                                //Lower the number = higher priority
	CharacterId string                             //associated character (if any)
	Data        any                                //Action specific data
	Callback    func(*Action, *WorldManager) error //Function to execute

	//For heap implementation
	index int
}

// ActionQueue implements a priority queue for actions
type ActionQueue []*Action

func (aq ActionQueue) Len() int {
	return len(aq)
}

func (aq ActionQueue) Less(i, j int) bool {
	//Execution time first
	if aq[i].ExecuteAt.Before(aq[j].ExecuteAt) {
		return true
	}
	if aq[i].ExecuteAt.After(aq[j].ExecuteAt) {
		return false
	}

	//If they are equal, use priority
	return aq[i].Priority < aq[j].Priority
}

func (aq ActionQueue) Swap(i, j int) {
	aq[i], aq[j] = aq[j], aq[i]
	aq[i].index = i
	aq[j].index = j
}

func (aq *ActionQueue) Push(x any) {
	n := len(*aq)
	action := x.(*Action)
	action.index = n
	*aq = append(*aq, action)
}

func (aq *ActionQueue) Pop() any {
	old := *aq
	n := len(old)
	//grab the oldest action
	action := old[n-1]
	old[n-1] = nil //nil it out to remove it
	action.index = -1
	*aq = old[0 : n-1] //set the slice with our popped item removed

	return action
}

// TickManager handles the action queue and tick processing
type TickManager struct {
	queue        ActionQueue
	mu           sync.Mutex
	nextActionId int64
	tickCount    int64
}

// Creates the new Tick Manager
func NewTickManager() *TickManager {
	tm := &TickManager{
		queue: make(ActionQueue, 0),
	}
	heap.Init(&tm.queue)
	return tm
}

// QueueAction adds an action to be executed at a specific time
func (tm *TickManager) QueueAction(action *Action) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	//If we don't give this a special id, just use the incrementer
	if action.Id == "" {
		tm.nextActionId++
		action.Id = fmt.Sprintf("action_%d", tm.nextActionId)
	}
	logger.Debug("Pushing action onto queue", "id", action.Id)
	heap.Push(&tm.queue, action)
}

// QueueDelayedAction queues an action to execute after a delay
func (tm *TickManager) QueueDelayedAction(actionType ActionType, delay time.Duration, characterId string, data any, callback func(*Action, *WorldManager) error) {
	action := &Action{
		Type:        actionType,
		ExecuteAt:   time.Now().Add(delay),
		Priority:    getPriorityForActionType(actionType),
		CharacterId: characterId,
		Data:        data,
		Callback:    callback,
	}
	tm.QueueAction(action)
}

// ProcessTick process all actions that are ready to execute
func (tm *TickManager) ProcessTick(wm *WorldManager) {

	tm.tickCount++
	now := time.Now()

	//Update balances for characters
	for _, c := range wm.characters {
		p, err := wm.userManager.GetUserById(c.Id)
		if err != nil {
			continue
		}
		bals := c.Balance.GetAndRestoreBalances()
		if len(bals) > 0 {
			for _, balanceMessage := range bals {
				p.SendText(wm.tmpl.Colorize(balanceMessage+"$n\n", false))
			}
		}
	}

	//Process everything in our queue
	for tm.queue.Len() > 0 {

		next := tm.queue[0]
		if next.ExecuteAt.After(now) {
			break //
		}
		tm.mu.Lock()
		action := heap.Pop(&tm.queue).(*Action) //Pop off queue and return type to Action
		tm.mu.Unlock()
		logger.Debug("Popped item off queue", "id", action.Id)
		if action.Callback != nil {
			if err := action.Callback(action, wm); err != nil {
				logger.Error("Error executing action", "action", action.Id, "err", err)
			}
		}
	}
}

// Get QueueSize returns the current # of queued actions
func (tm *TickManager) GetQueueSize() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.queue.Len()
}

// GetTickCount returns the total number of ticks processed
func (tm *TickManager) GetTickCount() int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.tickCount
}

// Helper function to get priority for action types.
// TODO Drive from config?
func getPriorityForActionType(actionType ActionType) int {
	switch actionType {
	case ActionPlayerCommand:
		return 10
	case ActionSpellEffect:
		return 20
	case ActionNPCAction:
		return 30
	case ActionBalanceRestore:
		return 40
	case ActionRegeneration:
		return 50
	case ActionHeartbeat:
		return 100
	default:
		return 50
	}
}

// ///////////////////////////////////////////////
// /////////      Callbacks     //////////////////
// ///////////////////////////////////////////////

// PlayerCommandData holds data for player command actions
type PlayerCommandData struct {
	Command string
	Args    []string
}

// PlayerCommandCallback executes a player command
func PlayerCommandCallback(action *Action, wm *WorldManager) error {
	data, ok := action.Data.(*PlayerCommandData)
	if !ok {
		return fmt.Errorf("invalid player command data")
	}

	// Reconstruct the full command
	fullCommand := data.Command
	if len(data.Args) > 0 {
		fullCommand += " " + strings.Join(data.Args, " ")
	}

	//we know our players have good ids
	id, err := strconv.ParseUint(action.CharacterId, 10, 64)
	if err != nil {
		logger.Error("Error converting an action character id for a player to uin64", "id", action.CharacterId, "err", err)
		id = 0
	}
	return wm.HandleInputImmediate(id, fullCommand)
}

// HeartbeatCallback handles periodic world updates
func HeartbeatCallback(action *Action, wm *WorldManager) error {
	// Perform periodic world maintenance

	// TODO: Add world-wide effects like:
	// - Weather changes
	// - Day/night cycle
	// - Spawn/despawn NPCs
	// - Clean up empty rooms
	// - Update area effects

	// Queue next heartbeat (every 30 seconds)
	nextHeartbeat := &Action{
		Type:      ActionHeartbeat,
		ExecuteAt: time.Now().Add(30 * time.Second),
		Priority:  getPriorityForActionType(ActionHeartbeat),
		Data:      nil,
		Callback:  HeartbeatCallback,
	}
	wm.tickManager.QueueAction(nextHeartbeat)

	return nil
}

// NPCActionData holds data for NPC actions
type NPCActionData struct {
	NPCID      string
	ActionType string
	TargetID   string
	Data       map[string]any
}

// NPCActionCallback handles NPC actions
func NPCActionCallback(action *Action, wm *WorldManager) error {
	data, ok := action.Data.(*NPCActionData)
	if !ok {
		return fmt.Errorf("invalid NPC action data")
	}

	// TODO: Implement NPC system
	// For now, just a placeholder that demonstrates the concept

	switch data.ActionType {
	case "wander":
		// NPC randomly moves around
		// npc := wm.GetNPC(data.NPCID)
		// randomDirection := getRandomDirection()
		// wm.MoveNPC(npc, randomDirection)

	case "speak":
		// NPC says something
		// npc := wm.GetNPC(data.NPCID)
		// message := data.Data["message"].(string)
		// wm.SendToRoom(npc.AreaID, npc.RoomID, npc.Name+" says: "+message)

	case "attack":
		// NPC attacks a player
		// target := wm.GetCharacter(data.TargetID)
		// wm.InitiateCombat(data.NPCID, data.TargetID)
	}

	// Queue next NPC action (random interval between 5-15 seconds)
	nextAction := time.Now().Add(time.Duration(5+rand.Intn(10)) * time.Second)
	wm.tickManager.QueueDelayedAction(ActionNPCAction, nextAction.Sub(time.Now()), "", data, NPCActionCallback)

	return nil
}
