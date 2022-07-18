package seed

import (
	"encoding/json"
	"github.com/harranali/authority"
	"start/models"
)

type AuthoritySeed struct {
	Name string
	Run  func(auth *authority.Authority) error
}

const UPDATE_CADDIE_CODE = "update_caddie"

var roles = map[string]models.RoleName{
	UPDATE_CADDIE_CODE: {
		Name:         "Update Caddie",
		Code:         UPDATE_CADDIE_CODE,
		CategoryName: "Caddie",
		CategoryCode: "caddie",
		Groups: []models.RoleGroup{
			{
				Name: "Admin",
				Code: "admin",
			},
		},
	},
}

var rolePermissionMapping = map[string][]string{
	UPDATE_CADDIE_CODE: {
		"PUT|/golf-cms/api/caddie/:id",
	},
}

func (_ AuthoritySeed) GetCreateRoles() []AuthoritySeed {
	return []AuthoritySeed{
		{
			Name: "Create Role " + UPDATE_CADDIE_CODE,
			Run: func(auth *authority.Authority) error {
				role := roles[UPDATE_CADDIE_CODE]

				roleJson, _ := json.Marshal(role)

				return auth.CreateRole(string(roleJson))
			},
		},
	}
}

func (_ AuthoritySeed) GetAssignPermissions() []AuthoritySeed {
	return []AuthoritySeed{
		{
			Name: "Assign Permissions For Role " + UPDATE_CADDIE_CODE,
			Run: func(auth *authority.Authority) error {
				role := roles[UPDATE_CADDIE_CODE]

				roleJson, _ := json.Marshal(role)

				return auth.AssignPermissions(string(roleJson), rolePermissionMapping[UPDATE_CADDIE_CODE])
			},
		},
	}
}
