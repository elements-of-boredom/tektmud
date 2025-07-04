package server

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"tektmud/internal/character"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/logger"
	"tektmud/internal/players"
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

	loginData := make(map[string]string)
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
			if player, success := s.processLoginInput(pc, input, loginData); success {
				//If we make it to here, we are through login (or player creation)
				//Add the player to the world manager
				//pass into our "game loop" that handles commands from the player
				//Add our player's character to the world
				if player.Char == nil {
					gender := loginData["gender"]
					race := loginData["race"]
					class := loginData["class"]
					characterName := loginData["character_name"]

					raceData := character.GetRaceByName(race)
					classData := character.GetClassByName(class)
					if raceData == nil || classData == nil {
						//We have a problem w/ the data files
						//Return that message to the player and start over.
						logger.Error("Unable to create a player because race or class data was invalid.")
						s.sendToPlayer(pc, "Unable to create your character, the game files for your chosen class or race are invalid. Please try again.")
						return
					}
					char := character.NewCharacter(player.Id, characterName, raceData.Id, classData.Id, gender) //TODO
					player.Char = char
					s.playerManager.UpdatePlayer(player)
				}
				var roles []character.AdminRole = []character.AdminRole{}
				//Reset their balances on re-entry
				//I -don't- think this will be abusable on disconnects... we'd have to see
				//If so i'll have to start storing balances in character files.
				player.Char.ResetBalances()
				//Map our player roles => character roles
				//TODO properly merge User & Character.
				if player.IsAdmin() {
					roles = append(roles, character.AdminRoleAdmin)
				}
				if player.IsBuilder() {
					roles = append(roles, character.AdminRoleBuilder)
				}
				if player.IsOwner() {
					roles = append(roles, character.AdminRoleOwner)
				}
				if len(roles) > 0 {
					player.Char.AdminCtx = character.NewAdminContext(roles...)
				}
				player.SetConnection(pc)
				s.worldManager.AddCharacter(player.Char, pc)

				loginData = nil
				logger.GetLogger().LogPlayerConnect(player.Id, player.Username, pc.Conn.RemoteAddr().String())

				s.handlePlayerSession(player.Id, pc)
				return //If we ever leave handlePlayerSession our defer will cleanup.
			}

			if pc.GetState() == connections.StateRejectedAuthentication {
				return
			}
		}
	}
}

