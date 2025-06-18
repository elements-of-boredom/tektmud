package usercommands

import (
	"tektmud/internal/rooms"
	"tektmud/internal/users"
)

var (
	UserHandlers = map[string]UserCommandHandler{
		`look`:  {Look, false},
		`l`:     {Look, false}, //provide simple shortcut for `look`
		`move`:  {Move, false},
		`quit`:  {Quit, false},
		`say`:   {Say, false},
		`'`:     {Say, false}, //Provide a shortcut for say using a single quote
		`score`: {Score, false},
		`tell`:  {Tell, false},

		`whisper`: {Tell, false}, //Provide an alias for tell
		`yell`:    {Yell, false},

		//Admin commands
		`templates`: {Templates, true},
		`tb`:        {TestBalance, true},
	}
)

type UserCommandHandler struct {
	Func           UserCommand
	IsAdminCommand bool
}

type UserCommand func(args string, user *users.UserRecord, room *rooms.Room) (bool, error)
