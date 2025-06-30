package playercommands

import (
	"strings"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
)

func Templates(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	//Reload them all
	if len(args) == 0 {
		player.SendText("Syntax for templates: templates <clear> <all|name>")
		return true, nil
	}

	arguments := strings.Fields(args)

	if len(arguments) > 0 {
		switch arguments[0] {
		case "clear":
			if len(arguments) > 1 {
				templates.ClearCache(arguments[1])
			} else {
				templates.ClearCache()
			}
		}

		return true, nil
	}

	return true, nil

}
