package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CBuggyCaddyFeeSetting struct{}

func (_ *CBuggyCaddyFeeSetting) GetBuggyCaddyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetBuggyCaddyFeeSetting{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	golfFee := models.GolfFee{
		GuestStyle: form.GuestStyle,
		CourseUid:  form.CourseUid,
		PartnerUid: form.PartnerUid,
	}
	golfFeeTotal := int64(0)
	caddieFee := int64(0)
	buggyFee := int64(0)
	greenFee := int64(0)

	handleFeeOnDay := func() {
		fee, _ := golfFee.GetGuestStyleOnDay(db)

		caddieFee = utils.GetFeeFromListFee(fee.CaddieFee, form.Hole)
		buggyFee = utils.GetFeeFromListFee(fee.BuggyFee, form.Hole)
		greenFee = utils.GetFeeFromListFee(fee.GreenFee, form.Hole)

		golfFeeTotal = caddieFee + greenFee + buggyFee
	}

	if form.AgencyId > 0 {

		course := models.Course{}
		course.Uid = form.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			response_message.BadRequest(c, "agency"+errCourse.Error())
			return
		}

		agency := models.Agency{}
		agency.Id = form.AgencyId
		errFindAgency := agency.FindFirst(db)
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencySpecialPriceR := models.AgencySpecialPrice{
			AgencyId: agency.Id,
		}
		// Tính lại giá riêng nếu thoả mãn các dk time
		agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
		if errFSP == nil && agencySpecialPrice.Id > 0 {
			// Tính lại giá riêng nếu thoả mãn các dk time,
			// List Booking GolfFee
			caddieFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.CaddieFee, course.RateGolfFee)
			buggyFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.BuggyFee, course.RateGolfFee)
			greenFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.GreenFee, course.RateGolfFee)
		} else {
			handleFeeOnDay()
		}
	} else {
		handleFeeOnDay()
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
		}
	}

	buggyFeeItemSettingR := models.BuggyFeeItemSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GuestStyle: form.GuestStyle,
		SettingId:  buggyFeeSetting.Id,
	}
	listSetting, _, _ := buggyFeeItemSettingR.FindAll(db)
	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	for _, item := range listSetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeItemSetting = item
		}
	}

	// Get Buggy Fee
	bookingCaddieFeeSettingR := models.BookingCaddyFeeSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	listBookingBuggyCaddySetting, _, _ := bookingCaddieFeeSettingR.FindList(db, models.Page{}, false)
	bookingCaddieFeeSetting := models.BookingCaddyFeeSetting{}
	for _, item := range listBookingBuggyCaddySetting {
		if item.Status == constants.STATUS_ENABLE {
			bookingCaddieFeeSetting = item
		}
	}

	res := map[string]interface{}{
		"golf_fee": golfFeeTotal,
		"buggy_fee": models.BuggyFeeItemSettingResponse{
			RentalFee:     buggyFeeItemSetting.RentalFee,
			PrivateCarFee: buggyFeeItemSetting.PrivateCarFee,
			OddCarFee:     buggyFeeItemSetting.OddCarFee,
		},
		"caddie_fee": models.BookingCaddyFeeSettingRes{
			Fee:  bookingCaddieFeeSetting.Fee,
			Name: bookingCaddieFeeSetting.Name,
		},
	}

	okResponse(c, res)
}
