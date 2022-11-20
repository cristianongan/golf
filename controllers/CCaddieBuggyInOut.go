package controllers

import (
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_gostarter "start/models/go-starter"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCaddieBuggyInOut struct{}

func (_ *CCaddieBuggyInOut) GetCaddieBuggyInOut(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCaddieBuggyInOut{}
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

	caddieBuggyInOut := model_gostarter.CaddieBuggyInOut{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		CaddieType: form.CaddieType,
		BuggyType:  form.BuggyType,
		BuggyCode:  form.BuggCode,
		CaddieCode: form.CaddieCode,
	}

	param := model_gostarter.CaddieBuggyInOutRequest{
		CaddieBuggyInOut: caddieBuggyInOut,
		Bag:              form.Bag,
		Date:             form.BookingDate,
		ShareBuggy:       form.ShareBuggy,
		BagOrBuggyCode:   form.BagOrBuggyCode,
	}
	list, total, err := caddieBuggyInOut.FindCaddieBuggyInOutWithBooking(db, page, param)
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