func (s *MudServer) processLoginInput(pc *connections.PlayerConnection, input string, stateData map[string]string) (*players.PlayerRecord, bool) {

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
			player, err := s.playerManager.GetPlayerByUsername(input)
			if err != nil {
				//The username is not known.
				//Tell them we don't know who that is, reshow the prompt
				s.sendToPlayer(pc, "That username is not known. If you meant to create a new user, please type 'new', otherwise try again.")
				return nil, false
			}

			pc.Username = player.Username
			stateData["playerid"] = fmt.Sprint(player.Id)
			s.sendToPlayer(pc, "Enter your password.")
			pc.SetState(connections.StatePassword)
		}

		return nil, false

		//This is only called if they are -creating- a new player.
	case connections.StateUsername:
		if input == "" {
			//Just resend them the username question
			s.sendToPlayer(pc, usernamePrompt)
			return nil, false
		}
		//Lookup the player
		_, err := s.playerManager.GetPlayerByUsername(input)
		if err != nil {
			//New player
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
		playerId, exists := stateData["playerid"]
		if !exists {
			//Should never get here but...
			s.sendToPlayer(pc, usernamePrompt)
			pc.SetState(connections.StateInitialPrompt)
			return nil, false
		}
		playerId64, err := strconv.ParseUint(playerId, 10, 64)
		if err != nil {
			logger.Error("Some how we cannot parse our playerId", "err", err)
			s.sendToPlayer(pc, "There was an error, lets start over.")
			s.sendToPlayer(pc, usernamePrompt)
			pc.SetState(connections.StateInitialPrompt)
			delete(stateData, "playerid")
			return nil, false
		}
		//Finally, validate their password against their existing password.
		if s.playerManager.ValidatePassword(input, playerId64) {
			s.sendToPlayer(pc, fmt.Sprintf("Welcome back, %s!\n", pc.Username))
			//Verify they have a character on their player. If not make one.
			ur, err := s.playerManager.GetPlayerByUsername(pc.Username)
			if err != nil {
				s.sendToPlayer(pc, "Unable to find your player file, please try again.")
				pc.SetState(connections.StateRejectedAuthentication)
				return nil, false
			}
			if ur.Char != nil {
				pc.SetState(connections.StateAuthenticated)
				return ur, true
			}

			//Ok, a portal with no character. initiate character creation
			if err := sendGenderPrompt(pc, s, mudData); err != nil {
				//there was an error we've already sent the error
				// and set the state to cause us to leave
				return nil, false
			}

			return nil, false //login complete
		} else {
			tries, ok := stateData["attempts"]
			var counter = 0
			if !ok {
				counter = 1
			}
			counter, err := strconv.Atoi(tries)
			if err != nil {
				counter = 3
			}
			if counter < 3 {
				s.sendToPlayer(pc, "Invalid password. Try again.")
				stateData["attempts"] = fmt.Sprint(counter)
			} else {
				s.sendToPlayer(pc, "Maximum password attempts met. Goodbye!")
				pc.SetState(connections.StateRejectedAuthentication)
			}
		}
		return nil, false //Something didn't work right

	case connections.StateNewPassword:
		//Validate the user's password. For now we are gonna auto-pass
		valid := s.playerManager.PasswordMeetsMinimums(input, pc.Username)
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
		passString := stateData["password"]

		ur, err := s.playerManager.CreatePlayer(pc.Username, passString, input)
		if err != nil {
			s.sendToPlayer(pc, "Error creating player, please try again.")
			pc.SetState(connections.StateRejectedAuthentication)
			return nil, false
		}
		logger.Info("Created user", "username", ur.Username)

		//Ok, initiate character creation
		sendGenderPrompt(pc, s, mudData)
		// if there was an error we've already sent the error
		// and set the state to cause us to leave
		// either way lets move on.
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

			s.sendToPlayer(pc, "Great! Next up is your choice of race.\n\n")

			sendRacePrompt(pc, s, mudData)

			return nil, false
		}
		s.sendToPlayer(pc, "Invalid input, try again.")
		return nil, false

	case connections.StateCharacterRaceChoice:
		if input == "" {
			return nil, false
		}

		cmd := strings.ToLower(input)

		//Not actually listed but if they forget its the go-to
		if cmd == `help` {
			sendRacePrompt(pc, s, mudData)
		}

		if cmd == `info` {
			tpl, err := s.templateManager.Process("creation/races/help", mudData)
			if err != nil {
				logger.Error("unable to load template", "tpl", "creation/races/help", "err", err)
			}
			s.sendToPlayer(pc, tpl)
		}
		if cmd == `races` {
			tpl, err := s.templateManager.Process("creation/races/default", mudData)
			if err != nil {
				logger.Error("unable to load template", "tpl", "creation/races/default", "err", err)
			}
			s.sendToPlayer(pc, tpl)
		}
		if strings.HasPrefix(cmd, "learn") {
			splits := strings.Split(cmd, " ")
			race, err := raceFromInput(splits)
			if err != nil {
				s.sendToPlayer(pc, "Invalid racial choice.")
				return nil, false
			}

			statData := character.GetStatsForRace(race)
			if statData == nil {
				logger.Error("why is our race stat data nil?", "race", race)
			}

			//TODO address nil pointer derefernce risk
			info := map[string]string{
				"Force":  fmt.Sprint(statData.Force),
				"Reflex": fmt.Sprint(statData.Reflex),
				"Acuity": fmt.Sprint(statData.Acuity),
				"Heart":  fmt.Sprint(statData.Heart),
			}

			tpl, err := s.templateManager.Process(fmt.Sprintf("creation/races/%s", strings.ToLower(race)), info)
			if err != nil {
				logger.Error("unable to load template", "tpl", fmt.Sprintf("creation/races/%s", strings.ToLower(race)), "err", err)
			}
			s.sendToPlayer(pc, tpl)
		}
		if strings.HasPrefix(cmd, "choose") {
			splits := strings.Split(cmd, " ")
			race, err := raceFromInput(splits)
			if err != nil {
				s.sendToPlayer(pc, "Invalid racial choice.")
				return nil, false
			}
			stateData["race"] = race
			//Race chosen, move on to class prompt
			sendClassPrompt(pc, s, mudData)
		}

		return nil, false

	case connections.StateCharacterClassChoice:
		if input == "" {
			return nil, false
		}
		cmd := strings.ToLower(input)

		if cmd == `help` {
			sendClassPrompt(pc, s, mudData)
		}

		if cmd == `classes` {
			tpl, err := s.templateManager.Process("creation/classes/allclasses", mudData)
			if err != nil {
				logger.Error("unable to load template", "tpl", "creation/classes/allclasses", "err", err)
			}
			s.sendToPlayer(pc, tpl)
		}

		if strings.HasPrefix(cmd, "choose") {
			splits := strings.Split(cmd, " ")
			class, err := classFromInput(splits)
			if err != nil {
				s.sendToPlayer(pc, "Invalid class choice.")
				return nil, false
			}
			stateData["class"] = class
			//class chosen, move on to name prompt
			mudData["Class"] = class
			mudData["Race"] = stateData["race"]
			mudData["Gender"] = stateData["gender"]

			sendNamePrompt(pc, s, mudData)
		}

		return nil, false

	case connections.StateCharacterNameChoice:
		if input == "" {
			return nil, false
		}

		if character.ValidateCharacterName(input) {
			ur, err := s.playerManager.GetPlayerByUsername(pc.Username)
			if err != nil {
				s.sendToPlayer(pc, "Error creating player, please try again.")
				pc.SetState(connections.StateRejectedAuthentication)
				return nil, false
			}
			stateData["character_name"] = input
			return ur, true
		}

	}

	//our catchall
	return nil, false
}

