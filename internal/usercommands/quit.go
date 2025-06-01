package usercommands

import (
	"tektmud/internal/commands"
	"tektmud/internal/rooms"
	"tektmud/internal/users"
)

func Quit(args []string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	//TODO
	//Want to show 2-3 messages before they can quit.
	//Any action should interrupt.
	//Shouldn't be able to be in combat.
	//Will probably need to handle with either a temporary buff, or some other
	//mechanism

	user.SendText(`
You tap the surface of your personal stasis cube, dropping it to the ground.
Stepping onto it you exhale slowly as nano bots begin to swarm over your body.
With one last glance at the world around you, you close your eyes and sleep.
	`)

	//
	commands.QueueGameCommand(user.Id, commands.PlayerQuit{
		UserId: user.Id,
	})

	return true, nil
}
