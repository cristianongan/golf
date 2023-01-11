package controllers

import (
	"errors"
	"log"
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

		round := models.Round{}
		round.BillCode = booking.BillCode
		totalRound, _ := round.CountWithBillCode(db)
		round.Index = int(totalRound + 1)
		round.Bag = booking.Bag
		round.PartnerUid = booking.PartnerUid
		round.CourseUid = booking.CourseUid
		round.GuestStyle = booking.GuestStyle
		round.Hole = hole
		round.MemberCardUid = booking.MemberCardUid
		round.TeeOffTime = booking.CheckInTime
		round.Pax = 1

		errCreateRound := round.Create(db)
		if errCreateRound != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// Update giá
		cRound.UpdateListFeePriceInRound(c, db, &booking, booking.GuestStyle, &round, hole)
		// Update lại bag_status của booking cũ
		booking.AddedRound = setBoolForCursor(true)
		booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
		booking.Update(db)

		// Tạo booking mới khi add round
		newBooking = cloneToBooking(booking)
		newBooking.BagStatus = constants.BAG_STATUS_WAITING
		newBooking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
		newBooking.FlightId = 0
		newBooking.HoleTimeOut = 0
		newBooking.TimeOutFlight = 0
		newBooking.CourseType = body.CourseType
		newBooking.ShowCaddieBuggy = setBoolForCursor(false)
		newBooking.AddedRound = setBoolForCursor(false)
		errCreateBooking := newBooking.Create(db, bUid)

		if errCreateBooking != nil {
			response_message.BadRequest(c, errCreateBooking.Error())
			return
		}
	}

	res := getBagDetailFromBooking(db, newBooking)

	okResponse(c, res)
}
func (cRound CRound) GetFeeOfRound(c *gin.Context, db *gorm.DB, booking *model_booking.Booking, guestStyle string, hole int) (int64, int64, int64, error) {
	caddieFee := int64(0)
	buggyFee := int64(0)
	greenFee := int64(0)

	if guestStyle != "" {
		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			GuestStyle: guestStyle,
		}
		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
		if errFindGF != nil {
			return 0, 0, 0, errors.New("golf fee err " + errFindGF.Error())
		}

		caddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, hole)
		buggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, hole)
		greenFee = utils.GetFeeFromListFee(golfFee.GreenFee, hole)

	} else {
		// Get config course
		course := models.Course{}
		course.Uid = booking.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			return 0, 0, 0, errCourse
		}
		// Lấy giá đặc biệt của member card
		if booking.MemberCardUid != "" {
			// Get Member Card
			memberCard := models.MemberCard{}
			memberCard.Uid = booking.MemberCardUid
			errFind := memberCard.FindFirst(db)
			if errFind != nil {
				return 0, 0, 0, errFind
			}

			if memberCard.PriceCode == 1 {
				caddieFee = utils.CalculateFeeByHole(hole, memberCard.CaddieFee, course.RateGolfFee)
				buggyFee = utils.CalculateFeeByHole(hole, memberCard.BuggyFee, course.RateGolfFee)
				greenFee = utils.CalculateFeeByHole(hole, memberCard.GreenFee, course.RateGolfFee)
			}
		}

		// Lấy giá đặc biệt của member card
		if booking.AgencyId > 0 {
			agency := models.Agency{}
			agency.Id = booking.AgencyId
			errFindAgency := agency.FindFirst(db)
			if errFindAgency != nil || agency.Id == 0 {
				return 0, 0, 0, errFindAgency
			}

			agencySpecialPrice := models.AgencySpecialPrice{
				AgencyId: agency.Id,
			}
			errFSP := agencySpecialPrice.FindFirst(db)
			if errFSP == nil && agencySpecialPrice.Id > 0 {
				// Tính lại giá
				// List Booking GolfFee
				caddieFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.CaddieFee, course.RateGolfFee)
				buggyFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.BuggyFee, course.RateGolfFee)
				greenFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.GreenFee, course.RateGolfFee)
			}
		}
	}
	return caddieFee, buggyFee, greenFee, nil
}

