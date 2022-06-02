package controllers

import (
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CBagsNote struct{}

func (_ *CBagsNote) GetListBagsNote(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBagNoteForm{}
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

	bagsNoteR := models.BagsNote{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GolfBag:    form.GolfBag,
	}
	list, total, err := bagsNoteR.FindList(page)
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
