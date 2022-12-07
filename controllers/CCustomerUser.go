package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"

	model_report "start/models/report"

	"github.com/gin-gonic/gin"
)

type CCustomerUser struct{}

func (_ *CCustomerUser) CreateCustomerUser(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

		errFind := cusTemp.FindFirst(db)
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

	// Check Identify
	if body.Identify != "" {
		cusTemp := models.CustomerUser{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			Identify:   body.Identify,
		}

		errFind := cusTemp.FindFirst(db)
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
		errFind := agency.FindFirst(db)
		if errFind != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency "+errFind.Error())
			return
		}

		customerUser.AgencyId = body.AgencyId
		customerUser.Type = constants.CUSTOMER_TYPE_AGENCY
	}

	errC := customerUser.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	go func() {
		body := request.CustomerBody{
			MaKh:      customerUser.Uid,
			TenKh:     customerUser.Name,
			MaSoThue:  customerUser.Mst,
			DiaChi:    customerUser.Address1,
			Tk:        "",
			DienThoai: customerUser.Phone,
			Fax:       customerUser.Fax,
			EMail:     customerUser.Email,
			DoiTac:    "",
			NganHang:  "",
			TkNh:      "",
		}
		callservices.CreateCustomer(body)
	}()

	okResponse(c, customerUser)
}

func (_ *CCustomerUser) GetListCustomerUser(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		Identify:   form.Identify,
	}
	list, total, err := customerUserGet.FindList(db, page, form.PartnerUid, form.CourseUid, form.Type, form.CustomerUid, form.Name)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.CustomerUser{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	body.PartnerUid = customerUser.PartnerUid
	body.CourseUid = customerUser.CourseUid
	if body.Phone != customerUser.Phone && body.IsDuplicated(db) {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	if body.Phone != "" {
		if customerUser.Phone != body.Phone {
			cusTemp := models.CustomerUser{
				PartnerUid: body.PartnerUid,
				CourseUid:  body.CourseUid,
				Phone:      body.Phone,
			}

			errFind := cusTemp.FindFirst(db)
			if errFind == nil || cusTemp.Uid != "" {
				// đã tồn tại
				res := map[string]interface{}{
					"message":     "Số điện thoại đã tồn tại",
					"status_code": 400,
					"user":        cusTemp,
				}
				c.JSON(400, res)
				return
			} else {
				customerUser.Phone = body.Phone
			}
		}
	}

	customerUser.CellPhone = body.CellPhone

	if body.Status != "" {
		customerUser.Status = body.Status
	}
	if body.Identify != "" {
		if customerUser.Identify != body.Identify {
			cusTemp := models.CustomerUser{
				PartnerUid: body.PartnerUid,
				CourseUid:  body.CourseUid,
				Identify:   body.Identify,
			}

			errFind := cusTemp.FindFirst(db)
			if errFind == nil || cusTemp.Uid != "" {
				// đã tồn tại
				res := map[string]interface{}{
					"message":     "Số chứng minh thư đã tồn tại",
					"status_code": 400,
					"user":        cusTemp,
				}
				c.JSON(400, res)
				return
			} else {
				customerUser.Identify = body.Identify
			}
		}
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
		errFind := agency.FindFirst(db)
		if errFind != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency "+errFind.Error())
			return
		}

		customerUser.AgencyId = body.AgencyId
		customerUser.Type = constants.CUSTOMER_TYPE_AGENCY
	}

	errUdp := customerUser.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, customerUser)
}

func (_ *CCustomerUser) DeleteCustomerUser(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := customerUser.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// Get chi tiết khách hàng
func (_ *CCustomerUser) GetCustomerUserDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	customerUserUidStr := c.Param("uid")
	if customerUserUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	customerUser := models.CustomerUser{}
	customerUser.Uid = customerUserUidStr
	errF := customerUser.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Get report play count
	reportCus := model_report.ReportCustomerPlay{
		CustomerUid: customerUserUidStr,
	}

	errFR := reportCus.FindFirst()
	if errFR != nil || reportCus.Id <= 0 {
		reportCus.CourseUid = customerUser.CourseUid
		reportCus.PartnerUid = customerUser.PartnerUid
		errRC := reportCus.Create()
		if errRC != nil {
			log.Println("GetCustomerUserDetail errRC", errRC.Error())
		}
	}

	reportData := map[string]interface{}{
		"total_paid":            reportCus.TotalPaid,
		"total_play_count":      reportCus.TotalPlayCount,
		"total_hour_play_count": reportCus.TotalHourPlayCount,
	}

	res := map[string]interface{}{
		"data":   customerUser,
		"report": reportData,
	}

	c.JSON(200, res)
}

// Delete agency customer
func (_ *CCustomerUser) DeleteAgencyCustomerUser(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		errF := cusUser.FindFirst(db)
		if errF == nil {
			cusUser.AgencyId = 0
			cusUser.GolfBag = ""
			errUdp := cusUser.Update(db)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	list, total, err := customer.FindCustomerList(db, page)

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
