package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddie struct{}

func (_ *CCaddie) CreateCaddie(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	if body.CourseUid == "" {
		response_message.BadRequest(c, "Course Uid not empty")
		return
	}

	if body.PartnerUid == "" {
		response_message.BadRequest(c, "Partner Uid not empty")
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, partnerErrFind.Error())
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.CourseUid = body.CourseUid
	caddieRequest.PartnerUid = body.PartnerUid
	caddieRequest.Code = body.Code // Id Caddie vận hành
	errExist := caddieRequest.FindFirst()

	if errExist == nil {
		response_message.BadRequest(c, "Caddie Id existed in course")
		return
	}

	Caddie := models.Caddie{
		Code:           body.Code,
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		Name:           body.Name,
		Avatar:         body.Avatar,
		Phone:          body.Phone,
		Address:        body.Address,
		Sex:            body.Sex,
		BirthDay:       body.BirthDay,
		IdentityCard:   body.IdentityCard,
		IssuedBy:       body.IssuedBy,
		ExpiredDate:    body.ExpiredDate,
		Group:          body.Group,
		StartedDate:    body.StartedDate,
		WorkingStatus:  body.WorkingStatus,
		Level:          body.Level,
		Note:           body.Note,
		PlaceOfOrigin:  body.PlaceOfOrigin,
		Email:          body.Email,
		IdHr:           body.IdHr,
		GroupId:        body.GroupId,
		ContractStatus: body.ContractStatus,
	}

	Caddie.GroupId, _ = strconv.ParseInt(body.Group, 10, 8)

	err := Caddie.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, Caddie)
}

func (_ *CCaddie) CreateCaddieBatch(c *gin.Context, prof models.CmsUser) {
	var body []request.CreateCaddieBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	if len(body) < 1 {
		response_message.BadRequest(c, "empty body")
		return
	}

	// base := models.Model{
	// 	Status: constants.STATUS_ENABLE,
	// }

	caddieRequest := models.Caddie{}
	caddieBatchRequest := []models.Caddie{}

	for _, body := range body {
		courseRequest := models.Course{}
		courseRequest.Uid = body.CourseUid
		errFind := courseRequest.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		caddieRequest := models.Caddie{}
		caddieRequest.CourseUid = body.CourseUid
		errExist := caddieRequest.FindFirst()
		if errExist == nil {
			response_message.BadRequest(c, "Caddie number existed in course")
			return
		}

		caddie := models.Caddie{
			PartnerUid:    body.PartnerUid,
			Code:          body.Code,
			CourseUid:     body.CourseUid,
			Name:          body.Name,
			Phone:         body.Phone,
			Address:       body.Address,
			Sex:           body.Sex,
			BirthDay:      body.BirthDay,
			IdentityCard:  body.IdentityCard,
			IssuedBy:      body.IssuedBy,
			ExpiredDate:   body.ExpiredDate,
			Group:         body.Group,
			StartedDate:   body.StartedDate,
			WorkingStatus: body.WorkingStatus,
			Level:         body.Level,
			Note:          body.Note,
			PlaceOfOrigin: body.PlaceOfOrigin,
			Email:         body.Email,
			IdHr:          body.IdHr,
		}

		caddieBatchRequest = append(caddieBatchRequest, caddie)
	}

	err := caddieRequest.CreateBatch(caddieBatchRequest)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddie) GetCaddieList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCaddieForm{}
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

	caddie := models.CaddieList{}

	if form.CourseId != "" {
		caddie.CourseUid = form.CourseId
	}

	if form.Level != "" {
		caddie.Level = form.Level
	}

	if form.Phone != "" {
		caddie.Phone = form.Phone
	}

	if form.Name != "" {
		caddie.CaddieName = form.Name
	}

	if form.Code != "" {
		caddie.CaddieCode = form.Code
	}

	if form.WorkingStatus != "" {
		caddie.WorkingStatus = form.WorkingStatus
	}

	if form.GroupId != "" {
		caddie.GroupId, _ = strconv.ParseInt(form.GroupId, 10, 64)
	}

	if form.IsInGroup != "" {
		caddie.IsInGroup = form.IsInGroup
	}

	if form.IsReadyForBooking != "" {
		caddie.IsReadyForBooking = form.IsReadyForBooking
	}

	if form.ContractStatus != "" {
		caddie.ContractStatus = form.ContractStatus
	}

	list, total, err := caddie.FindList(page)

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

func (_ *CCaddie) GetCaddieDetail(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("id")

	if caddieIdStr == "" {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id, _ = strconv.ParseInt(caddieIdStr, 10, 64)
	caddieDetail, errF := caddieRequest.FindCaddieDetail()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	okResponse(c, caddieDetail)
}

/*
TODO: chuyen ve id
*/
func (_ *CCaddie) DeleteCaddie(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("id")

	if caddieIdStr == "" {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id, _ = strconv.ParseInt(caddieIdStr, 10, 64)
	errF := caddieRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

/*
TODO: chuyen ve id
*/
func (_ *CCaddie) UpdateCaddie(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("id")

	var body request.UpdateCaddieBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if caddieIdStr == "" {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id, _ = strconv.ParseInt(caddieIdStr, 10, 64)

	errF := caddieRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	assignCaddieUpdate(&caddieRequest, body)

	err := caddieRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func assignCaddieUpdate(caddieRequest *models.Caddie, body request.UpdateCaddieBody) {
	if body.CourseId != nil {
		caddieRequest.CourseUid = *body.CourseId
	}
	if body.Name != nil {
		caddieRequest.Name = *body.Name
	}
	if body.Phone != nil {
		caddieRequest.Phone = *body.Phone
	}
	if body.Address != nil {
		caddieRequest.Address = *body.Address
	}
	if body.Sex != nil {
		caddieRequest.Sex = *body.Sex
	}
	if body.BirthDay != nil {
		caddieRequest.BirthDay = *body.BirthDay
	}
	if body.IdentityCard != nil {
		caddieRequest.IdentityCard = *body.IdentityCard
	}
	if body.IssuedBy != nil {
		caddieRequest.IssuedBy = *body.IssuedBy
	}
	if body.ExpiredDate != nil {
		caddieRequest.ExpiredDate = *body.ExpiredDate
	}
	if body.IdHr != nil {
		caddieRequest.IdHr = *body.IdHr
	}
	if body.Group != nil {
		caddieRequest.Group = *body.Group
	}
	if body.StartedDate != nil {
		caddieRequest.StartedDate = *body.StartedDate
	}
	if body.WorkingStatus != nil {
		caddieRequest.WorkingStatus = *body.WorkingStatus
	}
	if body.Level != nil {
		caddieRequest.Level = *body.Level
	}
	if body.Note != nil {
		caddieRequest.Note = *body.Note
	}
	if body.PlaceOfOrigin != nil {
		caddieRequest.Note = *body.PlaceOfOrigin
	}
	if body.Email != nil {
		caddieRequest.Note = *body.Email
	}
	if body.Avatar != nil {
		caddieRequest.Avatar = *body.Avatar
	}
}
