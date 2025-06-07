package rooms

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"tektmud/internal/character"
	"tektmud/internal/commands"
)

var (
	mu                                = sync.Mutex{}
	roomOccupants map[string][]uint64 = make(map[string][]uint64) //areaId:roomId => userid
)

// Direction represents movement directions
type Direction string

const (
	North     Direction = "north"
	South     Direction = "south"
	East      Direction = "east"
	West      Direction = "west"
	Northeast Direction = "northeast"
	Southeast Direction = "southeast"
	Southwest Direction = "southwest"
	Northwest Direction = "northwest"
	Up        Direction = "up"
	Down      Direction = "down"
	In        Direction = "in"
	Out       Direction = "out"
	Special   Direction = "special"
)

var Directions []Direction = []Direction{
	North, South, East, West,
	Northeast, Southeast, Southwest, Northwest,
	Up, Down, In, Out,
}

// DirectionAliases maps common abbreviations to directions
var DirectionAliases = map[string]Direction{
	"n":  North,
	"s":  South,
	"e":  East,
	"w":  West,
	"ne": Northeast,
	"se": Southeast,
	"sw": Southwest,
	"nw": Northwest,
	"u":  Up,
	"d":  Down,
	"i":  In,
	"o":  Out,
}

var ReverseDirections = map[Direction]Direction{
	North:     South,
	South:     North,
	East:      West,
	West:      East,
	Northeast: Southwest,
	Southwest: Northeast,
	Southeast: Northwest,
	Northwest: Southeast,
	Up:        Down,
	Down:      Up,
	In:        Out,
	Out:       In,
}

func GetReverseDirection(dir Direction) Direction {

	if reverse, exists := ReverseDirections[dir]; exists {
		return reverse
	}
	return Special // For special exits, return special
}

type AreasConfig struct {
	Areas []AreaDefinition `yaml:"areas"`
}

type AreaDefinition struct {
	Id          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Path        string            `yaml:"path"`
	Properties  map[string]string `yaml:"properties,omitempty"`
}

// Area represents a collection of rooms
type Area struct {
	Id          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Rooms       map[string]*Room  `yaml:"rooms"`
	Properties  map[string]string `yaml:"properties,omitempty"`
}

type Coordinates struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	Z int `yaml:"z"`
}

// PLACEHOLDERS
type RoomItem struct {
	Id         string `yaml:"id"`
	Quantity   int    `yaml:"quantity"`
	Respawn    *bool  `yaml:"respawn,omitempty"`
	ResetTimer *int   `yaml:"reset_timer,omitempty"`
}

type RoomNPC struct {
	Id         string `yaml:"id"`
	Quantity   int    `yaml:"quantity"`
	ResetTimer int    `yaml:"reset_timer"`
}

// Room represents a location in the game world
type Room struct {
	Id          string            `yaml:"id"`
	AreaId      string            `yaml:"-"`
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Coordinates Coordinates       `yaml:"coordinates"`
	Exits       []Exit            `yaml:"exits"`
	RoomType    string            `yaml:"room_type"`
	LightLevel  string            `yaml:"light_level"`
	Items       []RoomItem        `yaml:"items"`
	NPCs        []RoomNPC         `yaml:"npcs"`
	RoomFlags   []string          `yaml:"room_flags"`
	Scripts     []interface{}     `yaml:"scripts"`
	Triggers    []interface{}     `yaml:"triggers"`
	Properties  map[string]string `yaml:"properties,omitempty"` // Custom room properties
}

// Exit represents a connection between rooms
type Exit struct {
	Direction   Direction `yaml:"direction"`
	Destination string    `yaml:"destination"` // Room ID
	Hidden      bool      `yaml:"hidden"`      // For secret exits
	Description string    `yaml:"description,omitempty"`
	Keywords    []string  `yaml:"keywords,omitempty"` // For special exits
}

// Used just to see if the request is valid for attempting movement
// Doesn't return if they can actually go that way etc.
func (r *Room) IsExitCommand(input string) bool {
	//Check for special exits first.
	exists := false
	for i := range r.Exits {
		exists = slices.Contains(r.Exits[i].Keywords, input)
	}
	//We found a special exit, dont bother w/ the rest
	if exists {
		return exists
	}

	// Parse direction
	dirStr := strings.ToLower(input)
	_, exists = DirectionAliases[dirStr]
	if !exists {
		// Try full direction names
		for _, dir := range Directions {
			if string(dir) == dirStr {
				exists = true
				break
			}
		}
	}

	return exists
}

