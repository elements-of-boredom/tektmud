package listeners

import (
	"strings"
	"tektmud/internal/commands"
	"tektmud/internal/logger"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
)

type DisplayRoomListener struct {
	areaManager   *rooms.AreaManager
	playerManager *players.PlayerManager
	tmpl          *templates.TemplateManager
}

func NewDisplayRoomListener(am *rooms.AreaManager,
	pm *players.PlayerManager,
	template *templates.TemplateManager) *DisplayRoomListener {
	return &DisplayRoomListener{
		areaManager:   am,
		playerManager: pm,
		tmpl:          template,
	}
}

func (dr DisplayRoomListener) Priority() int { return 1 }
func (dr DisplayRoomListener) Name() string  { return `Message Handler` }

func (dr DisplayRoomListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	disp, ok := ctx.Command.(commands.DisplayRoom)
	if !ok {
		logger.Error("Command", "Expected", "DisplayRoom", "Actual", ctx.Command.Name())
		return commands.Continue
	}
	areaId, roomId := rooms.FromKey(disp.RoomKey)
	roomDesc := dr.areaManager.FormatRoom(areaId, roomId, dr.tmpl)

	if room := rooms.LoadRoom(areaId, roomId); room != nil {
		var others []string
		for _, p := range room.GetPlayers() {
			if ur, err := dr.playerManager.GetPlayerById(p); err == nil {
				if ur.Id != disp.PlayerId {
					others = append(others, ur.Char.Name)
				}
			}
		}
		if len(others) > 0 {
			roomDesc += "Also here: " + strings.Join(others, ", ")
		}
	}

	if player, err := dr.playerManager.GetPlayerById(disp.PlayerId); err == nil {
		player.SendText(roomDesc)
	}
	return commands.Continue
}
