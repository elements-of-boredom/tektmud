package playercommands

import (
	"fmt"
	"strings"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
)

func Say(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {
	if len(args) <= 0 {
		player.SendText("You attempt to speak, but nothing is said.")
		return true, nil
	}

	room.SendText(fmt.Sprintf("$C%s says, \"%s\"$n\n", player.Char.Name, args), player.Id)
	player.SendText(templates.Colorize(fmt.Sprintf("$CYou say, \"%s\"$n\n", args), false))
	return true, nil
}

func Tell(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {
	if len(args) <= 0 {
		player.SendText("You attempt to speak, but nothing is said.")
		return true, nil
	}

	//Ok, find our target (by name) by breaking at the first space
	parts := strings.SplitN(args, " ", 2)
	if len(parts) != 2 {
		player.SendText("You must specify who to tell.")
		return true, nil
	}

	if targetUser := players.GetByCharacterName(parts[0]); targetUser != nil {
		targetUser.SendText(templates.Colorize(fmt.Sprintf("$G%s tells you, \"%s\"$n\n", player.Char.Name, parts[1]), false))
		player.SendText(templates.Colorize(fmt.Sprintf("$GYou tell %s, \"%s\"$n\n", targetUser.Char.Name, parts[1]), false))
	} else {
		player.SendText(fmt.Sprintf("Unable to send a message to %s", parts[0]))
	}

	return true, nil
}

func Yell(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	if len(args) <= 0 {
		player.SendText("You open your mouth to yell, but nothing comes out.")
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

	room.SendAreaText(fmt.Sprintf("$Y%s yells, \"%s\"$n\n", player.Char.Name, args), player.Id)
	player.SendText(templates.Colorize(fmt.Sprintf("$YYou yell, \"%s\"$n\n", args), false))
	return true, nil
}