func (cRound CRound) UpdateListFeePriceInRound(c *gin.Context, db *gorm.DB, booking *model_booking.Booking, guestStyle string, round *models.Round, hole int) {
	caddieFee, buggyFee, greenFee, err := cRound.GetFeeOfRound(c, db, booking, guestStyle, hole)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if round != nil {
		round.CaddieFee = caddieFee
		round.BuggyFee = buggyFee
		round.GreenFee = greenFee

		if errRoundUdp := round.Update(db); errRoundUdp != nil {
			response_message.BadRequestDynamicKey(c, "UPDATE_ERROR", "")
			return
		}
	}

	if booking != nil {
		cRound.UpdateGolfFeeInBooking(booking, db)
	}
}

func (cRound CRound) SplitRound(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.SplitRoundBody
	var booking model_booking.Booking

	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	booking = model_booking.Booking{}
	booking.Uid = body.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequestDynamicKey(c, "BOOKING_NOT_FOUND", "")
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
	newRound.Id = 0
	newRound.Hole = int(body.Hole)
	newRound.Index = currentRound.Index + 1
	currentRound.Hole = currentRound.Hole - newRound.Hole

	// Update giá cho current round và new round
	updateListGolfFeeWithRound(db, &currentRound, booking, currentRound.GuestStyle, currentRound.Hole)
	updateListGolfFeeWithRound(db, &newRound, booking, newRound.GuestStyle, newRound.Hole)

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

	cRound.UpdateGolfFeeInBooking(&booking, db)
	// Update lại giá cho main bag

	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}

func (cRound CRound) MergeRound(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.MergeRoundBody
	var booking model_booking.Booking

	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	// validate := validator.New()

	// if err := validate.Struct(body); err != nil {
	// 	response_message.BadRequest(c, "Body format type error")
	// 	return
	// }

	booking = model_booking.Booking{}
	booking.Uid = body.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequestDynamicKey(c, "BOOKING_NOT_FOUND", "")
		return
	}

	roundR := models.Round{BillCode: booking.BillCode}
	listRound, _ := roundR.FindAll(db)

	if len(listRound) < 2 {
		response_message.BadRequestDynamicKey(c, "MERGE_ROUND_NOT_ENOUGH", "")
		return
	}
	// create round
	totalHoles := slices.Reduce(listRound, func(prev int, item models.Round) int {
		return prev + item.Hole
	})

	// Update giá cho current round và new round
	caddieFee, buggyFee, greenFee, _ := cRound.GetFeeOfRound(c, db, &booking, listRound[0].GuestStyle, totalHoles)

	newRound := models.Round{}
	newRound.Index = 1
	newRound.BillCode = booking.BillCode
	newRound.Bag = booking.Bag
	newRound.PartnerUid = booking.PartnerUid
	newRound.CourseUid = booking.CourseUid
	newRound.GuestStyle = listRound[0].GuestStyle
	newRound.Hole = totalHoles
	newRound.MemberCardUid = booking.MemberCardUid
	newRound.TeeOffTime = booking.CheckInTime
	newRound.CaddieFee = caddieFee
	newRound.BuggyFee = buggyFee
	newRound.GreenFee = greenFee
	newRound.Pax = 1

	if errCreateRound := newRound.Create(db); errCreateRound != nil {
		response_message.BadRequest(c, errCreateRound.Error())
		return
	}

	//Xóa các round cũ
	for _, item := range listRound {
		if errRound := item.Delete(db); errRound != nil {
			response_message.BadRequest(c, "Merge error")
			return
		}
	}

	// update fee booking
	cRound.UpdateGolfFeeInBooking(&booking, db)
	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}

