package usercommands

import (
	"fmt"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
	"tektmud/internal/users"
)

func Yell(args string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if len(args) <= 0 {
		user.SendText("You open your mouth to yell, but nothing comes out.")
		return true, nil
	}
	/*
		TODO:
		if user.silenced {
			user.SendText("You open your mouth to yell, but forget what you were doing.")
			return true, nil
		}
	*/

	//Get all rooms in the current area with people.

	room.SendAreaText(fmt.Sprintf("$Y%s yells, \"%s\"$n\n", user.Char.Name, args), user.Id)
	user.SendText(templates.Colorize(fmt.Sprintf("$YYou yell, \"%s\"$n\n", args), false))
	return true, nil
}
