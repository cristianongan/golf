package controllers

import (
	"errors"

	// "log"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	// "time"

	// "github.com/ez4o/go-try"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	// "github.com/twharmon/slices"
)

type CRound struct{}

func (_ CRound) validateBooking(bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(); err != nil {
		return booking, err
	}

	if booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		return booking, errors.New(booking.Bag + "Bag chưa TIME OUT ")
	}

	return booking, nil
}

func (_ CRound) createRound(booking model_booking.Booking, newHole int) error {
	// golfFeeModel := models.GolfFee{
	// 	PartnerUid: booking.PartnerUid,
	// 	CourseUid:  booking.CourseUid,
	// 	GuestStyle: booking.GuestStyle,
	// }

	// golfFee, err := golfFeeModel.GetGuestStyleOnDay()
	// if err != nil {
	// 	return models.Round{}, err
	// }

	round := models.Round{}
	round.BillCode = booking.BillCode
	totalRound, _ := round.Count()
	round.Index = int(totalRound + 1)
	round.Bag = booking.Bag
	round.PartnerUid = booking.PartnerUid
	round.CourseUid = booking.CourseUid
	round.GuestStyle = booking.GuestStyle
	round.Hole = newHole
	round.MemberCardUid = booking.MemberCardUid
	round.TeeOffTime = booking.CheckInTime
	round.Pax = 1
	if len(booking.ListGolfFee) > 0 {
		round.BuggyFee = booking.ListGolfFee[0].BuggyFee
		round.CaddieFee = booking.ListGolfFee[0].CaddieFee
		round.GreenFee = booking.ListGolfFee[0].GreenFee
	}

	errCreateRound := round.Create()
	if errCreateRound != nil {
		return errCreateRound
	}

	return nil
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

	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	for _, data := range body.BookUidList {
		// validate booking_uid
		booking, err = cRound.validateBooking(data)
		if err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// Tạo uid cho booking mới
		bookingUid := uuid.New()
		bUid := booking.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

		// Check giá guest style
		if booking.GuestStyle != "" {
			//Guest style
			golfFeeModel := models.GolfFee{
				PartnerUid: booking.PartnerUid,
				CourseUid:  booking.CourseUid,
				GuestStyle: booking.GuestStyle,
			}
			// Lấy phí bởi Guest style với ngày tạo
			golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay()
			if errFindGF != nil {
				response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
				return
			}

			getInitListGolfFeeForAddRound(&booking, golfFee, body.Hole)
		} else {
			// Get config course
			course := models.Course{}
			course.Uid = booking.CourseUid
			errCourse := course.FindFirst()
			if errCourse != nil {
				response_message.BadRequest(c, errCourse.Error())
				return
			}
			// Lấy giá đặc biệt của member card
			if booking.MemberCardUid != "" {
				// Get Member Card
				memberCard := models.MemberCard{}
				memberCard.Uid = booking.MemberCardUid
				errFind := memberCard.FindFirst()
				if errFind != nil {
					response_message.BadRequest(c, errFind.Error())
					return
				}

				if memberCard.PriceCode == 1 {
					getInitListGolfFeeWithOutGuestStyleForAddRound(&booking, course.RateGolfFee, memberCard.CaddieFee, memberCard.BuggyFee, memberCard.GreenFee, body.Hole)
				}
			}

			// Lấy giá đặc biệt của member card
			if booking.AgencyId > 0 {
				agency := models.Agency{}
				agency.Id = booking.AgencyId
				errFindAgency := agency.FindFirst()
				if errFindAgency != nil || agency.Id == 0 {
					response_message.BadRequest(c, "agency"+errFindAgency.Error())
					return
				}

				agencySpecialPrice := models.AgencySpecialPrice{
					AgencyId: agency.Id,
				}
				errFSP := agencySpecialPrice.FindFirst()
				if errFSP == nil && agencySpecialPrice.Id > 0 {
					// Tính lại giá
					// List Booking GolfFee
					getInitListGolfFeeWithOutGuestStyleForAddRound(&booking, course.RateGolfFee, agencySpecialPrice.CaddieFee, agencySpecialPrice.BuggyFee, agencySpecialPrice.GreenFee, body.Hole)
				}
			}
		}

		if len(booking.MainBags) > 0 {
			// Get data main bag
			bookingMain := model_booking.Booking{}
			bookingMain.Uid = booking.MainBags[0].BookingUid
			if err := bookingMain.FindFirst(); err != nil {
				return
			}

			for _, v1 := range bookingMain.MainBagPay {
				// TODO: Tính Fee cho sub bag fee
				if v1 == constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS {
					for i, v2 := range bookingMain.ListGolfFee {
						if v2.Bag == booking.Bag {
							bookingMain.ListGolfFee[i].BuggyFee = booking.ListGolfFee[0].BuggyFee
							bookingMain.ListGolfFee[i].CaddieFee = booking.ListGolfFee[0].CaddieFee
							bookingMain.ListGolfFee[i].GreenFee = booking.ListGolfFee[0].GreenFee
						}
					}
					// Update mush pay, current bag
					totalPayChange := booking.ListGolfFee[0].CaddieFee + booking.ListGolfFee[0].BuggyFee + booking.ListGolfFee[0].GreenFee

					bookingMain.MushPayInfo.MushPay += totalPayChange
					bookingMain.MushPayInfo.TotalGolfFee += totalPayChange

					errUpdateBooking := bookingMain.Update()

					if errUpdateBooking != nil {
						response_message.BadRequest(c, errUpdateBooking.Error())
						return
					}

					break
				}
			}

		}

		err = cRound.createRound(booking, body.Hole)
		if err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		booking.BagStatus = constants.BAG_STATUS_WAITING

		//Update mush pay, current bag
		if len(booking.ListGolfFee) > 0 {
			totalPayChange := booking.ListGolfFee[0].CaddieFee + booking.ListGolfFee[0].BuggyFee + booking.ListGolfFee[0].GreenFee

			booking.MushPayInfo.MushPay += totalPayChange
			booking.MushPayInfo.TotalGolfFee += totalPayChange
			booking.CurrentBagPrice.Amount += totalPayChange
			booking.CurrentBagPrice.GolfFee += totalPayChange
		}
		errCreateBooking := booking.Create(bUid)

		if errCreateBooking != nil {
			response_message.BadRequest(c, errCreateBooking.Error())
			return
		}
	}

	// create round and add round

	// booking.Rounds = append(booking.Rounds, newRound)

	// Golf fee for this Round
	// Thêm golf Fee cho Round mới
	// newRoundGolfFee := model_booking.BookingGolfFee{
	// 	BookingUid: booking.Uid,
	// 	PlayerName: booking.CustomerName,
	// 	Bag:        booking.Bag,
	// 	CaddieFee:  newRound.CaddieFee,
	// 	BuggyFee:   newRound.BuggyFee,
	// 	GreenFee:   newRound.GreenFee,
	// 	RoundIndex: newRound.Index,
	// }

	// // update list_golf_fee
	// booking.ListGolfFee = append(booking.ListGolfFee, newRoundGolfFee)

	// // update current_bag_price
	// booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
	// if err != nil {
	// 	try.ThrowOnError(err)
	// }

	// // update must_pay_info
	// booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
	// if err != nil {
	// 	try.ThrowOnError(err)
	// }

	// booking.CmsUser = prof.UserName
	// booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	// if err := booking.Update(); err != nil {
	// 	try.ThrowOnError(err)
	// }

	okResponse(c, booking)
}

