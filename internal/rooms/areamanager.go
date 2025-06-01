package rooms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"
	"tektmud/internal/templates"
)

var (
	areaManager = NewAreaManager()
)

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

func Initialize() (am *AreaManager, err error) {
	if err := areaManager.LoadAllAreas(); err != nil {
		return nil, fmt.Errorf("failed to load areas: %w", err)
	}

	//Validate the room connections
	if errors := areaManager.ValidateRoomConnections(); len(errors) > 0 {
		logger.Warn("Warning: Found rooms with connection errors:", "count", len(errors))
		for _, err := range errors {
			logger.Printf(" - %v", err)
		}
	}
	return areaManager, nil
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
func (am *AreaManager) FormatRoom(areaID, roomID string, tplm *templates.TemplateManager) string {
	room, exists := am.GetRoom(areaID, roomID)
	area, aExists := am.GetArea(areaID)
	if !exists || !aExists {
		return "You are in an empty void."
	}
	data := make(map[string]string)
	data["Title"] = room.Title
	data["AreaName"] = area.Name
	data["Description"] = room.Description

	// Add exits
	var visibleExits []string
	for _, exit := range room.Exits {
		if !exit.Hidden {
			visibleExits = append(visibleExits, string(exit.Direction))
		}
	}
	var exits string = ""
	if len(visibleExits) > 0 {
		var exitText string
		if len(visibleExits) == 1 {
			exitText = visibleExits[0]
		} else if len(visibleExits) == 2 {
			exitText = strings.Join(visibleExits, ", and ")
		} else {
			exitText = strings.Join(visibleExits[:len(visibleExits)-1], ", ") + ", and " + visibleExits[len(visibleExits)-1]
		}
		exits += fmt.Sprintf("You see exits to the %s", exitText)
	} else {
		exits += "There are no obvious exits."
	}

	data["Exits"] = exits

	output, err := tplm.Process("rooms/default", data)
	if err != nil {
		logger.Error("Unable to process template", "t", "rooms/default", "error", err)
	}
	return output
}
