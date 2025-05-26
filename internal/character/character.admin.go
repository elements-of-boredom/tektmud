package character

import (
	"slices"
	"sync"
)

// Represents the different permission levels
type AdminRole string

// TODO : Convert to flags, a user could be multiple roles, and admin != builder, but an admin COULD be a builder.
const (
	AdminRoleNone    AdminRole = ""
	AdminRoleBuilder AdminRole = "builder" //can create/edit rooms & areas
	AdminRoleAdmin   AdminRole = "admin"   //Full admin acces
	AdminRoleOwner   AdminRole = "owner"   //complete acces
)

var (
	builderPermissions = []string{"create_room", "edit_room", "create_area", "edit_area", "save_world"}
)

type AdminContext struct {
	Roles       []AdminRole `yaml:"roles"`
	Permissions []string    `yaml:"permissions,omitempty"`
	EditMode    bool        `yaml:"-"`

	mu sync.RWMutex
}

// NewAdminContext
func NewAdminContext(roles ...AdminRole) *AdminContext {
	var perms []string = []string{}
	if len(roles) > 0 {
		if slices.Contains(roles, AdminRoleBuilder) {
			perms = append(perms, builderPermissions...)
		}
		if slices.Contains(roles, AdminRoleAdmin) {
			perms = append(perms, "grant_roles", "revoke_roles", "shutdown", "reload")
		}
		if slices.Contains(roles, AdminRoleOwner) {
			perms = append(perms, "owner_only")
		}
	}
	return &AdminContext{
		Roles:       roles,
		Permissions: perms,
		EditMode:    false,
	}
}

// HasRole checks if the admin has a specific role
func (ac *AdminContext) HasRole(role AdminRole) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return slices.Contains(ac.Roles, role)
}

// HasPermission checks if the admin has a specific permission
func (ac *AdminContext) HasPermission(permission string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Owner has all permissions
	if ac.hasRoleUnsafe(AdminRoleOwner) {
		return true
	}

	// Check role-based permissions
	switch permission {
	case "create_room", "edit_room", "create_area", "edit_area", "save_world":
		if ac.hasRoleUnsafe(AdminRoleBuilder) {
			return true
		}
	case "grant_roles", "revoke_roles":
		if ac.hasRoleUnsafe(AdminRoleAdmin) {
			return true
		}
	case "shutdown", "reload":
		if ac.hasRoleUnsafe(AdminRoleAdmin) {
			return true
		}
	case "owner_only":
		return ac.hasRoleUnsafe(AdminRoleOwner)
	}

	// Check explicit permissions
	for _, perm := range ac.Permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// hasRoleUnsafe checks for role without locking (internal use only)
func (ac *AdminContext) hasRoleUnsafe(role AdminRole) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AddPermission adds an explicit permission
func (ac *AdminContext) AddPermission(permission string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check if permission already exists
	for _, perm := range ac.Permissions {
		if perm == permission {
			return
		}
	}

	ac.Permissions = append(ac.Permissions, permission)
}

// RemovePermission removes an explicit permission
func (ac *AdminContext) RemovePermission(permission string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for i, perm := range ac.Permissions {
		if perm == permission {
			ac.Permissions = append(ac.Permissions[:i], ac.Permissions[i+1:]...)
			return
		}
	}
}

// IsAdmin checks if the context has any admin roles
func (ac *AdminContext) IsAdmin() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.Roles) > 0
}

// SetEditMode toggles edit mode for the admin
func (ac *AdminContext) SetEditMode(enabled bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.EditMode = enabled
}

// IsEditMode checks if admin is in edit mode
func (ac *AdminContext) IsEditMode() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.EditMode
}
