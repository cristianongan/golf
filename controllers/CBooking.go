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
 Tạo Booking
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
	}

	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errDate == nil {
		booking.CreatedDate = dateDisplay
	} else {
		log.Println("booking date display err ", errDate.Error())
	}

	if booking.IsDuplicated() {
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

	if body.Bag != "" {
		booking.Bag = body.Bag
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

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
  Check out
*/

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

	booking.SubBags = body.SubBags
	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	booking.UpdateMushPay()

	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(body)
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

	booking.CmsUser = body.CmsUser
	booking.CmsUserLog = getBookingCmsUserLog(body.CmsUser, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}
