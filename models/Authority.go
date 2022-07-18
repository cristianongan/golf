package models

import (
	"github.com/harranali/authority"
	"start/datasources"
	"strings"
)

type Authority struct {
}

type Request struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type RoleName struct {
	Id           uint        `json:"id"`
	Name         string      `json:"name"`
	Code         string      `json:"code"`
	CategoryName string      `json:"category_name"`
	CategoryCode string      `json:"category_code"`
	Groups       []RoleGroup `json:"groups"`
}

type RoleGroup struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (_ Authority) GetPermissions(roleIds []int64) ([]Request, error) {
	var result []Request

	var permIds []int64

	if err := datasources.GetDatabase().Table("auth_role_permissions").Select("permission_id").Where("role_id IN ?", roleIds).Find(&permIds).Error; err != nil {
		return []Request{}, err
	}

	var perms []authority.Permission

	if err := datasources.GetDatabase().Table("auth_permissions").Where("id IN ?", permIds).Find(&perms).Error; err != nil {
		return []Request{}, err
	}

	for _, perm := range perms {
		nameSplit := strings.Split(perm.Name, "|")
		request := Request{
			Method: nameSplit[0],
			Path:   nameSplit[1],
		}
		result = append(result, request)
	}

	return result, nil
}

func (_ Authority) GetRoles(roleCodes []string) ([]authority.Role, error) {
	var result []authority.Role

	if err := datasources.GetDatabase().Table("auth_roles").Where("name->'$.code' IN ?", roleCodes).Find(&result).Error; err != nil {
		return []authority.Role{}, err
	}

	return result, nil
}

func (_ Authority) GetRolesByGroup(roleGroup string) ([]authority.Role, error) {
	var result []authority.Role

	if err := datasources.GetDatabase().Table("auth_roles").Where("JSON_SEARCH(`name`->'$.groups', 'one', ?, NULL, '$[*].code') IS NOT NULL", roleGroup).Find(&result).Error; err != nil {
		return []authority.Role{}, err
	}

	return result, nil
}

func (_ Authority) GetAllRole() ([]authority.Role, error) {
	var result []authority.Role

	if err := datasources.GetDatabase().Table("auth_roles").Find(&result).Error; err != nil {
		return []authority.Role{}, err
	}

	return result, nil
}

func (_ Authority) UpdateRole(roleId int64, roleName string) error {
	if err := datasources.GetDatabase().Table("auth_roles").Where("id = ?", roleId).Update("name", roleName).Error; err != nil {
		return err
	}
	return nil
}

func (_ Authority) GetAllGroupRole() ([]string, error) {
	var result []string

	if err := datasources.GetDatabase().Table("auth_roles").Distinct("groups").Select("name->'$.groups' as `groups`").Find(&result).Error; err != nil {
		return []string{}, err
	}

	return result, nil
}
