package controllers

import (
	"start/constants"
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
func GetGolfFeeInfoOfBag(c *gin.Context, mainBooking model_booking.Booking, bagDetail model_booking.BagDetail) model_booking.GolfFeeOfBag {
	db := datasources.GetDatabaseWithPartner(mainBooking.PartnerUid)
	// form := request.GetListBookingForm{}
	// if bindErr := c.ShouldBind(&form); bindErr != nil {
	// 	response_message.BadRequest(c, bindErr.Error())
	// 	return
	// }

	// if form.Bag == "" {
	// 	response_message.BadRequest(c, errors.New("Bag invalid").Error())
	// 	return
	// }

	// mainBooking := model_booking.Booking{}
	// mainBooking.PartnerUid = form.PartnerUid
	// mainBooking.CourseUid = form.CourseUid
	// mainBooking.Bag = form.Bag

	// if form.BookingDate != "" {
	// 	mainBooking.BookingDate = form.BookingDate
	// } else {
	// 	toDayDate, errD := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	// 	if errD != nil {
	// 		response_message.InternalServerError(c, errD.Error())
	// 		return
	// 	}
	// 	mainBooking.BookingDate = toDayDate
	// }

	// errF := mainBooking.FindFirst(db)
	// if errF != nil {
	// 	// response_message.InternalServerError(c, errF.Error())
	// 	response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
	// 	return
	// }

	// bagDetail := getBagDetailFromBooking(db, mainBooking)

	// // bagDetail := model_booking.BagDetail{
	// // 	Booking: mainBooking,
	// // }

	// // Get Rounds
	// if mainBooking.BillCode != "" {
	// 	round := models.Round{BillCode: mainBooking.BillCode}
	// 	listRound, _ := round.FindAll(db)

	// 	if len(listRound) > 0 {
	// 		bagDetail.Rounds = listRound
	// 	}
	// }

	golfFeeOfBag := model_booking.GolfFeeOfBag{
		BagDetail:         bagDetail,
		ListRoundOfSubBag: []model_booking.RoundOfBag{},
	}

	checkIsFirstRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	checkIsNextRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

	for _, item := range bagDetail.SubBags {
		subBooking := model_booking.Booking{}
		subBooking.Uid = item.BookingUid
		errF := subBooking.FindFirst(db)
		if errF != nil {
			// response_message.InternalServerError(c, errF.Error())
			response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
			return model_booking.GolfFeeOfBag{}
		}
		subRound := models.Round{BillCode: subBooking.BillCode}
		listRound, _ := subRound.FindAll(db)

		roundOfBag := model_booking.RoundOfBag{
			Bag:    subBooking.Bag,
			Rounds: []models.Round{},
		}

		if checkIsFirstRound > -1 && len(listRound) > 0 {
			roundOfBag.Rounds = append(roundOfBag.Rounds, listRound[0])
		}
		if checkIsNextRound > -1 && len(listRound) > 1 {
			roundOfBag.Rounds = append(roundOfBag.Rounds, listRound[1])
		}

		if len(listRound) > 0 {
			golfFeeOfBag.ListRoundOfSubBag = append(golfFeeOfBag.ListRoundOfSubBag, roundOfBag)
		}
	}

	return golfFeeOfBag
}
