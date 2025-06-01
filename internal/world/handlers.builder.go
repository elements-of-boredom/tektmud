package world

import (
	"fmt"
	"strings"
	"tektmud/internal/character"
	"tektmud/internal/logger"
	"tektmud/internal/rooms"
	"time"
)

// AdminHandler handles administrative commands
type BuilderHandler struct {
	BaseHandler
}

func NewBuilderHandler() *BuilderHandler {
	return &BuilderHandler{
		BaseHandler: NewBaseHandler("builder", 10), // High priority
	}
}

func (h *BuilderHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	// Check if character has admin context
	adminCtx := ctx.Character.GetAdminContext()
	if adminCtx == nil || adminCtx.HasRole(character.AdminRoleNone) {
		return HandlerContinue, nil
	}

	switch ctx.Command {
	case "acreate", "areacreate":
		return h.handleAreaCreate(ctx, adminCtx)
	case "rcreate", "roomcreate":
		return h.handleRoomCreate(ctx, adminCtx)
	case "redit", "roomedit":
		return h.handleRoomEdit(ctx, adminCtx)
	case "aedit", "areaedit":
		return h.handleAreaEdit(ctx, adminCtx)
	case "dig":
		return h.handleDig(ctx, adminCtx)
	case "asave", "areasave":
		return h.handleAreaSave(ctx, adminCtx)
	case "wsave", "worldsave":
		return h.handleWorldSave(ctx, adminCtx)
	case "editmode":
		return h.handleEditMode(ctx, adminCtx)
	case "astat", "areastat":
		return h.handleAreaStat(ctx, adminCtx)
	case "rstat", "roomstat":
		return h.handleRoomStat(ctx)
	default:
		return HandlerContinue, nil
	}
}

func (h *BuilderHandler) handleAreaCreate(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("create_area") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to create areas.")
		return HandlerStop, nil
	}

	if len(ctx.Args) < 2 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: acreate <area_id> <area_name>")
		return HandlerStop, nil
	}

	areaId := strings.ToLower(ctx.Args[0])
	areaName := strings.Join(ctx.Args[1:], " ")

	// Check if area already exists
	if _, exists := ctx.World.areaManager.GetArea(areaId); exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' already exists.", areaId))
		return HandlerStop, nil
	}

	// Create new area
	area := &rooms.Area{
		Id:          areaId,
		Name:        areaName,
		Description: "A newly created area.",
		Rooms:       make(map[string]*rooms.Room),
		Properties:  make(map[string]string),
	}

	// Add creator info
	area.Properties["created_by"] = ctx.Character.Name
	area.Properties["created_at"] = time.Now().Format(time.RFC3339)

	// Save to world
	ctx.World.areaManager.UpsertArea(area.Id, area)

	// Save to file
	if err := ctx.World.areaManager.SaveArea(areaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area created but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' created and saved successfully.", areaId))
	}
	logger.GetLogger().LogAreaCreation(ctx.Character.Id, ctx.Character.Name, areaId, areaName)

	return HandlerStop, nil
}

