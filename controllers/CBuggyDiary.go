package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBuggyDiary struct{}

func (_ *CBuggyDiary) CreateBuggyDiary(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateBuggyDiaryBody
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

	buggyRequest := models.Buggy{}
	buggyRequest.CourseUid = body.CourseId
	errExist := buggyRequest.FindFirst(db)

	if errExist != nil || buggyRequest.ModelId.Id < 1 {
		response_message.BadRequest(c, "Buggy number did not exist in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	diary := models.BuggyDiary{
		ModelId:       base,
		CourseId:      body.CourseId,
		BuggyNumber:   body.BuggyNumber,
		AccessoriesId: body.AccessoriesId,
		Amount:        body.Amount,
		Note:          body.Note,
	}

	if body.InputUser != "" {
		diary.InputUser = body.InputUser
	} else {
		diary.InputUser = prof.UserName
	}

	err := diary.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, diary)
}

func (_ *CBuggyDiary) GetBuggyDiaryList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBuggyDiaryForm{}
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

	diary := models.BuggyDiary{}

	if form.CourseId != "" {
		diary.CourseId = form.CourseId
	}
	if form.BuggyNumber != nil {
		diary.BuggyNumber = *form.BuggyNumber
	}

	list, total, err := diary.FindList(page, form.From, form.To)

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

func (_ *CBuggyDiary) DeleteBuggyDiary(c *gin.Context, prof models.CmsUser) {
	diaryIdStr := c.Param("uid")
	diaryId, errId := strconv.ParseInt(diaryIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	diary := models.BuggyDiary{}
	diary.Id = diaryId
	errF := diary.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := diary.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBuggyDiary) UpdateBuggyDiary(c *gin.Context, prof models.CmsUser) {
	diaryIdStr := c.Param("uid")
	diaryId, errId := strconv.ParseInt(diaryIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateBuggyDiaryBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	diaryRequest := models.BuggyDiary{}
	diaryRequest.Id = diaryId

	errF := diaryRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.AccessoriesId != nil {
		diaryRequest.AccessoriesId = *body.AccessoriesId
	}
	if body.Amount != nil {
		diaryRequest.Amount = *body.Amount
	}
	if body.Note != nil {
		diaryRequest.Note = *body.Note
	}

	err := diaryRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
