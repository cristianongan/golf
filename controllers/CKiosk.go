package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CKiosk struct{}
type KioskResponse struct {
	Type string                `json:"kiosk_type"`
	Data []model_service.Kiosk `json:"data"`
}

func (_ *CKiosk) GetListKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListKioskForm{}
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

	kioskR := model_service.Kiosk{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if form.KioskName != "" {
		kioskR.KioskName = form.KioskName
	}

	if form.Status != "" {
		kioskR.Status = form.Status
	}

	if form.KioskType != "" {
		kioskR.KioskType = form.KioskType
	}

	list, _, err := kioskR.FindList(db, page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	typeList := []string{constants.KIOSK_SETTING, constants.MINI_B_SETTING, constants.MINI_R_SETTING, constants.DRIVING_SETTING, constants.RENTAL_SETTING, constants.PROSHOP_SETTING, constants.RESTAURANT_SETTING}
	kioskList := []KioskResponse{}

	for _, typeD := range typeList {
		kioskItem := KioskResponse{
			Type: typeD,
			Data: []model_service.Kiosk{},
		}
		for _, data := range list {
			if data.KioskType == typeD {
				kioskItem.Data = append(kioskItem.Data, data)
			}
		}
		kioskList = append(kioskList, kioskItem)
	}

	okResponse(c, kioskList)
}

func (_ *CKiosk) GetListKioskForApp(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListKioskForm{}
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

	kioskR := model_service.Kiosk{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if form.KioskName != "" {
		kioskR.KioskName = form.KioskName
	}

	if form.Status != "" {
		kioskR.Status = form.Status
	}

	if form.KioskType != "" {
		kioskR.KioskType = form.KioskType
	}

	if form.IsColdBox != nil {
		kioskR.IsColdBox = form.IsColdBox
	}

	list, _, err := kioskR.FindList(db, page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, list)
}

func (_ *CKiosk) CreateKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := model_service.Kiosk{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" {
		response_message.BadRequest(c, "PartnerUid or CourseUid empty!")
		return
	}

	kiosk := model_service.Kiosk{}
	kiosk.PartnerUid = body.PartnerUid
	kiosk.CourseUid = body.CourseUid
	kiosk.KioskName = body.KioskName
	kiosk.ServiceType = body.ServiceType
	kiosk.KioskType = body.KioskType
	kiosk.Status = body.Status
	kiosk.IsColdBox = body.IsColdBox

	errC := kiosk.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SETTING_COURSE,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Function:    constants.OP_LOG_FUNCTION_POS_SYSTEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: kiosk},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okResponse(c, kiosk)
}

func (_ *CKiosk) UpdateKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	kioskIdStr := c.Param("id")
	kioskId, err := strconv.ParseInt(kioskIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && kioskId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	kiosk := model_service.Kiosk{}
	kiosk.Id = kioskId
	kiosk.PartnerUid = prof.PartnerUid
	kiosk.CourseUid = prof.CourseUid
	errF := kiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	valueOld := kiosk

	body := request.CreateKioskForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.KioskName != "" {
		kiosk.KioskName = body.KioskName
	}
	if body.Status != "" {
		kiosk.Status = body.Status
	}
	if body.KioskType != "" {
		kiosk.KioskType = body.KioskType
	}
	if body.ServiceType != "" {
		kiosk.ServiceType = body.ServiceType
	}
	if body.IsColdBox {
		kiosk.IsColdBox = setBoolForCursor(body.IsColdBox)
	}

	errUdp := kiosk.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SETTING_COURSE,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Function:    constants.OP_LOG_FUNCTION_POS_SYSTEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: valueOld},
		ValueNew:    models.JsonDataLog{Data: kiosk},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okResponse(c, kiosk)
}

func (_ *CKiosk) DeleteKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	kioskIdStr := c.Param("id")
	kioskId, err := strconv.ParseInt(kioskIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && kioskId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	kiosk := model_service.Kiosk{}
	kiosk.Id = kioskId
	kiosk.PartnerUid = prof.PartnerUid
	kiosk.CourseUid = prof.CourseUid
	errF := kiosk.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := kiosk.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  kiosk.PartnerUid,
		CourseUid:   kiosk.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SETTING_COURSE,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Function:    constants.OP_LOG_FUNCTION_POS_SYSTEM,
		Body:        models.JsonDataLog{Data: kioskIdStr},
		ValueOld:    models.JsonDataLog{Data: kiosk},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okRes(c)
}
