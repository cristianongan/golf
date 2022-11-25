package controllers

import (
	"errors"
	"log"
	"start/controllers/request"
	"start/models"
	model_role "start/models/role"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRole struct{}

/*
Create Role
*/
func (_ *CRole) CreateRole(c *gin.Context, prof models.CmsUser) {
	body := request.AddRoleBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	role := model_role.Role{}
	role.Name = body.Name
	role.PartnerUid = body.PartnerUid
	role.CourseUid = body.CourseUid

	errC := role.Create()
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//create role hierarchies
	roleHierarchy := model_role.RoleHierarchy{}
	roleHierarchy.ParentRoleUid = prof.RoleId
	roleHierarchy.RoleUid = role.Id
	errCreateHierarchy := roleHierarchy.Create()
	if errCreateHierarchy != nil {
		response_message.InternalServerError(c, errCreateHierarchy.Error())
		return
	}

	//Create role - permission
	if body.Permissions != nil && len(body.Permissions) > 0 {
		listRolePermission := []model_role.RolePermission{}
		for _, v := range body.Permissions {
			roleP := model_role.RolePermission{
				RoleId:        role.Id,
				PermissionUid: v,
			}
			listRolePermission = append(listRolePermission, roleP)
		}

		rolePR := model_role.RolePermission{}
		errC := rolePR.BatchInsert(listRolePermission)
		if errC != nil {
			log.Println("CreateRole errC", errC.Error())
		}
	}

	okResponse(c, role)
}

/*
Get list Role
*/
func (_ *CRole) GetListRole(c *gin.Context, prof models.CmsUser) {
	form := request.GetListRole{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if errPer := checkPermissionPartner(prof, form.PartnerUid, true); errPer != nil {
		response_message.PermissionDeny(c, "")
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	roleR := model_role.Role{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Name:       form.Search,
	}

	subRoles, err := model_role.GetAllSubRoleUids(int(prof.RoleId))
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	list, total, err := roleR.FindList(page, subRoles)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

/*
Update Role
*/
func (_ *CRole) UpdateRole(c *gin.Context, prof models.CmsUser) {
	roleIdStr := c.Param("id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || roleId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	role := model_role.Role{}
	role.Id = roleId
	errF := role.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.AddRoleBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		role.Name = body.Name
	}
	role.Description = body.Description

	// Role Permission Again
	if body.Permissions != nil && len(body.Permissions) > 0 {
		//Get All and del Role Permission
		roleP := model_role.RolePermission{
			RoleId: role.Id,
		}
		listRolePers, err1 := roleP.FindAll()
		if err1 == nil {
			roleDel := model_role.RolePermission{}
			roleDel.DeleteList(listRolePers)
		}

		// Add
		listRolePermission := []model_role.RolePermission{}
		for _, v := range body.Permissions {
			roleP := model_role.RolePermission{
				RoleId:        role.Id,
				PermissionUid: v,
			}
			listRolePermission = append(listRolePermission, roleP)
		}

		rolePR := model_role.RolePermission{}
		errC := rolePR.BatchInsert(listRolePermission)
		if errC != nil {
			log.Println("CreateRole errC", errC.Error())
		}
	}

	//update role hierarchies
	roleHierarchy := model_role.RoleHierarchy{}
	roleHierarchy.ParentRoleUid = prof.RoleId
	roleHierarchy.RoleUid = role.Id
	errDeleteHierarchy := roleHierarchy.FindFirst()
	if errDeleteHierarchy != nil {
		// response_message.InternalServerError(c, errDeleteHierarchy.Error())
		// return
		roleHierarchy.Create()
	}

	errUdp := role.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, role)
}

/*
Delete Role
*/
func (_ *CRole) DeleteRole(c *gin.Context, prof models.CmsUser) {
	roleIdStr := c.Param("id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || roleId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	role := model_role.Role{}
	role.Id = roleId
	errF := role.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Xoa Role Permission
	roleP := model_role.RolePermission{
		RoleId: role.Id,
	}
	listRolePers, err1 := roleP.FindAll()
	if err1 == nil {
		roleDel := model_role.RolePermission{}
		roleDel.DeleteList(listRolePers)
	}

	//delete role hierarchies
	roleHierarchy := model_role.RoleHierarchy{}
	roleHierarchy.ParentRoleUid = prof.RoleId
	roleHierarchy.RoleUid = role.Id
	errDeleteHierarchy := roleHierarchy.FindFirst()
	if errDeleteHierarchy != nil {
		// response_message.InternalServerError(c, errDeleteHierarchy.Error())
		// return
	} else {
		roleHierarchy.Delete()
	}

	errDel := role.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

/*
Update Role
*/
func (_ *CRole) GetRoleDetail(c *gin.Context, prof models.CmsUser) {
	roleIdStr := c.Param("id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || roleId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	role := model_role.Role{}
	role.Id = roleId
	errF := role.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Get list permission
	perR := model_role.RolePermission{
		RoleId: role.Id,
	}
	listPer, errL := perR.FindAll()
	listPerStr := utils.ListString{}
	if errL == nil {
		for _, v := range listPer {
			listPerStr = append(listPerStr, v.PermissionUid)
		}
	}

	roleDetail := model_role.RoleDetail{
		Role:        role,
		Permissions: listPerStr,
	}

	okResponse(c, roleDetail)
}
