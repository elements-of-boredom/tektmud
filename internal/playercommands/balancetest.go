package playercommands

import (
	"slices"
	"strconv"
	"strings"
	"tektmud/internal/character"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"time"
)

func TestBalance(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	if len(args) <= 0 {
		player.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	arguments := strings.Fields(args)
	if len(arguments) != 2 {
		player.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	if !slices.Contains([]string{"physical", "mental", "movement"}, strings.ToLower(arguments[0])) {
		player.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	if n, err := strconv.ParseFloat(arguments[1], 32); err == nil {
		player.Char.Balance.UseBalance(character.BalanceType(arguments[0]), time.Duration(n)*time.Second)
		return true, nil
	} else {
		player.SendText("Error :" + err.Error())
		return true, nil
	}
}
