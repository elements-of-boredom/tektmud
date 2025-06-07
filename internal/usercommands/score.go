package usercommands

import (
	"fmt"
	"strconv"
	"strings"
	"tektmud/internal/rooms"
	"tektmud/internal/templates"
	"tektmud/internal/users"
)

func Score(args []string, user *users.UserRecord, room *rooms.Room) (bool, error) {
	var template string = "playerinfo/score"
	if len(args) > 0 && strings.HasPrefix(strings.ToLower(args[0]), "full") {
		template = "playerinfo/score.full"
	}

	scoreData := map[string]string{
		"Name":      user.Char.Name,
		"Race":      strconv.Itoa(user.Char.RaceId),
		"Gender":    user.Char.Gender,
		"Class":     strconv.Itoa(user.Char.ClassId),
		"Age":       "18",
		"Level":     "0",
		"Hp":        "20/20",
		"Mana":      "15/15",
		"Endurance": "120/120",
		"ShipThing": "???",
		"Force":     strconv.Itoa(user.Char.Stats.Force),
		"Reflex":    strconv.Itoa(user.Char.Stats.Reflex),
		"Acuity":    strconv.Itoa(user.Char.Stats.Acuity),
		"Insight":   strconv.Itoa(user.Char.Stats.Insight),
		"Heart":     strconv.Itoa(user.Char.Stats.Heart),
	}
	var output string = "Error generating score data %s"

	if out, err := templates.Process(template, scoreData); err != nil {
		user.SendText(fmt.Sprintf(output, err.Error()))
	} else {
		user.SendText(out)
	}
	return true, nil
}
