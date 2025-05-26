package rooms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	configs "tektmud/internal/config"
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

type AreaManager struct {
	areas map[string]*Area
	mu    sync.RWMutex
}

func NewAreaManager() *AreaManager {
	return &AreaManager{
		areas: make(map[string]*Area),
	}
}

func getDataPath() string {
	c := configs.GetConfig()
	areaDataPath := filepath.Join(c.Paths.RootDataDir, c.Paths.WorldFiles)
	return areaDataPath
}

// LoadArea loads an area from file
func (am *AreaManager) LoadArea(areaID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	filename := filepath.Join(getDataPath(), areaID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read area file %s: %w", filename, err)
	}

	var area Area
	if err := json.Unmarshal(data, &area); err != nil {
		return fmt.Errorf("failed to parse area file %s: %w", filename, err)
	}

	// Initialize rooms map if nil
	if area.Rooms == nil {
		area.Rooms = make(map[string]*Room)
	}

	// Set area ID for all rooms
	for _, room := range area.Rooms {
		room.AreaId = area.Id
	}

	am.areas[area.Id] = &area
	return nil
}

// LoadAllAreas loads all area files from the data directory
func (am *AreaManager) LoadAllAreas() error {
	areaDataPath := getDataPath()
	dirEntries, err := os.ReadDir(areaDataPath)
	if err != nil {
		return fmt.Errorf("failed to read area directory %s, %w", areaDataPath, err)
	}

	var loadErrors []error
	for _, file := range dirEntries {
		if filepath.Ext(file.Name()) == ".json" {
			areaID := file.Name()[:len(file.Name())-5] // Remove .json extension
			if err := am.LoadArea(areaID); err != nil {
				loadErrors = append(loadErrors, err)
			}
		}
	}

	if len(loadErrors) > 0 {
		return fmt.Errorf("failed to load some areas: %v", loadErrors)
	}

	return nil
}

// GetArea returns an area by ID
func (am *AreaManager) GetArea(areaID string) (*Area, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	area, exists := am.areas[areaID]
	return area, exists
}

func (am *AreaManager) UpsertArea(areaId string, area *Area) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.areas[areaId] = area
}

// GetRoom returns a room by area and room ID
func (am *AreaManager) GetRoom(areaID, roomID string) (*Room, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	area, exists := am.areas[areaID]
	if !exists {
		return nil, false
	}

	room, exists := area.Rooms[roomID]
	return room, exists
}

// GetRoomExit finds an exit from a room in the given direction
func (am *AreaManager) GetRoomExit(areaID, roomID string, direction Direction) (*Exit, bool) {
	room, exists := am.GetRoom(areaID, roomID)
	if !exists {
		return nil, false
	}

	for i := range room.Exits {
		if room.Exits[i].Direction == direction {
			return &room.Exits[i], true
		}
	}

	return nil, false
}

// FindExitByKeyword finds a special exit by keyword
func (am *AreaManager) FindExitByKeyword(areaID, roomID, keyword string) (*Exit, bool) {
	room, exists := am.GetRoom(areaID, roomID)
	if !exists {
		return nil, false
	}

	for i := range room.Exits {
		for _, kw := range room.Exits[i].Keywords {
			if kw == keyword {
				return &room.Exits[i], true
			}
		}
	}

	return nil, false
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

// ValidateRoomConnections checks that all room exits point to valid destinations
func (am *AreaManager) ValidateRoomConnections() []error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var errors []error

	for areaID, area := range am.areas {
		for roomID, room := range area.Rooms {
			for _, exit := range room.Exits {
				// Parse destination (could be "areaID:roomID" or just "roomID")
				destAreaID := areaID
				destRoomID := exit.Destination

				if len(exit.Destination) > 0 && exit.Destination[0:1] != ":" {
					// Check if destination contains area specification
					parts := SplitDestination(exit.Destination)
					if len(parts) == 2 {
						destAreaID = parts[0]
						destRoomID = parts[1]
					}
				}

				// Validate destination exists
				if _, exists := am.GetRoom(destAreaID, destRoomID); !exists {
					errors = append(errors, fmt.Errorf(
						"room %s:%s has exit %s pointing to invalid destination %s:%s",
						areaID, roomID, exit.Direction, destAreaID, destRoomID,
					))
				}
			}
		}
	}

	return errors
}

// GetAreaList returns a list of all loaded area IDs
func (am *AreaManager) GetAreaList() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	areas := make([]string, 0, len(am.areas))
	for areaID := range am.areas {
		areas = append(areas, areaID)
	}
	return areas
}

// SaveArea saves an area to file
func (am *AreaManager) SaveArea(areaID string) error {
	am.mu.RLock()
	area, exists := am.areas[areaID]
	am.mu.RUnlock()

	if !exists {
		return fmt.Errorf("area %s not found", areaID)
	}
	areaDataPath := getDataPath()
	filename := filepath.Join(areaDataPath, areaID+".json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(areaDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", areaDataPath, err)
	}

	// Marshal area to JSON with pretty formatting
	data, err := json.MarshalIndent(area, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal area %s: %w", areaID, err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write area file %s: %w", filename, err)
	}

	return nil
}

// SaveAllAreas saves all loaded areas to files
func (am *AreaManager) SaveAllAreas() []error {
	am.mu.RLock()
	areaIDs := make([]string, 0, len(am.areas))
	for areaID := range am.areas {
		areaIDs = append(areaIDs, areaID)
	}
	am.mu.RUnlock()

	var errors []error
	for _, areaID := range areaIDs {
		if err := am.SaveArea(areaID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (am *AreaManager) GetRoomCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	count := 0
	for _, area := range am.areas {
		count += len(area.Rooms)
	}
	return count
}

// FormatRoom returns a formatted description of a room for display
// TODO: Shift to using a template so we can colorize simply.
func (am *AreaManager) FormatRoom(areaID, roomID string) string {
	room, exists := am.GetRoom(areaID, roomID)
	if !exists {
		return "You are in an empty void."
	}

	output := fmt.Sprintf("%s\n%s\n", room.Title, room.Description)

	// Add exits
	var visibleExits []string
	for _, exit := range room.Exits {
		if !exit.Hidden {
			visibleExits = append(visibleExits, string(exit.Direction))
		}
	}

	if len(visibleExits) > 0 {
		output += fmt.Sprintf("\nExits: %s", strings.Join(visibleExits, ", "))
	} else {
		output += "\nThere are no obvious exits."
	}

	return output
}
