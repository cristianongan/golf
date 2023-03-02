package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

/*
Get chi tiết Golf Fee của bag: Round, Sub bag
*/
func GetGolfFeeInfoOfBag(c *gin.Context, mainBooking model_booking.Booking) model_booking.GolfFeeOfBag {
	db := datasources.GetDatabaseWithPartner(mainBooking.PartnerUid)
	golfFeeOfBag := model_booking.GolfFeeOfBag{
		Booking:           mainBooking,
		ListRoundOfSubBag: []model_booking.RoundOfBag{},
		AgencyPaidAll:     0,
	}

	checkIsFirstRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	checkIsNextRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

	golfFeeOfBag.AgencyPaidAll = mainBooking.GetAgencyPaid()

	for _, subBooking := range mainBooking.SubBags {

		bookingR := model_booking.Booking{
			Model: models.Model{Uid: subBooking.BookingUid},
		}

		if eBookingR := bookingR.FindFirst(db); eBookingR != nil {
			log.Println(eBookingR.Error())
		}

		golfFeeOfBag.AgencyPaidAll += bookingR.GetAgencyPaid()

		if bookingR.CheckAgencyPaidAll() {
			break
		}

		subRound := models.Round{BillCode: subBooking.BillCode}
		listRound, _ := subRound.FindAll(db)

		roundOfBag := model_booking.RoundOfBag{
			Bag:         subBooking.GolfBag,
			BookingCode: subBooking.BookingCode,
			PlayerName:  subBooking.PlayerName,
			Rounds:      []models.Round{},
		}

		if checkIsFirstRound > -1 && len(listRound) > 0 && !bookingR.CheckAgencyPaidRound1() {
			round1 := models.Round{}
			for _, item := range listRound {
				if item.Index == 1 {
					round1 = item
				}
			}

			roundOfBag.Rounds = append(roundOfBag.Rounds, round1)
		}

		if checkIsNextRound > -1 && len(listRound) > 1 {
			round2 := models.Round{}
			for _, item := range listRound {
				if item.Index == 2 {
					round2 = item
				}
			}
			roundOfBag.Rounds = append(roundOfBag.Rounds, round2)
		}

		if len(listRound) > 0 {
			golfFeeOfBag.ListRoundOfSubBag = append(golfFeeOfBag.ListRoundOfSubBag, roundOfBag)
		}
	}

	return golfFeeOfBag
}

func GetGolfFeeInfoOfBagForBill(c *gin.Context, mainBooking model_booking.Booking) model_booking.GolfFeeOfBag {
	db := datasources.GetDatabaseWithPartner(mainBooking.PartnerUid)
	golfFeeOfBag := model_booking.GolfFeeOfBag{
		Booking:           mainBooking,
		ListRoundOfSubBag: []model_booking.RoundOfBag{},
		AgencyPaidAll:     0,
	}

	// checkIsFirstRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	// checkIsNextRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

	golfFeeOfBag.AgencyPaidAll = mainBooking.GetAgencyPaid()

	for _, subBooking := range mainBooking.SubBags {

		bookingR := model_booking.Booking{
			Model: models.Model{Uid: subBooking.BookingUid},
		}

		booking, eBookingR := bookingR.FindFirstByUId(db)
		if eBookingR != nil {
			log.Println(eBookingR.Error())
		}

		golfFeeOfBag.AgencyPaidAll += booking.GetAgencyPaid()

		// if bookingR.CheckAgencyPaidAll() {
		// 	break
		// }

		subRound := models.Round{BillCode: subBooking.BillCode}
		listRound, _ := subRound.FindAll(db)

		roundOfBag := model_booking.RoundOfBag{
			Bag:         subBooking.GolfBag,
			BookingCode: subBooking.BookingCode,
			PlayerName:  subBooking.PlayerName,
			Rounds:      []models.Round{},
			AgencyPaid:  booking.AgencyPaid,
		}

		for _, item := range listRound {
			roundOfBag.Rounds = append(roundOfBag.Rounds, item)
		}

		if len(listRound) > 0 {
			golfFeeOfBag.ListRoundOfSubBag = append(golfFeeOfBag.ListRoundOfSubBag, roundOfBag)
		}
	}

	return golfFeeOfBag
}

/*
Get Round Bag trong ngày
*/
func (_ *CBooking) GetRoundOfBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.GolfBag == "" {
		response_message.BadRequest(c, errors.New("Bag invalid").Error())
		return
	}

	booking := model_booking.BookingList{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.GolfBag = form.GolfBag
	booking.BookingDate = form.BookingDate

	if form.BookingDate != "" {
		booking.BookingDate = form.BookingDate
	} else {
		toDayDate, errD := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		if errD != nil {
			response_message.InternalServerError(c, errD.Error())
			return
		}
		booking.BookingDate = toDayDate
	}

	db, total, err := booking.FindAllBookingList(db)

	db = db.Order("created_at asc")
	db = db.Preload("CaddieBuggyInOut")

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}
	okResponse(c, res)
}
