package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_role "start/models/role"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CPermission struct{}

/*
Create Permission
*/
func (_ *CPermission) CreatePermission(c *gin.Context, prof models.CmsUser) {
	body := request.CreatePermissionBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	for _, v := range body.PermissionList {
		permission := model_role.Permission{}
		permission.Uid = v.Uid
		errFind := permission.FindFirst()
		if errFind != nil {
			errCreate := v.Create()
			if errCreate != nil {
				response_message.InternalServerError(c, errCreate.Error())
				return
			}
		} else {
			response_message.DuplicateRecord(c, "Duplicated")
			return
		}
	}
	okRes(c)
}

/*
Update Permission
*/
func (_ *CPermission) UpdatePermission(c *gin.Context, prof models.CmsUser) {
	idStr := c.Param("id")
	if len(idStr) == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	permission := model_role.Permission{}
	permission.Uid = idStr

	errF := permission.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_role.Permission{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		permission.Name = body.Name
	}
	if body.Category != "" {
		permission.Category = body.Category
	}
	if body.Description != "" {
		permission.Description = body.Description
	}
	if body.Status != "" {
		permission.Status = body.Status
	}
	if len(body.Resources) > 0 {
		permission.Resources = body.Resources
	}

	errUdp := permission.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, permission)
}

/*
Delete Permission
*/
func (_ *CPermission) DeletePermissions(c *gin.Context, prof models.CmsUser) {
	body := request.DeletePermissionBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	for _, v := range body.PermissionUidList {
		permission := model_role.Permission{}
		permission.Uid = v
		errFind := permission.FindFirst()
		if errFind != nil {
			// response_message.InternalServerError(c, errFind.Error())
			// return
		} else {
			errDelete := permission.Delete()
			if errDelete != nil {
				response_message.InternalServerError(c, errFind.Error())
				return
			}
		}
	}
	okRes(c)
}

func (_ *CPermission) GetPermissionDetail(c *gin.Context, prof models.CmsUser) {
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

	if prof.PartnerUid != constants.ROOT_PARTNER_UID && (role.PartnerUid != prof.PartnerUid || role.CourseUid != prof.CourseUid) {
		response_message.Forbidden(c, "forbidden")
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
