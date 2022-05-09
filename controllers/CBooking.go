package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CBooking struct{}

/// --------- Booking ----------

/*
 Tạo Booking rồi Check In luôn
*/
func (_ *CBooking) CreateBookingCheckIn(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// Check validated, Check đã tạo
	if body.Bag == "" || body.CustomerName == "" || body.GuestStyle == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.Hole <= 0 {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Get GolfFee
	golfFeeGet := models.GolfFee{
		GuestStyle: body.GuestStyle,
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
	}
	golfFee, errFind := golfFeeGet.GetGuestStyleOnDay()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// Booking
	booking := model_booking.Booking{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Bag:        body.Bag,
		Hole:       body.Hole,
	}

	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errDate == nil {
		booking.CreatedDate = dateDisplay
	} else {
		log.Println("booking date display err ", errDate.Error())
	}

	//Check duplicated
	isDuplicated, errDupli := booking.IsDuplicated(false, true)
	if isDuplicated {
		if errDupli != nil {
			response_message.DuplicateRecord(c, errDupli.Error())
			return
		}
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	/*
		Data cần xử lý:
		CurrentBagPrice
		ListGolfFee
		MushPayInfo
		Rounds
		ListServiceItems (?)
		MainBags (?)
		SubBags (?)
		MainBagNoPay (?)
	*/

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

	// List Booking GolfFee
	listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(bUid, body, golfFee)
	booking.ListGolfFee = listBookingGolfFee

	// Current Bag Price Detail
	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	booking.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(booking)
	booking.MushPayInfo = mushPayInfo

	// Check in Time
	checkInTime := time.Now().Unix()

	// Rounds: Init First
	listRounds := initListRound(booking, bookingGolfFee, checkInTime)
	booking.Rounds = listRounds

	// Check in out
	booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_IN
	booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
	booking.CheckInTime = checkInTime

	errC := booking.Create(bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Tạo Booking từ TeeSheet
*/
func (_ *CBooking) CreateBooking(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		TeeType:    body.TeeType,
		TeePath:    body.TeePath,
		TeeTime:    body.TeeTime,
		TeeOffTime: body.TeeTime,
		TurnTime:   body.TurnTime,
		RowIndex:   body.RowIndex,
		CmsUser:    body.CmsUser,
		Hole:       body.Hole,
	}

	if body.Bag != "" {
		booking.Bag = body.Bag
	}

	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errDate == nil {
		booking.CreatedDate = dateDisplay
	} else {
		log.Println("booking date display err ", errDate.Error())
	}

	//Check duplicated
	isDuplicated, errDupli := booking.IsDuplicated(true, true)
	if isDuplicated {
		if errDupli != nil {
			response_message.DuplicateRecord(c, errDupli.Error())
			return
		}
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner()
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return
		}

		booking.MemberCardUid = body.MemberCardUid
		booking.CardId = memberCard.CardId
		booking.CustomerName = owner.Name
		booking.CustomerUid = owner.Uid

	} else {
		booking.CustomerName = body.CustomerName
	}

	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	//Guest style
	golfFeeModel := models.GolfFee{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		GuestStyle: body.GuestStyle,
	}
	// Lấy phí bởi Guest style với ngày tạo
	golfFee, errFind := golfFeeModel.GetGuestStyleOnDay()
	if errFind != nil {
		response_message.InternalServerError(c, "golf fee err "+errFind.Error())
		return
	}
	booking.GuestStyle = body.GuestStyle
	booking.GuestStyleName = golfFee.GuestStyleName

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

	// List Booking GolfFee
	listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(bUid, body, golfFee)
	booking.ListGolfFee = listBookingGolfFee

	// Current Bag Price Detail
	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	booking.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(booking)
	booking.MushPayInfo = mushPayInfo

	// Check In Out
	checkInTime := time.Now().Unix()
	booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_INIT
	booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING

	// Rounds: Init First
	listRounds := initListRound(booking, bookingGolfFee, checkInTime)
	booking.Rounds = listRounds

	errC := booking.Create(bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Get booking Detail With Uid
*/
func (_ *CBooking) GetBookingDetail(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Get booking by Bag
 Get Booking Bag trong ngày
*/
func (_ *CBooking) GetBookingByBag(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.Bag == "" {
		response_message.BadRequest(c, errors.New("Bag invalid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.Bag = form.Bag
	toDayDate, errD := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errD != nil {
		response_message.InternalServerError(c, errD.Error())
		return
	}
	booking.CreatedDate = toDayDate

	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Danh sách booking
*/
func (_ *CBooking) GetListBooking(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	bookingR := model_booking.Booking{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := bookingR.FindList(page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

/*
 Cập nhật booking
 Thêm Service item
*/
func (_ *CBooking) UpdateBooking(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.Booking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Không udp lại Bag
	// if body.Bag != "" {
	// 	booking.Bag = body.Bag
	// }
	if body.ListServiceItems != nil {
		countItems := len(body.ListServiceItems)
		if countItems > 0 {
			for i := 0; i < countItems; i++ {
				body.ListServiceItems[i].BookingUid = booking.Uid
				body.ListServiceItems[i].PlayerName = booking.CustomerName
				body.ListServiceItems[i].Bag = booking.Bag
			}
		}
	}

	if body.GuestStyle != "" {
		booking.GuestStyle = body.GuestStyle
	}

	//Upd Main Pay for Sub
	if body.MainBagNoPay != nil {
		booking.MainBagNoPay = body.MainBagNoPay
	}

	//Update service items
	booking.ListServiceItems = body.ListServiceItems

	// Tính lại giá
	booking.UpdatePriceDetailCurrentBag()
	booking.UpdateMushPay()

	// Nếu có MainBag thì udp lại giá cho MainBag
	// Update Lại GolfFee, Service items
	if booking.MainBags != nil && len(booking.MainBags) > 0 {
		// Chỉ có 1 main bags
		errBookingMainBag := booking.UpdateBookingMainBag()
		if errBookingMainBag != nil {
			response_message.BadRequest(c, errBookingMainBag.Error())
			return
		}
	}

	// Udp Log Tracking
	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
  Check in
*/
func (_ *CBooking) CheckIn(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.CheckInBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	booking.Hole = body.Hole

	if body.Bag != "" {
		booking.Bag = body.Bag
		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(false, true)
		if isDuplicated {
			if errDupli != nil {
				response_message.DuplicateRecord(c, errDupli.Error())
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		// Cập nhật lại info Bag
		booking.UpdateBagGolfFee()
	}

	if body.Locker != "" {
		booking.Locker = body.Locker
	}

	if body.Hole > 0 {
		booking.Hole = body.Hole
	}

	if body.Note != "" {
		booking.Note = body.Note
	}

	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())
	booking.CheckInTime = time.Now().Unix()
	booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_IN

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
  Check out: c
*/
func (_ *CBooking) CheckOut(c *gin.Context, prof models.CmsUser) {

}

/*
 Add Sub bag to Booking
*/
func (_ *CBooking) AddSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.SubBags == nil {
		response_message.BadRequest(c, "Subbags invalid nil")
		return
	}

	if len(body.SubBags) == 0 {
		response_message.BadRequest(c, "Subbags invalid empty")
		return
	}

	if booking.SubBags == nil {
		booking.SubBags = utils.ListSubBag{}
	}
	if booking.ListServiceItems == nil {
		booking.ListServiceItems = utils.ListBookingServiceItems{}
	}

	// Check lại SubBag
	// Có thể udp thêm vào hoặc remove đi
	// Check exits
	for _, v := range body.SubBags {
		if checkCheckSubBagDupli(v.BookingUid, booking) {
			// Có rồi không thêm nữa
			log.Println("AddSubBagToBooking dupli book", v.BookingUid)
		} else {
			subBooking := model_booking.Booking{}
			subBooking.Uid = v.BookingUid
			err1 := subBooking.FindFirst()
			if err1 == nil {
				//Subbag
				subBag := utils.BookingSubBag{
					BookingUid: v.BookingUid,
					GolfBag:    subBooking.Bag,
				}
				booking.SubBags = append(booking.SubBags, subBag)

				//Udp List GolfFee
				subBagGolfFee := subBooking.GetCurrentBagGolfFee()
				if booking.ListGolfFee == nil {
					booking.ListGolfFee = model_booking.ListBookingGolfFee{}
				}
				booking.ListGolfFee = append(booking.ListGolfFee, subBagGolfFee)

				//Udp lại Sub service items
				if subBooking.ListServiceItems != nil {
					booking.ListServiceItems = append(booking.ListServiceItems, subBooking.ListServiceItems...)
				}
			} else {
				log.Println("AddSubBagToBooking err1", err1.Error())
			}
		}
	}

	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	// Tính lại giá
	booking.UpdateMushPay()

	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(body, booking.Bag)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Booking Rounds: Vòng Round
 Thêm Round cho Booking
*/
func (_ *CBooking) AddRound(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddRoundBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, errors.New("Uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Handle Round Logic
	// Tính lại giá

	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}
