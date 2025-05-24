package users

import "slices"

var (
	RoleUser    string = "user"
	RoleAdmin   string = "admin"
	RoleCreator string = "creator"
)

type UserRecord struct {
	Id       uint64   `yaml:"id"`
	Username string   `yaml:"username"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`

	connectionId string
	inputBlocked bool
}

func (ur *UserRecord) IsAdmin() bool {
	return slices.Contains(ur.Roles, RoleAdmin)
}
func (ur *UserRecord) IsCreator() bool {
	return slices.Contains(ur.Roles, RoleCreator)
}
func (ur *UserRecord) HasRole(role string) bool {
	return slices.Contains(ur.Roles, role)
}
