package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"tektmud/internal/character"
	configs "tektmud/internal/config"
	"tektmud/internal/connections"
	"tektmud/internal/language"
	"tektmud/internal/logger"
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

func NewMudServer() (*MudServer, error) {

	config := configs.GetConfig()
	//create context to manage graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	//initialize our UserManager
	userDir := filepath.Join(config.Paths.RootDataDir, config.Paths.UserData)
	userManager, err := users.NewUserManager("users.idx", userDir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create a usermanager: %w", err)
	}

	connMgr, err := connections.NewConnectionManager(&config)
	if err != nil {
		//No real value in calling cancel
		//but the compiler doesn't like we don't use it in this path
		cancel()
		return nil, fmt.Errorf("failed to create a connection manager %w", err)
	}

	//Templates & Localization
	language.Initialize()        //make sure i18n support is setup
	tm := templates.Initialize() //make sure Templates are setup.

	//bootup our world manager
	wm := world.NewWorldManager(userManager, tm)
	if err := wm.Initialize(); err != nil {
		//We need to bail if this errored.
		cancel()
		return nil, fmt.Errorf("failed to initialize the world manager %w", err)
	}

	//Initalize server components
	server := &MudServer{
		config:            &config,
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
	logger.Info("Initializing ...")

	//Create any of our data directories that may be empty.

	character.InitializeRaceData()
	character.InitializeClassData()

	//load any required things
	s.worldManager.Start()
	return nil
}

// Starts the MUD server with new listners
func (s *MudServer) Start() error {
	logger.Info(fmt.Sprintf("Starting %s on ports %v", s.config.Server.Name, s.config.Server.Ports))

	//Start our listeners on all configured ports
	for _, port := range s.config.Server.Ports {
		if err := s.startListener(port); err != nil {
			//Close any listeners already started
			s.stopListeners()
			return fmt.Errorf("failed to start listener on port %d: %w", port, err)
		}
	}

	//Start background tasks

	logger.Info("Server started successfully", "port(s)", len(s.listeners))
	return nil
}

func (s *MudServer) Shutdown() error {
	logger.Warn("Shutting down server...")

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
		logger.Warn("Server shutdown complete")
	case <-time.After(10 * time.Second):
		logger.Warn("Shutdown timeout reached, forcing exit")
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

	logger.Info("Listener started", "port", port)
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
					logger.Warn("Error accepting connection.", "port", port, "error", err)
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

	logger.Info("New connection added",
		"from", conn.RemoteAddr().String(),
		"port", port)

	//Hand off to the login flow
	s.handlePlayerLogin(playerConn)
}

// This is our ultimate game loop for input mgmt.
func (s *MudServer) handlePlayerSession(playerId uint64, pc *connections.PlayerConnection) {
	//For now we are just going to tell us how to exit
	//and otehrwise echo stuff back.
	s.sendToPlayer(pc, "You are now in the game! type 'quit' to exit.\n")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			line, err := pc.Reader.ReadString('\n')
			if err != nil {
				logger.Error("Unknown error reading connection", "user", pc.Username, "error", err)
				return
			}

			input := strings.TrimSpace(line)
			if err := s.worldManager.HandleInput(playerId, input); err != nil {
				logger.Error("Unknowing handling error", "user", pc.Username, "error", err)
				pc.Send("$RUnknown error handling input. Disconnecting!")
				return
			}
		}
	}
}

func (s *MudServer) sendToPlayer(playerConn *connections.PlayerConnection, message string) {
	//need to lock to prevent different requests writing at the same time
	playerConn.Mu.Lock()
	defer playerConn.Mu.Unlock()

	/*
		if !strings.HasSuffix(message, "\r\n") {
			message += "\r\n"
		}
	*/
	playerConn.Writer.WriteString(message)
	playerConn.Writer.Flush()
}
