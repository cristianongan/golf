package controllers

import (
	"errors"
	"log"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/ez4o/go-try"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/twharmon/slices"
)

type CRound struct{}

func (_ CRound) validateBooking(bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(); err != nil {
		return booking, err
	}

	return booking, nil
}

func (_ CRound) createRound(booking model_booking.Booking, newHole int) (model_booking.BookingRound, error) {
	golfFeeModel := models.GolfFee{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		GuestStyle: booking.GuestStyle,
	}

	golfFee, err := golfFeeModel.GetGuestStyleOnDay()
	if err != nil {
		return model_booking.BookingRound{}, err
	}

	round := model_booking.BookingRound{}
	round.GuestStyle = booking.GuestStyle
	round.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, newHole)
	round.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, newHole)
	round.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, newHole)
	round.Hole = newHole
	round.MemberCardUid = booking.MemberCardUid
	round.TeeOffTime = booking.CheckInTime
	round.Pax = 1
	if booking.Rounds != nil {
		round.Index = len(booking.Rounds) + 1
	}

	return round, nil
}

// func (_ CRound) updateListGolfFee(booking model_booking.Booking, currentGolfFee *model_booking.BookingGolfFee) (model_booking.ListBookingGolfFee, error) {
// 	currentGolfFee.CaddieFee = slices.Reduce(booking.Rounds, func(prev int64, item model_booking.BookingRound) int64 {
// 		return prev + item.CaddieFee
// 	})

// 	currentGolfFee.BuggyFee = slices.Reduce(booking.Rounds, func(prev int64, item model_booking.BookingRound) int64 {
// 		return prev + item.BuggyFee
// 	})

// 	currentGolfFee.GreenFee = slices.Reduce(booking.Rounds, func(prev int64, item model_booking.BookingRound) int64 {
// 		return prev + item.GreenFee
// 	})

// 	return slices.Splice(booking.ListGolfFee, 0, 1, *currentGolfFee), nil
// }

func (_ CRound) updateCurrentBagPrice(booking model_booking.Booking, golfFee int64) (model_booking.BookingCurrentBagPriceDetail, error) {
	currentBagPriceDetail := booking.CurrentBagPrice
	currentBagPriceDetail.GolfFee += golfFee
	currentBagPriceDetail.UpdateAmount()

	return currentBagPriceDetail, nil
}

func (_ CRound) updateMustPayInfo(booking model_booking.Booking) (model_booking.BookingMushPay, error) {
	mustPayInfo := booking.MushPayInfo
	mustPayInfo.TotalGolfFee = booking.GetTotalGolfFee()
	mustPayInfo.MushPay = mustPayInfo.TotalGolfFee + mustPayInfo.TotalServiceItem
	return mustPayInfo, nil
}

func (cRound CRound) AddRound(c *gin.Context, prof models.CmsUser) {
	var body request.AddRoundBody
	var booking model_booking.Booking
	var err error
	var hasError = false

	try.Try(func() {
		if err := c.BindJSON(&body); err != nil {
			log.Print("AddRound BindJSON error")
			try.ThrowOnError(err)
		}

		validate := validator.New()

		if err := validate.Struct(body); err != nil {
			try.ThrowOnError(err)
		}

		// validate booking_uid
		booking, err = cRound.validateBooking(body.BookingUid)
		if err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.BadRequest(c, "")
		hasError = true
	})

	if hasError {
		return
	}

	try.Try(func() {
		// create round and add round
		newRound, err := cRound.createRound(booking, 18)
		if err != nil {
			try.ThrowOnError(err)
		}

		booking.Rounds = append(booking.Rounds, newRound)

		// Golf fee for this Round
		// Thêm golf Fee cho Round mới
		newRoundGolfFee := model_booking.BookingGolfFee{
			BookingUid: booking.Uid,
			PlayerName: booking.CustomerName,
			Bag:        booking.Bag,
			CaddieFee:  newRound.CaddieFee,
			BuggyFee:   newRound.BuggyFee,
			GreenFee:   newRound.GreenFee,
			RoundIndex: newRound.Index,
		}

		// update list_golf_fee
		booking.ListGolfFee = append(booking.ListGolfFee, newRoundGolfFee)

		// update current_bag_price
		booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
		if err != nil {
			try.ThrowOnError(err)
		}

		// update must_pay_info
		booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
		if err != nil {
			try.ThrowOnError(err)
		}

		booking.CmsUser = prof.UserName
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
		if err := booking.Update(); err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.InternalServerError(c, err.Error())
		hasError = true
	})

	if hasError {
		return
	}

	okResponse(c, booking)
}

