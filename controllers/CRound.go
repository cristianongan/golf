package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/twharmon/slices"
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

func (_ CRound) createRound(booking model_booking.Booking, newHole int, isMerge bool) error {
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
	if isMerge {
		round.Index = 0
	} else {
		round.BillCode = booking.BillCode
		totalRound, _ := round.CountWithBillCode()
		round.Index = int(totalRound + 1)
	}
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
							bookingMain.ListGolfFee[i].BookingUid = booking.Uid
							bookingMain.ListGolfFee[i].BuggyFee = booking.ListGolfFee[0].BuggyFee
							bookingMain.ListGolfFee[i].CaddieFee = booking.ListGolfFee[0].CaddieFee
							bookingMain.ListGolfFee[i].GreenFee = booking.ListGolfFee[0].GreenFee

							break
						}
					}
					for i, v2 := range bookingMain.SubBags {
						if v2.GolfBag == booking.Bag {
							bookingMain.SubBags[i].BookingUid = booking.Uid

							break
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

		err = cRound.createRound(booking, body.Hole, false)
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

	okRes(c)
}

func (cRound CRound) SplitRound(c *gin.Context, prof models.CmsUser) {
	var body request.SplitRoundBody
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

	booking, err = cRound.validateBooking(body.BookingUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	currentRound := models.Round{BillCode: booking.BillCode, Index: body.RoundIndex}
	errRound := currentRound.FindFirst()

	if errRound != nil {
		response_message.BadRequest(c, errRound.Error())
		return
	}

	if currentRound.Hole <= 9 {
		response_message.BadRequest(c, errors.New("Hole invalid for split").Error())
		return
	}
	newRound := currentRound
	newRound.Hole = int(body.Hole)
	newRound.Index = body.RoundIndex + 1
	currentRound.Hole = currentRound.Hole - newRound.Hole

	// Update giá cho current round và new round
	updateListGolfFeeWithRound(&currentRound, &booking, currentRound.Hole)
	updateListGolfFeeWithRound(&newRound, &booking, newRound.Hole)

	errUpdate := currentRound.Update()
	if errUpdate != nil {
		response_message.BadRequest(c, errUpdate.Error())
		return
	}

	errCreate := newRound.Create()
	if errCreate != nil {
		response_message.BadRequest(c, errCreate.Error())
		return
	}

	// Update lại giá cho main bag
	if len(booking.MainBags) > 0 {
		//init fee
		var greenFee int64 = 0
		var buggyFee int64 = 0
		var caddieFee int64 = 0
		var totalFeeAfter int64 = 0

		// Get data main bag
		bookingMain := model_booking.Booking{}
		bookingMain.Uid = booking.MainBags[0].BookingUid
		if err := bookingMain.FindFirst(); err != nil {
			return
		}

		// Check loại tính tiền của main bag
		checkIsFirstRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
		checkIsNextRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

		// get all round
		round := models.Round{
			BillCode: booking.BillCode,
		}

		rounds, errR := round.FindAll()
		if errR != nil {
			response_message.BadRequest(c, errR.Error())
			return
		}

		if checkIsFirstRound > -1 && checkIsNextRound > -1 {
			for _, v := range rounds {
				greenFee += v.GreenFee
				buggyFee += v.BuggyFee
				caddieFee += v.CaddieFee
			}

		} else if checkIsFirstRound > -1 {
			for _, v := range rounds {
				greenFee += v.GreenFee
				buggyFee += v.BuggyFee
				caddieFee += v.CaddieFee

				break
			}

		} else if checkIsNextRound > -1 {
			for i, v := range rounds {
				if i != 0 {
					greenFee += v.GreenFee
					buggyFee += v.BuggyFee
					caddieFee += v.CaddieFee
				}
			}
		}

		for i, v2 := range bookingMain.ListGolfFee {
			if v2.Bag == booking.Bag {
				totalFeeAfter += v2.BuggyFee + v2.CaddieFee + v2.GreenFee
				bookingMain.ListGolfFee[i].BuggyFee = buggyFee
				bookingMain.ListGolfFee[i].CaddieFee = caddieFee
				bookingMain.ListGolfFee[i].GreenFee = greenFee

				break
			}
		}
		// Update mush pay, current bag
		totalFeeBefore := buggyFee + caddieFee + greenFee

		bookingMain.MushPayInfo.MushPay += totalFeeBefore - totalFeeAfter
		bookingMain.MushPayInfo.TotalGolfFee += totalFeeBefore - totalFeeAfter

		errUpdateBooking := bookingMain.Update()

		if errUpdateBooking != nil {
			response_message.BadRequest(c, errUpdateBooking.Error())
			return
		}

	}

	okRes(c)
}

func (cRound CRound) MergeRound(c *gin.Context, prof models.CmsUser) {
	var body request.MergeRoundBody
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

	booking, err = cRound.validateBooking(body.BookingUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	round := models.Round{BillCode: booking.BillCode}
	listRound, _ := round.FindAll()

	// create round
	totalHoles := slices.Reduce(listRound, func(prev int, item models.Round) int {
		return prev + item.Hole
	})

	// Update giá cho current round và new round
	updateListGolfFeeWithRound(&round, &booking, totalHoles)

	err = cRound.createRound(booking, totalHoles, true)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}