func (h *BuilderHandler) handleAreaEdit(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("edit_area") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to edit areas.")
		return HandlerStop, nil
	}

	// Usage: aedit [area_id] [field] [value...]
	// Examples:
	//   aedit                           - Show current area info
	//   aedit name New Area Name        - Change current area name
	//   aedit village name Village Name - Change specific area name
	//   aedit desc A new description for this area
	//   aedit property level_range 10-20

	currentAreaID, _ := ctx.Character.GetLocation()
	targetAreaID := currentAreaID
	argOffset := 0

	// Check if first argument is an area ID
	if len(ctx.Args) > 0 {
		if _, exists := ctx.World.areaManager.GetArea(ctx.Args[0]); exists {
			targetAreaID = ctx.Args[0]
			argOffset = 1 // Skip the area ID argument
		}
	}

	// Get the target area
	area, exists := ctx.World.areaManager.GetArea(targetAreaID)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' not found.", targetAreaID))
		return HandlerStop, nil
	}

	// If no field specified, show area information
	if len(ctx.Args) <= argOffset {
		return h.showAreaEditInfo(ctx, area)
	}

	field := strings.ToLower(ctx.Args[argOffset])

	// Ensure we have enough arguments for the field
	if len(ctx.Args) <= argOffset+1 {
		return h.showAreaEditUsage(ctx, field)
	}

	// Get the new value (remaining arguments joined)
	newValue := strings.Join(ctx.Args[argOffset+1:], " ")

	// Edit the field
	switch field {
	case "name":
		return h.editAreaName(ctx, area, newValue)
	case "desc", "description":
		return h.editAreaDescription(ctx, area, newValue)
	case "property", "prop":
		return h.editAreaProperty(ctx, area, ctx.Args[argOffset+1:])
	default:
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Unknown field '%s'. Valid fields: name, desc, property", field))
		return HandlerStop, nil
	}
}

func (h *BuilderHandler) showAreaEditInfo(ctx *InputContext, area *rooms.Area) (HandlerResult, error) {
	var properties []string
	for key, value := range area.Properties {
		properties = append(properties, fmt.Sprintf("%s=%s", key, value))
	}
	//TODO: Template
	info := fmt.Sprintf(`Area Edit Information:
ID: %s
Name: %s
Description: %s
Room Count: %d

Properties: %s

Usage:
  aedit name <new name>
  aedit desc <new description>
  aedit property <key> <value>`,
		area.Id,
		area.Name,
		area.Description,
		len(area.Rooms),
		strings.Join(properties, ", "))

	ctx.World.SendToCharacter(ctx.Character, info)
	return HandlerStop, nil
}

func (h *BuilderHandler) showAreaEditUsage(ctx *InputContext, field string) (HandlerResult, error) {
	var usage string
	switch field {
	case "name":
		usage = "Usage: aedit name <new name>"
	case "desc", "description":
		usage = "Usage: aedit desc <new description>"
	case "property", "prop":
		usage = "Usage: aedit property <key> <value>"
	default:
		usage = "Valid fields: name, desc, property"
	}

	ctx.World.SendToCharacter(ctx.Character, usage)
	return HandlerStop, nil
}

func (h *BuilderHandler) editAreaName(ctx *InputContext, area *rooms.Area, newName string) (HandlerResult, error) {
	oldName := area.Name

	area.Name = newName
	area.Properties["last_edited_by"] = ctx.Character.Name
	area.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(area.Id); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Name changed but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area name changed from '%s' to '%s' and saved.", oldName, newName))
	}
	logger.GetLogger().LogAdminAction(ctx.Character.Id, ctx.Character.Name, "edit_area_name", area.Id, "previous_name", oldName)
	return HandlerStop, nil
}

func (h *BuilderHandler) editAreaDescription(ctx *InputContext, area *rooms.Area, newDesc string) (HandlerResult, error) {
	oldDesc := area.Description
	area.Description = newDesc
	area.Properties["last_edited_by"] = ctx.Character.Name
	area.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(area.Id); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Description changed but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, "Area description updated and saved.")
	}
	logger.GetLogger().LogAdminAction(ctx.Character.Id, ctx.Character.Name, "edit_area_description", area.Id, "previous_desc", oldDesc)
	return HandlerStop, nil
}

func (h *BuilderHandler) editAreaProperty(ctx *InputContext, area *rooms.Area, args []string) (HandlerResult, error) {
	if len(args) < 2 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: aedit property <key> <value>")
		return HandlerStop, nil
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	if area.Properties == nil {
		area.Properties = make(map[string]string)
	}
	area.Properties[key] = value
	area.Properties["last_edited_by"] = ctx.Character.Name
	area.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(area.Id); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Property set but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area property '%s' set to '%s' and saved.", key, value))
	}
	logger.GetLogger().LogAdminAction(ctx.Character.Id, ctx.Character.Name, "edit_area_property", key, "area", area.Id)
	return HandlerStop, nil
}

