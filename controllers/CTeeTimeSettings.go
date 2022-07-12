package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CTeeTimeSettings struct{}

func (_ *CTeeTimeSettings) CreateTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTeeTimeSettings{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	teeTimeSetting := models.TeeTimeSettings{
		TeeTime:    body.TeeTime,
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
	}

	errFind := teeTimeSetting.FindFirst()
	teeTimeSetting.TeeTimeStatus = body.TeeTimeStatus
	teeTimeSetting.Note = body.Note

	if errFind == nil {
		errC := teeTimeSetting.Update()
		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	} else {
		errC := teeTimeSetting.Create()
		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	}
	okResponse(c, teeTimeSetting)
}
func (_ *CTeeTimeSettings) GetTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	query := request.GetListTeeTimeSettings{}
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

	teeTimeSetting := models.TeeTimeSettings{}

	if query.TeeTime != "" {
		teeTimeSetting.TeeTime = query.TeeTime
	}

	if query.TeeTimeStatus != "" {
		teeTimeSetting.TeeTimeStatus = query.TeeTimeStatus
	}

	list, total, err := teeTimeSetting.FindList(page)

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
