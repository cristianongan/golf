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
	"strings"
	"time"

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

	// Check token
	checkToken := "CHILINH_TEST" + body.CourseCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"

		okResponse(c, dataRes)
		return
	}

	// Check course code
	course := models.Course{}

	course.Uid = body.CourseCode

	err := course.FindFirst(db)
	if err != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = err.Error()

		okResponse(c, dataRes)
		return
	}

	// Get list caddie
	caddie := models.CaddieList{
		PartnerUid:     course.PartnerUid,
		CourseUid:      body.CourseCode,
		IsReadyForJoin: "1",
	}

	listCaddie, totalCaddie, errC := caddie.FindAllCaddieReadyOnDayListOTA(db, time.Now().Format("02/01/2006"))

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

	//Update data res
	dataRes.Result.Status = http.StatusOK
	dataRes.Result.Infor = fmt.Sprintf("%d Rentals,%d Caddies", totalRental, totalCaddie)
	dataRes.RentalList = listRental
	dataRes.CaddieList = listCaddie
	dataRes.Token = body.Token

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

	// Check token
	checkToken := "CHILINH_TEST" + body.CourseCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"
		dataRes.CourseCode = body.CourseCode

		okResponse(c, dataRes)
		return
	}

	// Check course code
	course := models.Course{}

	course.Uid = body.CourseCode

	errCourse := course.FindFirst(db)
	if errCourse != nil {
		dataRes.Result.Status = 1000
		dataRes.Result.Infor = errCourse.Error()

		okResponse(c, dataRes)
		return
	}

	// Check Caddie

	if body.CaddieNo > 0 {
		// Get caddie
		caddie := models.Caddie{
			PartnerUid: course.PartnerUid,
			CourseUid:  body.CourseCode,
		}

		caddie.Id = body.CaddieNo
		errCad := course.FindFirst(db)
		if errCad != nil {
			dataRes.Result.Status = 1000
			dataRes.Result.Infor = errCad.Error()

			okResponse(c, dataRes)
			return
		}

		if caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_READY {
			dataRes.Result.Status = http.StatusInternalServerError
			dataRes.Result.Infor = fmt.Sprintf("Caddie %d status invalid", body.CaddieNo)

			okResponse(c, dataRes)
			return
		}

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
			dataRes.Result.Infor = errI.Error()

			okResponse(c, dataRes)
			return
		}

		// Kiểm tra số lượng hàng tồn trong kho
		if body.Qty > inventory.Quantity {
			dataRes.Result.Status = http.StatusInternalServerError
			dataRes.Result.Infor = fmt.Sprintf("Item %s is not enough", body.RenTalCode)

			okResponse(c, dataRes)
			return
		}
	}

	//Update data res
	dataRes.Result.Status = http.StatusOK

	okResponse(c, dataRes)
}
