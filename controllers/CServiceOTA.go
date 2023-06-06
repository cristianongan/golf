package controllers

import (
	"fmt"
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CCServiceOTA struct{}

// Get list course
func (_ *CCServiceOTA) GetServiceOTA(c *gin.Context) {
	db := datasources.GetDatabase()
	body := request.ServiceGolfDataBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	// Data res
	dataRes := response.GetServiceRes{
		CourseCode: body.CourseCode,
	}

	// Check course code
	course := models.Course{}

	course.Uid = body.CourseCode
	errCourse := course.FindFirstHaveKey()
	if errCourse != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = fmt.Sprintf("Course %s %s", body.CourseCode, errCourse.Error())

		okResponse(c, dataRes)
		return
	}

	// Check token
	checkToken := course.ApiKey + body.CourseCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"

		okResponse(c, dataRes)
		return
	}

	// Get list caddie
	caddie := models.CaddieList{
		PartnerUid:     course.PartnerUid,
		CourseUid:      body.CourseCode,
		IsReadyForJoin: "1",
	}

	listCaddie, totalCaddie, errC := caddie.FindAllCaddieReadyOnDayListOTA(db, "")

	if errC != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errC.Error()

		okResponse(c, dataRes)
		return
	}

	rental := model_service.Rental{}
	rental.PartnerUid = course.PartnerUid
	rental.CourseUid = body.CourseCode

	listRental, totalRental, errR := rental.FindALL(db)
	if errR != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errC.Error()

		okResponse(c, dataRes)
		return
	}

	// craete caddie Res
	dataRentalRes := []response.RentalRes{}

	for _, item := range listRental {
		itemRes := response.RentalRes{
			Code:  item.RentalId,
			Name:  item.VieName,
			Unit:  item.Unit,
			Price: item.Price,
		}

		dataRentalRes = append(dataRentalRes, itemRes)
	}

	//Update data res
	dataRes.Result.Status = http.StatusOK
	dataRes.Result.Infor = fmt.Sprintf("%d Rentals,%d Caddies", totalRental, totalCaddie)
	dataRes.RentalList = dataRentalRes
	dataRes.CaddieList = listCaddie
	dataRes.Token = "???"

	okResponse(c, dataRes)
}

// Get list course
func (_ *CCServiceOTA) CheckServiceOTA(c *gin.Context) {
	db := datasources.GetDatabase()
	body := request.CheckServiceGolfBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	// Data res
	dataRes := response.CheckServiceRes{
		RenTalCode: body.RenTalCode,
		CaddieNo:   body.CaddieNo,
		DateStr:    body.DateStr,
		TeeOffStr:  body.TeeOffStr,
		CourseCode: body.CourseCode,
		Qty:        body.Qty,
	}

	// Check course code
	course := models.Course{}
	course.Uid = body.CourseCode
	errCourse := course.FindFirstHaveKey()
	if errCourse != nil {
		dataRes.Result.Status = 1000
		dataRes.Result.Infor = fmt.Sprintf("Course %s %s", body.CourseCode, errCourse.Error())

		okResponse(c, dataRes)
		return
	}

	checkToken := course.ApiKey + body.CourseCode + body.CaddieNo + body.RenTalCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"
		dataRes.CourseCode = body.CourseCode

		okResponse(c, dataRes)
		return
	}

	// Check Caddie

	if body.CaddieNo != "" {
		// Get caddie
		caddie := models.Caddie{
			PartnerUid: course.PartnerUid,
			CourseUid:  body.CourseCode,
		}
		//Parse
		id, _ := strconv.ParseInt(body.CaddieNo, 10, 64)
		caddie.Id = id
		errCad := caddie.FindFirst(db)
		if errCad != nil {
			dataRes.Result.Status = 1000
			dataRes.Result.Infor = fmt.Sprintf("Caddie %s %s", body.CaddieNo, errCad.Error())

			okResponse(c, dataRes)
			return
		}

		if caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_READY &&
			caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_FINISH &&
			caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_FINISH_R2 &&
			caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_FINISH_R3 {
			dataRes.Result.Status = http.StatusInternalServerError
			dataRes.Result.Infor = fmt.Sprintf("Caddie %s status invalid", body.CaddieNo)

			okResponse(c, dataRes)
			return
		}

		//Update data res
		dataRes.Result.Status = http.StatusOK
		dataRes.Result.Infor = "Caddie validate"
		dataRes.Token = "???"

		okResponse(c, dataRes)
		return
	}

	if body.RenTalCode != "" {
		// Get id rental
		kioskR := model_service.Kiosk{
			KioskType: constants.RENTAL_SETTING,
		}

		errK := kioskR.FindFirst(db)
		if errK != nil {
			dataRes.Result.Status = http.StatusInternalServerError
			dataRes.Result.Infor = errK.Error()

			okResponse(c, dataRes)
			return
		}

		// validate quantity
		inventory := kiosk_inventory.InventoryItem{}
		inventory.PartnerUid = course.PartnerUid
		inventory.CourseUid = body.CourseCode
		inventory.ServiceId = kioskR.Id
		inventory.Code = body.RenTalCode

		if errI := inventory.FindFirst(db); errI != nil {
			dataRes.Result.Status = 1000
			dataRes.Result.Infor = fmt.Sprintf("Item %s %s", body.RenTalCode, errI.Error())

			okResponse(c, dataRes)
			return
		}

		// Kiểm tra số lượng hàng tồn trong kho
		if body.Qty > inventory.Quantity {
			dataRes.Result.Status = http.StatusInternalServerError
			dataRes.Result.Infor = fmt.Sprintf("Quantity %s is not enough", body.RenTalCode)

			okResponse(c, dataRes)
			return
		}

		//Update data res
		dataRes.Result.Status = http.StatusOK
		dataRes.Result.Infor = "Rental validate"
		dataRes.Token = "???"

		okResponse(c, dataRes)
		return
	}
}

