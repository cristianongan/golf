package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCustomerUser struct{}

func (_ *CCustomerUser) CreateCustomerUser(c *gin.Context, prof models.CmsUser) {
	body := models.CustomerUser{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.CourseUid == "" {
		response_message.BadRequest(c, errors.New("course uid invalid").Error())
		return
	}

	customerUser := models.CustomerUser{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		Name:        body.Name,
		Dob:         body.Dob,
		Sex:         body.Sex,
		Avatar:      body.Avatar,
		Nationality: body.Nationality,
		Phone:       body.Phone,
		CellPhone:   body.CellPhone,
		Fax:         body.Fax,
		Email:       body.Email,
		Address1:    body.Address1,
		Address2:    body.Address2,
		Job:         body.Job,
		Position:    body.Position,
		CompanyName: body.CompanyName,
		Mst:         body.Mst,
		Note:        body.Note,
		Identify:    body.Identify,
	}

	errC := customerUser.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, customerUser)
}

func (_ *CCustomerUser) GetListCustomerUser(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCustomerUserForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	customerUserGet := models.CustomerUser{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := customerUserGet.FindList(page, form.PartnerUid, form.CourseUid, form.Type, form.CustomerUid, form.Name)
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

func (_ *CCustomerUser) UpdateCustomerUser(c *gin.Context, prof models.CmsUser) {
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.CustomerUser{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Status != "" {
		customerUser.Status = body.Status
	}
	if body.Identify != "" {
		customerUser.Identify = body.Identify
	}
	if body.Name != "" {
		customerUser.Name = body.Name
	}
	if body.Address1 != "" {
		customerUser.Address1 = body.Address1
	}
	if body.Address2 != "" {
		customerUser.Address2 = body.Address2
	}
	if body.Note != "" {
		customerUser.Note = body.Note
	}
	if body.Avatar != "" {
		customerUser.Avatar = body.Avatar
	}
	if body.Nationality != "" {
		customerUser.Nationality = body.Nationality
	}
	if body.Fax != "" {
		customerUser.Fax = body.Fax
	}
	if body.Email != "" {
		customerUser.Email = body.Email
	}
	if body.Job != "" {
		customerUser.Job = body.Job
	}
	if body.Position != "" {
		customerUser.Position = body.Position
	}
	if body.CompanyName != "" {
		customerUser.CompanyName = body.CompanyName
	}
	if body.CompanyId > 0 {
		customerUser.CompanyId = body.CompanyId
	}

	customerUser.Sex = body.Sex

	errUdp := customerUser.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, customerUser)
}

func (_ *CCustomerUser) DeleteCustomerUser(c *gin.Context, prof models.CmsUser) {
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := customerUser.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// Get chi tiết khách hàng
func (_ *CCustomerUser) GetCustomerUserDetail(c *gin.Context, prof models.CmsUser) {
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	c.JSON(200, customerUser)
}
