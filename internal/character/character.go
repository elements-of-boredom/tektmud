package character

import (
	"slices"
	"time"
)

type Character struct {
	Id      uint64   `yaml:"id"`
	Name    string   `yaml:"name"`
	RoomId  string   `yaml:"room_id"`
	AreaId  string   `yaml:"area_id"`
	Balance *Balance `yaml:"balance"`

	Handlers []string
	AdminCtx *AdminContext

	// Persistence facade - these would be saved/loaded
	SavedHandlers []string `yaml:"saved_handlers,omitempty"`
	LastLocation  string   `yaml:"last_location,omitempty"`
}

// NewCharacter creates a new character
func NewCharacter(id uint64, name string) *Character {
	char := &Character{
		Id:       id,
		Name:     name,
		Balance:  NewBalance(),
		Handlers: make([]string, 0),
		AdminCtx: nil, // No admin rights by default
	}

	// Set default balance cooldowns
	char.Balance.SetCooldown(AttackBalance, 2*time.Second)
	char.Balance.SetCooldown(HealingBalance, 4*time.Second)
	char.Balance.SetCooldown(MovementBalance, 200*time.Millisecond)

	return char
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
