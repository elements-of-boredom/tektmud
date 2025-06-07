package listeners

import (
	"fmt"
	"strings"
	"tektmud/internal/commands"
	"tektmud/internal/logger"
	"tektmud/internal/rooms"
	"tektmud/internal/usercommands"
	"tektmud/internal/users"
)

type InputListener struct {
	areaManager *rooms.AreaManager
	userManager *users.UserManager
}

func NewInputListener(am *rooms.AreaManager, um *users.UserManager) *InputListener {
	return &InputListener{
		areaManager: am,
		userManager: um,
	}
}

func (il InputListener) Priority() int { return 100 }
func (il InputListener) Name() string  { return `Input Handler` }

func (il InputListener) Handle(ctx *commands.CommandContext) commands.CommandResult {

	input, ok := ctx.Command.(commands.Input)
	if !ok {
		logger.Error("Command", "Expected", "Input", "Actual", ctx.Command.Name())
	}

	//Check to see if we are ignoring commands for this user.
	//If so pitch it.
	user, err := il.userManager.GetUserById(input.UserId)
	if err != nil {
		logger.Error("User not found", "UserId", input.UserId, "err", err)
		return commands.Continue
	}

	handled := false

	if len(input.Text) > 0 {
		//before all else check to see if this is movement.
		//simplest way to do this is see if the input is a known room exit.

		parts := strings.Fields((strings.TrimSpace(input.Text)))

		room, exists := il.areaManager.GetRoom(user.Char.AreaId, user.Char.RoomId)
		if !exists {
			logger.Error("Room not found", "AreaId", user.Char.AreaId, "RoomId", user.Char.RoomId)
			return commands.Continue
		}

		var cmd string = ""
		var args []string = make([]string, max(len(parts)-1, 1))
		if len(parts) > 0 {
			cmd = parts[0]
			args = parts[1:]
		}

		isExit := room.IsExitCommand(cmd)
		if isExit {
			//This is a movement command
			args[0] = cmd
			cmd = `move`
		}

		cmdHandler, ok := usercommands.UserHandlers[cmd]
		if ok {
			//If this is an admin command and they aren't an admin just act like we dont
			//know this command exists.
			if cmdHandler.IsAdminCommand && !user.HasRole(users.RoleAdmin) {
				logger.Warn("User attempted admin command but is not an Admin", "userId", user.Id, "cmd", cmd, "args", fmt.Sprintf("[%s]", strings.Join(args, " ")))
				//TODO do we tell the user we failed here?
				return commands.Continue
			}
			//Otherwise run the command
			handled, err = cmdHandler.Func(args, user, room)
			if err != nil {
				logger.Error("CmdHandler.Func", "err", err, "cmd", cmd, "args", "args", fmt.Sprintf("[%s]", strings.Join(args, " ")))
			}
		}

		//Its not a general user command see
		//if this is a special class command
		//TODO: Check user skills first (often cheaper/free)
		if !handled {
		}

		//If we make it here, nothing above properly handled this.
		//Throw the "huh?" equivalent
		if !handled {
			user.SendText(fmt.Sprintf("%s is not a valid command.", cmd))
		}

	} else {
		//They just hit enter... Resend the prompt for now
	}

	return commands.Continue
}
