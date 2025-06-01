package users

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

type UserRecord struct {
	Id       uint64               `yaml:"id"`
	Username string               `yaml:"username"`
	Email    string               `yaml:"email"`
	Password string               `yaml:"password"`
	Roles    []string             `yaml:"roles"`
	Char     *character.Character `yaml:"character"`

	//isDisabled bool
	conn *connections.PlayerConnection
}

func (ur *UserRecord) IsAdmin() bool {
	return slices.Contains(ur.Roles, RoleAdmin)
}
func (ur *UserRecord) IsBuilder() bool {
	return slices.Contains(ur.Roles, RoleBuilder)
}
func (ur *UserRecord) IsOwner() bool {
	return slices.Contains(ur.Roles, RoleOwner)
}
func (ur *UserRecord) HasRole(role string) bool {
	return slices.Contains(ur.Roles, role)
}
func (ur *UserRecord) IsDisabled() bool {
	return false //TODO impelment this
}

func (ur *UserRecord) SetConnection(c *connections.PlayerConnection) {
	ur.conn = c
}

func (ur *UserRecord) SendText(input string) {
	//Enqueue message
	if ur.conn != nil {
		ur.conn.Conn.Write([]byte(input))
	}
}