// func (cRound CRound) SplitRound(c *gin.Context, prof models.CmsUser) {
// 	var body request.SplitRoundBody
// 	var booking model_booking.Booking
// 	var err error
// 	var hasError = false

// 	try.Try(func() {
// 		if err := c.BindJSON(&body); err != nil {
// 			log.Print("SplitRound BindJSON error")
// 			try.ThrowOnError(err)
// 		}

// 		validate := validator.New()

// 		if err := validate.Struct(body); err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		// validate booking_uid
// 		booking, err = cRound.validateBooking(body.BookingUid)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}
// 	}).Catch(func(e error, st *try.StackTrace) {
// 		response_message.BadRequest(c, "")
// 		hasError = true
// 	})

// 	if hasError {
// 		return
// 	}

// 	try.Try(func() {
// 		currentRound := booking.Rounds[body.RoundIndex]
// 		if currentRound.Hole <= 9 {
// 			try.ThrowOnError(errors.New("Hole invalid for split"))
// 			return
// 		}
// 		newRound := currentRound
// 		newRound.Hole = int(body.Hole)
// 		newRound.Index = len(booking.Rounds) + 1
// 		currentRound.Hole = currentRound.Hole - newRound.Hole
// 		booking.Rounds[body.RoundIndex] = currentRound
// 		booking.Rounds = append(booking.Rounds, newRound)