func (h *BuilderHandler) handleAreaSave(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("save_world") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to save areas.")
		return HandlerStop, nil
	}

	// Determine which area to save
	var areaID string
	if len(ctx.Args) > 0 {
		areaID = ctx.Args[0]
		// Verify area exists
		if _, exists := ctx.World.areaManager.GetArea(areaID); !exists {
			ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' not found.", areaID))
			return HandlerStop, nil
		}
	} else {
		// Save current area
		areaID, _ = ctx.Character.GetLocation()
	}

	// Save the area
	if err := ctx.World.areaManager.SaveArea(areaID); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Failed to save area '%s': %v", areaID, err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' saved successfully.", areaID))
	}

	return HandlerStop, nil
}

func (h *BuilderHandler) handleRoomCreate(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("create_room") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to create rooms.")
		return HandlerStop, nil
	}

	if len(ctx.Args) < 2 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: rcreate <room_id> <room_title>")
		return HandlerStop, nil
	}

	areaId, _ := ctx.Character.GetLocation()
	roomId := strings.ToLower(ctx.Args[0])
	roomTitle := strings.Join(ctx.Args[1:], " ")

	// Get current area
	area, exists := ctx.World.areaManager.GetArea(areaId)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, "Current area not found.")
		return HandlerStop, nil
	}

	// Check if room already exists
	if _, exists := area.Rooms[roomId]; exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' already exists in this area.", roomId))
		return HandlerStop, nil
	}

	// Create new room
	room := &rooms.Room{
		Id:          roomId,
		Title:       roomTitle,
		Description: "A newly created room.",
		AreaId:      areaId,
		Exits:       []rooms.Exit{},
		Properties:  make(map[string]string),
	}

	// Add creator info
	room.Properties["created_by"] = ctx.Character.Name
	room.Properties["created_at"] = time.Now().Format(time.RFC3339)

	// Add to area
	area.Rooms[roomId] = room

	// Save area
	if err := ctx.World.areaManager.SaveArea(areaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room created but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' created and saved successfully.", roomId))
	}
	logger.GetLogger().LogRoomCreation(ctx.Character.Id, ctx.Character.Name, areaId, roomId, roomTitle)
	return HandlerStop, nil
}

func (h *BuilderHandler) handleRoomEdit(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("edit_room") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to edit rooms.")
		return HandlerStop, nil
	}

	// Usage: redit [room_id] [field] [value...]
	// Examples:
	//   redit                    - Show current room info
	//   redit title New Title    - Change current room title
	//   redit center title New Center Title - Change specific room title
	//   redit desc A new description for this room
	//   redit property safe_zone true

	currentAreaID, currentRoomID := ctx.Character.GetLocation()
	targetRoomID := currentRoomID
	argOffset := 0

	// Check if first argument is a room ID
	if len(ctx.Args) > 0 {
		// Try to find a room with this ID in current area
		if _, exists := ctx.World.areaManager.GetRoom(currentAreaID, ctx.Args[0]); exists {
			targetRoomID = ctx.Args[0]
			argOffset = 1 // Skip the room ID argument
		}
	}

	// Get the target room
	room, exists := ctx.World.areaManager.GetRoom(currentAreaID, targetRoomID)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' not found in area '%s'.", targetRoomID, currentAreaID))
		return HandlerStop, nil
	}

	// If no field specified, show room information
	if len(ctx.Args) <= argOffset {
		return h.showRoomEditInfo(ctx, room)
	}

	field := strings.ToLower(ctx.Args[argOffset])

	// Ensure we have enough arguments for the field
	if len(ctx.Args) <= argOffset+1 {
		return h.showRoomEditUsage(ctx, field)
	}

	// Get the new value (remaining arguments joined)
	newValue := strings.Join(ctx.Args[argOffset+1:], " ")

	// Edit the field
	switch field {
	case "title":
		return h.editRoomTitle(ctx, room, newValue)
	case "desc", "description":
		return h.editRoomDescription(ctx, room, newValue)
	case "property", "prop":
		return h.editRoomProperty(ctx, room, ctx.Args[argOffset+1:])
	case "exit":
		return h.editRoomExit(ctx, room, ctx.Args[argOffset+1:])
	default:
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Unknown field '%s'. Valid fields: title, desc, property, exit", field))
		return HandlerStop, nil
	}
}

