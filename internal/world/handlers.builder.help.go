package world

import (
	"slices"
	"tektmud/internal/character"
)

// AdminHelpHandler provides help for admin commands
type BuilderHelpHandler struct {
	BaseHandler
}

func NewBuilderHelpHandler() *BuilderHelpHandler {
	return &BuilderHelpHandler{
		BaseHandler: NewBaseHandler("builder_help", 950),
	}
}

func (h *BuilderHelpHandler) Handle(ctx *InputContext) (HandlerResult, error) {
	if ctx.Command != "help" {
		return HandlerContinue, nil
	}

	// Check if character has admin context
	adminCtx := ctx.Character.GetAdminContext()
	if adminCtx == nil || adminCtx.HasRole(character.AdminRoleNone) {
		return HandlerContinue, nil
	}

	if len(ctx.Args) == 0 || (len(ctx.Args) > 0 && ctx.Args[0] != "buliding") {
		return HandlerContinue, nil
	}

	helpText := h.getBuilderHelp(adminCtx.Roles)
	ctx.World.SendToCharacter(ctx.Character, helpText)

	return HandlerStop, nil
}

func (h *BuilderHelpHandler) getBuilderHelp(roles []character.AdminRole) string {
	var builderHelp string = ""
	var adminHelp string = ""
	if slices.Contains(roles, character.AdminRoleBuilder) {
		builderHelp = `
Building Commands Available:

WORLD BUILDING:
  acreate <id> <name>     - Create a new area
  rcreate <id> <title>    - Create a new room in current area
  dig <dir> <id> <title>  - Create a room and connect it in direction
  
EDITING:
  editmode [on|off]       - Toggle edit mode
  redit [room_id]         - Edit room properties (when implemented)
  aedit [area_id]         - Edit area properties (when implemented)
  
INFORMATION:
  astat [area_id]         - Show area statistics  
  rstat [room_id]         - Show room statistics
  
SAVING:
  asave                   - Save current area to file
  wsave                   - Save all areas to files

EXAMPLES:
  acreate forest "The Dark Forest" 
  rcreate clearing "A Peaceful Clearing"
  dig north meadow "A Grassy Meadow"
  astat
  wsave
`
	}

	//TODO Move to owner help.
	if slices.Contains(roles, character.AdminRoleAdmin) {
		adminHelp = `
OWNER COMMANDS:
  grant <player> <role>   - Grant admin role to player (when implemented)
  revoke <player>         - Remove admin role from player (when implemented)
  shutdown                - Shutdown the server (when implemented)
`
	}

	return builderHelp + adminHelp
}
