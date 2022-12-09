package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRental struct{}

func (_ *CRental) CreateRental(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateRentalBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
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

	if body.GroupCode == "" {
		response_message.BadRequest(c, "Group Code not empty")
		return
	}

	servicesRequest := model_service.GroupServices{}
	servicesRequest.GroupCode = body.GroupCode
	servicesErrFind := servicesRequest.FindFirst(db)
	if servicesErrFind != nil {
		response_message.BadRequest(c, "GroupCode not existed")
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, "Partner not existed")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, "Course not existed")
		return
	}

	rentalRequest := model_service.Rental{}
	rentalRequest.CourseUid = body.CourseUid
	rentalRequest.PartnerUid = body.PartnerUid
	rentalRequest.RentalId = body.RentalId
	errExist := rentalRequest.FindFirst(db)

	if errExist == nil {
		response_message.BadRequest(c, "Rental Id existed in course")
		return
	}

	name := "" // tên default của proshop

	if body.EnglishName != "" {
		name = body.EnglishName
	} else {
		name = body.VieName
	}

	rental := model_service.Rental{
		RentalId:    body.RentalId,
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		EnglishName: body.EnglishName,
		RenPos:      body.RenPos,
		VieName:     body.VieName,
		GroupCode:   body.GroupCode,
		Unit:        body.Unit,
		Price:       body.Price,
		ByHoles:     body.ByHoles,
		ForPos:      body.ForPos,
		OnlyForRen:  body.OnlyForRen,
		InputUser:   body.InputUser,
		Name:        name,
		IsDriving:   body.IsDriving,
		Rate:        body.Rate,
		Type:        body.Type,
		AccountCode: body.AccountCode,
	}
	rental.Status = body.Status

	err := rental.Create(db)
	if err != nil {
		log.Print("Caddie.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, rental)
}

func (_ *CRental) GetListRental(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListRentalForm{}
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

	rentalR := model_service.RentalRequest{}
	rentalR.PartnerUid = form.PartnerUid
	rentalR.CourseUid = form.CourseUid
	rentalR.EnglishName = form.EnglishName
	rentalR.VieName = form.VieName
	rentalR.GroupCode = form.GroupCode
	rentalR.Type = form.Type
	rentalR.CodeOrName = form.CodeOrName
	rentalR.IsDriving = form.IsDriving

	list, total, err := rentalR.FindList(db, page)
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

func (_ *CRental) UpdateRental(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	rentalIdStr := c.Param("id")
	rentalId, errId := strconv.ParseInt(rentalIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	rental := model_service.Rental{}
	rental.Id = rentalId
	errF := rental.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdateRentalBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.EnglishName != "" {
		rental.EnglishName = body.EnglishName
	}
	if body.VieName != "" {
		rental.VieName = body.VieName
	}
	if body.SystemCode != "" {
		rental.SystemCode = body.SystemCode
	}
	if body.Unit != "" {
		rental.Unit = body.Unit
	}
	if body.RenPos != "" {
		rental.RenPos = body.RenPos
	}
	if body.Price != nil {
		rental.Price = *body.Price
	}
	if body.ByHoles != nil {
		rental.ByHoles = *body.ByHoles
	}
	if body.ForPos != nil {
		rental.ForPos = *body.ForPos
	}
	if body.OnlyForRen != nil {
		rental.OnlyForRen = *body.OnlyForRen
	}
	if body.Status != "" {
		rental.Status = body.Status
	}
	if body.InputUser != "" {
		rental.InputUser = body.InputUser
	}
	if body.Type != "" {
		rental.Type = body.Type
	}
	if body.GroupName != "" {
		rental.GroupName = body.GroupName
	}
	if body.GroupCode != "" {
		rental.GroupCode = body.GroupCode
	}
	if body.IsDriving != nil {
		rental.IsDriving = body.IsDriving
	}
	if body.Rate != "" {
		rental.Rate = body.Rate
	}
	errUdp := rental.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, rental)
}

func (_ *CRental) DeleteRental(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	rentalIdStr := c.Param("id")
	rentalId, errId := strconv.ParseInt(rentalIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	rental := model_service.Rental{}
	rental.Id = rentalId
	errF := rental.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := rental.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CRental) GetGolfClubRental(c *gin.Context, prof models.CmsUser) {
	form := request.GetListRentalForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}
	rentalR := model_service.Rental{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		IsDriving:  form.IsDriving,
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	rentalList, _, err := rentalR.FindALL(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Get Buggy Fee
	buggyFeeSettingR := models.BuggyFeeSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	listBuggySetting, _, _ := buggyFeeSettingR.FindAll(db)
	buggyFeeSetting := models.BuggyFeeSetting{}
	for _, item := range listBuggySetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeSetting = item
			break
		}
	}

	buggyFeeItemSettingR := models.BuggyFeeItemSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		SettingId:  buggyFeeSetting.Id,
		GuestStyle: form.GuestStyle,
		ModelId: models.ModelId{
			Status: constants.STATUS_ENABLE,
		},
	}
	listSetting, _, _ := buggyFeeItemSettingR.FindAllToday(db)
	buggyFeeItemSetting := models.BuggyFeeItemSettingResForRental{}
	for _, v := range listSetting {
		// Ưu tiên All Guest Style (GuestStyle = "")
		if v.GuestStyle == "" {
			buggyFeeItemSetting = v
			break
		} else if v.GuestStyle == form.GuestStyle {
			buggyFeeItemSetting = v
			break
		}
	}

	// Get Buggy Fee
	bookingCaddieFeeSettingR := models.BookingCaddyFeeSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	listBookingBuggyCaddySetting, _, _ := bookingCaddieFeeSettingR.FindList(db, models.Page{}, false)
	bookingCaddieFeeSetting := models.BookingCaddyFeeSettingRes{}
	for _, item := range listBookingBuggyCaddySetting {
		if item.Status == constants.STATUS_ENABLE {
			bookingCaddieFeeSetting = models.BookingCaddyFeeSettingRes{
				Fee:  item.Fee,
				Name: item.Name,
			}
		}
	}

	if form.IsDriving != nil && *form.IsDriving {
		res := map[string]interface{}{
			"rentals": rentalList,
		}
		okResponse(c, res)
		return
	}
	res := map[string]interface{}{
		"rentals":        rentalList,
		"booking_buggy":  buggyFeeItemSetting,
		"booking_caddie": bookingCaddieFeeSetting,
	}
	okResponse(c, res)
}