func (cRound CRound) ChangeGuestyleOfRound(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.ChangeGuestyleRound

	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "Body format type error")
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequestDynamicKey(c, "BOOKING_NOT_FOUND", "")
		return
	}

	round := models.Round{}
	round.Id = body.RoundId
	round.BillCode = booking.BillCode

	if errRound := round.FindFirst(db); errRound != nil {
		response_message.BadRequestDynamicKey(c, "ROUND_NOT_FOUND", "")
		return
	}

	if body.GuestStyle != "" {
		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			GuestStyle: body.GuestStyle,
		}

		if errGS := golfFeeModel.FindFirst(db); errGS != nil {
			response_message.BadRequestDynamicKey(c, "GUEST_STYLE_NOT_FOUND", "")
			return
		}
	}
	round.GuestStyle = body.GuestStyle

	// Update lại GS booking
	go func() {
		Rround := models.Round{}
		Rround.BillCode = booking.BillCode
		list, _ := Rround.FindAll(db)

		if round.Index == len(list) {
			booking.GuestStyle = body.GuestStyle
			golfFee := models.GolfFee{
				GuestStyle: body.GuestStyle,
			}

			if golfFeeFind := golfFee.FindFirst(db); golfFeeFind == nil {
				booking.GuestStyleName = golfFee.GuestStyleName
				booking.Update(db)
			}
		}
	}()
	// Update giá
	cRound.UpdateListFeePriceInRound(c, db, &booking, body.GuestStyle, &round, round.Hole)

	okResponse(c, round)
}
func (cRound CRound) UpdateListFeePriceInBookingAndRound(c *gin.Context, db *gorm.DB, booking model_booking.Booking, guestStyle string, hole int) {
	round := models.Round{
		BillCode: booking.BillCode,
	}

	if errFindRound := round.LastRound(db); errFindRound != nil {
		response_message.BadRequestDynamicKey(c, "ROUND_NOT_FOUND", "")
		log.Println("Round not found")
		return
	}

	// if round.Hole != hole {
	// Update số hole của Round
	round.Hole = hole

	// Update lại giá của Round theo số hố
	cRound1 := CRound{}
	cRound1.UpdateListFeePriceInRound(c, db, &booking, round.GuestStyle, &round, hole)
	// }
}

// Khi changeToMain thì reset lại các round đã trả bởi main bag trước đó
func (cRound CRound) ResetRoundPaidByMain(billCode string, db *gorm.DB) {
	round1 := models.Round{BillCode: billCode, Index: 1}
	if errRound1 := round1.FindFirst(db); errRound1 == nil {
		round1.MainBagPaid = setBoolForCursor(false)
		round1.Update(db)
	}
	round2 := models.Round{BillCode: billCode, Index: 2}
	if errRound2 := round2.FindFirst(db); errRound2 == nil {
		round2.MainBagPaid = setBoolForCursor(false)
		round2.Update(db)
	}
}

// Update lại bag cho round 1 khi check in
func (cRound CRound) UpdateBag(booking model_booking.Booking, db *gorm.DB) {
	round1 := models.Round{BillCode: booking.BillCode, Index: 1}
	if errRound1 := round1.FindFirst(db); errRound1 == nil {
		if round1.Bag == "" {
			round1.Bag = booking.Bag
			round1.Update(db)
		}
	}
}

// Update lại giá sau khi add, merge, split round
func (cRound CRound) UpdateGolfFeeInBooking(booking *model_booking.Booking, db *gorm.DB) {
	booking.UpdatePriceDetailCurrentBag(db)
	booking.UpdateMushPay(db)
	booking.Update(db)
	go handlePayment(db, *booking)

	if len(booking.MainBags) > 0 {
		// Get data main bag
		bookingMain := model_booking.Booking{
			CourseUid:   booking.CourseUid,
			PartnerUid:  booking.PartnerUid,
			Bag:         booking.MainBags[0].GolfBag,
			BookingDate: booking.BookingDate,
		}
		if err := bookingMain.FindFirst(db); err != nil {
			return
		}

		bookingMain.UpdatePriceDetailCurrentBag(db)
		bookingMain.UpdateMushPay(db)
		bookingMain.Update(db)
		go handlePayment(db, bookingMain)
	}
}
