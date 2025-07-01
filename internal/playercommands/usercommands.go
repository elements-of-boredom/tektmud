package playercommands

import (
	"tektmud/internal/players"
	"tektmud/internal/rooms"
)

var (
	PlayerHandlers = map[string]PlayerCommandHandler{
		`look`:  {Look, false},
		`l`:     {Look, false}, //provide simple shortcut for `look`
		`move`:  {Move, false},
		`quit`:  {Quit, false},
		`say`:   {Say, false},
		`'`:     {Say, false}, //Provide a shortcut for say using a single quote //TODO: Handle 'Hi vs ' Hi
		`score`: {Score, false},
		`sc`:    {Score, false}, //Provide shortcut for score
		`tell`:  {Tell, false},

		`whisper`: {Tell, false}, //Provide an alias for tell
		`yell`:    {Yell, false},

		//Admin commands
		`templates`: {Templates, true},
		`tb`:        {TestBalance, true},
		`doto`:      {DoTo, true},
	}
)

type PlayerCommandHandler struct {
	Func           PlayerCommand
	IsAdminCommand bool
}

type PlayerCommand func(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error)
