package usercommands

import (
	"tektmud/internal/rooms"
	"tektmud/internal/users"
)

func Look(args []string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	room.ShowRoom(user.Id)

	return true, nil
}
