package controllers

import (
	"log"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CCaddieWorkingTime struct{}

func (_ *CCaddieWorkingTime) CaddieCheckInWorkingTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CaddieCheckInWorkingTimeBody

	if bindErr := c.BindJSON(&body); bindErr != nil {
		log.Print("BindJSON CaddieNote error")
		response_message.BadRequest(c, "")
		return
	}

	caddieRequest := models.Caddie{}
	// caddieRequest.Uid = body.CaddieId
	errExist := caddieRequest.FindFirst(db)

	if errExist != nil {
		response_message.BadRequest(c, "Caddie IdentityCard did not exist")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}

	checkInTime := time.Now().Unix()

	caddieWorkingTime := models.CaddieWorkingTime{
		ModelId:     base,
		CaddieId:    body.CaddieId,
		CheckInTime: checkInTime,
	}

	err := caddieWorkingTime.Create(db)
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieWorkingTime)
}

func (_ *CCaddieWorkingTime) CaddieCheckOutWorkingTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CaddieCheckOutWorkingTimeBody

	if bindErr := c.BindJSON(&body); bindErr != nil {
		log.Print("BindJSON CaddieNote error")
		response_message.BadRequest(c, "")
		return
	}

	caddieWorkingTimeRequest := models.CaddieWorkingTime{}
	caddieWorkingTimeRequest.Id = body.Id
	response := caddieWorkingTimeRequest.FindCaddieWorkingTimeDetail(db)

	if response == nil {
		response_message.BadRequest(c, "Caddie IdentityCard did not exist")
		return
	}

	checkOutTime := time.Now().Unix()

	caddieWorkingTime := models.CaddieWorkingTime{
		ModelId:      response.ModelId,
		CaddieId:     response.CaddieId,
		CheckInTime:  response.CheckInTime,
		CheckOutTime: checkOutTime,
	}

	err := caddieWorkingTime.Update(db)
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieWorkingTime)
}

func (_ *CCaddieWorkingTime) GetCaddieWorkingTimeDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCaddieWorkingTimeForm{}
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

	caddieWorkingTimeRequest := models.CaddieWorkingTimeRequest{}

	if form.CaddieId != "" {
		caddieWorkingTimeRequest.CaddieId = form.CaddieId
	}

	if form.CaddieName != "" {
		caddieWorkingTimeRequest.CaddieName = form.CaddieName
	}

	list, total, err := caddieWorkingTimeRequest.FindList(db, page, form.From, form.To)

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

func (_ *CCaddieWorkingTime) DeleteCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idRequest := c.Param("id")
	caddieIdIncrement, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieWorkingTimeRequest := models.CaddieWorkingTime{}
	caddieWorkingTimeRequest.Id = caddieIdIncrement
	errF := caddieWorkingTimeRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieWorkingTimeRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingTime) UpdateCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idStr := c.Param("id")
	caddieId, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieWorkingTimeBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieRequest := models.CaddieWorkingTime{}
	caddieRequest.Id = caddieId

	errF := caddieRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.CaddieId != nil {
		caddieRequest.CaddieId = *body.CaddieId
	}
	if body.CheckInTime != nil {
		caddieRequest.CheckOutTime = *body.CheckOutTime
	}
	if body.CheckInTime != nil {
		caddieRequest.CheckInTime = *body.CheckInTime
	}

	err := caddieRequest.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingTime) GetListCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	form := request.GetListCaddieWorkingTimeBody{}
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

	caddie := models.Caddie{}
	caddie.PartnerUid = form.PartnerUid
	caddie.CourseUid = form.CourseUid
	caddie.Name = form.CaddieName
	caddie.Code = form.CaddieId

	list, total, err := caddie.FindList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, data := range list {
		//find detail working of caddie in week
		body := request.GetDetalCaddieWorkingSyncBody{}
		body.PartnerUid = form.PartnerUid
		body.CourseUid = form.CourseUid
		body.EmployeeId = "A1"
		body.Week = form.Week

		_, listWorking := callservices.GetDetailCaddieWorking(body)

		// Add infor to response
		listData[i] = map[string]interface{}{
			"caddie_infor":  data,
			"working_times": listWorking.Data,
		}
	}

	res := response.PageResponse{
		Total: total,
		Data:  listData,
	}

	okResponse(c, res)
}
