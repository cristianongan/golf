package controllers

import (
	"log"
	"start/controllers/request"
	"start/models"
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
