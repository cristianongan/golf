package controllers

import (
	"net/http"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils"

	"github.com/gin-gonic/gin"
)

type Cekyc struct{}

/*
Get List member card for eKyc
*/
func (_ *Cekyc) GetListMemberForEkycList(c *gin.Context) {
	body := request.EkycGetMemberCardList{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	responseBaseModel := response.EkycBaseResponse{
		Code: "00",
		Desc: "Success",
	}

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseUid
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseBaseModel.Code = "01"
		responseBaseModel.Desc = "Course Uid not found"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	checkCheckSum := course.ApiKey + body.PartnerUid + body.CourseUid
	token := utils.GetSHA256Hash(checkCheckSum)

	if token != body.CheckSum {
		responseBaseModel.Code = "02"
		responseBaseModel.Desc = "Checksum incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	memberCardR := models.MemberCard{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
	}

	list, errL := memberCardR.FindListForEkyc(db)
	if errL != nil {
		responseBaseModel.Code = "03"
		responseBaseModel.Desc = "Error"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	responseBaseModel.Data = list

	c.JSON(http.StatusOK, responseBaseModel)
}
