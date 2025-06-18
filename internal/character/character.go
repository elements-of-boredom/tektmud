package character

import (
	"slices"
	"time"
)

type Stats struct {
	Force   int `yaml:"force"`
	Reflex  int `yaml:"reflex"`
	Acuity  int `yaml:"acuity"`
	Insight int `yaml:"insight"`
	Heart   int `yaml:"heart"`
}

type Character struct {
	Id      uint64 `yaml:"id"`
	Name    string `yaml:"name"`
	RaceId  int    `yaml:"race_id"`
	Stats   Stats  `yaml:"stats"`
	ClassId int    `yaml:"class_id"`
	Gender  string `yaml:"gender"`

	Balance *Balance `yaml:"balance"`

	//Location information
	RoomId string `yaml:"room_id"`
	AreaId string `yaml:"area_id"`

	Handlers []string
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
		Handlers: make([]string, 0),
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

func (c *Character) AddHandler(name string) {
	c.Handlers = append(c.Handlers, name)
}

func (c *Character) RemoveHandler(name string) {
	c.Handlers = slices.DeleteFunc(c.Handlers, func(n string) bool {
		return n == name
	})
}

func (c *Character) SetLocation(areaId, roomId string) {
	c.AreaId = areaId
	c.RoomId = roomId
}

func (c *Character) GetLocation() (areaId string, roomId string) {
	return c.AreaId, c.RoomId
}
func (c *Character) GetAdminContext() *AdminContext {
	return c.AdminCtx
}

// TODO: Implement
func ValidateCharacterName(input string) bool {
	return len(input) > 1 //Allow names like Xi etc
	//Needs to validate against all known character names
	//Needs to validate against NPC name list.
}
