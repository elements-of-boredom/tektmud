package listeners

import (
	"tektmud/internal/commands"
	"tektmud/internal/logger"
	"tektmud/internal/players"
)

type PromptListener struct {
	playerManager *players.PlayerManager
}

func NewPromptListener(pm *players.PlayerManager) *PromptListener {
	return &PromptListener{
		playerManager: pm,
	}
}

func (pl PromptListener) Priority() int { return 1 }
func (pl PromptListener) Name() string  { return `Prompt Handler` }

func (pl PromptListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	pq, ok := ctx.Command.(commands.SendPrompt)
	if !ok {
		logger.Error("Command", "Expected", "SendPrompt", "Actual", ctx.Command.Name())
		return commands.Continue
	}

	if player, err := pl.playerManager.GetPlayerById(pq.PlayerId); err == nil {
		player.SendPrompt()
	} else {
		logger.Error("Command", "SendPrompt", "Error", err)
	}
	return commands.Continue
}
