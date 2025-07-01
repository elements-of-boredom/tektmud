package character

import (
	"time"
)

type CharacterActionState int

const (
	Unset         CharacterActionState = iota
	Dead                               //Fully dead - In embrace flow.
	Downed                             //Not dead (embrace death) but not alive.
	Incapacitated                      //accepts almost no player input
	QuestFrozen                        // used to prevent all actions for quest purposes
	Sleeping                           //Sleeping - regenerating health, mana & Endurance - must WAKE, prevents -ALL- but wake
	Stunned                            // Prevents all movement/ability actions. Allows Score etc
	Meditating                         //Meditating - regenerating mana & willpower - Cancels on any action
	Prone                              //Sitting, laying, anything preventing standing
	Standing                           //The best state.
)

type Stats struct {
	Force  int `yaml:"force"`
	Reflex int `yaml:"reflex"`
	Acuity int `yaml:"acuity"`
	Heart  int `yaml:"heart"`
}

type Character struct {
	Id      uint64   `yaml:"id"`
	Name    string   `yaml:"name"`
	RaceId  int      `yaml:"race_id"`
	Stats   Stats    `yaml:"stats"`
	ClassId int      `yaml:"class_id"`
	Gender  string   `yaml:"gender"`
	Balance *Balance `yaml:"balance"`

	//Combat/Life related info
	Level        int                  `yaml:"level"`
	Xp           uint32               `yaml:"xp"`
	xpPercent    int                  `yaml:"-"`
	Hp           int                  `yaml:"hp"`
	MaxHp        int                  `yaml:"max_hp"`
	Mana         int                  `yaml:"mana"`
	MaxMana      int                  `yaml:"max_mana"`
	Willpower    int                  `yaml:"willpower"`
	MaxWillpower int                  `yaml:"max_willpower"`
	Endurance    int                  `yaml:"endurance"`
	MaxEndurance int                  `yaml:"max_endurance"`
	ActionState  CharacterActionState `yaml:"action_state"`

	//Location information
	RoomId string `yaml:"room_id"`
	AreaId string `yaml:"area_id"`

	AdminCtx *AdminContext

	// Persistence facade - these would be saved/loaded
	SavedHandlers []string `yaml:"saved_handlers,omitempty"`
	LastLocation  string   `yaml:"last_location,omitempty"`
}

// NewCharacter creates a new character
func NewCharacter(id uint64, name string, raceId, classId int, gender string) *Character {
	char := &Character{
		Id:       id,
		Name:     name,
		RaceId:   raceId,
		Stats:    RacesById[raceId].Stats,
		ClassId:  classId,
		Gender:   gender,
		AdminCtx: nil, // No admin rights by default
	}

	char.ResetBalances()
	return char
}

func (c *Character) ResetBalances() {
	c.Balance = NewBalance()
	// Set default balance cooldowns
	c.Balance.SetCooldown(PhysicalBalance, 2*time.Second)
	c.Balance.SetCooldown(MentalBalance, 2*time.Second)
	c.Balance.SetCooldown(MovementBalance, 100*time.Millisecond)
}

func (c *Character) SetLocation(areaId, roomId string) {
	c.AreaId = areaId
	c.RoomId = roomId
}
func (c *Character) GetXpAsPercentOfLevel() int {
	return c.xpPercent
}
func (c *Character) GetLocation() (areaId string, roomId string) {
	return c.AreaId, c.RoomId
}
func (c *Character) GetAdminContext() *AdminContext {
	return c.AdminCtx
}

func (c *Character) SetXpTo(xp uint32) (levelChangedBy int) {
	c.Xp = xp
	return c.xpChanged()
}

func (c *Character) ApplyXp(xp int) (levelChangedBy int) {
	if xp > 0 {
		c.Xp += uint32(xp)
	} else {
		c.Xp -= uint32((xp * -1))
	}

	return c.xpChanged()
}

func (c *Character) xpChanged() int {
	//Calculate new level
	originalLvl := c.Level
	adjustedLvl := findLevelFromXp(c.Xp)
	c.Level = adjustedLvl

	//calculate xp% once vs everytime we look at it.
	xpRange := xpTable[c.Level+1] - xpTable[c.Level]
	xpAt := c.Xp - xpTable[c.Level]
	c.xpPercent = int((float64(xpAt) / float64(xpRange)) * 100)

	if originalLvl != adjustedLvl {
		c.updateMaxStats()
	}
	//We leveled up. set current stats to max value on level up
	if originalLvl < adjustedLvl {
		c.Hp = c.MaxHp
		c.Mana = c.MaxMana
		c.Endurance = c.MaxEndurance
		c.Willpower = c.MaxWillpower
	} else {
		c.clampStats()
	}

	return c.Level - originalLvl
}

func (c *Character) clampStats() {
	c.Hp = max(0, min(c.Hp, c.MaxHp))
	c.Mana = max(0, min(c.Mana, c.MaxMana))
	c.Endurance = max(0, min(c.Endurance, c.MaxEndurance))
	c.Willpower = max(0, min(c.Willpower, c.MaxWillpower))
}

func findLevelFromXp(xpValue uint32) int {

	for level, xpRequired := range xpTable {
		adj := xpValue + 1
		if adj > xpRequired {
			continue
		}
		return max(1, min(100, level-1))
	}

	return 1 //We can never be lower than level 1
}

// Performs stat setup etc.
func (c *Character) Validate() bool {
	//Everyone starts at level 1. Could start at 0 i guess but eww.
	if c.Level <= 0 {
		c.Level = 1
	}

	if c.Xp <= 0 {
		c.Xp = 0
	}

	//Probably first time ever being created. Either way, setup their stats
	if c.MaxHp == 0 || c.MaxMana == 0 || c.MaxEndurance == 0 || c.MaxWillpower == 0 {
		c.updateMaxStats()
	}

	//If we are unset, this is the first time we've attempted to load
	//the character with stats
	if c.Hp == 0 && c.ActionState == Unset {
		c.Hp = c.MaxHp
		c.Mana = c.MaxMana
		c.Endurance = c.MaxEndurance
		c.Willpower = c.MaxWillpower
		c.ActionState = Standing
	}

	return true
}

func (c *Character) updateMaxStats() {
	//Very basic formulas. Need to update this at some point
	//TODO: move multipliers to config?

	//Character reaches their max hp/mana values by lvl 80
	//This is to front-load their survivability.
	if c.Level <= 80 {
		//Each character starts out at Heart * 26, and then gains slighly more or less HP as they level based on their Heart stat
		//12 is the equalizer and will be a flat 60 per level
		//For a 9 heart (Corven) max hp unadjusted is 4,632, for a Stoneheart is 5,752
		//ALL stat bonuses are capped at 25 for effect
		c.MaxHp = (min(c.Stats.Heart, 25) * 26) + ((60 + (min(c.Stats.Heart, 25)-12)*2) * c.Level)

		c.MaxMana = (min(c.Stats.Acuity, 25) * 26) + ((70 + (min(c.Stats.Heart, 25)-12)*2) * c.Level)
	}

	//These still grow per level
	c.MaxEndurance = min(c.Stats.Force, 25) * 20 * c.Level
	c.MaxWillpower = min(c.Stats.Acuity, 25) * 20 * c.Level

}

// TODO: Implement
func ValidateCharacterName(input string) bool {
	return len(input) > 1 //Allow names like Xi etc
	//Needs to validate against all known character names
	//Needs to validate against NPC name list.
}
