package listeners

import (
	"tektmud/internal/commands"
	"tektmud/internal/logger"
)

type HandlesRemoval interface {
	RemoveCharacter(uint64)
}

type QuitListener struct {
	Remover HandlesRemoval
}

func NewQuitListener(remover HandlesRemoval) *QuitListener {
	return &QuitListener{
		Remover: remover,
	}
}

func (ql QuitListener) Priority() int { return 1 }
func (ql QuitListener) Name() string  { return `Message Handler` }

func (ql QuitListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	pq, ok := ctx.Command.(commands.PlayerQuit)
	if !ok {
		logger.Error("Command", "Expected", "PlayerQuit", "Actual", ctx.Command.Name())
		return commands.Continue
	}

	ql.Remover.RemoveCharacter(pq.UserId)
	return commands.Continue
}