// TODO - Use game files instead of hard coded
func classFromInput(input []string) (string, error) {
	if len(input) != 2 {
		return "", fmt.Errorf("not a valid class choice")
	}

	class := input[1]

	if strings.HasPrefix(class, "an") {
		return "Animist", nil
	}
	if strings.HasPrefix(class, "au") {
		return "Augur", nil
	}
	if strings.HasPrefix(class, "di") {
		return "Distortionist", nil
	}
	if strings.HasPrefix(class, "fa") {
		return "Fabricator", nil
	}
	if strings.HasPrefix(class, "ha") {
		return "Harmonist", nil
	}
	if strings.HasPrefix(class, "me") {
		return "Mentalist", nil
	}
	if strings.HasPrefix(class, "an") {
		return "Animist", nil
	}
	if strings.HasPrefix(class, "sy") {
		return "Symbiont", nil
	}
	if strings.HasPrefix(class, "sc") {
		return "Scavenger", nil
	}
	if strings.HasPrefix(class, "va") {
		return "Vanguard", nil
	}
	if strings.HasPrefix(class, "wa") {
		return "Warden", nil
	}
	return "", fmt.Errorf("not a valid class choice")
}

func raceFromInput(input []string) (string, error) {
	if len(input) != 2 {
		return "", fmt.Errorf("not a valid race choice")
	}

	race := strings.ToLower(input[1])
	//We want to try and attempt to handle typo's and laziness
	//So this wont be pretty
	if strings.HasPrefix(race, "hu") {
		return "Human", nil
	}
	if strings.HasPrefix(race, "sy") {
		return "Synthetic", nil
	}
	if strings.HasPrefix(race, "um") {
		return "Umbran", nil
	}
	if strings.HasPrefix(race, "ve") {
		return "Verdani", nil
	}
	if strings.HasPrefix(race, "vo") {
		return "Voidborn", nil
	}
	if strings.HasPrefix(race, "st") {
		return "Stoneheart", nil
	}
	if strings.HasPrefix(race, "co") {
		return "Corvan", nil
	}

	//If we made it here they typed something weird.
	return "", fmt.Errorf("not a valid race choice")
}

func sendGenderPrompt(pc *connections.PlayerConnection, s *MudServer, mudData map[string]string) error {
	//Put them through Character Creation flow
	tpl, err := s.templateManager.Process("creation/gender", mudData)
	if err != nil {
		s.sendToPlayer(pc, err.Error())
		pc.SetState(connections.StateRejectedAuthentication)
		return err
	}
	s.sendToPlayer(pc, tpl)
	pc.SetState(connections.StateCharacterGenderChoice)
	return nil
}

func sendRacePrompt(pc *connections.PlayerConnection, s *MudServer, mudData map[string]string) error {
	tpl, err := s.templateManager.Process("creation/race", mudData)
	if err != nil {
		s.sendToPlayer(pc, err.Error())
		pc.SetState(connections.StateRejectedAuthentication)
		return err
	}
	s.sendToPlayer(pc, tpl)
	pc.SetState(connections.StateCharacterRaceChoice)
	return nil
}

func sendClassPrompt(pc *connections.PlayerConnection, s *MudServer, mudData map[string]string) error {
	tpl, err := s.templateManager.Process("creation/classes", mudData)
	if err != nil {
		s.sendToPlayer(pc, err.Error())
		pc.SetState(connections.StateRejectedAuthentication)
		return err
	}
	s.sendToPlayer(pc, tpl)
	pc.SetState(connections.StateCharacterClassChoice)
	return nil
}

func sendNamePrompt(pc *connections.PlayerConnection, s *MudServer, mudData map[string]string) error {
	tpl, err := s.templateManager.Process("creation/pickaname", mudData)
	if err != nil {
		s.sendToPlayer(pc, err.Error())
		pc.SetState(connections.StateRejectedAuthentication)
		return err
	}
	s.sendToPlayer(pc, tpl)
	pc.SetState(connections.StateCharacterNameChoice)
	return nil
}
