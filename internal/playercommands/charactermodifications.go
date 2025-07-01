package playercommands

import (
	"fmt"
	"strconv"
	"strings"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
)

// Expected usage: doto <playername> <action> <arguments...>
func DoTo(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	if len(args) == 0 {
		return false, fmt.Errorf("received a GrantXp command with no arguments")
	}

	arguments := strings.Fields(args)
	if len(arguments) < 2 {
		return false, fmt.Errorf("expected usage: doto <playername> <action> <arguments...>, received %s", args)
	}
	playerName := arguments[0]

	targetPlayer := players.GetByCharacterName(playerName)
	if targetPlayer == nil {
		return false, fmt.Errorf("unable to find player with name:%s , are they in the realm currently?", playerName)
	}

	action := arguments[1]
	var impacted int
	var err error
	switch strings.ToLower(action) {
	case "grantxp":
		impacted, err = GrantXp(arguments[2:], targetPlayer)
	case "removexp":
		impacted, err = RemoveXp(arguments[2:], targetPlayer)
	case "setxp":
		impacted, err = SetXp(arguments[2:], targetPlayer)
	default:
		return false, fmt.Errorf("unknown action: %s", action)
	}
	if err != nil {
		return false, err
	}

	if impacted > 0 {
		targetPlayer.SendText(fmt.Sprintf("*** Congratulations! You have reached level %d! ***\n", targetPlayer.Char.Level))
	} else if impacted < 0 {
		targetPlayer.SendText(fmt.Sprintf("*** You sigh deeply, feeling the loss of knowledge as you fall to level %d ***\n", targetPlayer.Char.Level))
	} else {
		targetPlayer.SendText("You feel your experience shift in unknown ways.\n")
	}

	return true, nil
}

func GrantXp(args []string, player *players.PlayerRecord) (int, error) {

	if len(args) != 1 {
		return 0, fmt.Errorf("invalid argument count to GrantXp, expects 1")
	}
	if xp, err := strconv.ParseInt(args[0], 10, 64); err == nil {
		var impacted = player.Char.ApplyXp(int(xp))
		return impacted, nil

	} else {
		return 0, err
	}
}

func RemoveXp(args []string, player *players.PlayerRecord) (int, error) {

	if len(args) != 1 {
		return 0, fmt.Errorf("invalid argument count to RemoveXp, expects 1")
	}
	if xp, err := strconv.ParseInt(args[0], 10, 64); err == nil {
		var impacted = player.Char.ApplyXp(int(xp) * -1)
		return impacted, nil

	} else {
		return 0, err
	}
}

func SetXp(args []string, player *players.PlayerRecord) (int, error) {

	if len(args) != 1 {
		return 0, fmt.Errorf("invalid argument count to RemoveXp, expects 1")
	}
	if xp, err := strconv.ParseInt(args[0], 10, 64); err == nil {
		var impacted = player.Char.SetXpTo(uint32(xp))
		return impacted, nil

	} else {
		return 0, err
	}
}
