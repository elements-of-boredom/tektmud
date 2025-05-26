package server

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/language"
	"tektmud/internal/templates"
	"tektmud/internal/users"
	"tektmud/internal/world"
	"time"
)

// Just our general server structure
type MudServer struct {
	config            *configs.Config
	connectionManager *connections.ConnectionManager
	userManager       *users.UserManager
	worldManager      *world.WorldManager
	templateManager   *templates.TemplateManager
	listeners         []net.Listener
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

func NewMudServer(configPath string) (*MudServer, error) {

	config, err := configs.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find config at %s,  %w", configPath, err)
	}

	//create context to manage graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	//initialize our UserManager
	userDir := filepath.Join(config.Paths.RootDataDir, config.Paths.UserData)
	userManager, err := users.NewUserManager("users.idx", userDir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create a usermanager: %w", err)
	}

	connMgr, err := connections.NewConnectionManager(config)
	if err != nil {
		//No real value in calling cancel
		//but the compiler doesn't like we don't use it in this path
		cancel()
		return nil, fmt.Errorf("failed to create a connection manager %w", err)
	}

	//Templates & Localization
	language.Initialize() //make sure i18n support is setup
	tm := templates.NewTemplateManager()

	//bootup our world manager
	wm := world.NewWorldManager()
	if err := wm.Initialize(userManager); err != nil {
		//We need to bail if this errored.
		cancel()
		return nil, fmt.Errorf("failed to initialize the world manager %w", err)
	}

	//Initalize server components
	server := &MudServer{
		config:            config,
		connectionManager: connMgr,
		userManager:       userManager,
		templateManager:   tm,
		worldManager:      wm,
		ctx:               ctx,
		cancel:            cancel,
	}

	return server, nil
}

func (s *MudServer) Initialize() error {
	slog.Info("Initializing ...")

	//Create any of our data directories that may be empty.
	//load any required things
	s.worldManager.Start()
	return nil
}

// Starts the MUD server with new listners
func (s *MudServer) Start() error {
	slog.Info("Starting %s on ports %v", s.config.Server.Name, s.config.Server.Ports)

	//Start our listeners on all configured ports
	for _, port := range s.config.Server.Ports {
		if err := s.startListener(port); err != nil {
			//Close any listeners already started
			s.stopListeners()
			return fmt.Errorf("failed to start listener on port %d: %w", port, err)
		}
	}

	//Start background tasks

	slog.Info("Server started successfully", "port(s)", len(s.listeners))
	return nil
}

func (s *MudServer) Shutdown() error {
	slog.Warn("Shutting down server...")

	//Cancel the context to signal a shutdown
	s.cancel()

	//Close all listeners
	s.stopListeners()

	//close all connections
	s.connectionManager.CloseAll(func(c *connections.PlayerConnection) {
		s.sendToPlayer(c, "Server is shuttding down. Goodbyte!")
		c.Conn.Close()
	})

	//Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	//Wait for graceful shutdown, or timeout and die
	select {
	case <-done:
		slog.Warn("Server shutdown complete")
	case <-time.After(10 * time.Second):
		slog.Warn("Shutdown timeout reached, forcing exit")
	}

	return nil
}

// closes all TCP listeners
func (s *MudServer) stopListeners() {
	for _, listener := range s.listeners {
		listener.Close()
	}
	s.listeners = nil
}

func (s *MudServer) startListener(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s.listeners = append(s.listeners, listener)

	//Start accepting connections in a new goroutine
	s.wg.Add(1)
	go s.acceptConnections(listener, port)

	slog.Info("Listener started", "port", port)
	return nil
}

func (s *MudServer) acceptConnections(listener net.Listener, port int) {
	defer s.wg.Done()

	//We are off in our own goroutine, eternally wait to accept,
	//unless context tells us we are shutting down
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				//due to timing, we could be in the middle of shutting down.
				//if we are just bail, otherwise, do nothing which will reject the connection
				select {
				case <-s.ctx.Done():
					return
				default:
					slog.Warn("Error accepting connection.", "port", port, "error", err)
				}
			}

			//spinup a new goroutine (with this model -EVERY- player gets their own goroutine)
			//there is for sure a limit to how well this scales. I'm also sure i'll never hit it
			//but if i do, go here: https://github.com/maurice2k/tcpserver
			go s.handleNewConnection(conn, port)
		}

	}
}

func (s *MudServer) handleNewConnection(conn net.Conn, port int) {
	//Check to see if we are at max capacity
	currentCount := s.connectionManager.GetConnectionCount()
	if currentCount >= s.config.Server.MaxPlayers {
		conn.Write([]byte("Sorry, the server is full. Please try again later."))
		conn.Close()
		return
	}

	//Create a player connection
	playerConn := &connections.PlayerConnection{
		Id:         connections.GenerateConnectionId(),
		Conn:       conn,
		Reader:     bufio.NewReader(conn),
		Writer:     bufio.NewWriter(conn),
		LastActive: time.Now(),
	}
	playerConn.SetState(connections.StateConnected)

	s.connectionManager.Add(playerConn)

	slog.Info("New connection added",
		"from", conn.RemoteAddr().String(),
		"port", port)

	//Hand off to the login flow
	s.handlePlayerLogin(playerConn)
}

