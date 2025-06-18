package usercommands

import (
	"slices"
	"strconv"
	"strings"
	"tektmud/internal/character"
	"tektmud/internal/rooms"
	"tektmud/internal/users"
	"time"
)

func TestBalance(args string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if len(args) <= 0 {
		user.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	arguments := strings.Fields(args)
	if len(arguments) != 2 {
		user.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	if !slices.Contains([]string{"physical", "mental", "movement"}, strings.ToLower(arguments[0])) {
		user.SendText("To use: tb <physical|mental|movement> <N.N>")
		return true, nil
	}

	if n, err := strconv.ParseFloat(arguments[1], 32); err == nil {
		user.Char.Balance.UseBalance(character.BalanceType(arguments[0]), time.Duration(n)*time.Second)
		return true, nil
	} else {
		user.SendText("Error :" + err.Error())
		return true, nil
	}
}
