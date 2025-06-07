package usercommands

import (
	"fmt"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
	"tektmud/internal/users"
)

func Say(args string, user *users.UserRecord, room *rooms.Room) (bool, error) {
	if len(args) <= 0 {
		user.SendText("You attempt to speak, but nothing is said.")
		return true, nil
	}

	room.SendText(fmt.Sprintf("$C%s says, \"%s\"$n\n", user.Char.Name, args), user.Id)
	user.SendText(templates.Colorize(fmt.Sprintf("$CYou say, \"%s\"$n\n", args), false))
	return true, nil
}