func (h *BuilderHandler) showRoomEditInfo(ctx *InputContext, room *rooms.Room) (HandlerResult, error) {
	// Show detailed room information for editing
	var properties []string
	for key, value := range room.Properties {
		properties = append(properties, fmt.Sprintf("%s=%s", key, value))
	}

	var exits []string
	for _, exit := range room.Exits {
		exitInfo := fmt.Sprintf("%s->%s", exit.Direction, exit.Destination)
		if exit.Hidden {
			exitInfo += " (hidden)"
		}
		if len(exit.Keywords) > 0 {
			exitInfo += fmt.Sprintf(" [%s]", strings.Join(exit.Keywords, ","))
		}
		exits = append(exits, exitInfo)
	}

	info := fmt.Sprintf(`Room Edit Information:
ID: %s
Title: %s
Area: %s
Description: %s

Properties: %s
Exits: %s

Usage:
  redit title <new title>
  redit desc <new description>  
  redit property <key> <value>
  redit exit add <direction> <destination> [hidden] [keywords...]
  redit exit remove <direction>`,
		room.Id,
		room.Title,
		room.AreaId,
		room.Description,
		strings.Join(properties, ", "),
		strings.Join(exits, ", "))

	ctx.World.SendToCharacter(ctx.Character, info)
	return HandlerStop, nil
}

func (h *BuilderHandler) showRoomEditUsage(ctx *InputContext, field string) (HandlerResult, error) {
	var usage string
	switch field {
	case "title":
		usage = "Usage: redit title <new title>"
	case "desc", "description":
		usage = "Usage: redit desc <new description>"
	case "property", "prop":
		usage = "Usage: redit property <key> <value>"
	case "exit":
		usage = `Usage: 
  redit exit add <direction> <destination> [hidden] [keywords...]
  redit exit remove <direction>`
	default:
		usage = "Valid fields: title, desc, property, exit"
	}

	ctx.World.SendToCharacter(ctx.Character, usage)
	return HandlerStop, nil
}

func (h *BuilderHandler) editRoomTitle(ctx *InputContext, room *rooms.Room, newTitle string) (HandlerResult, error) {
	oldTitle := room.Title

	room.Title = newTitle
	room.Properties["last_edited_by"] = ctx.Character.Name
	room.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(room.AreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Title changed but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room title changed from '%s' to '%s' and saved.", oldTitle, newTitle))
	}

	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, room.AreaId, room.Id, "title", oldTitle, newTitle)

	// If this is the current room, show the updated room to the character
	currentAreaID, currentRoomID := ctx.Character.GetLocation()
	if room.AreaId == currentAreaID && room.Id == currentRoomID {
		room.ShowRoom(ctx.Character.Id)
	}

	return HandlerStop, nil
}

func (h *BuilderHandler) editRoomDescription(ctx *InputContext, room *rooms.Room, newDesc string) (HandlerResult, error) {
	oldDesc := room.Description
	room.Description = newDesc
	room.Properties["last_edited_by"] = ctx.Character.Name
	room.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(room.AreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Description changed but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, "Room description updated and saved.")
	}

	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, room.AreaId, room.Id, "description", oldDesc, newDesc)

	// If this is the current room, show the updated room to the character
	currentAreaId, currentRoomId := ctx.Character.GetLocation()
	if room.AreaId == currentAreaId && room.Id == currentRoomId {
		room.ShowRoom(ctx.Character.Id)
	}

	return HandlerStop, nil
}

