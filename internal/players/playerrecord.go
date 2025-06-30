package players

import (
	"slices"
	"tektmud/internal/character"
	"tektmud/internal/connections"
)

var (
	RoleUser    string = "user"
	RoleAdmin   string = "admin"
	RoleBuilder string = "builder"
	RoleOwner   string = "owner"
)

type PlayerRecord struct {
	Id       uint64               `yaml:"id"`
	Username string               `yaml:"username"`
	Email    string               `yaml:"email"`
	Password string               `yaml:"password"`
	Roles    []string             `yaml:"roles"`
	Char     *character.Character `yaml:"character"`

	//isDisabled bool
	conn *connections.PlayerConnection
}

func (ur *PlayerRecord) IsAdmin() bool {
	return slices.Contains(ur.Roles, RoleAdmin)
}
func (ur *PlayerRecord) IsBuilder() bool {
	return slices.Contains(ur.Roles, RoleBuilder)
}
func (ur *PlayerRecord) IsOwner() bool {
	return slices.Contains(ur.Roles, RoleOwner)
}
func (ur *PlayerRecord) HasRole(role string) bool {
	return slices.Contains(ur.Roles, role)
}
func (ur *PlayerRecord) IsDisabled() bool {
	return false //TODO impelment this
}

func (ur *PlayerRecord) SetConnection(c *connections.PlayerConnection) {
	ur.conn = c
}

func (ur *PlayerRecord) SendText(input string) {
	//Enqueue message
	if ur.conn != nil {
		ur.conn.Conn.Write([]byte(input))
	}
}
