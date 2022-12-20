package controllers

import (
	"errors"
	"log"
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
func (_ *CBooking) GetListAgencyCancelBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.BookingDate == "" {
		response_message.BadRequest(c, errors.New("Chưa chọn ngày").Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	booking := model_booking.Booking{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.BookingDate = form.BookingDate
	booking.BookingCode = form.BookingCode

	list, total, err := booking.FindAgencyCancelBooking(db, page)

	res := response.PageResponse{}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res = response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

/*
Get booking Detail With Uid
*/
func (_ *CBooking) GetBookingDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.PartnerUid != prof.PartnerUid || booking.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
		return
	}

	bagDetail := getBagDetailFromBooking(db, booking)
	okResponse(c, bagDetail)
}

/*
Get booking payment
*/
func (_ *CBooking) GetBookingPaymentDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	bookingR := model_booking.Booking{}
	bookingR.Uid = bookingIdStr
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.PartnerUid != prof.PartnerUid || booking.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
		return
	}

	bagDetail := getBagDetailFromBooking(db, booking)

	// Get List Round Of Sub Bag
	listRoundOfSub := []model_booking.RoundOfBag{}
	if len(booking.SubBags) > 0 {
		res := GetGolfFeeInfoOfBag(c, booking)
		listRoundOfSub = res.ListRoundOfSubBag
	}

	res := model_booking.PaymentOfBag{
		BagDetail:         bagDetail,
		ListRoundOfSubBag: listRoundOfSub,
	}

	okResponse(c, res)
}

/*
Get booking by Bag
Get Booking Bag trong ngày
*/
func (_ *CBooking) GetBookingByBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	bagDetail := getBagDetailFromBooking(db, booking)
	okResponse(c, bagDetail)
}

func (_ *CBooking) GetBookingFeeOfBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	// Get List Round Of Main Bag
	mainPaidRound1 := false
	mainPaidRound2 := false
	mainCheckOutTime := int64(0)

	// Tính giá của khi có main bag
	if len(booking.MainBags) > 0 {
		mainBook := model_booking.Booking{
			CourseUid:   booking.CourseUid,
			PartnerUid:  booking.PartnerUid,
			Bag:         booking.MainBags[0].GolfBag,
			BookingDate: booking.BookingDate,
		}
		errFMB := mainBook.FindFirst(db)
		if errFMB != nil {
			log.Println("UpdateMushPay-"+booking.Bag+"-Find Main Bag", errFMB.Error())
		}
		mainCheckOutTime = mainBook.CheckOutTime
		mainPaidRound1 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND) > -1
		mainPaidRound2 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS) > -1
	}
	listRoundOfMain := []models.RoundPaidByMainBag{}
	if booking.BillCode != "" {
		round := models.Round{BillCode: booking.BillCode}
		listRound, _ := round.FindAllRoundPaidByMain(db)
		listRoundOfMain = listRound

		if mainCheckOutTime > 0 {
			for index, round := range listRoundOfMain {
				if round.Index == 1 && mainPaidRound1 {
					listRoundOfMain[index].IsPaid = true
				}
				if round.Index == 2 && mainPaidRound2 {
					listRoundOfMain[index].IsPaid = true
				}
			}
		}
	}

	// Get List Service Item
	listServices := booking.FindServiceItemsWithPaidInfo(db)

	// Get List Round Of Sub Bag
	listRoundOfSub := []model_booking.RoundOfBag{}
	if len(booking.SubBags) > 0 {
		res := GetGolfFeeInfoOfBag(c, booking)
		listRoundOfSub = res.ListRoundOfSubBag
	}

	feeResponse := model_booking.BookingFeeOfBag{
		AgencyPaid:        booking.AgencyPaid,
		SubBags:           booking.SubBags,
		MushPayInfo:       booking.MushPayInfo,
		ListServiceItems:  listServices,
		ListRoundOfSubBag: listRoundOfSub,
		Rounds:            listRoundOfMain,
	}

	okResponse(c, feeResponse)
}