func (h *BuilderHandler) editRoomProperty(ctx *InputContext, room *rooms.Room, args []string) (HandlerResult, error) {
	if len(args) < 2 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: redit property <key> <value>")
		return HandlerStop, nil
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	if room.Properties == nil {
		room.Properties = make(map[string]string)
	}
	room.Properties[key] = value
	room.Properties["last_edited_by"] = ctx.Character.Name
	room.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(room.AreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Property set but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room property '%s' set to '%s' and saved.", key, value))
	}
	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, room.AreaId, room.Id, key, "???", value)
	return HandlerStop, nil
}

func (h *BuilderHandler) editRoomExit(ctx *InputContext, room *rooms.Room, args []string) (HandlerResult, error) {
	if len(args) < 1 {
		ctx.World.SendToCharacter(ctx.Character, `Usage:
  redit exit add <direction> <destination> [hidden] [keywords...]
  redit exit remove <direction>`)
		return HandlerStop, nil
	}

	action := strings.ToLower(args[0])

	switch action {
	case "add":
		return h.addRoomExit(ctx, room, args[1:])
	case "remove", "delete":
		return h.removeRoomExit(ctx, room, args[1:])
	default:
		ctx.World.SendToCharacter(ctx.Character, "Exit action must be 'add' or 'remove'")
		return HandlerStop, nil
	}
}

func (h *BuilderHandler) addRoomExit(ctx *InputContext, room *rooms.Room, args []string) (HandlerResult, error) {
	if len(args) < 2 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: redit exit add <direction> <destination> [hidden] [keywords...]")
		return HandlerStop, nil
	}

	dirStr := strings.ToLower(args[0])
	destination := args[1]

	// Parse direction
	direction, exists := rooms.DirectionAliases[dirStr]
	if !exists {
		// Try full direction names
		for _, dir := range rooms.Directions {
			if string(dir) == dirStr {
				direction = dir
				exists = true
				break
			}
		}
	}

	if !exists {
		ctx.World.SendToCharacter(ctx.Character, "Invalid direction: "+dirStr)
		return HandlerStop, nil
	}

	// Check if exit already exists
	for _, exit := range room.Exits {
		if exit.Direction == direction {
			ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit %s already exists. Use 'redit exit remove %s' first.", direction, direction))
			return HandlerStop, nil
		}
	}

	// Create new exit
	newExit := rooms.Exit{
		Direction:   direction,
		Destination: destination,
		Hidden:      false,
		Description: fmt.Sprintf("An exit leads %s.", direction),
		Keywords:    []string{},
	}

	// Parse optional parameters
	for i := 2; i < len(args); i++ {
		arg := strings.ToLower(args[i])
		switch arg {
		case "hidden":
			newExit.Hidden = true
		default:
			// Treat as keyword
			newExit.Keywords = append(newExit.Keywords, arg)
		}
	}

	// Add the exit
	room.Exits = append(room.Exits, newExit)
	room.Properties["last_edited_by"] = ctx.Character.Name
	room.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)

	// Save the area
	if err := ctx.World.areaManager.SaveArea(room.AreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit added but failed to save: %v", err))
	} else {
		hiddenStr := ""
		if newExit.Hidden {
			hiddenStr = " (hidden)"
		}
		keywordStr := ""
		if len(newExit.Keywords) > 0 {
			keywordStr = fmt.Sprintf(" with keywords: %s", strings.Join(newExit.Keywords, ", "))
		}
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit %s -> %s added%s%s and saved.", direction, destination, hiddenStr, keywordStr))
	}

	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, room.AreaId, room.Id, "exits", "", string(direction))

	return HandlerStop, nil
}

