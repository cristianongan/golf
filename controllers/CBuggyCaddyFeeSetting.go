package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"time"

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

	golfFeeTotal := int64(0)
	caddieFee := int64(0)
	buggyFee := int64(0)
	greenFee := int64(0)
	guestStyle := ""

	handleGolfFeeOnDay := func(gs string) {
		bookingDate, _ := time.Parse(constants.DATE_FORMAT_1, form.BookingDate)
		golfFee := models.GolfFee{
			GuestStyle: gs,
			CourseUid:  form.CourseUid,
			PartnerUid: form.PartnerUid,
		}
		fee, _ := golfFee.GetGuestStyleOnTime(db, bookingDate)

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
			AgencyId:   agency.Id,
			CourseUid:  form.CourseUid,
			PartnerUid: form.PartnerUid,
		}
		// Tính lại giá riêng nếu thoả mãn các dk time
		agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
		if errFSP == nil && agencySpecialPrice.Id > 0 {
			// Tính lại giá riêng nếu thoả mãn các dk time,
			// List Booking GolfFee
			caddieFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.CaddieFee, course.RateGolfFee)
			buggyFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.BuggyFee, course.RateGolfFee)
			greenFee = utils.CalculateFeeByHole(form.Hole, agencySpecialPrice.GreenFee, course.RateGolfFee)
			golfFeeTotal = caddieFee + greenFee + buggyFee
		} else {
			handleGolfFeeOnDay(form.GuestStyle)
		}
		guestStyle = agency.GuestStyle
	} else {
		handleGolfFeeOnDay(form.GuestStyle)
	}

	if form.GuestStyle == "" {
		form.GuestStyle = guestStyle
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
		GuestStyle: form.GuestStyle,
		SettingId:  buggyFeeSetting.Id,
	}
	listSetting, _ := buggyFeeItemSettingR.FindBuggyFeeOnDate(db, form.BookingDate)
	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	for _, item := range listSetting {
		if item.GuestStyle != "" {
			buggyFeeItemSetting = item
			break
		} else {
			buggyFeeItemSetting = item
		}
	}

	rentalFee := utils.GetFeeFromListFee(buggyFeeItemSetting.RentalFee, form.Hole)
	privateCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.PrivateCarFee, form.Hole)
	oddCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.OddCarFee, form.Hole)

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
			RentalFee:     rentalFee,
			PrivateCarFee: privateCarFee,
			OddCarFee:     oddCarFee,
		},
		"caddie_fee": models.BookingCaddyFeeSettingRes{
			Fee:  bookingCaddieFeeSetting.Fee,
			Name: bookingCaddieFeeSetting.Name,
		},
	}

	okResponse(c, res)
}
