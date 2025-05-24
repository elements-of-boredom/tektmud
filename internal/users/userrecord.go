package users

type UserRecord struct {
	Id       uint64   `yaml:"id"`
	Username string   `yaml:"username"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`

	connectionId string
	inputBlocked bool
}
