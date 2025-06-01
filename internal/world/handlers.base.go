package world

// This file contains all the basic handles a user always has.

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
