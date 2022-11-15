package controllers

import (
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"

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
			Bag:    subBooking.GolfBag,
			Rounds: []models.Round{},
		}

		if checkIsFirstRound > -1 && len(listRound) > 0 {
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