func (cRound CRound) SplitRound(c *gin.Context, prof models.CmsUser) {
	var body request.SplitRoundBody
	var booking model_booking.Booking
	var err error
	var hasError = false

	try.Try(func() {
		if err := c.BindJSON(&body); err != nil {
			log.Print("SplitRound BindJSON error")
			try.ThrowOnError(err)
		}

		validate := validator.New()

		if err := validate.Struct(body); err != nil {
			try.ThrowOnError(err)
		}

		// validate booking_uid
		booking, err = cRound.validateBooking(body.BookingUid)
		if err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.BadRequest(c, "")
		hasError = true
	})

	if hasError {
		return
	}

	try.Try(func() {
		currentRound := booking.Rounds[body.RoundIndex]
		if currentRound.Hole <= 9 {
			try.ThrowOnError(errors.New("Hole invalid for split"))
			return
		}
		newRound := currentRound
		newRound.Hole = int(body.Hole)
		newRound.Index = len(booking.Rounds) + 1
		currentRound.Hole = currentRound.Hole - newRound.Hole
		booking.Rounds[body.RoundIndex] = currentRound
		booking.Rounds = append(booking.Rounds, newRound)

		// Thêm golf Fee cho Round mới
		newRoundGolfFee := model_booking.BookingGolfFee{
			BookingUid: booking.Uid,
			PlayerName: booking.CustomerName,
			Bag:        booking.Bag,
			CaddieFee:  newRound.CaddieFee,
			BuggyFee:   newRound.BuggyFee,
			GreenFee:   newRound.GreenFee,
			RoundIndex: newRound.Index,
		}

		// update list_golf_fee
		booking.ListGolfFee = append(booking.ListGolfFee, newRoundGolfFee)

		// update current_bag_price
		booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
		if err != nil {
			try.ThrowOnError(err)
		}

		// update must_pay_info
		booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
		if err != nil {
			try.ThrowOnError(err)
		}

		booking.CmsUser = prof.UserName
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
		if err := booking.Update(); err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.InternalServerError(c, err.Error())
		return
	})

	if hasError {
		return
	}

	okResponse(c, booking)
}

func (cRound CRound) MergeRound(c *gin.Context, prof models.CmsUser) {
	var body request.MergeRoundBody
	var booking model_booking.Booking
	var err error
	var hasError = false

	try.Try(func() {
		if err := c.BindJSON(&body); err != nil {
			log.Print("MergeRound BindJSON error")
			try.ThrowOnError(err)
		}

		validate := validator.New()

		if err := validate.Struct(body); err != nil {
			try.ThrowOnError(err)
		}

		// validate booking_uid
		booking, err = cRound.validateBooking(body.BookingUid)
		if err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.BadRequest(c, "")
		hasError = true
	})

	if hasError {
		return
	}

	try.Try(func() {
		// create round
		totalHoles := slices.Reduce(booking.Rounds, func(prev int, item model_booking.BookingRound) int {
			return prev + item.Hole
		})

		newRound, err := cRound.createRound(booking, totalHoles)
		newRound.Index = 0
		if err != nil {
			try.ThrowOnError(err)
		}

		booking.Rounds = append(model_booking.ListBookingRound{}, newRound)

		newRoundGolfFee := model_booking.BookingGolfFee{
			BookingUid: booking.Uid,
			PlayerName: booking.CustomerName,
			Bag:        booking.Bag,
			CaddieFee:  newRound.CaddieFee,
			BuggyFee:   newRound.BuggyFee,
			GreenFee:   newRound.GreenFee,
			RoundIndex: newRound.Index,
		}

		listGolfFeeTemp := model_booking.ListBookingGolfFee{}
		listGolfFeeTemp = append(listGolfFeeTemp, newRoundGolfFee)
		for i, v := range booking.ListGolfFee {
			if i > 0 && v.BookingUid != booking.Uid {
				listGolfFeeTemp = append(listGolfFeeTemp, v)
			}
		}

		// Udp golf fee
		booking.ListGolfFee = listGolfFeeTemp

		// update current_bag_price
		booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
		if err != nil {
			try.ThrowOnError(err)
		}

		// update must_pay_info
		booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
		if err != nil {
			try.ThrowOnError(err)
		}

		booking.CmsUser = prof.UserName
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
		if err := booking.Update(); err != nil {
			try.ThrowOnError(err)
		}
	}).Catch(func(e error, st *try.StackTrace) {
		response_message.InternalServerError(c, err.Error())
		return
	})

	if hasError {
		return
	}

	okResponse(c, booking)
}
