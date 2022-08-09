package controllers

import (
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CLocker struct{}

func (_ *CLocker) GetListLocker(c *gin.Context, prof models.CmsUser) {
	form := request.GetListLockerForm{}
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

	lockerR := models.Locker{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Locker:     form.Locker,
		GolfBag:    form.GolfBag,
	}

	if form.PageRequest.Limit == 0 {
		// Lấy full theo ngày hôm nay
		list, total, err := lockerR.FindList(page, form.From, form.To, true)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)

		return
	}

	list, total, err := lockerR.FindList(page, form.From, form.To, false)
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
