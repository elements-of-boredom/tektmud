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
	Special   Direction = "special"
)

var Directions []Direction = []Direction{
	North, South, East, West,
	Northeast, Southeast, Southwest, Northwest,
	Up, Down,
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
}

func GetReverseDirection(dir Direction) Direction {

	if reverse, exists := ReverseDirections[dir]; exists {
		return reverse
	}
	return Special // For special exits, return special
}

// Area represents a collection of rooms
type Area struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rooms       map[string]*Room  `json:"rooms"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// Room represents a location in the game world
type Room struct {
	Id          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Exits       []Exit            `json:"exits"`
	AreaId      string            `json:"area_id"`
	Properties  map[string]string `json:"properties,omitempty"` // Custom room properties
}

// Exit represents a connection between rooms
type Exit struct {
	Direction   Direction `json:"direction"`
	Destination string    `json:"destination"` // Room ID
	Hidden      bool      `json:"hidden"`      // For secret exits
	Description string    `json:"description,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"` // For special exits
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
	//command.SendMessage to user
	commands.QueueGameCommand(0, commands.Message{
		RoomKey:         MakeKey(r.AreaId, r.Id),
		ExcludedUserIds: toExclude,
		Text:            message,
	})
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
