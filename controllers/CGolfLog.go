package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CGolfLog struct{}

func (_ *CGolfLog) GetGolfLogList(c *gin.Context, prof models.CmsUser) {
	form := request.GetOperationLogForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.BookingDate == "" || len(form.BookingDate) < 8 {
		response_message.BadRequest(c, "GetGolfLogList err booking date invalid")
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	if page.Limit <= 0 || page.Limit > 20 {
		page.Limit = 20
	}

	opLog := models.OperationLog{}
	opLog.CourseUid = form.CourseUid
	opLog.PartnerUid = form.PartnerUid
	opLog.BookingDate = form.BookingDate
	opLog.Bag = form.Bag
	opLog.Function = form.Function
	opLog.Module = form.Module

	list, total, err := opLog.FindList(page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if form.Bag != "" {
		// Get thêm lúc tạo booking ở page cuối
		if int64(form.Page*form.Limit) >= total {
			booking := model_booking.Booking{
				Bag:         form.Bag,
				BookingDate: form.BookingDate,
			}
			db := datasources.GetDatabaseWithPartner(form.PartnerUid)
			errFF := booking.FindFirst(db)
			if errFF == nil {
				opLogBag := models.OperationLog{
					BookingUid: booking.Uid,
					Action:     constants.OP_LOG_ACTION_CREATE,
				}
				errFS := opLogBag.FindFirst()
				if errFS == nil {
					listTemp := []models.OperationLog{}
					listTemp = append(listTemp, opLogBag)
					list = append(list, listTemp...)
					total += 1
				}
			}
		}

	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func createOperationLog(opLog models.OperationLog) {
	errC := opLog.Create()
	if errC != nil {
		log.Print("createOperationLog errC", errC.Error())
	}
}
