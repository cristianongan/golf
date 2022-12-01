package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

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
	}

	checkIsFirstRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	checkIsNextRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

	for _, subBooking := range mainBooking.SubBags {
		subRound := models.Round{BillCode: subBooking.BillCode}
		listRound, _ := subRound.FindAll(db)

		roundOfBag := model_booking.RoundOfBag{
			Bag:         subBooking.GolfBag,
			BookingCode: subBooking.BookingCode,
			PlayerName:  subBooking.PlayerName,
			Rounds:      []models.Round{},
		}

		if checkIsFirstRound > -1 && len(listRound) > 0 {
			round1 := models.Round{}
			for _, item := range listRound {
				if item.Index == 1 {
					round1 = item
				}
			}

			isAgencyPaid := false
			for _, v := range subBooking.AgencyPaid {
				if v.Type == constants.BOOKING_AGENCY_GOLF_FEE && v.Fee > 0 {
					isAgencyPaid = true
				}
			}

			if !isAgencyPaid {
				roundOfBag.Rounds = append(roundOfBag.Rounds, round1)
			}
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
		toDayDate, errD := utils.GetBookingDateFromTimestamp(time.Now().Unix())
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
