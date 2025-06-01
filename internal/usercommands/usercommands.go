package usercommands

import (
	"tektmud/internal/rooms"
	"tektmud/internal/users"
)

var (
	UserHandlers = map[string]UserCommandHandler{
		`look`: {Look, false},
		`l`:    {Look, false}, //provide simple shortcut for `look`
		`move`: {Move, false},
		`quit`: {Quit, false},
	}
)

type UserCommandHandler struct {
	Func           UserCommand
	IsAdminCommand bool
}

type UserCommand func(args []string, user *users.UserRecord, room *rooms.Room) (bool, error)
