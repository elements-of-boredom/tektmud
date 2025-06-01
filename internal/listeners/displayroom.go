package listeners

import (
	"strings"
	"tektmud/internal/commands"
	"tektmud/internal/logger"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
	"tektmud/internal/users"
)

type DisplayRoomListener struct {
	areaManager *rooms.AreaManager
	userManager *users.UserManager
	tmpl        *templates.TemplateManager
}

func NewDisplayRoomListener(am *rooms.AreaManager,
	um *users.UserManager,
	template *templates.TemplateManager) *DisplayRoomListener {
	return &DisplayRoomListener{
		areaManager: am,
		userManager: um,
		tmpl:        template,
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
			if ur, err := dr.userManager.GetUserById(p); err == nil {
				if ur.Id != disp.UserId {
					others = append(others, ur.Char.Name)
				}
			}
		}
		if len(others) > 0 {
			roomDesc += "Also here: " + strings.Join(others, ", ")
		}
	}

	if user, err := dr.userManager.GetUserById(disp.UserId); err == nil {
		user.SendText(roomDesc)
	}
	return commands.Continue
}
