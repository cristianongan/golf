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
	model_report "start/models/report"
	model_service "start/models/service"
	"start/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CCServiceOTA struct{}

/*
Check member card avaible
*/
func (cServiceOTA *CCServiceOTA) CheckMemberCard(c *gin.Context) {
	db := datasources.GetDatabase()
	body := request.CheckMemberCardBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	// Data res
	dataRes := response.OtaGeneralRes{
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
	checkToken := course.ApiKey + body.CourseCode + body.OtaCode + body.CardId
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"

		okResponse(c, dataRes)
		return
	}

	// Find memberCard
	memberCard := models.MemberCard{
		PartnerUid: course.PartnerUid,
		CourseUid:  course.Uid,
		CardId:     body.CardId,
	}

	errF := memberCard.FindFirst(db)
	if errF != nil {
		dataRes.Result.Status = 100 // not found member
		dataRes.Result.Infor = "Khong tim thay member"

		okResponse(c, dataRes)
		return
	}

	// Get Owner
	owner, errOwner := memberCard.GetOwner(db)
	if errOwner != nil {
		dataRes.Result.Status = 101 // Not found owner
		dataRes.Result.Infor = "Khong tim thay thong tin chu the"

		okResponse(c, dataRes)
		return
	}

	if memberCard.AnnualType == constants.ANNUAL_TYPE_LIMITED {
		// Validate số lượt chơi còn lại của memeber
		reportCustomer := model_report.ReportCustomerPlay{
			CustomerUid: owner.Uid,
		}

		if errF := reportCustomer.FindFirst(); errF == nil {
			playCountRemain := memberCard.AdjustPlayCount - reportCustomer.TotalPlayCount
			if playCountRemain <= 0 {
				dataRes.Result.Status = 102 // Thẻ hết lượt chơi
				dataRes.Result.Infor = "the het lout choi"

				okResponse(c, dataRes)
				return
			}
		}
	}

	dataRes.Result.Status = 200
	dataRes.Result.Infor = fmt.Sprintf("card id %s ok", body.CardId)
	dataRes.Data = memberCard
	okResponse(c, dataRes)
}

/*
Check Caddie avaible
*/
func (cServiceOTA *CCServiceOTA) CheckCaddie(c *gin.Context) {
	db := datasources.GetDatabase()
	body := request.CheckCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	// Data res
	dataRes := response.OtaGeneralRes{
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
	checkToken := course.ApiKey + body.CourseCode + body.OtaCode + body.CaddieCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"

		okResponse(c, dataRes)
		return
	}

	// Find memberCard
	caddie := models.Caddie{
		PartnerUid: course.PartnerUid,
		CourseUid:  course.Uid,
		Code:       body.CaddieCode,
	}
	caddie.Status = constants.STATUS_ENABLE

	errF := caddie.FindFirst(db)
	if errF != nil {
		dataRes.Result.Status = 100 // not found caddie
		dataRes.Result.Infor = "Khong tim thay caddie"

		okResponse(c, dataRes)
		return
	}

	//TODO: check caddie avaible ngay book

	dataRes.Result.Status = 200
	dataRes.Result.Infor = fmt.Sprintf("caddie code %s ok", body.CaddieCode)
	dataRes.Data = caddie
	okResponse(c, dataRes)
}

// Get list course
func (cServiceOTA *CCServiceOTA) GetServiceOTA(c *gin.Context) {
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
func (cServiceOTA *CCServiceOTA) CheckServiceOTA(c *gin.Context) {
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
