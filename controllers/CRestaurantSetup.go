package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service/restaurant_setup"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRestaurantSetup struct{}

// Table Setup
func (_ *CRestaurantSetup) GetRestaurantSetupList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetSetupListForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	tableSetupkiosk := model_service.RestaurantTableSetup{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	tableSetupList, _, _ := tableSetupkiosk.FindList(db)

	timeSetupkiosk := model_service.RestaurantTimeSetup{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	tableTimeList, _, _ := timeSetupkiosk.FindList(db)

	res := map[string]interface{}{
		"table_set_up": tableSetupList,
		"time_set_up":  tableTimeList,
	}
	okResponse(c, res)
}

func (_ *CRestaurantSetup) CreateRestaurantTableSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := model_service.RestaurantTableSetup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" {
		response_message.BadRequest(c, "data not valid")
		return
	}

	tableSetupkiosk := model_service.RestaurantTableSetup{}
	tableSetupkiosk.PartnerUid = body.PartnerUid
	tableSetupkiosk.CourseUid = body.CourseUid
	tableSetupkiosk.NumberOfFloor = body.NumberOfFloor
	tableSetupkiosk.NumberOfTables = body.NumberOfTables
	tableSetupkiosk.MaxPersonInTable = body.MaxPersonInTable
	tableSetupkiosk.Status = body.Status

	errC := tableSetupkiosk.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, tableSetupkiosk)
}

func (_ *CRestaurantSetup) UpdateRestaurantTableSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	kioskId, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && kioskId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	tableSetupkiosk := model_service.RestaurantTableSetup{}
	tableSetupkiosk.Id = kioskId
	tableSetupkiosk.PartnerUid = prof.PartnerUid
	tableSetupkiosk.CourseUid = prof.CourseUid
	errF := tableSetupkiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_service.RestaurantTableSetup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.NumberOfFloor > 0 {
		tableSetupkiosk.NumberOfFloor = body.NumberOfFloor
	}
	if body.Status != "" {
		tableSetupkiosk.Status = body.Status
	}
	if body.NumberOfTables > 0 {
		tableSetupkiosk.NumberOfTables = body.NumberOfTables
	}
	if body.MaxPersonInTable > 0 {
		tableSetupkiosk.MaxPersonInTable = body.MaxPersonInTable
	}

	errUdp := tableSetupkiosk.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, tableSetupkiosk)
}

func (_ *CRestaurantSetup) DeleteRestaurantTableSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	kioskId, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && kioskId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	tableSetupkiosk := model_service.RestaurantTableSetup{}
	tableSetupkiosk.Id = kioskId
	tableSetupkiosk.PartnerUid = prof.PartnerUid
	tableSetupkiosk.CourseUid = prof.CourseUid
	errF := tableSetupkiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := tableSetupkiosk.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// Time Setup
func (_ *CRestaurantSetup) CreateRestaurantTimeSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := model_service.RestaurantTimeSetup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" {
		response_message.BadRequest(c, "data not valid")
		return
	}

	timeSetupkiosk := model_service.RestaurantTimeSetup{}
	timeSetupkiosk.PartnerUid = body.PartnerUid
	timeSetupkiosk.CourseUid = body.CourseUid
	timeSetupkiosk.Minutes = body.Minutes
	timeSetupkiosk.SetupType = body.SetupType
	timeSetupkiosk.Status = body.Status

	errC := timeSetupkiosk.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, timeSetupkiosk)
}

func (_ *CRestaurantSetup) UpdateRestaurantTimeSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	timeSetupkiosk := model_service.RestaurantTimeSetup{}
	timeSetupkiosk.Id = Id
	timeSetupkiosk.PartnerUid = prof.PartnerUid
	timeSetupkiosk.CourseUid = prof.CourseUid
	errF := timeSetupkiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_service.RestaurantTimeSetup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Minutes > 0 {
		timeSetupkiosk.Minutes = body.Minutes
	}
	if body.Status != "" {
		timeSetupkiosk.Status = body.Status
	}
	if body.SetupType != "" {
		timeSetupkiosk.SetupType = body.SetupType
	}

	errUdp := timeSetupkiosk.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, timeSetupkiosk)
}

func (_ *CRestaurantSetup) DeleteRestaurantTimeSetup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	kioskId, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && kioskId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	timeSetupkiosk := model_service.RestaurantTimeSetup{}
	timeSetupkiosk.Id = kioskId
	timeSetupkiosk.PartnerUid = prof.PartnerUid
	timeSetupkiosk.CourseUid = prof.CourseUid
	errF := timeSetupkiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := timeSetupkiosk.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
