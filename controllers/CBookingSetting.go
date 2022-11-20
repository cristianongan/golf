package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CBookingSetting struct{}

// / --------- Booking Setting Group ----------
func (_ *CBookingSetting) CreateBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := model_booking.BookingSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated(db) {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Name:       body.Name,
		FromDate:   body.FromDate,
		ToDate:     body.ToDate,
	}

	errC := bookingSettingGroup.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, bookingSettingGroup)
}

func (_ *CBookingSetting) GetListBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	list, total, err := bookingSettingGroupR.FindList(db, page, 0, 0)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingSettingGroupIdStr := c.Param("id")
	bookingSettingGroupId, err := strconv.ParseInt(bookingSettingGroupIdStr, 10, 64)
	if err != nil || bookingSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{}
	bookingSettingGroup.Id = bookingSettingGroupId
	errF := bookingSettingGroup.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.BookingSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if bookingSettingGroup.Name != body.Name || bookingSettingGroup.FromDate != body.FromDate || bookingSettingGroup.ToDate != body.ToDate {
		if body.IsDuplicated(db) {
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
	}

	if body.Name != "" {
		bookingSettingGroup.Name = body.Name
	}
	if body.Status != "" {
		bookingSettingGroup.Status = body.Status
	}
	bookingSettingGroup.FromDate = body.FromDate
	bookingSettingGroup.ToDate = body.ToDate

	errUdp := bookingSettingGroup.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, bookingSettingGroup)
}

func (_ *CBookingSetting) DeleteBookingSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingSettingGroupIdStr := c.Param("id")
	bookingSettingGroupId, err := strconv.ParseInt(bookingSettingGroupIdStr, 10, 64)
	if err != nil || bookingSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSettingGroup := model_booking.BookingSettingGroup{}
	bookingSettingGroup.Id = bookingSettingGroupId
	errF := bookingSettingGroup.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := bookingSettingGroup.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

/// --------- Booking Setting ----------

func (_ *CBookingSetting) CreateBookingSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := model_booking.BookingSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated(db) {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	//Check Group Id avaible
	bSettingGroup := model_booking.BookingSettingGroup{}
	bSettingGroup.Id = body.GroupId
	errFind := bSettingGroup.FindFirst(db)
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
		Part1TeeType:   body.Part1TeeType,
		Part2TeeType:   body.Part2TeeType,
		Part3TeeType:   body.Part3TeeType,
		IncludeDays:    body.IncludeDays,
	}

	bookingSetting.Status = body.Status

	errC := bookingSetting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, bookingSetting)
}

func (_ *CBookingSetting) GetListBookingSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	list, total, err := bookingSettingR.FindList(db, page)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingSettingIdStr := c.Param("id")
	bookingSettingId, err := strconv.ParseInt(bookingSettingIdStr, 10, 64)
	if err != nil || bookingSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSetting := model_booking.BookingSetting{}
	bookingSetting.Id = bookingSettingId
	errF := bookingSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.BookingSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Dow != bookingSetting.Dow {
		if body.IsDuplicated(db) {
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
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

	bookingSetting.Part1TeeType = body.Part1TeeType
	bookingSetting.Part2TeeType = body.Part2TeeType
	bookingSetting.Part3TeeType = body.Part3TeeType

	errUdp := bookingSetting.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, bookingSetting)
}

func (_ *CBookingSetting) DeleteBookingSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingSettingIdStr := c.Param("id")
	bookingSettingId, err := strconv.ParseInt(bookingSettingIdStr, 10, 64)
	if err != nil || bookingSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	bookingSetting := model_booking.BookingSetting{}
	bookingSetting.Id = bookingSettingId
	errF := bookingSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := bookingSetting.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// Get booking Config của ngày
func (item *CBookingSetting) GetListBookingSettingOnDate(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingSettingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	list, bookingSettingGroup, errDel := item.GetSettingOnDate(db, form)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	res := map[string]interface{}{
		"booking-setting-group": bookingSettingGroup,
		"data":                  list,
	}

	okResponse(c, res)
}

func (_ *CBookingSetting) GetSettingOnDate(db *gorm.DB, form request.GetListBookingSettingForm) ([]model_booking.BookingSetting, *model_booking.BookingSettingGroup, error) {
	if form.OnDate == "" {
		log.Println("on-date empty", form)
		form.OnDate = utils.GetCurrentDay1()
	}

	from := int64(0)
	to := int64(0)

	if form.OnDate != "" {
		// Lấy ngày hiện tại
		fromInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.OnDate)
		from = fromInt
		to = from + 24*60*60
	}

	if from == 0 {
		return nil, nil, errors.New("Tee 1 is closed")
	}

	// Get booking Group
	page := models.Page{
		Limit:   20,
		Page:    1,
		SortBy:  "created_at",
		SortDir: "desc",
	}

	bookingSettingGroupR := model_booking.BookingSettingGroup{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	listBSG, _, errLBSG := bookingSettingGroupR.FindList(db, page, from, to)
	if errLBSG != nil || len(listBSG) == 0 {
		return nil, nil, errors.New("Not found booking setting group")
	}
	bookingSettingGroup := listBSG[0]
	bookingSettingGroupId := bookingSettingGroup.Id

	page1 := models.Page{
		Limit:   200,
		Page:    1,
		SortBy:  "created_at",
		SortDir: "desc",
	}

	bookingSettingR := model_booking.BookingSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GroupId:    bookingSettingGroupId,
	}
	list, _, err := bookingSettingR.FindList(db, page1)
	if err != nil {
		return nil, nil, err
	}
	return list, &bookingSettingGroup, nil
}

func (_ *CBookingSetting) ValidateClose1ST(db *gorm.DB, BookingDate string, PartnerUid string, CourseUid string) error {
	bookingSetting := model_booking.BookingSettingGroup{
		PartnerUid: PartnerUid,
		CourseUid:  CourseUid,
	}
	from := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, BookingDate)
	to := from + 24*60*60
	page := models.Page{
		Limit:   20,
		Page:    1,
		SortBy:  "created_at",
		SortDir: "desc",
	}
	listBSG, _, errLBSG := bookingSetting.FindList(db, page, from, to)
	if errLBSG != nil || len(listBSG) == 0 {
		return nil
	}
	bookingSettingGroup := listBSG[0]
	if bookingSettingGroup.Status == constants.STATUS_ENABLE {
		teeTypeClose := models.TeeTypeClose{
			PartnerUid:       bookingSettingGroup.PartnerUid,
			CourseUid:        bookingSettingGroup.CourseUid,
			BookingSettingId: bookingSettingGroup.Id,
			DateTime:         BookingDate,
		}
		err := teeTypeClose.FindFirst(db)
		if err == nil {
			return errors.New("Tee 1 is closed")
		}
	}
	return nil
}
