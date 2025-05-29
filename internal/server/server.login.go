package server

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"tektmud/internal/character"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/logger"
	"tektmud/internal/users"
	"time"
)

// Begins our login process for a player. The actual state machine is in
// `procesLoginInput`
func (s *MudServer) handlePlayerLogin(pc *connections.PlayerConnection) {
	//Feels weird, but if we leave this method, we are leaving it all
	defer func(id connections.ConnectionId) {
		s.connectionManager.Remove(id)
		logger.Info("Closed connection", "connId", pc.Id)
	}(pc.Id)

	//Send the initial welcome message/Splash text
	output, err := s.templateManager.Process("login/welcome-splash")
	if err != nil {
		logger.Error("Prompt template error", "template", "login/welcome-splash", "error", err)
		output = fmt.Sprintf("Error generating prompt template '%s'", "splash")
		s.sendToPlayer(pc, output)
		return
	}

	s.sendToPlayer(pc, output)

	s.sendToPlayer(pc, fmt.Sprintf("Welcome to %s!\n", s.config.Server.Name))
	pc.SetState(connections.StateInitialPrompt)

	loginData := make(map[string]any)
	//Handle login loop.
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			//set read deadline for idle timeout.
			if s.config.Server.IdleTimeout > 0 {
				pc.Conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Server.IdleTimeout) * time.Minute))
			}

			line, err := pc.Reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					s.sendToPlayer(pc, "Connection timed out due to inactivity.")
				}
				logger.Warn("Connection timeout due to inactivity at login", "connId", pc.Id)
				return
			}

			pc.LastActive = time.Now()
			input := strings.TrimSpace(line)

			//This returns false until they can get through login
			//either with existing account, or new.
			if user, success := s.processLoginInput(pc, input, loginData); success {
				//If we make it to here, we are through login (or user creation)
				//Add the player to the world manager
				//pass into our "game loop" that handles commands from the player
				//Add our user to the world
				char := character.NewCharacter(user.Id, pc.Username)
				var roles []character.AdminRole = []character.AdminRole{}

				//Map our user roles => character roles
				//TODO properly merge User & Character.
				if user.IsAdmin() {
					roles = append(roles, character.AdminRoleAdmin)
				}
				if user.IsBuilder() {
					roles = append(roles, character.AdminRoleBuilder)
				}
				if user.IsOwner() {
					roles = append(roles, character.AdminRoleOwner)
				}
				if len(roles) > 0 {
					char.AdminCtx = character.NewAdminContext(roles...)
				}
				s.worldManager.AddCharacter(char, pc)

				loginData = nil
				logger.GetLogger().LogPlayerConnect(user.Id, user.Username, pc.Conn.RemoteAddr().String())

				s.handlePlayerSession(user.Id, pc)
				return //If we ever leave handlePlayerSession our defer will cleanup.
			}

			if pc.GetState() == connections.StateRejectedAuthentication {
				return
			}
		}
	}
}

