package controllers

import (
	"fmt"
	"net/http"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCourseOTA struct{}

// Get list course
func (_ *CCourseOTA) GetListCourseOTA(c *gin.Context) {
	body := request.GetListCourseOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	courseR := models.Course{
		PartnerUid: body.PartnerUid,
	}

	list, total, err := courseR.FindALL()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	dataRes := []response.CourseOTARes{}

	for _, item := range list {
		itemRes := response.CourseOTARes{
			Code:        item.Uid,
			Name:        item.Name,
			Description: "",
			Image:       "",
			Logo:        item.Icon,
			Holes:       item.Hole,
		}

		dataRes = append(dataRes, itemRes)
	}

	resultDetail := map[string]interface{}{
		"status": http.StatusOK,
		"infor":  fmt.Sprintf("%d CHI-LINH Courses", total),
	}

	res := map[string]interface{}{
		"result": resultDetail,
		"data":   dataRes,
	}

	okResponse(c, res)
}
