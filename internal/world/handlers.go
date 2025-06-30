package world

import (
	"strings"
	"tektmud/internal/character"
)

// InputHandler defines the interface for command handlers
type InputHandler interface {
	Name() string
	Handle(ctx *InputContext) (HandlerResult, error)
	Priority() int // For future sorting if needed
}

// HandlerResult indicates how the handler chain should proceed
type HandlerResult int

const (
	HandlerContinue HandlerResult = iota // Continue to next handler
	HandlerStop                          // Stop processing, command handled
	HandlerError                         // Stop processing due to error
)

// InputContext contains information about the command being processed
type InputContext struct {
	Character *character.Character
	RawInput  string
	Command   string
	Args      []string
	World     *WorldManager
}

// BaseHandler provides common handler functionality
type BaseHandler struct {
	name     string
	priority int
}

func NewBaseHandler(name string, priority int) BaseHandler {
	return BaseHandler{
		name:     name,
		priority: priority,
	}
}
func (h BaseHandler) Name() string  { return h.name }
func (h BaseHandler) Priority() int { return h.priority }

func ProcessInput(character *character.Character, rawInput string, world *WorldManager) error {
	//ctx := ParseInput(character, rawInput, world)

	return nil
}

func ParseInput(character *character.Character, rawInput string, world *WorldManager) *InputContext {
	parts := strings.Fields((strings.TrimSpace(rawInput)))
	ic := &InputContext{
		Character: character,
		RawInput:  rawInput,
		World:     world,
	}
	if len(parts) == 0 {
		ic.Command = ""
		ic.Args = []string{}
		return ic
	}

	ic.Command = strings.ToLower(parts[0])
	ic.Args = parts[1:]
	return ic
}
