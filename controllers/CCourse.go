package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCourse struct{}

func (_ *CCourse) CreateCourse(c *gin.Context, prof models.CmsUser) {
	body := models.Course{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	partner := models.Partner{}
	partner.Uid = body.PartnerUid

	//Check Partner Exits
	errFind := partner.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	//Check Course Exits
	course := models.Course{}
	course.Uid = udpCourseUid(body.Uid, body.PartnerUid)

	errFind1 := course.FindFirst()
	if errFind1 == nil || course.Name != "" {
		response_message.DuplicateRecord(c, errors.New("Duplicate uid").Error())
		return
	}

	// Create Course
	course.Name = body.Name
	course.Status = body.Status
	course.PartnerUid = body.PartnerUid
	course.Address = body.Address
	course.Lat = body.Lat
	course.Lng = body.Lng
	course.Icon = body.Icon
	course.Hole = body.Hole

	errC := course.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, course)
}

func (_ *CCourse) GetListCourse(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCourseForm{}
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

	courseR := models.Course{
		PartnerUid: form.PartnerUid,
	}
	list, total, err := courseR.FindList(page)
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

func (_ *CCourse) UpdateCourse(c *gin.Context, prof models.CmsUser) {
	courseUidStr := c.Param("uid")
	if courseUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	//Check tồn tại
	course := models.Course{}
	course.Uid = courseUidStr
	errF := course.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.Course{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		course.Name = body.Name
	}
	if body.Status != "" {
		course.Status = body.Status
	}
	if body.Hole > 0 {
		course.Hole = body.Hole
	}
	if body.Address != "" {
		course.Address = body.Address
	}
	if body.Lat > 0 {
		course.Lat = body.Lat
	}
	if body.Lng > 0 {
		course.Lng = body.Lng
	}
	if body.Icon != "" {
		course.Icon = body.Icon
	}

	errUdp := course.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, course)
}

func (_ *CCourse) DeleteCourse(c *gin.Context, prof models.CmsUser) {
	courseUidStr := c.Param("uid")
	if courseUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	course := models.Course{}
	course.Uid = courseUidStr
	errF := course.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := course.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
