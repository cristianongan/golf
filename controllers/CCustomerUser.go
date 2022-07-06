package controllers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
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

	// Check Customer
	if body.Phone != "" {
		cusTemp := models.CustomerUser{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			Phone:      body.Phone,
		}

		errFind := cusTemp.FindFirst()
		if errFind == nil || cusTemp.Uid != "" {
			// đã tồn tại
			res := map[string]interface{}{
				"message":     "Khách hàng đã tồn tại",
				"status_code": 400,
				"user":        cusTemp,
			}
			c.JSON(400, res)
			return
		}
	}

	customerUser := models.CustomerUser{}
	dataByte, _ := json.Marshal(&body)
	_ = json.Unmarshal(dataByte, &customerUser)

	if body.AgencyId > 0 {
		// Check agency Valid
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFind := agency.FindFirst()
		if errFind != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency "+errFind.Error())
			return
		}

		customerUser.AgencyId = body.AgencyId
		customerUser.Type = constants.CUSTOMER_TYPE_AGENCY
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
		AgencyId:   form.AgencyId,
		Phone:      form.Phone,
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
	if body.Type != "" {
		customerUser.Type = body.Type
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
	if body.Dob > 0 {
		customerUser.Dob = body.Dob
	}

	customerUser.Sex = body.Sex

	if body.AgencyId > 0 {
		// Check agency Valid
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFind := agency.FindFirst()
		if errFind != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency "+errFind.Error())
			return
		}

		customerUser.AgencyId = body.AgencyId
		customerUser.Type = constants.CUSTOMER_TYPE_AGENCY
	}

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

// Delete agency customer
func (_ *CCustomerUser) DeleteAgencyCustomerUser(c *gin.Context, prof models.CmsUser) {
	body := request.DeleteAgencyCustomerUser{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if len(body.CusUserUids) == 0 {
		okRes(c)
		return
	}

	for _, v := range body.CusUserUids {
		cusUser := models.CustomerUser{}
		cusUser.Uid = v
		errF := cusUser.FindFirst()
		if errF == nil {
			cusUser.AgencyId = 0
			cusUser.GolfBag = ""
			errUdp := cusUser.Update()
			if errUdp != nil {
				log.Println("DeleteAgencyCustomerUser errUdp", errUdp.Error())
			}
		} else {
			log.Println("DeleteAgencyCustomerUser errF", errF.Error())
		}
	}

	okRes(c)
}

func (_ CCustomerUser) GetBirthday(c *gin.Context, prof models.CmsUser) {
	query := request.GetBirthdayList{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	customer := models.CustomerUserList{}
	customer.CourseUid = prof.CourseUid
	customer.FromBirthDate = query.FromDate
	customer.ToBirthDate = query.ToDate

	list, total, err := customer.FindCustomerList(page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}
