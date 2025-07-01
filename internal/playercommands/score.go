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
		"Class":     character.GetClassNameById(player.Char.RaceId),
		"Age":       "18",
		"Level":     fmt.Sprintf("%d (%d%%)", player.Char.Level, player.Char.GetXpAsPercentOfLevel()), //"64 (1%)"
		"Health":    fmt.Sprintf("%d/%d", player.Char.Hp, player.Char.MaxHp),
		"Mana":      fmt.Sprintf("%d/%d", player.Char.Mana, player.Char.MaxMana),
		"Endurance": fmt.Sprintf("%d/%d", player.Char.Endurance, player.Char.MaxEndurance),
		"Willpower": fmt.Sprintf("%d/%d", player.Char.Willpower, player.Char.MaxWillpower),
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