/*
Danh sách booking
*/
func (_ *CBooking) GetListBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		PartnerUid:   form.PartnerUid,
		CourseUid:    form.CourseUid,
		BookingDate:  form.BookingDate,
		BookingCode:  form.BookingCode,
		AgencyId:     form.AgencyId,
		BagStatus:    form.BagStatus,
		CustomerName: form.PlayerName,
		Bag:          form.Bag,
		FlightId:     form.FlightId,
	}

	list, total, err := bookingR.FindList(db, page, form.From, form.To, form.AgencyType)
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
Danh sách booking với select
*/
func (_ *CBooking) GetListBookingWithSelect(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
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

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = form.BookingDate
	bookings.GolfBag = form.GolfBag
	bookings.BookingCode = form.BookingCode
	bookings.InitType = form.InitType
	bookings.IsAgency = form.IsAgency
	bookings.AgencyId = form.AgencyId
	bookings.Status = form.Status
	bookings.FromDate = form.FromDate
	bookings.ToDate = form.ToDate
	bookings.IsToday = form.IsToday
	bookings.BookingUid = form.BookingUid
	bookings.IsFlight = form.IsFlight
	bookings.BagStatus = form.BagStatus
	bookings.HaveBag = form.HaveBag
	bookings.HasBuggy = form.HasBuggy
	bookings.HasCaddie = form.HasCaddie
	bookings.CaddieCode = form.CaddieCode
	bookings.HasBookCaddie = form.HasBookCaddie
	bookings.CustomerName = form.PlayerName
	bookings.HasCaddieInOut = form.HasCaddieInOut
	bookings.FlightId = form.FlightId
	bookings.TeeType = form.TeeType
	bookings.CourseType = form.CourseType
	bookings.IsCheckIn = form.IsCheckIn
	bookings.GuestStyleName = form.GuestStyleName
	bookings.PlayerOrBag = form.PlayerOrBag
	bookings.CustomerUid = form.CustomerUid
	bookings.CustomerType = form.CustomerType
	bookings.BuggyCode = form.BuggyCode
	bookings.GuestStyle = form.GuestStyle
	bookings.CaddieName = form.CaddieName

	db, total, err := bookings.FindBookingListWithSelect(db, page, form.IsGroupBillCode)

	if form.HasCaddieInOut != "" {
		db = db.Preload("CaddieBuggyInOut")
	}

	res := response.PageResponse{}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)
	res = response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

/*
Danh sách booking với thông tin flight
*/
func (_ *CBooking) GetListBookingWithFightInfo(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
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

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = form.BookingDate
	bookings.GolfBag = form.GolfBag
	bookings.BookingCode = form.BookingCode
	bookings.InitType = form.InitType
	bookings.IsAgency = form.IsAgency
	bookings.AgencyId = form.AgencyId
	bookings.Status = form.Status
	bookings.FromDate = form.FromDate
	bookings.ToDate = form.ToDate
	bookings.IsToday = form.IsToday
	bookings.BookingUid = form.BookingUid
	bookings.IsFlight = form.IsFlight
	bookings.BagStatus = form.BagStatus
	bookings.HaveBag = form.HaveBag
	bookings.CaddieCode = form.CaddieCode
	bookings.HasBookCaddie = form.HasBookCaddie
	bookings.CustomerName = form.PlayerName
	bookings.HasFlightInfo = form.HasFlightInfo

	db, total, err := bookings.FindBookingListWithSelect(db, page, false)
	res := response.PageResponse{}
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.FlyInfoResponse
	db = db.Joins("JOIN flights ON flights.id = bookings.flight_id")
	db = db.Select("bookings.*, flights.tee_off as tee_off_flight," +
		"flights.tee as tee_flight, flights.date_display as date_display_flight," +
		"flights.group_name as group_name_flight")
	db.Find(&list)
	res = response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

/*
TODO: Update lại api này
Danh sách Booking với thông tin service item
*/

func (_ *CBooking) GetListBookingWithListServiceItems(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithListServiceItems{}
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

	booking := model_booking.Booking{}
	param := model_booking.GetListBookingWithListServiceItems{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		FromDate:    form.FromDate,
		ToDate:      form.ToDate,
		ServiceType: form.Type,
		GolfBag:     form.GolfBag,
		PlayerName:  form.PlayerName,
	}
	list, total, err := booking.FindListServiceItems(db, param, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

/*
Danh sách booking tee time
*/
func (_ *CBooking) GetListBookingTeeTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingTeeTimeForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
		TeeTime:     form.TeeTime,
	}

	list, total, err := bookingR.FindBookingTeeTimeList(db)
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
Lấy các slot còn trống của Tee Time
*/
func (cBooking *CBooking) GetSlotRemainInTeeTime(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.TeeTime == "" {
		response_message.BadRequestFreeMessage(c, "TeeTime empty!")
		return
	}

	if form.TeeType == "" {
		response_message.BadRequestFreeMessage(c, "TeeType empty!")
		return
	}

	if form.CourseType == "" {
		response_message.BadRequestFreeMessage(c, "CourseType empty!")
		return
	}

	if form.BookingDate == "" {
		response_message.BadRequestFreeMessage(c, "BookingDate empty!")
		return
	}

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = form.BookingDate
	bookings.TeeTime = form.TeeTime
	bookings.TeeType = form.TeeType
	bookings.CourseType = form.CourseType

	_, total, _ := bookings.FindAllBookingList(db)

	res := map[string]interface{}{
		"total": constants.SLOT_TEE_TIME - total,
	}

	okResponse(c, res)
}

/*
Get Bag Not Check Out
*/
func (cBooking *CBooking) GetBagNotCheckOut(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	now := time.Now().Format(constants.DATE_FORMAT_1)
	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = now
	bookings.IsCheckIn = "1"

	db, total, err := bookings.FindAllBookingList(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)
	res := map[string]interface{}{
		"total": total,
	}

	okResponse(c, res)
}
