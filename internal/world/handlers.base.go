package world

import (
	"strings"
	"tektmud/internal/character"
	"tektmud/internal/rooms"
)

// This file contains all the basic handles a user always has.

type MovementHandler struct {
	BaseHandler
}

func NewMovementHandler() *MovementHandler {
	return &MovementHandler{
		BaseHandler: NewBaseHandler("movement", 100),
	}
}

func (h *MovementHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	//check if its a movement command
	direction, isMovement := rooms.DirectionAliases[ctx.Command]
	if !isMovement {
		for _, dir := range rooms.Directions {
			if string(dir) == ctx.Command {
				direction = dir
				isMovement = true
				break
			}
		}
	}

	if !isMovement {
		return HandlerContinue, nil
	}

	//Check movement balance
	if !ctx.Character.Balance.HasBalance(character.MovementBalance) {
		ctx.World.SendToCharacter(ctx.Character, "You must wait before moving again.")
		return HandlerStop, nil
	}

	// Attempt movement
	success, message := ctx.World.MoveCharacter(ctx.Character, direction)
	if success {
		ctx.Character.Balance.UseBalance(character.MovementBalance)
		// Look at new room automatically
		ctx.World.ShowRoom(ctx.Character)
	} else {
		ctx.World.SendToCharacter(ctx.Character, message)
	}

	return HandlerStop, nil
}

type LookHandler struct {
	BaseHandler
}

func NewLookHandler() *LookHandler {
	return &LookHandler{
		BaseHandler: NewBaseHandler("look", 50),
	}
}

func (h *LookHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	if ctx.Command != "look" && ctx.Command != "l" {
		return HandlerContinue, nil
	}

	if len(ctx.Args) == 0 {
		// Look at current room
		ctx.World.ShowRoom(ctx.Character)
	} else {
		// Look at specific thing (not implemented yet)
		// TODO
		ctx.World.SendToCharacter(ctx.Character, "You don't see that here.")
	}

	return HandlerStop, nil
}

type SayHandler struct {
	BaseHandler
}

func NewSayHandler() *SayHandler {
	return &SayHandler{
		BaseHandler: NewBaseHandler("say", 50),
	}
}

func (h *SayHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	if ctx.Command != "say" && ctx.Command != "'" {
		return HandlerContinue, nil
	}

	if len(ctx.Args) == 0 {
		ctx.World.SendToCharacter(ctx.Character, "Say what?")
		return HandlerStop, nil
	}

	//TODO: Template
	message := strings.Join(ctx.Args, " ")
	ctx.World.SendToRoom(ctx.Character.AreaId, ctx.Character.RoomId,
		ctx.Character.Name+" says: "+message, ctx.Character.Id)
	ctx.World.SendToCharacter(ctx.Character, "You say: "+message)

	return HandlerStop, nil
}

type QuitHandler struct {
	BaseHandler
}

func NewQuitHandler() *QuitHandler {
	return &QuitHandler{
		BaseHandler: NewBaseHandler("quit", 1000), // High priority
	}
}

func (h *QuitHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	if ctx.Command != "quit" && ctx.Command != "q" {
		return HandlerContinue, nil
	}

	ctx.World.RemoveCharacter(ctx.Character.Id)
	return HandlerStop, nil
}

type DefaultHandler struct {
	BaseHandler
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{
		BaseHandler: NewBaseHandler("default", -1), // Lowest priority
	}
}

func (h *DefaultHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	if ctx.Command == "" {
		return HandlerStop, nil // Empty command, do nothing
	}

	ctx.World.SendToCharacter(ctx.Character, "I don't understand that command.")
	return HandlerStop, nil
}
