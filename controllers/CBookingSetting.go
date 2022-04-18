package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBookingSetting struct{}

/// --------- Booking Setting Group ----------
func (_ *CBookingSetting) CreateBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	body := model_booking.BookingSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Name:       body.Name,
		From:       body.From,
		To:         body.To,
	}

	errC := bookingSettingGroup.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, bookingSettingGroup)
}

func (_ *CBookingSetting) GetListBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingSettingGroupForm{}
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

	bookingSettingGroupR := model_booking.BookingSettingGroup{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := bookingSettingGroupR.FindList(page)
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

func (_ *CBookingSetting) UpdateBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	bookingSettingGroupIdStr := c.Param("id")
	bookingSettingGroupId, err := strconv.ParseInt(bookingSettingGroupIdStr, 10, 64)
	if err != nil || bookingSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{}
	bookingSettingGroup.Id = bookingSettingGroupId
	errF := bookingSettingGroup.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.BookingSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	if body.Name != "" {
		bookingSettingGroup.Name = body.Name
	}
	if body.Status != "" {
		bookingSettingGroup.Status = body.Status
	}
	bookingSettingGroup.From = body.From
	bookingSettingGroup.To = body.To

	errUdp := bookingSettingGroup.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, bookingSettingGroup)
}

func (_ *CBookingSetting) DeleteBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	bookingSettingGroupIdStr := c.Param("id")
	bookingSettingGroupId, err := strconv.ParseInt(bookingSettingGroupIdStr, 10, 64)
	if err != nil || bookingSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{}
	bookingSettingGroup.Id = bookingSettingGroupId
	errF := bookingSettingGroup.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := bookingSettingGroup.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

/// --------- Booking Setting ----------

func (_ *CBookingSetting) CreateBookingSetting(c *gin.Context, prof models.CmsUser) {
	body := model_booking.BookingSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	//Check Group Id avaible
	bSettingGroup := model_booking.BookingSettingGroup{}
	bSettingGroup.Id = body.GroupId
	errFind := bSettingGroup.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	bookingSetting := model_booking.BookingSetting{
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		Dow:            body.Dow,
		GroupId:        body.GroupId,
		TeeMinutes:     body.TeeMinutes,
		TurnLength:     body.TurnLength,
		IsHideTeePart1: body.IsHideTeePart1,
		IsHideTeePart2: body.IsHideTeePart2,
		IsHideTeePart3: body.IsHideTeePart3,
		StartPart1:     body.StartPart1,
		StartPart2:     body.StartPart2,
		StartPart3:     body.StartPart3,
		EndPart1:       body.EndPart1,
		EndPart2:       body.EndPart2,
		EndPart3:       body.EndPart3,
	}

	errC := bookingSetting.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, bookingSetting)
}

func (_ *CBookingSetting) GetListBookingSetting(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingSettingForm{}
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

	bookingSettingR := model_booking.BookingSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GroupId:    form.GroupId,
	}
	list, total, err := bookingSettingR.FindList(page)
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

func (_ *CBookingSetting) UpdateBookingSetting(c *gin.Context, prof models.CmsUser) {
	bookingSettingIdStr := c.Param("id")
	bookingSettingId, err := strconv.ParseInt(bookingSettingIdStr, 10, 64)
	if err != nil || bookingSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSetting := model_booking.BookingSetting{}
	bookingSetting.Id = bookingSettingId
	errF := bookingSetting.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.BookingSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	if body.Status != "" {
		bookingSetting.Status = body.Status
	}
	if body.Dow != "" {
		bookingSetting.Dow = body.Dow
	}

	bookingSetting.TeeMinutes = body.TeeMinutes
	bookingSetting.TurnLength = body.TurnLength

	bookingSetting.IsHideTeePart1 = body.IsHideTeePart1
	bookingSetting.IsHideTeePart2 = body.IsHideTeePart2
	bookingSetting.IsHideTeePart3 = body.IsHideTeePart3

	bookingSetting.StartPart1 = body.StartPart1
	bookingSetting.StartPart2 = body.StartPart2
	bookingSetting.StartPart3 = body.StartPart3

	bookingSetting.EndPart1 = body.EndPart1
	bookingSetting.EndPart2 = body.EndPart2
	bookingSetting.EndPart3 = body.EndPart3

	errUdp := bookingSetting.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, bookingSetting)
}

func (_ *CBookingSetting) DeleteBookingSetting(c *gin.Context, prof models.CmsUser) {
	bookingSettingIdStr := c.Param("id")
	bookingSettingId, err := strconv.ParseInt(bookingSettingIdStr, 10, 64)
	if err != nil || bookingSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSetting := model_booking.BookingSetting{}
	bookingSetting.Id = bookingSettingId
	errF := bookingSetting.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := bookingSetting.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}