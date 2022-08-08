package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/harranali/authority"
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strings"
)

type CAuthority struct{}

func (cAuthority CAuthority) validateUser(userUid string) (models.CmsUser, error) {
	cmsUser := models.CmsUser{}
	cmsUser.Uid = userUid
	if err := cmsUser.FindFirst(); err != nil {
		return cmsUser, err
	}
	return cmsUser, nil
}

func (cAuthority CAuthority) AssignRoles(c *gin.Context, prof models.CmsUser) {
	var body request.AssignRolesBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AssignRoles BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate user_uid
	cmsUser, err := cAuthority.validateUser(body.UserUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	roles, err := auth1.GetRoles(body.Roles)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth2 := authority.New(authority.Options{
		TablesPrefix: "auth_",
		DB:           datasources.GetDatabase(),
	})

	for _, role := range roles {
		roleJson, _ := json.Marshal(role.Name)

		if err := auth2.AssignRole(cmsUser.AuthorityId, strings.Replace(strings.Trim(string(roleJson), "\""), "\\", "", -1)); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (cAuthority CAuthority) RevokeRoles(c *gin.Context, prof models.CmsUser) {
	var body request.RevokeRolesBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("RevokeRoles BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate user_uid
	cmsUser, err := cAuthority.validateUser(body.UserUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	roles, err := auth1.GetRoles(body.Roles)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth2 := authority.New(authority.Options{
		TablesPrefix: "auth_",
		DB:           datasources.GetDatabase(),
	})

	for _, role := range roles {
		roleJson, _ := json.Marshal(role.Name)
		if err := auth2.RevokeRole(cmsUser.AuthorityId, strings.Replace(strings.Trim(string(roleJson), "\""), "\\", "", -1)); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (_ CAuthority) CreateGroupRole(c *gin.Context, prof models.CmsUser) {
	var body request.CreateGroupRoleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateGroupRole BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	roles, err := auth1.GetRoles(body.Roles)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, role := range roles {
		var roleTemp models.RoleName
		if err := json.Unmarshal([]byte(role.Name), &roleTemp); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		roleTemp.Groups = append(roleTemp.Groups, models.RoleGroup{
			Code: body.GroupRoleCode,
			Name: body.GroupRoleName,
		})

		roleJson, _ := json.Marshal(roleTemp)
		if err := auth1.UpdateRole(int64(role.ID), string(roleJson)); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (_ CAuthority) DeleteGroupRole(c *gin.Context, prof models.CmsUser) {
	var body request.DeleteGroupRoleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("DeleteGroupRole BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	roles, err := auth1.GetRolesByGroup(body.GroupRoleCode)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, role := range roles {
		var roleTemp models.RoleName
		if err := json.Unmarshal([]byte(role.Name), &roleTemp); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// remove group role
		removeIndex := utils.SliceIndex(len(roleTemp.Groups), func(i int) bool {
			return roleTemp.Groups[i].Code == body.GroupRoleCode
		})

		roleTemp.Groups = append(roleTemp.Groups[:removeIndex], roleTemp.Groups[removeIndex+1:]...)

		roleJson, _ := json.Marshal(roleTemp)
		if err := auth1.UpdateRole(int64(role.ID), string(roleJson)); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (cAuthority CAuthority) AssignGroupRole(c *gin.Context, prof models.CmsUser) {
	var body request.AssignGroupRoleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AssignGroupRole BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate user_uid
	cmsUser, err := cAuthority.validateUser(body.UserUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	roles, err := auth1.GetRolesByGroup(body.GroupRole)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth2 := authority.New(authority.Options{
		TablesPrefix: "auth_",
		DB:           datasources.GetDatabase(),
	})

	// revoke all role
	allRole, err := auth1.GetAllRole()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, role := range allRole {
		roleJson, _ := json.Marshal(role.Name)
		if err := auth2.RevokeRole(cmsUser.AuthorityId, strings.Replace(strings.Trim(string(roleJson), "\""), "\\", "", -1)); err != nil {
			continue
		}
	}

	// assign roles
	for _, role := range roles {
		roleJson, _ := json.Marshal(role.Name)
		if err := auth2.AssignRole(cmsUser.AuthorityId, strings.Replace(strings.Trim(string(roleJson), "\""), "\\", "", -1)); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (_ CAuthority) GetRoles(c *gin.Context, prof models.CmsUser) {
	total := int64(-1)
	query := request.GetRoles{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//page := models.Page{
	//	Limit:   query.PageRequest.Limit,
	//	Page:    query.PageRequest.Page,
	//	SortBy:  query.PageRequest.SortBy,
	//	SortDir: query.PageRequest.SortDir,
	//}

	auth1 := models.Authority{}
	list, err := auth1.GetAllRole()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	var roles []models.RoleName
	for _, item := range list {
		var role models.RoleName
		if err := json.Unmarshal([]byte(item.Name), &role); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
		role.Id = item.ID
		roles = append(roles, role)
	}

	res := response.PageResponse{
		Total: total,
		Data:  roles,
	}

	c.JSON(200, res)
}

func (_ CAuthority) GetGroupRoles(c *gin.Context, prof models.CmsUser) {
	total := int64(-1)
	query := request.GetGroupRoles{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	auth1 := models.Authority{}
	list, err := auth1.GetAllGroupRole()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	var groupRoles []models.RoleGroup

	for _, item := range list {
		var groupRolesTemp []models.RoleGroup
		if err := json.Unmarshal([]byte(item), &groupRolesTemp); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
		groupRoles = append(groupRoles, groupRolesTemp...)
	}

	// remove duplicate
	var uniqueGroupRoles []models.RoleGroup
	unique := map[string]bool{}

	for _, groupRole := range groupRoles {
		if _, ok := unique[groupRole.Name]; !ok {
			unique[groupRole.Name] = true
			uniqueGroupRoles = append(uniqueGroupRoles, groupRole)
		}

	}

	res := response.PageResponse{
		Total: total,
		Data:  uniqueGroupRoles,
	}

	c.JSON(200, res)
}
