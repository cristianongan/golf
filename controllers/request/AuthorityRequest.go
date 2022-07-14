package request

type AssignRolesBody struct {
	UserUid string   `json:"user_uid" validate:"required"`
	Roles   []string `json:"roles" validate:"required"`
}

type RevokeRolesBody struct {
	UserUid string   `json:"user_uid" validate:"required"`
	Roles   []string `json:"roles" validate:"required"`
}

type AssignGroupRoleBody struct {
	UserUid   string `json:"user_uid" validate:"required"`
	GroupRole string `json:"group_role" validate:"required"`
}

type RevokeAllBody struct {
	UserUid string `json:"user_uid" validate:"required"`
}

type CreateGroupRoleBody struct {
	GroupRoleName string   `json:"group_role_name" validate:"required"`
	GroupRoleCode string   `json:"group_role_code" validate:"required"`
	Roles         []string `json:"roles" validate:"required"`
}

type DeleteGroupRoleBody struct {
	GroupRoleCode string `json:"group_role_code" validate:"required"`
}

type GetRoles struct {
	PageRequest
}

type GetGroupRoles struct {
	PageRequest
}