// Get list fee service booking
func (_ *CCServiceOTA) GetListFeeServiceOTA(c *gin.Context) {
	body := request.GetListFeeServiceOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	// Data res
	dataRes := response.ServiceFeeRes{
		DateStr:    body.DateStr,
		CourseCode: body.CourseCode,
	}

	// Check course code
	course := models.Course{}
	course.Uid = body.CourseCode
	errCourse := course.FindFirstHaveKey()
	if errCourse != nil {
		dataRes.Result.Status = 1000
		dataRes.Result.Infor = fmt.Sprintf("Course %s %s", body.CourseCode, errCourse.Error())

		okResponse(c, dataRes)
		return
	}

	checkToken := course.ApiKey + body.CourseCode + body.OTACode + body.DateStr
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"
		dataRes.CourseCode = body.CourseCode

		okResponse(c, dataRes)
		return
	}

	db := datasources.GetDatabaseWithPartner(course.PartnerUid)

	// Validate agency code
	agencyR := models.Agency{
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
		AgencyId:   body.OTACode,
	}

	errAF := agencyR.FindFirst(db)
	if errAF != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errAF.Error()

		okResponse(c, dataRes)
		return
	}

	// Get Buggy Fee
	buggyFeeSettingR := models.BuggyFeeSetting{
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
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
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
		SettingId:  buggyFeeSetting.Id,
	}
	listSetting, _ := buggyFeeItemSettingR.FindBuggyFeeOnDate(db, body.DateStr)
	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	for _, item := range listSetting {
		if item.GuestStyle != "" {
			buggyFeeItemSetting = item
			break
		} else {
			buggyFeeItemSetting = item
		}
	}

	rentalFee := utils.GetFeeFromListFee(buggyFeeItemSetting.RentalFee, body.Hole)
	privateCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.PrivateCarFee, body.Hole)
	oddCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.OddCarFee, body.Hole)

	// Get Caddie Fee
	bookingCaddieFeeSettingR := models.BookingCaddyFeeSetting{
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
	}

	listBookingBuggyCaddySetting, _, _ := bookingCaddieFeeSettingR.FindList(db, models.Page{}, false)
	bookingCaddieFeeSetting := models.BookingCaddyFeeSetting{}
	for _, item := range listBookingBuggyCaddySetting {
		if item.Status == constants.STATUS_ENABLE {
			bookingCaddieFeeSetting = item
		}
	}

	//Update data res
	dataRes.Result.Status = http.StatusOK

	dataRes.RentalFee = response.ServiceInfor{
		Fee:  rentalFee,
		Name: "Thuê xe",
	}
	dataRes.PrivateCarFee = response.ServiceInfor{
		Fee:  privateCarFee,
		Name: "Thuê xe riêng",
	}
	dataRes.OddCarFee = response.ServiceInfor{
		Fee:  oddCarFee,
		Name: "Thuê lẻ xe",
	}
	dataRes.CaddieFee = response.ServiceInfor{
		Fee:  bookingCaddieFeeSetting.Fee,
		Name: bookingCaddieFeeSetting.Name,
	}

	okResponse(c, dataRes)

}