// Begins our login process for a player. The actual state machine is in
// `procesLoginInput`
func (s *MudServer) handlePlayerLogin(pc *connections.PlayerConnection) {
	//Feels weird, but if we leave this method, we are leaving it all
	defer func(id connections.ConnectionId) {
		s.connectionManager.Remove(id)
		slog.Info("Closed connection", "connId", pc.Id)
	}(pc.Id)

	//Send the initial welcome message/Splash text
	output, err := s.templateManager.Process("login/welcome-splash")
	if err != nil {
		slog.Error("Prompt template error", "template", "login/welcome-splash", "error", err)
		output = fmt.Sprintf("Error generating propt template '%s'", "splash")
	}
	s.sendToPlayer(pc, output)
	s.sendToPlayer(pc, fmt.Sprintf("Welcome to %s!", s.config.Server.Name))
	s.sendToPlayer(pc, "What is your name?")

	pc.SetState(connections.StateUsername)

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
				slog.Warn("Connection timeout due to inactivity at login", "connId", pc.Id)
				return
			}

			pc.LastActive = time.Now()
			input := strings.TrimSpace(line)

			//This returns false until they can get through login
			//either with existing account, or new.
			if s.processLoginInput(pc, input, loginData) {
				//If we make it to here, we are through login (or user creation)
				//Add the player to the world manager
				//pass into our "game loop" that handles commands from the player
				loginData = nil
				slog.Info("Player logged in.", "player", pc.Username, "id", pc.Id)

				s.handlePlayerSession(pc)
				return //If we ever leave handlePlayerSession our defer will cleanup.
			}

			if pc.GetState() == connections.StateRejectedAuthentication {
				return
			}
		}
	}

}

func (s *MudServer) processLoginInput(pc *connections.PlayerConnection, input string, stateData map[string]any) bool {
	switch pc.GetState() {
	case connections.StateUsername:
		if input == "" {
			//Just resend them the username question
			s.sendToPlayer(pc, "What is your name?")
			return false
		}
		pc.Username = input

		//Lookup the player
		user, err := s.userManager.GetUserByUsername(input)
		if err != nil {
			//New User
			s.sendToPlayer(pc, "Welcome new player! Choose a password.")
			pc.SetState(connections.StateNewPassword)
		} else {
			//Existing user
			stateData["userid"] = user.Id
			s.sendToPlayer(pc, "Enter your password.")
			pc.SetState(connections.StatePassword)
		}
		return false //We haven't completed login

	case connections.StatePassword:
		userId, exists := stateData["userid"]
		if !exists {
			//Should never get here but...
			s.sendToPlayer(pc, "What is your name?")
			return false
		}
		userId64, ok := userId.(uint64)
		if !ok {
			slog.Error("Some how we cannot parse our userId")
			s.sendToPlayer(pc, "What is your name?")
			return false
		}
		//Finally, validate their password against their existing password.
		if s.userManager.ValidatePassword(input, userId64) {
			s.sendToPlayer(pc, fmt.Sprintf("Welcome back, %s!", pc.Username))
			pc.SetState(connections.StateAuthenticated)
			return true //login complete
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
		return false //Something didn't work right

	case connections.StateNewPassword:
		//Validate the user's password. For now we are gonna auto-pass TODO
		valid := s.userManager.PasswordMeetsMinimums(input, pc.Username)
		if !valid {
			s.sendToPlayer(pc, "Passwords must be at least 6 characters, and cannot be your username.\r\nChoose a password:")
			return false
		}
		//password was good, throw it in cache, and confirm it.
		stateData["password"] = input
		s.sendToPlayer(pc, fmt.Sprintf("Please confirm your password, %s", pc.Username))
		pc.SetState(connections.StateConfirmPassword)
		return false

	case connections.StateConfirmPassword:
		if input == stateData["password"] {
			s.sendToPlayer(pc, "Would you like to associate an email address? Without one, you will be unable to recover your account if you forget your password. If so, enter one now, or simply press enter to continue. \r\n(You may also set one later in game.)")
			pc.SetState(connections.StateCollectEmail)
			return false
		}

		s.sendToPlayer(pc, "Those passwords did not match. Please try again.")
		s.sendToPlayer(pc, "Welcome new player! Choose a password.")
		pc.SetState(connections.StateNewPassword)
		return false

	case connections.StateCollectEmail:
		passString := stateData["password"].(string)

		s.userManager.CreateUser(pc.Username, passString, input)
		//set authenticated
		pc.SetState(connections.StateAuthenticated)
		return true
	}

	//our catchall
	return false
}

// This is our ultimate game loop for input mgmt.
func (s *MudServer) handlePlayerSession(pc *connections.PlayerConnection) {
	//For now we are just going to tell us how to exit
	//and otehrwise echo stuff back.
	s.sendToPlayer(pc, "Your are now in the game! type 'quit' to exit.")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			line, err := pc.Reader.ReadString('\n')
			if err != nil {
				slog.Error("Unknown error reading connection", "user", pc.Username, "error", err)
				return
			}

			input := strings.TrimSpace(line)
			if input == "quit" {
				s.sendToPlayer(pc, "Goodbye!")
				return
			}

			//Otherwise echo back
			s.sendToPlayer(pc, fmt.Sprintf("You said: '%s'", input))
		}
	}
}

func (s *MudServer) sendToPlayer(playerConn *connections.PlayerConnection, message string) {
	//need to lock to prevent different requests writing at the same time
	playerConn.Mu.Lock()
	defer playerConn.Mu.Unlock()

	if !strings.HasSuffix(message, "\r\n") {
		message += "\r\n"
	}
	playerConn.Writer.WriteString(message)
	playerConn.Writer.Flush()
}
