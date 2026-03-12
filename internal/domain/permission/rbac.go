package permission

import "github.com/studio/platform/internal/domain/user"

// RolePermissions maps roles to their permissions
var RolePermissions = map[user.Role][]Permission{
	user.RoleSuperAdmin: {
		// All permissions
		GameReleaseCreate, GameReleaseUpdate, GameReleaseDelete, GameView, GameManage,
		CommentCreate, CommentDeleteOwn, CommentDeleteAny, CommentUpdate,
		OSTView, OSTDownloadHiFi, OSTManage,
		DashboardView,
		UserView, UserUpdate, UserDelete, UserManage,
	},
	user.RoleAdmin: {
		GameReleaseCreate, GameReleaseUpdate, GameReleaseDelete, GameView, GameManage,
		CommentCreate, CommentDeleteOwn, CommentDeleteAny, CommentUpdate,
		OSTView, OSTDownloadHiFi, OSTManage,
		DashboardView,
		UserView, UserUpdate,
	},
	user.RoleModerator: {
		GameView,
		CommentCreate, CommentDeleteOwn, CommentDeleteAny, CommentUpdate,
		OSTView, OSTDownloadHiFi,
		UserView,
	},
	user.RoleCreator: {
		GameView,
		CommentCreate, CommentDeleteOwn, CommentUpdate,
		OSTView, OSTDownloadHiFi,
		UserView,
	},
	user.RoleSupporter: {
		GameView,
		CommentCreate, CommentDeleteOwn,
		OSTView, OSTDownloadHiFi,
		UserView,
	},
	user.RolePremium: {
		GameView,
		CommentCreate, CommentDeleteOwn,
		OSTView, OSTDownloadHiFi,
		UserView,
	},
	user.RoleMember: {
		GameView,
		CommentCreate, CommentDeleteOwn,
		OSTView,
		UserView,
	},
	user.RolePlayer: {
		GameView,
		CommentCreate, CommentDeleteOwn,
		OSTView,
		UserView,
	},
	user.RoleGuest: {
		GameView,
		OSTView,
	},
}

// GetPermissions returns permissions for a given role
func GetPermissions(role user.Role) []Permission {
	perms, ok := RolePermissions[role]
	if !ok {
		return RolePermissions[user.RoleGuest]
	}
	return perms
}

// HasPermission checks if a role has a specific permission
func HasPermission(role user.Role, permission Permission) bool {
	perms := GetPermissions(role)
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissionStrings converts permissions to string slice
func GetPermissionStrings(role user.Role) []string {
	perms := GetPermissions(role)
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = string(p)
	}
	return result
}