// Searches a room for an exit.
func (r *Room) FindExit(input string) *Exit {

	if len(r.Exits) == 0 {
		return nil
	}

	//Check for special exits first.
	for i := range r.Exits {
		if slices.Contains(r.Exits[i].Keywords, input) {
			return &r.Exits[i]
		}
	}

	// Parse direction
	dirStr := strings.ToLower(input)
	direction, exists := DirectionAliases[dirStr]
	if !exists {
		// Try full direction names
		for _, dir := range Directions {
			if string(dir) == dirStr {
				direction = dir
				exists = true
				break
			}
		}
	}

	//Check known exits
	for i := range r.Exits {
		if r.Exits[i].Direction == direction {
			return &r.Exits[i]
		}
	}

	return nil
}

// splitDestination parses destination strings like "area:room" or just "room"
func SplitDestination(destination string) []string {
	for i, char := range destination {
		if char == ':' {
			return []string{destination[:i], destination[i+1:]}
		}
	}
	return []string{destination}
}

func LoadRoom(areaId, roomId string) *Room {
	if r, exists := areaManager.GetRoom(areaId, roomId); !exists {
		return nil
	} else {
		return r
	}
}

func MoveToRoom(user *character.Character, origin *Room, destination *Room) error {

	RemoveFromRoom(user.Id, origin.AreaId, origin.Id)

	user.SetLocation(destination.AreaId, destination.Id)

	AddToRoom(user.Id, destination.AreaId, destination.Id)

	return nil
}

func (r *Room) GetPlayers() []uint64 {
	roomKey := MakeKey(r.AreaId, r.Id)
	mu.Lock()
	defer mu.Unlock()
	if len(roomOccupants[roomKey]) == 0 {
		return []uint64{}
	}
	return roomOccupants[MakeKey(r.AreaId, r.Id)]
}

// Used to find the direction the person enters from
func (r *Room) FindExitTo(areaId, roomId string) string {
	for _, exit := range r.Exits {
		if exit.Destination == roomId ||
			exit.Destination == fmt.Sprintf("%s:%s", areaId, roomId) {
			return string(exit.Direction)
		}
	}

	return ""
}

func (r *Room) ShowRoom(userId uint64) {
	//command.ShowRoom to user - will need a template manager
	commands.QueueGameCommand(0, commands.DisplayRoom{
		UserId:  userId,
		RoomKey: MakeKey(r.AreaId, r.Id),
	})
}

func (r *Room) SendText(message string, toExclude ...uint64) {
	commands.QueueGameCommand(0, commands.Message{
		RoomKey:         MakeKey(r.AreaId, r.Id),
		ExcludedUserIds: toExclude,
		Text:            message,
	})
}

func (r *Room) SendAreaText(message string, toExclude ...uint64) {
	for key := range roomOccupants {
		if strings.HasPrefix(key, r.AreaId) {
			commands.QueueGameCommand(0, commands.Message{
				RoomKey:         key,
				ExcludedUserIds: toExclude,
				Text:            message,
			})
		}
	}
}

func RemoveFromRoom(userId uint64, areaId, roomId string) {
	roomKey := MakeKey(areaId, roomId)
	mu.Lock()
	if characters, exists := roomOccupants[roomKey]; exists {
		roomOccupants[roomKey] = slices.DeleteFunc(characters, func(x uint64) bool { return x == userId })
	}
	mu.Unlock()
}

func AddToRoom(userId uint64, areaId, roomId string) {
	destKey := MakeKey(areaId, roomId)
	mu.Lock()
	if roomOccupants[destKey] == nil {
		roomOccupants[destKey] = []uint64{}
	}
	roomOccupants[destKey] = append(roomOccupants[destKey], userId)
	mu.Unlock()
}

func MakeKey(areaId, roomId string) string {
	return fmt.Sprintf("%s:%s", areaId, roomId)
}

func FromKey(roomKey string) (areaId string, roomId string) {
	parts := SplitDestination(roomKey)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", parts[0]
}
