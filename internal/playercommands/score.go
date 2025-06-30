package playercommands

import (
	"fmt"
	"strconv"
	"strings"
	"tektmud/internal/character"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
)

func Score(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {
	var template string = "playerinfo/score"
	if len(args) > 0 && strings.HasPrefix(strings.ToLower(args), "full") {
		template = "playerinfo/score.full"
	}

	scoreData := map[string]string{
		"Name":      player.Char.Name,
		"Race":      character.GetRaceNameById(player.Char.RaceId),
		"Gender":    player.Char.Gender,
		"Class":     character.GetRaceNameById(player.Char.RaceId),
		"Age":       "18",
		"Level":     "0",
		"Hp":        "20/20",
		"Mana":      "15/15",
		"Endurance": "120/120",
		"ShipThing": "???",
		"Force":     strconv.Itoa(player.Char.Stats.Force),
		"Reflex":    strconv.Itoa(player.Char.Stats.Reflex),
		"Acuity":    strconv.Itoa(player.Char.Stats.Acuity),
		"Heart":     strconv.Itoa(player.Char.Stats.Heart),
	}
	var output string = "Error generating score data %s"

	if out, err := templates.Process(template, scoreData); err != nil {
		player.SendText(fmt.Sprintf(output, err.Error()))
	} else {
		player.SendText(out)
	}
	return true, nil
}
