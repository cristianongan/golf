package controllers

import (
	"fmt"
	"log"
	"net/http"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCourseOTA struct{}

// Get list course
func (cCourseOTA *CCourseOTA) GetListTeeTypeInfo(c *gin.Context) {
	body := request.GetListTeeTypeInfoOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// Data res
	dataRes := response.OtaGeneralRes{
		CourseCode: body.CourseCode,
	}

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseCode
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Loi he thong"
		okResponse(c, dataRes)
		return
	}

	// Get list Tee Type info
	teeTypeR := models.TeeTypeInfo{
		PartnerUid: course.PartnerUid,
		CourseUid:  course.Uid,
	}
	listTeeType, errLTP := teeTypeR.FindALL()
	if errLTP != nil || len(listTeeType) == 0 {
		log.Println("GetListTeeTypeInfo errLTP error or empty")
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Loi he thong"

		okResponse(c, dataRes)
		return
	}

	// agency paid
	dataRes.Result.Status = 200
	dataRes.Result.Infor = "ok"
	dataRes.Data = listTeeType
	okResponse(c, dataRes)
}

// Get list course
func (cCourseOTA *CCourseOTA) GetListCourseOTA(c *gin.Context) {
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
