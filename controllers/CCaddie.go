package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type CCaddie struct{}

func (_ *CCaddie) CreateCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	caddieRequest.Phone = body.Phone
	errPhoneExist := caddieRequest.FindFirst(db)

	if errPhoneExist == nil {
		response_message.BadRequest(c, "Số điện thoại đã tồn tại")
		return
	}

	caddieRequest.Phone = ""
	caddieRequest.Code = body.Code // Id Caddie vận hành
	errCodeExist := caddieRequest.FindFirst(db)

	if errCodeExist == nil {
		response_message.BadRequest(c, "Caddie Code đã tồn tại")
		return
	}

	Caddie := models.Caddie{
		Code:         body.Code,
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		Name:         body.Name,
		Avatar:       body.Avatar,
		Phone:        body.Phone,
		Address:      body.Address,
		Sex:          body.Sex,
		BirthDay:     body.BirthDay,
		IdentityCard: body.IdentityCard,
		IssuedBy:     body.IssuedBy,
		ExpiredDate:  body.ExpiredDate,
		Group:        body.Group,
		StartedDate:  body.StartedDate,
		//WorkingStatus: body.WorkingStatus,
		CurrentStatus:  constants.CADDIE_CURRENT_STATUS_READY,
		Level:          body.Level,
		Note:           body.Note,
		PlaceOfOrigin:  body.PlaceOfOrigin,
		Email:          body.Email,
		IdHr:           body.IdHr,
		GroupId:        body.GroupId,
		ContractStatus: body.ContractStatus,
	}

	err := Caddie.Create(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, Caddie)
}

func (_ *CCaddie) CreateCaddieBatch(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		errExist := caddieRequest.FindFirst(db)
		if errExist == nil {
			response_message.BadRequest(c, "Caddie number existed in course")
			return
		}

		caddie := models.Caddie{
			PartnerUid:   body.PartnerUid,
			Code:         body.Code,
			CourseUid:    body.CourseUid,
			Name:         body.Name,
			Phone:        body.Phone,
			Address:      body.Address,
			Sex:          body.Sex,
			BirthDay:     body.BirthDay,
			IdentityCard: body.IdentityCard,
			IssuedBy:     body.IssuedBy,
			ExpiredDate:  body.ExpiredDate,
			Group:        body.Group,
			StartedDate:  body.StartedDate,
			//WorkingStatus: body.WorkingStatus,
			Level:          body.Level,
			Note:           body.Note,
			PlaceOfOrigin:  body.PlaceOfOrigin,
			Email:          body.Email,
			IdHr:           body.IdHr,
			ContractStatus: body.ContractStatus,
		}

		caddieBatchRequest = append(caddieBatchRequest, caddie)
	}

	err := caddieRequest.CreateBatch(db, caddieBatchRequest)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddie) GetCaddieList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	if form.PartnerUid != "" {
		caddie.PartnerUid = form.PartnerUid
	}

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

	if form.CurrentStatus != "" {
		caddie.CurrentStatus = form.CurrentStatus
	}

	if form.IsReadyForJoin != "" {
		caddie.IsReadyForJoin = form.IsReadyForJoin
	}

	if form.IsBooked != "" {
		caddie.IsBooked = form.IsBooked
	}

	list, total, err := caddie.FindList(db, page)

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

func (_ *CCaddie) GetCaddieReadyOnDay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCaddieReady{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddie := models.CaddieList{}

	if form.PartnerUid != "" {
		caddie.PartnerUid = form.PartnerUid
	}

	if form.CourseId != "" {
		caddie.CourseUid = form.CourseId
	}

	caddie.IsReadyForJoin = "1"
	list, _, err := caddie.FindAllCaddieReadyOnDayList(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"data": list,
	}

	c.JSON(200, res)
}

func (_ *CCaddie) GetCaddiGroupDayOffByDate(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetCaddiGroupDayOffByDateForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Get group caddie work today
	dateConvert, _ := time.Parse(constants.DATE_FORMAT_1, form.Date)
	applyDate1 := datatypes.Date(dateConvert)
	idDayOff1 := true

	// get caddie work sechedule
	caddieWCN := models.CaddieWorkingSchedule{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
		ApplyDate:  &(applyDate1),
		IsDayOff:   &idDayOff1,
	}

	listCWS, err := caddieWCN.FindListWithoutPage(db)
	if err != nil {
		log.Println("Find list caddie working schedule today", err.Error())
	}

	//get all group
	caddieGroup := models.CaddieGroup{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
	}

	listCaddieGroup, err := caddieGroup.FindListWithoutPage(db)
	if err != nil {
		log.Println("Find frist caddie working schedule", err.Error())
	}

	var groupDayOff []int64

	//add group caddie
	for _, item := range listCWS {
		id := getIdGroup(listCaddieGroup, item.CaddieGroupCode)

		groupDayOff = append(groupDayOff, id)
	}

	//Get caddie list
	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	caddie := models.CaddieList{}

	if form.PartnerUid != "" {
		caddie.PartnerUid = form.PartnerUid
	}

	if form.CourseId != "" {
		caddie.CourseUid = form.CourseId
	}

	if form.Name != "" {
		caddie.CaddieName = form.Name
	}

	if form.Code != "" {
		caddie.CaddieCode = form.Code
	}

	if form.GroupId != "" {
		caddie.GroupId, _ = strconv.ParseInt(form.GroupId, 10, 64)
	}

	if len(groupDayOff) > 0 {
		caddie.GroupList = groupDayOff
	}

	list, total, err := caddie.FindList(db, page)

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieIdStr := c.Param("id")

	if caddieIdStr == "" {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id, _ = strconv.ParseInt(caddieIdStr, 10, 64)
	caddieDetail, errF := caddieRequest.FindCaddieDetail(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if caddieDetail.PartnerUid != prof.PartnerUid || caddieDetail.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
		return
	}

	okResponse(c, caddieDetail)
}

/*
 */
func (_ *CCaddie) DeleteCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieIdStr := c.Param("id")

	if caddieIdStr == "" {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Id, _ = strconv.ParseInt(caddieIdStr, 10, 64)
	caddieRequest.PartnerUid = prof.PartnerUid
	caddieRequest.CourseUid = prof.CourseUid
	errF := caddieRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieRequest.SolfDelete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

/*
 */
func (_ *CCaddie) UpdateCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	caddieRequest.PartnerUid = prof.PartnerUid
	caddieRequest.CourseUid = prof.CourseUid

	errF := caddieRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	assignCaddieUpdate(&caddieRequest, body)

	err := caddieRequest.Update(db)
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
	if body.GroupId > 0 {
		caddieRequest.GroupId = body.GroupId
	}
	if body.ContractStatus != nil {
		caddieRequest.ContractStatus = *body.ContractStatus
	}
}
