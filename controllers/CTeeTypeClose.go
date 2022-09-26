package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CTeeTypeClose struct{}

func (_ *CTeeTypeClose) CreateTeeTypeClose(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateTeeTypeClose{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	bookingSetting := model_booking.BookingSettingGroup{}
	bookingSetting.Id = body.BookingSettingId
	errBookingSettingFind := bookingSetting.FindFirst(db)
	if errBookingSettingFind != nil {
		response_message.InternalServerError(c, "booking setting id incorrect")
		return
	}

	teeTypeClose := models.TeeTypeClose{
		BookingSettingId: body.BookingSettingId,
		DateTime:         body.DateTime,
		CourseUid:        body.CourseUid,
		PartnerUid:       body.PartnerUid,
	}

	errFind := teeTypeClose.FindFirst(db)
	teeTypeClose.Note = body.Note

	if errFind != nil {
		errC := teeTypeClose.Create(db)
		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	} else {
		response_message.BadRequest(c, "Tee Type đã tạo!")
		return
	}
	okResponse(c, teeTypeClose)
}
func (_ *CTeeTypeClose) GetTeeTypeClose(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetListTeeTypeClose{}
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

	teeTypeClose := models.TeeTypeClose{}

	if query.BookingSettingId != nil {
		teeTypeClose.BookingSettingId = *query.BookingSettingId
	}

	if query.DateTime != "" {
		teeTypeClose.DateTime = query.DateTime
	}

	list, total, err := teeTypeClose.FindList(db, page)

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
func (_ *CTeeTypeClose) DeleteTeeTypeClose(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	teeTypeCloseIdStr := c.Param("id")
	teeTypeCloseId, err := strconv.ParseInt(teeTypeCloseIdStr, 10, 64)
	if err != nil || teeTypeCloseId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	teeTypeClose := models.TeeTypeClose{}
	teeTypeClose.Id = teeTypeCloseId
	errF := teeTypeClose.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := teeTypeClose.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
