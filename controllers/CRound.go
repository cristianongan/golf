package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/twharmon/slices"
	"gorm.io/gorm"
)

type CRound struct{}

func (_ CRound) validateBooking(db *gorm.DB, bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(db); err != nil {
		return booking, err
	}

	if booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		return booking, errors.New("Lỗi Add Round")
	}

	return booking, nil
}

func (_ CRound) createRound(db *gorm.DB, booking model_booking.Booking, newHole int, isMerge bool) error {
	round := models.Round{}
	if isMerge {
		round.Index = 0
	} else {
		totalRound, _ := round.CountWithBillCode(db)
		round.Index = int(totalRound + 1)
	}
	round.BillCode = booking.BillCode
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

	errCreateRound := round.Create(db)
	if errCreateRound != nil {
		return errCreateRound
	}

	return nil
}

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.AddRoundBody
	var booking model_booking.Booking
	var newBooking model_booking.Booking
	var err error

	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	hole := 0
	if body.Hole == nil {
		hole = 18
	} else {
		hole = *body.Hole
	}

	for _, data := range body.BookUidList {
		// validate booking_uid
		booking, err = cRound.validateBooking(db, data)
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
			golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
			if errFindGF != nil {
				response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
				return
			}

			getInitListGolfFeeForAddRound(&booking, golfFee, hole)
		} else {
			// Get config course
			course := models.Course{}
			course.Uid = booking.CourseUid
			errCourse := course.FindFirst(db)
			if errCourse != nil {
				response_message.BadRequest(c, errCourse.Error())
				return
			}
			// Lấy giá đặc biệt của member card
			if booking.MemberCardUid != "" {
				// Get Member Card
				memberCard := models.MemberCard{}
				memberCard.Uid = booking.MemberCardUid
				errFind := memberCard.FindFirst(db)
				if errFind != nil {
					response_message.BadRequest(c, errFind.Error())
					return
				}

				if memberCard.PriceCode == 1 {
					getInitListGolfFeeWithOutGuestStyleForAddRound(&booking, course.RateGolfFee, memberCard.CaddieFee, memberCard.BuggyFee, memberCard.GreenFee, hole)
				}
			}

			// Lấy giá đặc biệt của member card
			if booking.AgencyId > 0 {
				agency := models.Agency{}
				agency.Id = booking.AgencyId
				errFindAgency := agency.FindFirst(db)
				if errFindAgency != nil || agency.Id == 0 {
					response_message.BadRequest(c, "agency"+errFindAgency.Error())
					return
				}

				agencySpecialPrice := models.AgencySpecialPrice{
					AgencyId: agency.Id,
				}
				errFSP := agencySpecialPrice.FindFirst(db)
				if errFSP == nil && agencySpecialPrice.Id > 0 {
					// Tính lại giá
					// List Booking GolfFee
					getInitListGolfFeeWithOutGuestStyleForAddRound(&booking, course.RateGolfFee, agencySpecialPrice.CaddieFee, agencySpecialPrice.BuggyFee, agencySpecialPrice.GreenFee, hole)
				}
			}
		}

		if len(booking.MainBags) > 0 {
			// Get data main bag
			bookingMain := model_booking.Booking{}
			bookingMain.Uid = booking.MainBags[0].BookingUid
			if err := bookingMain.FindFirst(db); err != nil {
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

					errUpdateBooking := bookingMain.Update(db)

					if errUpdateBooking != nil {
						response_message.BadRequest(c, errUpdateBooking.Error())
						return
					}

					break
				}
			}

		}

		err = cRound.createRound(db, booking, hole, false)
		if err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		//Update mush pay, current bag
		if len(booking.ListGolfFee) > 0 {
			totalPayChange := booking.ListGolfFee[0].CaddieFee + booking.ListGolfFee[0].BuggyFee + booking.ListGolfFee[0].GreenFee

			booking.MushPayInfo.MushPay += totalPayChange
			booking.MushPayInfo.TotalGolfFee += totalPayChange
			booking.CurrentBagPrice.Amount += totalPayChange
			booking.CurrentBagPrice.GolfFee += totalPayChange
		}

		// Tạo booking mới khi add round
		newBooking = cloneToBooking(booking)
		newBooking.BagStatus = constants.BAG_STATUS_WAITING
		newBooking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
		newBooking.FlightId = 0
		newBooking.TimeOutFlight = 0
		newBooking.CourseType = body.CourseType
		newBooking.ShowCaddieBuggy = newTrue(false)
		errCreateBooking := newBooking.Create(db, bUid)

		if errCreateBooking != nil {
			response_message.BadRequest(c, errCreateBooking.Error())
			return
		}

		// Update lại bag_status của booking cũ
		booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
		go booking.Update(db)
	}

	res := getBagDetailFromBooking(db, newBooking)

	okResponse(c, res)
}

