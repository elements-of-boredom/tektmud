package listeners

import (
	"fmt"
	"strings"
	"tektmud/internal/commands"
	"tektmud/internal/logger"
	"tektmud/internal/playercommands"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
)

type InputListener struct {
	areaManager   *rooms.AreaManager
	playerManager *players.PlayerManager
}

func NewInputListener(am *rooms.AreaManager, um *players.PlayerManager) *InputListener {
	return &InputListener{
		areaManager:   am,
		playerManager: um,
	}
}

func (il InputListener) Priority() int { return 100 }
func (il InputListener) Name() string  { return `Input Handler` }

func (il InputListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	input, ok := ctx.Command.(commands.Input)
	if !ok {
		logger.Error("Command", "Expected", "Input", "Actual", ctx.Command.Name())
	}

	//Check to see if we are ignoring commands for this player.
	//If so pitch it.
	player, err := il.playerManager.GetPlayerById(input.PlayerId)
	if err != nil {
		logger.Error("player not found", "PlayerId", input.PlayerId, "err", err)
		return commands.Continue
	}

	handled := false

	if len(input.Text) > 0 {

		parts := strings.SplitN(input.Text, " ", 2)

		room, exists := il.areaManager.GetRoom(player.Char.AreaId, player.Char.RoomId)
		if !exists {
			logger.Error("Room not found", "AreaId", player.Char.AreaId, "RoomId", player.Char.RoomId)
			return commands.Continue
		}

		var cmd string = ""

		if len(parts) > 0 {
			cmd = parts[0]
		}

		//before all else check to see if this is movement.
		//simplest way to do this is see if the input is a known room exit.
		isExit := room.IsExitCommand(cmd)
		if isExit {
			//This is a movement command
			cmd = `move`
		}

		cmdHandler, ok := playercommands.PlayerHandlers[cmd]
		if ok {

			arguments := strings.Replace(input.Text, fmt.Sprintf("%s ", cmd), "", 1)

			//If this is an admin command and they aren't an admin just act like we dont
			//know this command exists.
			if cmdHandler.IsAdminCommand && !player.HasRole(players.RoleAdmin) {
				logger.Warn("Player attempted admin command but is not an Admin", "player.Id", player.Id, "cmd", cmd, "args", fmt.Sprintf("[%s]", arguments))
				//TODO do we tell the player we failed here?
				return commands.Continue
			}
			//Otherwise run the command
			handled, err = cmdHandler.Func(arguments, player, room)
			if err != nil {
				logger.Error("CmdHandler.Func", "err", err, "cmd", cmd, "args", "args", fmt.Sprintf("[%s]", arguments))
			}
		}

		//Its not a general player command see
		//if this is a special class command
		//TODO: Check player skills first (often cheaper/free)
		if !handled {
		}

		//If we make it here, nothing above properly handled this.
		//Throw the "huh?" equivalent
		if !handled {
			player.SendText(fmt.Sprintf("%s is not a valid command.", cmd))
		}

	} else {
		//They just hit enter... Resend the prompt for now
	}

	return commands.Continue
}