func (h *BuilderHandler) removeRoomExit(ctx *InputContext, room *rooms.Room, args []string) (HandlerResult, error) {
	if len(args) < 1 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: redit exit remove <direction>")
		return HandlerStop, nil
	}

	dirStr := strings.ToLower(args[0])

	// Parse direction
	direction, exists := rooms.DirectionAliases[dirStr]
	if !exists {
		for _, dir := range rooms.Directions {
			if string(dir) == dirStr {
				direction = dir
				exists = true
				break
			}
		}
	}

	if !exists {
		ctx.World.SendToCharacter(ctx.Character, "Invalid direction: "+dirStr)
		return HandlerStop, nil
	}

	// Find and remove the exit
	found := false
	for i, exit := range room.Exits {
		if exit.Direction == direction {
			// Remove this exit
			room.Exits = append(room.Exits[:i], room.Exits[i+1:]...)
			room.Properties["last_edited_by"] = ctx.Character.Name
			room.Properties["last_edited_at"] = time.Now().Format(time.RFC3339)
			found = true
			break
		}
	}

	if !found {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("No exit found in direction %s.", direction))
		return HandlerStop, nil
	}

	// Save the area
	if err := ctx.World.areaManager.SaveArea(room.AreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit removed but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit %s removed and saved.", direction))
	}

	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, room.AreaId, room.Id, "remove_exit", "", string(direction))

	return HandlerStop, nil
}

func (h *BuilderHandler) handleDig(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("create_room") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to dig new rooms.")
		return HandlerStop, nil
	}

	if len(ctx.Args) < 3 {
		ctx.World.SendToCharacter(ctx.Character, "Usage: dig <direction> <new_room_id> <room_title>")
		return HandlerStop, nil
	}

	// Parse direction
	dirStr := strings.ToLower(ctx.Args[0])
	direction, exists := rooms.DirectionAliases[dirStr]
	if !exists {
		// Try full direction names
		for _, dir := range rooms.Directions {
			if string(dir) == dirStr {
				direction = dir
				exists = true
				break
			}
		}
	}

	if !exists {
		ctx.World.SendToCharacter(ctx.Character, "Invalid direction: "+dirStr)
		return HandlerStop, nil
	}

	currentAreaId, currentRoomId := ctx.Character.GetLocation()
	newRoomID := strings.ToLower(ctx.Args[1])
	newRoomTitle := strings.Join(ctx.Args[2:], " ")

	// Get current room
	currentRoom, exists := ctx.World.areaManager.GetRoom(currentAreaId, currentRoomId)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, "Current room not found.")
		return HandlerStop, nil
	}

	// Check if exit already exists
	if _, exists := ctx.World.areaManager.GetRoomExit(currentAreaId, currentRoomId, direction); exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Exit %s already exists.", direction))
		return HandlerStop, nil
	}

	// Create new room (similar to rcreate)
	area, _ := ctx.World.areaManager.GetArea(currentAreaId)
	if _, exists := area.Rooms[newRoomID]; exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' already exists.", newRoomID))
		return HandlerStop, nil
	}

	newRoom := &rooms.Room{
		Id:          newRoomID,
		Title:       newRoomTitle,
		Description: "A newly dug room.",
		AreaId:      currentAreaId,
		Exits:       []rooms.Exit{},
		Properties:  make(map[string]string),
	}
	newRoom.Properties["created_by"] = ctx.Character.Name
	newRoom.Properties["created_at"] = time.Now().Format(time.RFC3339)

	// Create exits in both directions
	forwardExit := rooms.Exit{
		Direction:   direction,
		Destination: newRoomID,
		Description: fmt.Sprintf("An exit leads %s.", direction),
	}

	// Determine reverse direction
	reverseDir := rooms.GetReverseDirection(direction)
	backwardExit := rooms.Exit{
		Direction:   reverseDir,
		Destination: currentRoomId,
		Description: fmt.Sprintf("An exit leads %s.", reverseDir),
	}

	// Add exits
	currentRoom.Exits = append(currentRoom.Exits, forwardExit)
	newRoom.Exits = append(newRoom.Exits, backwardExit)
	area.Rooms[newRoomID] = newRoom

	// Save area
	if err := ctx.World.areaManager.SaveArea(currentAreaId); err != nil {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room dug but failed to save: %v", err))
	} else {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' dug %s and saved successfully.", newRoomID, direction))
	}

	logger.GetLogger().LogRoomEdit(ctx.Character.Id, ctx.Character.Name, currentAreaId, currentRoomId, "dig", string(direction), newRoomID)

	return HandlerStop, nil
}