func (cRound CRound) SplitRound(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	booking, err = cRound.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	currentRound := models.Round{}
	currentRound.Id = body.RoundId
	errRound := currentRound.FindFirst(db)

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
	newRound.Index = currentRound.Index + 1
	currentRound.Hole = currentRound.Hole - newRound.Hole

	// Update giá cho current round và new round
	updateListGolfFeeWithRound(db, &currentRound, booking, currentRound.Hole)
	updateListGolfFeeWithRound(db, &newRound, booking, newRound.Hole)

	errUpdate := currentRound.Update(db)
	if errUpdate != nil {
		response_message.BadRequest(c, errUpdate.Error())
		return
	}

	errCreate := newRound.Create(db)
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
		if err := bookingMain.FindFirst(db); err != nil {
			return
		}

		// Check loại tính tiền của main bag
		checkIsFirstRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
		checkIsNextRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

		// get all round
		round := models.Round{
			BillCode: booking.BillCode,
		}

		rounds, errR := round.FindAll(db)
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

		errUpdateBooking := bookingMain.Update(db)

		if errUpdateBooking != nil {
			response_message.BadRequest(c, errUpdateBooking.Error())
			return
		}

	}

	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}

func (cRound CRound) MergeRound(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	booking, err = cRound.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	round := models.Round{BillCode: booking.BillCode}
	listRound, _ := round.FindAll(db)

	// create round
	totalHoles := slices.Reduce(listRound, func(prev int, item models.Round) int {
		return prev + item.Hole
	})

	// Update giá cho current round và new round
	updateListGolfFeeWithRound(db, &round, booking, totalHoles)

	// update fee booking
	totalPayChange := round.BuggyFee + round.CaddieFee + round.CaddieFee

	booking.ListGolfFee[0].BuggyFee = round.BuggyFee
	booking.ListGolfFee[0].CaddieFee = round.CaddieFee
	booking.ListGolfFee[0].GreenFee = round.GreenFee

	booking.MushPayInfo.MushPay += totalPayChange
	booking.MushPayInfo.TotalGolfFee += totalPayChange
	booking.CurrentBagPrice.Amount += totalPayChange
	booking.CurrentBagPrice.GolfFee += totalPayChange

	err = cRound.createRound(db, booking, totalHoles, true)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	err = booking.Update(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update lại giá cho main bag
	if len(booking.MainBags) > 0 {
		//init fee

		var totalFeeAfter int64 = 0

		// Get data main bag
		bookingMain := model_booking.Booking{}
		bookingMain.Uid = booking.MainBags[0].BookingUid
		if err := bookingMain.FindFirst(db); err != nil {
			return
		}

		// Check loại tính tiền của main bag
		checkIsFirstRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)

		if checkIsFirstRound > -1 {
			for i, v1 := range bookingMain.ListGolfFee {
				if v1.Bag == booking.Bag {
					totalFeeAfter += v1.BuggyFee + v1.CaddieFee + v1.GreenFee
					bookingMain.ListGolfFee[i].BuggyFee = round.BuggyFee
					bookingMain.ListGolfFee[i].CaddieFee = round.CaddieFee
					bookingMain.ListGolfFee[i].GreenFee = round.GreenFee

					break
				}
			}
			// Update mush pay, current bag
			totalFeeBefore := round.BuggyFee + round.CaddieFee + round.GreenFee

			bookingMain.MushPayInfo.MushPay += totalFeeBefore - totalFeeAfter
			bookingMain.MushPayInfo.TotalGolfFee += totalFeeBefore - totalFeeAfter

		} else {
			for _, v1 := range bookingMain.ListGolfFee {
				if v1.Bag == booking.Bag {
					totalFeeAfter += v1.BuggyFee + v1.CaddieFee + v1.GreenFee

					break
				}
			}

			// remove sub bag
			subBags := utils.ListSubBag{}
			for _, v2 := range bookingMain.SubBags {
				if v2.GolfBag != booking.Bag {
					subBags = append(subBags, v2)
				}
			}
			bookingMain.SubBags = subBags

			// Update mush pay, current bag
			bookingMain.MushPayInfo.MushPay -= totalFeeAfter
			bookingMain.MushPayInfo.TotalGolfFee -= totalFeeAfter
		}

		errUpdateBooking := bookingMain.Update(db)

		if errUpdateBooking != nil {
			response_message.BadRequest(c, errUpdateBooking.Error())
			return
		}

	}

	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}