func (s *MudServer) processLoginInput(pc *connections.PlayerConnection, input string, stateData map[string]any) (*users.UserRecord, bool) {

	//see if at any point they entered "quit"
	if strings.ToLower(input) == `quit` {
		pc.SetState(connections.StateRejectedAuthentication)
		return nil, false

	}

	c := configs.GetConfig()
	mudData := map[string]string{
		"WorldName": c.Server.WorldName,
		"Name":      c.Server.Name,
	}

	usernamePrompt := s.templateManager.Colorize("Please enter your desired username. \r\nThis is $W**NOT**$n your character name, which will you will define later.", false)

	switch pc.GetState() {
	case connections.StateInitialPrompt:
		if input == "" {
			s.sendToPlayer(pc, "Enter your username or type 'new' to create a new character.")
			return nil, false
		}

		if strings.ToLower(input) == `new` {
			//Show New Username prompt
			s.sendToPlayer(pc, usernamePrompt)
			pc.SetState(connections.StateUsername)

		} else {
			//Lookup the player
			user, err := s.userManager.GetUserByUsername(input)
			if err != nil {
				//The username is not known.
				//Tell them we don't know who that is, reshow the prompt
				s.sendToPlayer(pc, "That username is not known. If you meant to create a new user, please type 'new', otherwise try again.")
				return nil, false
			}

			pc.Username = user.Username
			stateData["userid"] = user.Id
			s.sendToPlayer(pc, "Enter your password.")
			pc.SetState(connections.StatePassword)
		}

		return nil, false

		//This is only called if they are -creating- a new user.
	case connections.StateUsername:
		if input == "" {
			//Just resend them the username question
			s.sendToPlayer(pc, usernamePrompt)
			return nil, false
		}
		//Lookup the player
		_, err := s.userManager.GetUserByUsername(input)
		if err != nil {
			//New User
			pc.Username = input
			s.sendToPlayer(pc, s.templateManager.Colorize(
				"Welcome new player! Please enter a password. \r\n$y[Passwords must be at least 6 characters long]$n",
				false))
			pc.SetState(connections.StateNewPassword)
		} else {
			//This is a problem. They are entering a username already taken.
			//resend the username prompt after letting them know
			s.sendToPlayer(pc, "This username is not available, please try again.")
			s.sendToPlayer(pc, usernamePrompt)
			return nil, false
		}
		return nil, false //We haven't completed login

	case connections.StatePassword:
		userId, exists := stateData["userid"]
		if !exists {
			//Should never get here but...
			s.sendToPlayer(pc, usernamePrompt)
			pc.SetState(connections.StateInitialPrompt)
			return nil, false
		}
		userId64, ok := userId.(uint64)
		if !ok {
			logger.Error("Some how we cannot parse our userId")
			s.sendToPlayer(pc, "There was an error, lets start over.")
			s.sendToPlayer(pc, usernamePrompt)
			pc.SetState(connections.StateInitialPrompt)
			delete(stateData, "userid")
			return nil, false
		}
		//Finally, validate their password against their existing password.
		if s.userManager.ValidatePassword(input, userId64) {
			s.sendToPlayer(pc, fmt.Sprintf("Welcome back, %s!", pc.Username))
			//Verify they have a character on their user. If not make one.
			ur, err := s.userManager.GetUserByUsername(pc.Username)
			if err != nil {
				s.sendToPlayer(pc, "Unable to find your user file, please try again.")
				pc.SetState(connections.StateRejectedAuthentication)
				return nil, false
			}
			if ur.Char != nil {
				pc.SetState(connections.StateAuthenticated)
				return ur, true
			}

			s.sendToPlayer(pc, "It appears you've not finished character creation, lets do that now. Press any key to continue")
			pc.SetState(connections.StateCharacterEval)
			return nil, false //login complete
		} else {
			tries, ok := stateData["attempts"]
			if !ok {
				tries = 1
			}
			triesInt := tries.(int)
			if triesInt < 3 {
				s.sendToPlayer(pc, "Invalid password. Try again.")
			} else {
				s.sendToPlayer(pc, "Maximum password attempts met. Goodbye!")
				pc.SetState(connections.StateRejectedAuthentication)
			}
		}
		return nil, false //Something didn't work right

	case connections.StateNewPassword:
		//Validate the user's password. For now we are gonna auto-pass
		valid := s.userManager.PasswordMeetsMinimums(input, pc.Username)
		if !valid {
			s.sendToPlayer(pc, "Passwords must be at least 6 characters, and cannot be your username.\r\nChoose a password:")
			return nil, false
		}
		//password was good, throw it in cache, and confirm it.
		stateData["password"] = input
		s.sendToPlayer(pc, fmt.Sprintf("Please confirm your password, %s", pc.Username))
		pc.SetState(connections.StateConfirmPassword)
		return nil, false

	case connections.StateConfirmPassword:
		if input == stateData["password"] {
			s.sendToPlayer(pc, "Would you like to associate an email address? Without one, you will be unable to recover your account if you forget your password. If so, enter one now, or simply press enter to continue. \r\n(You may also set one later in game.)")
			pc.SetState(connections.StateCollectEmail)
			return nil, false
		}

		s.sendToPlayer(pc, "Those passwords did not match. Please try again.")
		s.sendToPlayer(pc, s.templateManager.Colorize(
			"Welcome new player! Please enter a password. \r\n$y[Passwords must be at least 6 characters long]$n",
			false))
		pc.SetState(connections.StateNewPassword)
		return nil, false

	case connections.StateCollectEmail:
		passString := stateData["password"].(string)

		ur, err := s.userManager.CreateUser(pc.Username, passString, input)
		if err != nil {
			s.sendToPlayer(pc, "Error creating player, please try again.")
			pc.SetState(connections.StateRejectedAuthentication)
			return nil, false
		}
		logger.Info("Created user", "username", ur.Username)
		//set authenticated
		pc.SetState(connections.StateCharacterEval)
		s.sendToPlayer(pc, "Ok, press enter to begin character creation.")
		return nil, false

	case connections.StateCharacterEval:
		//Verify they have a character on their user. If not make one.
		ur, err := s.userManager.GetUserByUsername(pc.Username)
		if err != nil {
			s.sendToPlayer(pc, "Unable to find your user file, please try again.")
			pc.SetState(connections.StateRejectedAuthentication)
			return nil, false
		}
		if ur.Char != nil {
			//They have a character already, lets get into the world
			pc.SetState(connections.StateAuthenticated)
			return ur, true
		} else {
			//Put them through Character Creation flow
			tpl, err := s.templateManager.Process("creation/gender", mudData)
			if err != nil {
				s.sendToPlayer(pc, err.Error())
				pc.SetState(connections.StateRejectedAuthentication)
				return nil, false
			}
			s.sendToPlayer(pc, tpl)
			pc.SetState(connections.StateCharacterGenderChoice)
		}
		return nil, false

	case connections.StateCharacterGenderChoice:
		if input == "" {
			return nil, false
		}
		gender := strings.ToLower(input)
		if regexp.MustCompile(`^[a-z0-9]+$`).MatchString(gender) {
			//No point in doing much parsing effort for this. its its letters or numbers
			if gender == `1` || strings.HasPrefix(gender, "m") {
				stateData["gender"] = "male"
			} else if gender == `2` || strings.HasPrefix(gender, "f") {
				stateData["gender"] = "female"
			} else {
				s.sendToPlayer(pc, "Invalid input, try again.")
				return nil, false
			}

			//TODO get Race prompt.

			pc.SetState(connections.StateCharacterRaceChoice)
			return nil, false
		}
		s.sendToPlayer(pc, "Invalid input, try again.")
		return nil, false

	case connections.StateCharacterRaceChoice:

	case connections.StateCharacterClassChoice:

	case connections.StateCharacterNameChoice:

	}

	//our catchall
	return nil, false
}
