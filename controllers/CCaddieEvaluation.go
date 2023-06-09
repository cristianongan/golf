package controllers

import (
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CCaddieEvaluation struct{}

func (_ *CCaddieEvaluation) validateBooking(c *gin.Context, db *gorm.DB, bookingUid string, caddieUid string, caddieCode string) (model_booking.Booking, error) {
	bookingList := model_booking.BookingList{}
	bookingList.BookingUid = bookingUid
	bookingList.CaddieUid = caddieUid
	bookingList.CaddieCode = caddieCode

	return bookingList.FindFirst(db)
}

func (cCaddieEvaluation *CCaddieEvaluation) CreateCaddieEvaluation(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieEvaluationBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieEvaluation BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate booking_uid, caddie_uid, caddie_code
	booking, err := cCaddieEvaluation.validateBooking(c, db, body.BookingUid, body.CaddieUid, body.CaddieCode)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate duplicate
	caddieEvaluationList := models.CaddieEvaluationList{}
	caddieEvaluationList.BookingUid = body.BookingUid
	if _, err = caddieEvaluationList.FindFirst(db); err == nil {
		response_message.BadRequest(c, "record duplicate")
		return
	}

	caddieEvaluation := models.CaddieEvaluation{
		BookingUid:  body.BookingUid,
		BookingCode: "",
		BookingDate: datatypes.Date(utils.GetTimeNow()),
		CaddieUid:   body.CaddieUid,
		CaddieCode:  body.CaddieCode,
		CaddieName:  booking.CaddieInfo.Name,
		CourseUid:   prof.CourseUid,
		PartnerUid:  prof.PartnerUid,
		GolfBag:     booking.Bag,
		Hole:        booking.Hole,
		RankType:    body.RankType,
	}

	if err := caddieEvaluation.Create(db); err != nil {
		log.Print("CaddieEvaluation.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieEvaluation)
}

func (_ *CCaddieEvaluation) GetCaddieEvaluationList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetCaddieEvaluationList{}
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

	caddieEvaluation := models.CaddieEvaluationList{}

	caddieEvaluation.CourseUid = prof.CourseUid
	caddieEvaluation.CaddieName = query.CaddieName
	caddieEvaluation.CaddieCode = query.CaddieCode
	caddieEvaluation.Month = query.Month

	list, total, err := caddieEvaluation.FindList(db, page)

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

func (_ *CCaddieEvaluation) UpdateCaddieEvaluation(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.UpdateCaddieEvaluationBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieEvaluation BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	caddieEvaluation := models.CaddieEvaluation{}
	caddieEvaluation.BookingUid = body.BookingUid
	caddieEvaluation.CaddieUid = body.CaddieUid
	caddieEvaluation.CaddieCode = body.CaddieCode
	caddieEvaluation.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	caddieEvaluation.PartnerUid = prof.PartnerUid
	caddieEvaluation.CourseUid = prof.CourseUid

	if err := caddieEvaluation.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieEvaluation.RankType = body.RankType

	if err := caddieEvaluation.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
