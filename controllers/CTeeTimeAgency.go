package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/models"

	"github.com/gin-gonic/gin"
)

type CTeeTimeAgency struct{}

func (_ *CTeeTimeAgency) FindTeeTimeList(c *gin.Context, prof models.CmsUser) {
	body := request.GetTeeTimeAgencyList{}

	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	response := response.GetTeeTimeAgencyResponse{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseUid:    body.CourseUid,
		Date:         body.Date,
	}

	listTeeTime, _, err := searchTeeTimeList(body.CourseUid, body.Date, body.Token,
		body.AgencyId, body.Hole, body.TeeType, body.IsMainCourse)

	if err != nil {
		badRequest(c, err.Error())
		return
	}

	response.Data = listTeeTime

	okResponse(c, response)
}