// 		// Thêm golf Fee cho Round mới
// 		newRoundGolfFee := model_booking.BookingGolfFee{
// 			BookingUid: booking.Uid,
// 			PlayerName: booking.CustomerName,
// 			Bag:        booking.Bag,
// 			CaddieFee:  newRound.CaddieFee,
// 			BuggyFee:   newRound.BuggyFee,
// 			GreenFee:   newRound.GreenFee,
// 			RoundIndex: newRound.Index,
// 		}

// 		// update list_golf_fee
// 		booking.ListGolfFee = append(booking.ListGolfFee, newRoundGolfFee)

// 		// update current_bag_price
// 		booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		// update must_pay_info
// 		booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		booking.CmsUser = prof.UserName
// 		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
// 		if err := booking.Update(); err != nil {
// 			try.ThrowOnError(err)
// 		}
// 	}).Catch(func(e error, st *try.StackTrace) {
// 		response_message.InternalServerError(c, err.Error())
// 		return
// 	})

// 	if hasError {
// 		return
// 	}

// 	okResponse(c, booking)
// }

// func (cRound CRound) MergeRound(c *gin.Context, prof models.CmsUser) {
// 	var body request.MergeRoundBody
// 	var booking model_booking.Booking
// 	var err error
// 	var hasError = false

// 	try.Try(func() {
// 		if err := c.BindJSON(&body); err != nil {
// 			log.Print("MergeRound BindJSON error")
// 			try.ThrowOnError(err)
// 		}

// 		validate := validator.New()

// 		if err := validate.Struct(body); err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		// validate booking_uid
// 		booking, err = cRound.validateBooking(body.BookingUid)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}
// 	}).Catch(func(e error, st *try.StackTrace) {
// 		response_message.BadRequest(c, "")
// 		hasError = true
// 	})

// 	if hasError {
// 		return
// 	}

// 	try.Try(func() {
// 		// create round
// 		totalHoles := slices.Reduce(booking.Rounds, func(prev int, item model_booking.BookingRound) int {
// 			return prev + item.Hole
// 		})

// 		newRound, err := cRound.createRound(booking, totalHoles)
// 		newRound.Index = 0
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		// booking.Rounds = append(model_booking.ListBookingRound{}, newRound)

// 		newRoundGolfFee := model_booking.BookingGolfFee{
// 			BookingUid: booking.Uid,
// 			PlayerName: booking.CustomerName,
// 			Bag:        booking.Bag,
// 			CaddieFee:  newRound.CaddieFee,
// 			BuggyFee:   newRound.BuggyFee,
// 			GreenFee:   newRound.GreenFee,
// 			RoundIndex: newRound.Index,
// 		}

// 		listGolfFeeTemp := model_booking.ListBookingGolfFee{}
// 		listGolfFeeTemp = append(listGolfFeeTemp, newRoundGolfFee)
// 		for i, v := range booking.ListGolfFee {
// 			if i > 0 && v.BookingUid != booking.Uid {
// 				listGolfFeeTemp = append(listGolfFeeTemp, v)
// 			}
// 		}

// 		// Udp golf fee
// 		booking.ListGolfFee = listGolfFeeTemp

// 		// update current_bag_price
// 		booking.CurrentBagPrice, err = cRound.updateCurrentBagPrice(booking, newRoundGolfFee.CaddieFee+newRoundGolfFee.BuggyFee+newRoundGolfFee.GreenFee)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		// update must_pay_info
// 		booking.MushPayInfo, err = cRound.updateMustPayInfo(booking)
// 		if err != nil {
// 			try.ThrowOnError(err)
// 		}

// 		booking.CmsUser = prof.UserName
// 		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
// 		if err := booking.Update(); err != nil {
// 			try.ThrowOnError(err)
// 		}
// 	}).Catch(func(e error, st *try.StackTrace) {
// 		response_message.InternalServerError(c, err.Error())
// 		return
// 	})

// 	if hasError {
// 		return
// 	}

// 	okResponse(c, booking)
// }
