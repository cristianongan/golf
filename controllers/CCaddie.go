package controllers

import (
	"github.com/gin-gonic/gin"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"
)

type CCaddie struct{}

func (_ *CCaddie) CreateCaddie(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseId
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Num = body.Num
	caddieRequest.CourseId = body.CourseId
	errExist := caddieRequest.FindFirst()

	if errExist == nil && caddieRequest.ModelId.Id > 0 {
		response_message.BadRequest(c, "Caddie number existed in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	Caddie := models.Caddie{
		ModelId:        base,
		CourseId:       body.CourseId,
		Num:            body.Num,
		Name:           body.Name,
		Phone:          body.Phone,
		Address:        body.Address,
		Sex:            body.Sex,
		BirthDay:       body.BirthDay,
		IdentityCard:   body.IdentityCard,
		IssuedBy:       body.IssuedBy,
		IssuedDate:     body.IssuedDate,
		EducationLevel: body.EducationLevel,
		FingerPrint:    body.FingerPrint,
		HrCode:         body.HrCode,
		HrPosition:     body.HrPosition,
		Group:          body.Group,
		Row:            body.Row,
		StartedDate:    body.StartedDate,
		RaisingChild:   body.RaisingChild,
		TempAbsent:     body.TempAbsent,
		FullTime:       body.FullTime,
		WEWork:         body.WEWork,
		Level:          body.Level,
		Note:           body.Note,
	}

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

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}

	caddieRequest := models.Caddie{}
	caddieBatchRequest := []models.Caddie{}

	for _, b := range body {
		courseRequest := models.Course{}
		courseRequest.Uid = b.CourseId
		errFind := courseRequest.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		caddieRequest := models.Caddie{}
		caddieRequest.Num = b.Num
		caddieRequest.CourseId = b.CourseId
		errExist := caddieRequest.FindFirst()
		if errExist == nil && caddieRequest.ModelId.Id > 0 {
			response_message.BadRequest(c, "Caddie number existed in course")
			return
		}

		caddie := models.Caddie{
			ModelId:        base,
			CourseId:       b.CourseId,
			Num:            b.Num,
			Name:           b.Name,
			Phone:          b.Phone,
			Address:        b.Address,
			Sex:            b.Sex,
			BirthDay:       b.BirthDay,
			IdentityCard:   b.IdentityCard,
			IssuedBy:       b.IssuedBy,
			IssuedDate:     b.IssuedDate,
			EducationLevel: b.EducationLevel,
			FingerPrint:    b.FingerPrint,
			HrCode:         b.HrCode,
			HrPosition:     b.HrPosition,
			Group:          b.Group,
			Row:            b.Row,
			StartedDate:    b.StartedDate,
			RaisingChild:   b.RaisingChild,
			TempAbsent:     b.TempAbsent,
			FullTime:       b.FullTime,
			WEWork:         b.WEWork,
			Level:          b.Level,
			Note:           b.Note,
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

	caddieRequest := models.Caddie{}

	if form.CourseId != "" {
		caddieRequest.CourseId = form.CourseId
	}

	list, total, err := caddieRequest.FindList(page)

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

func (_ *CCaddie) DeleteCaddie(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("uid")
	caddieId, errId := strconv.ParseInt(caddieIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id = caddieId
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

func (_ *CCaddie) UpdateCaddie(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("uid")
	caddieId, errId := strconv.ParseInt(caddieIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id = caddieId

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
	if body.Num != nil {
		caddieRequest.Num = *body.Num
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
	if body.BirthPlace != nil {
		caddieRequest.BirthPlace = *body.BirthPlace
	}
	if body.IdentityCard != nil {
		caddieRequest.IdentityCard = *body.IdentityCard
	}
	if body.IssuedBy != nil {
		caddieRequest.IssuedBy = *body.IssuedBy
	}
	if body.IssuedDate != nil {
		caddieRequest.IssuedDate = *body.IssuedDate
	}
	if body.EducationLevel != nil {
		caddieRequest.EducationLevel = *body.EducationLevel
	}
	if body.FingerPrint != nil {
		caddieRequest.FingerPrint = *body.FingerPrint
	}
	if body.HrCode != nil {
		caddieRequest.HrCode = *body.HrCode
	}
	if body.HrPosition != nil {
		caddieRequest.HrPosition = *body.HrPosition
	}
	if body.Group != nil {
		caddieRequest.Group = *body.Group
	}
	if body.Row != nil {
		caddieRequest.Row = *body.Row
	}
	if body.StartedDate != nil {
		caddieRequest.StartedDate = *body.StartedDate
	}
	if body.RaisingChild != nil {
		caddieRequest.RaisingChild = *body.RaisingChild
	}
	if body.TempAbsent != nil {
		caddieRequest.TempAbsent = *body.TempAbsent
	}
	if body.FullTime != nil {
		caddieRequest.FullTime = *body.FullTime
	}
	if body.WEWork != nil {
		caddieRequest.WEWork = *body.WEWork
	}
	if body.Level != nil {
		caddieRequest.Level = *body.Level
	}
	if body.Note != nil {
		caddieRequest.Note = *body.Note
	}
}