func (h *BuilderHandler) handleEditMode(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if len(ctx.Args) == 0 {
		status := "off"
		if adminCtx.IsEditMode() {
			status = "on"
		}
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Edit mode is %s.", status))
		return HandlerStop, nil
	}

	switch strings.ToLower(ctx.Args[0]) {
	case "on", "true", "1":
		adminCtx.SetEditMode(true)
		ctx.World.SendToCharacter(ctx.Character, "Edit mode enabled.")
	case "off", "false", "0":
		adminCtx.SetEditMode(false)
		ctx.World.SendToCharacter(ctx.Character, "Edit mode disabled.")
	default:
		ctx.World.SendToCharacter(ctx.Character, "Usage: editmode [on|off]")
	}

	return HandlerStop, nil
}

func (h *BuilderHandler) handleWorldSave(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	if !adminCtx.HasPermission("save_world") {
		ctx.World.SendToCharacter(ctx.Character, "You don't have permission to save the world.")
		return HandlerStop, nil
	}

	ctx.World.SendToCharacter(ctx.Character, "Saving all areas...")

	errors := ctx.World.areaManager.SaveAllAreas()
	if len(errors) > 0 {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("World saved with %d errors:", len(errors)))
		for _, err := range errors {
			ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("  - %v", err))
		}
	} else {
		ctx.World.SendToCharacter(ctx.Character, "World saved successfully.")
	}

	return HandlerStop, nil
}

// Additional handler methods for room/area editing and stats would go here...
func (h *BuilderHandler) handleAreaStat(ctx *InputContext, adminCtx *character.AdminContext) (HandlerResult, error) {
	areaID := ""
	if len(ctx.Args) > 0 {
		areaID = ctx.Args[0]
	} else {
		areaID, _ = ctx.Character.GetLocation()
	}

	area, exists := ctx.World.areaManager.GetArea(areaID)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Area '%s' not found.", areaID))
		return HandlerStop, nil
	}

	//TODO: template
	stats := fmt.Sprintf(`Area Statistics:
ID: %s
Name: %s
Description: %s
Room Count: %d
Created By: %s
Created At: %s`,
		area.Id,
		area.Name,
		area.Description,
		len(area.Rooms),
		area.Properties["created_by"],
		area.Properties["created_at"])

	ctx.World.SendToCharacter(ctx.Character, stats)
	return HandlerStop, nil
}

func (h *BuilderHandler) handleRoomStat(ctx *InputContext) (HandlerResult, error) {
	areaID, roomID := ctx.Character.GetLocation()

	if len(ctx.Args) > 0 {
		roomID = ctx.Args[0]
	}

	room, exists := ctx.World.areaManager.GetRoom(areaID, roomID)
	if !exists {
		ctx.World.SendToCharacter(ctx.Character, fmt.Sprintf("Room '%s' not found.", roomID))
		return HandlerStop, nil
	}

	exitList := make([]string, len(room.Exits))
	for i, exit := range room.Exits {
		exitList[i] = fmt.Sprintf("%s -> %s", exit.Direction, exit.Destination)
	}
	//TODO: Template
	stats := fmt.Sprintf(`Room Statistics:
ID: %s
Title: %s
Area: %s
Description: %s
Exits: %s
Created By: %s
Created At: %s`,
		room.Id,
		room.Title,
		room.AreaId,
		room.Description,
		strings.Join(exitList, ", "),
		room.Properties["created_by"],
		room.Properties["created_at"])

	ctx.World.SendToCharacter(ctx.Character, stats)
	return HandlerStop, nil
}
