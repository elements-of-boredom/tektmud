package listeners

import (
	"slices"
	"tektmud/internal/commands"
	"tektmud/internal/connections"
	"tektmud/internal/logger"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
)

type MessageListener struct {
	areaManager   *rooms.AreaManager
	playerManager *players.PlayerManager
	pcs           map[uint64]*connections.PlayerConnection
	tmpl          *templates.TemplateManager
}

func NewMessageListener(am *rooms.AreaManager,
	pm *players.PlayerManager,
	c map[uint64]*connections.PlayerConnection,
	template *templates.TemplateManager) *MessageListener {
	return &MessageListener{
		areaManager:   am,
		playerManager: pm,
		pcs:           c,
		tmpl:          template,
	}
}

func (il MessageListener) Priority() int { return 1 }
func (il MessageListener) Name() string  { return `Message Handler` }

func (il MessageListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	msg, ok := ctx.Command.(commands.Message)
	if !ok {
		logger.Error("Command", "Expected", "Message", "Actual", ctx.Command.Name())
	}

	//Message to a specific player
	if msg.PlayerId > 0 {

	}

	//Room wide message
	if len(msg.RoomKey) > 0 {
		room, exists := il.areaManager.GetRoom(rooms.FromKey(msg.RoomKey))
		if !exists {
			logger.Warn("Received a message for a room that doesn't exist.", "roomKey", msg.RoomKey, "msg", msg.Text)
			return commands.Continue
		}

		for _, playerId := range room.GetPlayers() {

			//Don't send messages to the "sender"
			if msg.PlayerId == playerId {
				continue
			}

			//Dont send messages to exlcuded Ids
			excluded := len(msg.ExcludedPlayerIds)
			if excluded > 0 {
				if slices.Contains(msg.ExcludedPlayerIds, playerId) {
					continue
				}
			}

			if player, err := il.playerManager.GetPlayerById(playerId); err == nil {
				/* TODO
				if msg.IsCommunication && user.IsDeaf {
					continue
				}
				*/
				text := il.tmpl.Colorize(msg.Text, false)
				if conn, exists := il.pcs[player.Id]; exists {
					conn.Send(text)
				}
			}
		}

	}

	return commands.Continue
}
