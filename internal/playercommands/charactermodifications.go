package playercommands

import (
	"fmt"
	"slices"
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

	action := strings.ToLower(arguments[1])
	if slices.Contains([]string{"grantxp", "removexp", "setxp"}, action) {
		var impacted int
		var err error
		switch action {
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
	}

	if slices.Contains([]string{"harm", "heal"}, action) {
		var err error
		switch action {
		case "harm":
			err = DoHarm(arguments[2:], targetPlayer)
		case "heal":
			err = DoHeal(arguments[2:], targetPlayer)
		}
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func DoHarm(args []string, player *players.PlayerRecord) error {
	if len(args) != 3 {
		return fmt.Errorf("incorrect use of DoHarm. Expects doto <player> harm <hp|mana|end|wp> <intvalue> <damagetype if hp>")
	}

	if !slices.Contains([]string{"hp", "mana", "end", "wp"}, args[0]) {
		return fmt.Errorf("incorrect use of DoHarm. Expects doto <player> harm <hp|mana|end|wp> <intvalue> <damagetype if hp>")
	}
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	if args[0] == "hp" {
		if !slices.Contains([]string{"fire", "cold", "electrical", "blunt", "slashing", "poison", "radiation", "sonic", "suffocation"}, args[2]) {
			return fmt.Errorf("incorrect use of DoHarm. Expects doto <player> harm <hp|mana|end|wp> <intvalue> <damagetype if hp>")
		}

		applied := player.Char.ApplyDamage(amount, args[2])
		impact := "attacked"
		if applied < 0 {
			impact = "healed"
		}
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your %s %s for %d\n", args[0], impact, applied))

	}

	if args[0] == "mana" {
		player.Char.Mana = min(player.Char.MaxMana, max(0, player.Char.Mana-amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your %s %s for %d\n", args[0], "reduced", amount))
	}
	if args[0] == "end" {
		player.Char.Endurance = min(player.Char.MaxEndurance, max(0, player.Char.Endurance-amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your endurance %s for %d\n", "reduced", amount))
	}
	if args[0] == "wp" {
		player.Char.Willpower = min(player.Char.MaxWillpower, max(0, player.Char.Willpower-amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your willpower %s for %d\n", "reduced", amount))
	}

	return nil
}

func DoHeal(args []string, player *players.PlayerRecord) error {

	if len(args) != 2 {
		return fmt.Errorf("incorrect use of DoHeal. Expects doto <player> heal <hp|mana|end|wp> <intvalue>")
	}

	if !slices.Contains([]string{"hp", "mana", "end", "wp"}, args[0]) {
		return fmt.Errorf("incorrect use of DoHeal. Expects doto <player> heal <hp|mana|end|wp> <intvalue>")
	}
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	if args[0] == "hp" {
		player.Char.Hp = min(player.Char.MaxHp, max(0, player.Char.Hp+amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your %s %s for %d\n", args[0], "increased", amount))
	}

	if args[0] == "mana" {
		player.Char.Mana = min(player.Char.MaxMana, max(0, player.Char.Mana+amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your %s %s for %d\n", args[0], "increased", amount))
	}
	if args[0] == "end" {
		player.Char.Endurance = min(player.Char.MaxEndurance, max(0, player.Char.Endurance+amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your endurance %s for %d\n", "increased", amount))
	}
	if args[0] == "wp" {
		player.Char.Willpower = min(player.Char.MaxWillpower, max(0, player.Char.Willpower+amount))
		player.SendText(fmt.Sprintf("From out of nowhere, you feel your willpower %s for %d\n", "increased", amount))
	}

	return nil
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
