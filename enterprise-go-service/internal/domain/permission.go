package domain

type Permission string

const (
	PermissionReadProfile   Permission = "read:profile"
	PermissionUpdateProfile Permission = "update:profile"
	PermissionDeleteUser    Permission = "delete:user"
	PermissionManageSystem  Permission = "manage:system"
)

var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionReadProfile,
		PermissionUpdateProfile,
		PermissionDeleteUser,
		PermissionManageSystem,
	},
	RoleUser: {
		PermissionReadProfile,
		PermissionUpdateProfile,
	},
}

func HasPermission(role Role, required Permission) bool {
	perms, exists := RolePermissions[role]
	if !exists {
		return false
	}
	for _, p := range perms {
		if p == required {
			return true
		}
	}
	return false
}