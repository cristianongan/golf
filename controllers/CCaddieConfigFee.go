package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type CCaddieConfigFee struct{}

func (_ *CCaddieConfigFee) CreateCaddieConfigFee(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieConfigFeeBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, partnerErrFind.Error())
		return
	}

	caddieConfigFeeRequest := models.CaddieConfigFee{}
	caddieConfigFeeRequest.CourseUid = body.CourseUid
	caddieConfigFeeRequest.PartnerUid = body.PartnerUid
	errExist := caddieConfigFeeRequest.FindFirst()

	if errExist == nil && caddieConfigFeeRequest.ModelId.Id > 0 {
		response_message.BadRequest(c, "Code existed in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}

	validDate, _ := time.Parse("2006-01-02", body.ValidDate)
	expDate, _ := time.Parse("2006-01-02", body.ExpDate)

	caddieConfigFee := models.CaddieConfigFee{
		ModelId:    base,
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
		Type:       body.Type,
		FeeDetail:  body.FeeDetail,
		ValidDate:  datatypes.Date(validDate),
		ExpDate:    datatypes.Date(expDate),
	}

	err := caddieConfigFee.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieConfigFee)
}

func (_ *CCaddieConfigFee) GetCaddieConfigFee(c *gin.Context, prof models.CmsUser) {
	form := request.GetCaddieConfigFeeList{}
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

	caddieConfigFeeRequest := models.CaddieConfigFee{}

	list, total, err := caddieConfigFeeRequest.FindList(page)

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

func (_ *CCaddieConfigFee) DeleteCaddieConfigFee(c *gin.Context, prof models.CmsUser) {
	caddieConfigFeeIdStr := c.Param("id")
	caddieConfigFeeId, errId := strconv.ParseInt(caddieConfigFeeIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieConfigFeeRequest := models.CaddieConfigFee{}
	caddieConfigFeeRequest.Id = caddieConfigFeeId

	errF := caddieConfigFeeRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieConfigFeeRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieConfigFee) UpdateCaddieConfigFee(c *gin.Context, prof models.CmsUser) {
	caddieConfigFeeIdStr := c.Param("id")
	caddieConfigFeeId, errId := strconv.ParseInt(caddieConfigFeeIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieConfigFeeBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieConfigFeeRequest := models.CaddieConfigFee{}
	caddieConfigFeeRequest.Id = caddieConfigFeeId

	errF := caddieConfigFeeRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.CourseUid != nil {
		courseRequest := models.Course{}
		courseRequest.Uid = *body.CourseUid
		errFind := courseRequest.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, "course_uid not found")
			return
		}
	}

	if body.PartnerUid != nil {
		partnerRequest := models.Partner{}
		partnerRequest.Uid = *body.PartnerUid
		partnerErrFind := partnerRequest.FindFirst()
		if partnerErrFind != nil {
			response_message.BadRequest(c, "partner_uid not found")
			return
		}
	}

	if body.CourseUid != nil {
		caddieConfigFeeRequest.CourseUid = *body.CourseUid
	}
	if body.PartnerUid != nil {
		caddieConfigFeeRequest.PartnerUid = *body.PartnerUid
	}

	err := caddieConfigFeeRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
