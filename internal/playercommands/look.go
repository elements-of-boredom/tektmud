package playercommands

import (
	"tektmud/internal/players"
	"tektmud/internal/rooms"
)

func Look(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	room.ShowRoom(player.Id)

	return true, nil
}
