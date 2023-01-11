package request

import (
	"database/sql/driver"
	"encoding/json"
	model_role "start/models/role"
)

type CreatePermissionBody struct {
	PermissionList ListCreatePermissionBody `json:"permission_list"`
}

type ListCreatePermissionBody []model_role.Permission

func (item *ListCreatePermissionBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListCreatePermissionBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type DeletePermissionBody struct {
	PermissionUidList ListDeletePermissionBody `json:"permission_uids"`
}

type ListDeletePermissionBody []string

func (item *ListDeletePermissionBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListDeletePermissionBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}
