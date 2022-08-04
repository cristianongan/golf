package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CBooking struct{}

/// --------- Booking ----------
/*
 Tạo Booking từ TeeSheet
*/
func (cBooking *CBooking) CreateBooking(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// validate caddie_code
	var caddie models.Caddie
	var err error
	if body.CaddieCode != "" {
		caddie, err = cBooking.validateCaddie(prof.CourseUid, body.CaddieCode)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// check trạng thái Tee Time
	if body.TeeTime != "" {
		teeTime := models.TeeTimeSettings{}
		teeTime.TeeTime = body.TeeTime
		teeTime.CourseUid = body.CourseUid
		teeTime.PartnerUid = body.PartnerUid
		teeTime.DateTime = body.BookingDate
		errFind := teeTime.FindFirst()
		if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_LOCKED) {
			response_message.BadRequest(c, "Tee Time đã bị khóa")
			return
		}
		// if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_DELETED) {
		// 	response_message.BadRequest(c, "Tee Time đã bị xóa")
		// 	return
		// }
	}

	//check Booking Source with date time rule
	if body.BookingSourceId != "" {
		errorTime := cBooking.validateTimeRuleInBookingSource(body.BookingSourceId, c, body.BookingDate)
		if errorTime != nil {
			response_message.BadRequest(c, errorTime.Error())
			return
		}
	}

	teePartList := []string{"MORNING", "NOON", "NIGHT"}

	if !checkStringInArray(teePartList, body.TeePath) {
		response_message.BadRequest(c, "Tee Part not in (MORNING, NOON, NIGHT)")
		return
	}

	booking := model_booking.Booking{
		PartnerUid:        body.PartnerUid,
		CourseUid:         body.CourseUid,
		TeeType:           body.TeeType,
		TeePath:           body.TeePath,
		TeeTime:           body.TeeTime,
		TeeOffTime:        body.TeeTime,
		TurnTime:          body.TurnTime,
		RowIndex:          body.RowIndex,
		CmsUser:           prof.UserName,
		Hole:              body.Hole,
		BookingRestaurant: body.BookingRestaurant,
		BookingRetal:      body.BookingRetal,
	}

	// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

	if body.Bag != "" {
		booking.Bag = body.Bag
	}

	if body.BookingDate != "" {
		booking.BookingDate = body.BookingDate
	} else {
		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
		if errDate == nil {
			booking.BookingDate = dateDisplay
		} else {
			log.Println("booking date display err ", errDate.Error())
		}
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

	// Member Card
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
		booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)
		if memberCard.PriceCode == 1 {
			// TODO: Giá riêng không theo guest style

		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle()
		}
	} else {
		if strings.TrimSpace(body.CustomerName) != "" {
			booking.CustomerName = body.CustomerName
		} else {
			response_message.BadRequest(c, "CustomerName not empty")
			return
		}
	}

	//Agency id
	if body.AgencyId > 0 {
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst()
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)

		booking.AgencyInfo = agencyBooking
		body.GuestStyle = agency.GuestStyle
		//TODO: check giá đặc biệt của agency

	}

	// Có thông tin khách hàng
	/*
		Chọn khách hàng từ agency
	*/
	if body.CustomerUid != "" {
		//check customer
		customer := models.CustomerUser{}
		customer.Uid = body.CustomerUid
		errFindCus := customer.FindFirst()
		if errFindCus != nil || customer.Uid == "" {
			response_message.BadRequest(c, "customer"+errFindCus.Error())
			return
		}

		booking.CustomerName = customer.Name
		booking.CustomerInfo = cloneToCustomerBooking(customer)
		booking.CustomerUid = body.CustomerUid
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

	// Checkin Time
	checkInTime := time.Now().Unix()

	// GuestStyle
	if body.GuestStyle != "" {
		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			GuestStyle: body.GuestStyle,
		}
		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay()
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			return
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		// List Booking GolfFee
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(bUid, body, golfFee)
		booking.ListGolfFee = listBookingGolfFee

		// Current Bag Price Detail
		currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
		currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
		currentBagPriceDetail.UpdateAmount()
		booking.CurrentBagPrice = currentBagPriceDetail

		// MushPayInfo
		mushPayInfo := initBookingMushPayInfo(booking)
		booking.MushPayInfo = mushPayInfo

		// Rounds: Init First
		listRounds := initListRound(booking, bookingGolfFee, checkInTime)
		booking.Rounds = listRounds
	}

	// Check In Out
	if body.IsCheckIn {
		// Tạo booking check in luôn
		booking.BagStatus = constants.BAG_STATUS_IN
		booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
		booking.CheckInTime = checkInTime
	} else {
		// Tạo booking
		booking.BagStatus = constants.BAG_STATUS_INIT
		booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
	}

	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT

	// Update caddie
	if body.CaddieCode != "" {
		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

		// Udp Note
		caddieInOutNote := model_gostarter.CaddieInOutNote{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			BookingUid: booking.Uid,
			CaddieId:   booking.CaddieId,
			Type:       constants.STATUS_IN,
			Note:       "",
		}

		go addCaddieInOutNote(caddieInOutNote)
	}
	if body.CustomerBookingName != "" {
		booking.CustomerBookingName = body.CustomerBookingName
	} else {
		booking.CustomerBookingName = booking.CustomerName
	}

	if body.CustomerBookingPhone != "" {
		booking.CustomerBookingPhone = body.CustomerBookingPhone
	} else {
		booking.CustomerBookingPhone = booking.CustomerInfo.Phone
	}

	bookingCode := utils.HashCodeUuid(bookingUid.String())
	booking.BookingCode = bookingCode

	errC := booking.Create(bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, booking)
}
func (_ CBooking) validateTimeRuleInBookingSource(BookingSourceId string, c *gin.Context, BookingDate string) error {
	bookingSourceId, err := strconv.ParseInt(BookingSourceId, 10, 64)
	if err != nil {
		return err
	}
	bookingSource := model_booking.BookingSource{}
	bookingSource.Id = bookingSourceId
	errF := bookingSource.FindFirst()
	if errF != nil {
		return errors.New("BookingSource not found")
	}
	currentDInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, utils.GetCurrentDay1())
	lastDInt := currentDInt + bookingSource.NumberOfDays*24*60*60

	bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, BookingDate)

	if bookingDateInt >= currentDInt && bookingDateInt <= lastDInt {
		return nil
	}

	return errors.New("BookingDate không nằm trong ngày quy định của Booking Source")
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

	errF := booking.FindFirst()
	if errF != nil {
		// response_message.InternalServerError(c, errF.Error())
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
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
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
		BookingCode: form.BookingCode,
		AgencyId:    form.AgencyId,
	}

	list, total, err := bookingR.FindList(page, form.From, form.To)
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
	bookings.GolfBag = form.Bag
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

	var list []model_booking.Booking
	db, total, err := bookings.FindBookingListWithSelect(page)
	db.Find(&list)

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

	list, total, err := bookingR.FindBookingTeeTimeList()
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

func (_ CBooking) validateCaddie(courseUid string, caddieCode string) (models.Caddie, error) {
	caddieList := models.CaddieList{}
	caddieList.CourseUid = courseUid
	caddieList.CaddieCode = caddieCode
	caddieNew, err := caddieList.FindFirst()

	if err != nil {
		return caddieNew, err
	}

	return caddieNew, nil
}

/*
 Cập nhật booking
 Thêm Service item
*/
func (cBooking *CBooking) UpdateBooking(c *gin.Context, prof models.CmsUser) {
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

	body := request.UpdateBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate caddie_code
	var caddie models.Caddie
	var err error
	if body.CaddieCode != "" {
		caddie, err = cBooking.validateCaddie(prof.CourseUid, body.CaddieCode)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
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

	if body.LockerNo == "" {
		booking.LockerNo = body.LockerNo
	}

	if body.ReportNo == "" {
		booking.ReportNo = body.ReportNo
	}

	if body.CustomerBookingName != "" {
		booking.CustomerBookingName = body.CustomerBookingName
	}

	if body.CustomerBookingPhone != "" {
		booking.CustomerBookingPhone = body.CustomerBookingPhone
	}

	//Agency id
	if body.AgencyId > 0 {
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst()
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyId = body.AgencyId
		booking.AgencyInfo = agencyBooking
		body.GuestStyle = agency.GuestStyle
		//TODO: check giá đặc biệt của agency

	}

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
		booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)
		if memberCard.PriceCode == 1 {
			// TODO: Giá riêng không theo guest style

		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle()
		}
	} else {
		if body.CustomerName != "" {
			booking.CustomerName = body.CustomerName
		}
	}

	//Find Booking Code
	list, _ := booking.FindListWithBookingCode()
	if len(list) == 1 {
		booking.CustomerBookingName = booking.CustomerName
		booking.CustomerBookingPhone = booking.CustomerInfo.Phone
	}

	//Update service items
	booking.ListServiceItems = body.ListServiceItems

	//Update service items cho table booking_service_items
	// cBooking.UpdateBookServiceList(body.ListServiceItems)

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

	// Booking Note
	if body.NoteOfBag != "" && body.NoteOfBag != booking.NoteOfBag {
		booking.NoteOfBag = body.NoteOfBag
		go createBagsNoteNoteOfBag(booking)
	}

	if body.NoteOfBooking != "" && body.NoteOfBooking != booking.NoteOfBooking {
		booking.NoteOfBooking = body.NoteOfBooking
		go createBagsNoteNoteOfBooking(booking)
	}

	// Update caddie
	if body.CaddieCode != "" {
		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

		// Udp Note
		caddieInOutNote := model_gostarter.CaddieInOutNote{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			BookingUid: booking.Uid,
			CaddieId:   booking.CaddieId,
			Type:       constants.STATUS_IN,
			Note:       "",
		}

		go addCaddieInOutNote(caddieInOutNote)
	}

	// Udp Log Tracking
	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
 update book service list
*/

func (_ *CBooking) UpdateBookServiceList(serviceList model_booking.ListBookingServiceItems) {
	for _, v := range serviceList {
		bookingServiceItem := model_booking.BookingServiceItem{
			BookingUid: v.BookingUid,
			GroupCode:  v.GroupCode,
			Bag:        v.Bag,
		}

		errFind := bookingServiceItem.FindFirst()
		bookingServiceItemUpdate := model_booking.BookingServiceItem{
			BookingUid:    v.BookingUid,
			GroupCode:     v.GroupCode,
			ServiceId:     v.ServiceId,
			PlayerName:    v.PlayerName,
			Bag:           v.Bag,
			Type:          v.Type,
			Order:         v.Order,
			Name:          v.Name,
			Quality:       v.Quality,
			UnitPrice:     v.UnitPrice,
			DiscountType:  v.DiscountType,
			DiscountValue: v.DiscountValue,
			Amount:        v.Amount,
			Input:         v.Input,
		}
		if errFind == nil {
			bookingServiceItemUpdate.ModelId = bookingServiceItem.ModelId
			bookingServiceItemUpdate.Update()
		} else {
			bookingServiceItem.Create()
		}
	}
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
		booking.LockerNo = body.Locker
	}

	if body.Hole > 0 {
		booking.Hole = body.Hole
	}

	if body.GuestStyle != "" {
		// Tính giá
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
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

		// List Booking GolfFee
		bodyCreate := request.CreateBookingBody{
			Hole:         booking.Hole,
			CustomerName: booking.CustomerName,
			Bag:          booking.Bag,
		}
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(booking.Uid, bodyCreate, golfFee)
		booking.ListGolfFee = listBookingGolfFee

		// Current Bag Price Detail
		currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
		currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
		currentBagPriceDetail.UpdateAmount()
		booking.CurrentBagPrice = currentBagPriceDetail

		// MushPayInfo
		mushPayInfo := initBookingMushPayInfo(booking)
		booking.MushPayInfo = mushPayInfo

		// Rounds: Init First
		checkInTime := time.Now().Unix()
		listRounds := initListRound(booking, bookingGolfFee, checkInTime)
		booking.Rounds = listRounds
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CheckInTime = time.Now().Unix()
	booking.BagStatus = constants.BAG_STATUS_IN

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
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
		booking.ListServiceItems = model_booking.ListBookingServiceItems{}
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
					PlayerName: subBooking.CustomerName,
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

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	// Tính lại giá
	booking.UpdateMushPay()

	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(body, booking.Bag, booking.CustomerName)
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
 Edit Sub bag to Booking
*/
func (_ *CBooking) EditSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.EditSubBagToBooking{}
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
		response_message.BadRequest(c, "subbag empty")
		return
	}

	// Khi có xoá note thì Udp lại giá
	isUpdPrice := false

	for i, v := range body.SubBags {
		// Get Booking Detail
		subBooking := model_booking.Booking{}
		subBooking.Uid = v.BookingUid
		errFSB := subBooking.FindFirst()

		if errFSB != nil {
			log.Println("EditSubBagToBooking errFSB", errF.Error())
		}

		if v.IsOut == true {
			//remove di
			// Remove main bag
			subBooking.MainBags = utils.ListSubBag{}
			errSBUdp := subBooking.Update()
			if errSBUdp != nil {
				log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
			}
			isUpdPrice = true
			// Udp sub bag for booking
			// remove sub bag
			subBags := utils.ListSubBag{}
			for _, v1 := range booking.SubBags {
				if v1.BookingUid != v.BookingUid {
					subBags = append(subBags, v1)
				}
			}
			booking.SubBags = subBags
			// remove list golf fee
			listGolfFees := model_booking.ListBookingGolfFee{}
			for _, v1 := range booking.ListGolfFee {
				if v1.BookingUid != v.BookingUid {
					listGolfFees = append(listGolfFees, v1)
				}
			}
			booking.ListGolfFee = listGolfFees
			// remove list service items
			listServiceItems := model_booking.ListBookingServiceItems{}
			for _, v1 := range booking.ListServiceItems {
				if v1.BookingUid != v.BookingUid {
					listServiceItems = append(listServiceItems, v1)
				}
			}
			booking.ListServiceItems = listServiceItems
		} else {
			isCanUdp := false

			if subBooking.SubBagNote != v.SubBagNote {
				subBooking.SubBagNote = v.SubBagNote
				isCanUdp = true
			}
			if subBooking.CustomerName != v.PlayerName {
				subBooking.CustomerName = v.PlayerName
				if i < len(booking.SubBags) {
					booking.SubBags[i].PlayerName = v.PlayerName
				}
				isCanUdp = true
			}

			if isCanUdp {
				errSBUdp := subBooking.Update()
				if errSBUdp != nil {
					log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
				}
			}
		}
	}

	if isUpdPrice {
		booking.UpdateMushPay()
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

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
//func (_ *CBooking) AddRound(c *gin.Context, prof models.CmsUser) {
//	// Body request
//	body := request.AddRoundBody{}
//	if bindErr := c.ShouldBind(&body); bindErr != nil {
//		response_message.BadRequest(c, bindErr.Error())
//		return
//	}
//
//	if body.BookingUid == "" {
//		response_message.BadRequest(c, errors.New("Uid not valid").Error())
//		return
//	}
//
//	booking := model_booking.Booking{}
//	booking.Uid = body.BookingUid
//	errF := booking.FindFirst()
//	if errF != nil {
//		response_message.InternalServerError(c, errF.Error())
//		return
//	}
//
//	// Handle Round Logic
//	// Tính lại giá
//
//	booking.CmsUser = prof.UserName
//	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
//
//	errUdp := booking.Update()
//	if errUdp != nil {
//		response_message.InternalServerError(c, errUdp.Error())
//		return
//	}
//
//	okResponse(c, booking)
//}

/*
 Danh sách booking cho Add sub bag
*/
func (_ *CBooking) GetListBookingForAddSubBag(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" || form.CourseUid == "" || form.Bag == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		BagStatus:  constants.BAG_STATUS_IN,
	}
	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errDate == nil {
		bookingR.BookingDate = dateDisplay
	} else {
		log.Println("GetListBookingForAddSubBag booking date display err ", errDate.Error())
	}

	list, errF := bookingR.FindListForSubBag()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	listResponse := []model_booking.BookingForSubBag{}

	if len(list) == 0 {
		okResponse(c, listResponse)
		return
	}

	for _, v := range list {
		if v.Bag != form.Bag {
			listResponse = append(listResponse, v)
		}
	}

	okResponse(c, listResponse)
}

func (_ *CBooking) GetSubBagDetail(c *gin.Context, prof models.CmsUser) {
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

	list := []model_booking.Booking{}

	if booking.SubBags == nil || len(booking.SubBags) == 0 {
		okResponse(c, list)
		return
	}

	for _, v := range booking.SubBags {
		bookingTemp := model_booking.Booking{}
		bookingTemp.Uid = v.BookingUid
		errFind := bookingTemp.FindFirst()
		if errFind != nil {
			log.Println("GetListSubBagDetail err", errFind.Error())
		}
		list = append(list, bookingTemp)
	}

	okResponse(c, list)
}

/*
	Other Paid
*/
func (_ *CBooking) AddOtherPaid(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddOtherPaidBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, errors.New("Uid not valid").Error())
		return
	}

	if body.OtherPaids == nil || len(body.OtherPaids) == 0 {
		response_message.BadRequest(c, errors.New("other paid empty").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// list service items
	// Remove cái cũ
	listServiceItems := model_booking.ListBookingServiceItems{}
	for _, v := range booking.ListServiceItems {
		if v.Type != constants.BOOKING_OTHER_FEE {
			listServiceItems = append(listServiceItems, v)
		}
	}

	// add cái mới
	for _, v := range body.OtherPaids {
		serviceItem := model_booking.BookingServiceItem{
			Type:       constants.BOOKING_OTHER_FEE,
			Amount:     v.Amount,
			Name:       v.Reason,
			PlayerName: booking.CustomerName,
			Bag:        booking.Bag,
		}
		listServiceItems = append(listServiceItems, serviceItem)
	}

	booking.ListServiceItems = listServiceItems
	booking.UpdateMushPay()

	booking.OtherPaids = body.OtherPaids

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
	Cancel Booking
	- check chưa check-in mới cancel dc
*/
func (_ *CBooking) CancelBooking(c *gin.Context, prof models.CmsUser) {
	body := request.CancelBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, "Booking Uid not empty")
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_INIT {
		response_message.InternalServerError(c, "This booking did check in")
		return
	}

	booking.BagStatus = constants.BAG_STATUS_CANCEL
	booking.CancelNote = body.Note
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
	Moving Booking
	- check chưa check-in mới moving dc
*/
func (_ *CBooking) MovingBooking(c *gin.Context, prof models.CmsUser) {
	body := request.MovingBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if len(body.BookUidList) == 0 {
		response_message.BadRequest(c, "Booking invalid empty")
		return
	}

	if len(body.BookUidList) > 4 {
		response_message.BadRequest(c, "The number of Bookings cannot exceed 4")
		return
	}

	for _, BookingUid := range body.BookUidList {
		if BookingUid == "" {
			response_message.BadRequest(c, "Booking Uid not empty")
			return
		}

		booking := model_booking.Booking{}
		booking.Uid = BookingUid
		errF := booking.FindFirst()
		if errF != nil {
			response_message.InternalServerError(c, errF.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_INIT {
			response_message.InternalServerError(c, booking.Uid+" did check in")
			return
		}
		if body.TeeTime != "" {
			booking.TeeTime = body.TeeTime
		}
		if body.TeeType != "" {
			booking.TeeType = body.TeeType
		}
		if body.BookingDate != "" {
			booking.BookingDate = body.BookingDate
		}

		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(true, false)
		if isDuplicated {
			if errDupli != nil {
				response_message.DuplicateRecord(c, errDupli.Error())
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		if body.Hole != 0 {
			booking.Hole = body.Hole
		}

		errUdp := booking.Update()
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}

	okRes(c)
}

func (_ CBooking) validateBooking(bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(); err != nil {
		return booking, err
	}

	return booking, nil
}

func (cBooking *CBooking) Checkout(c *gin.Context, prof models.CmsUser) {
	body := request.CheckoutBody{}
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cBooking.validateBooking(body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if booking.Bag != body.GolfBag {
		response_message.InternalServerError(c, "Booking uid and golf bag do not match")
		return
	}

	booking.BagStatus = constants.BAG_STATUS_OUT
	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (cBooking *CBooking) CreateBatchBooking(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for _, body := range bodyRequest.BookingList {
		// validate caddie_code
		var caddie models.Caddie
		var err error
		if body.CaddieCode != "" {
			caddie, err = cBooking.validateCaddie(prof.CourseUid, body.CaddieCode)
			if err != nil {
				response_message.InternalServerError(c, err.Error())
				return
			}
		}

		// check trạng thái Tee Time
		if body.TeeTime != "" {
			teeTime := models.TeeTimeSettings{}
			teeTime.TeeTime = body.TeeTime
			teeTime.CourseUid = body.CourseUid
			teeTime.PartnerUid = body.PartnerUid
			teeTime.DateTime = body.BookingDate
			errFind := teeTime.FindFirst()
			if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_LOCKED) {
				response_message.BadRequest(c, "Tee Time đã bị khóa")
				return
			}
			// if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_DELETED) {
			// 	response_message.BadRequest(c, "Tee Time đã bị xóa")
			// 	return
			// }
		}

		teePartList := []string{"MORNING", "NOON", "NIGHT"}

		if !checkStringInArray(teePartList, body.TeePath) {
			response_message.BadRequest(c, "Tee Part not in (MORNING, NOON, NIGHT)")
			return
		}

		booking := model_booking.Booking{
			PartnerUid:           body.PartnerUid,
			CourseUid:            body.CourseUid,
			TeeType:              body.TeeType,
			TeePath:              body.TeePath,
			TeeTime:              body.TeeTime,
			TeeOffTime:           body.TeeTime,
			TurnTime:             body.TurnTime,
			RowIndex:             body.RowIndex,
			CmsUser:              prof.UserName,
			Hole:                 body.Hole,
			BookingCode:          bookingCode,
			BookingRestaurant:    body.BookingRestaurant,
			BookingRetal:         body.BookingRetal,
			CustomerBookingName:  body.CustomerBookingName,
			CustomerBookingPhone: body.CustomerBookingPhone,
		}

		// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

		if body.Bag != "" {
			booking.Bag = body.Bag
		}

		if body.BookingDate != "" {
			booking.BookingDate = body.BookingDate
		} else {
			dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
			if errDate == nil {
				booking.BookingDate = dateDisplay
			} else {
				log.Println("booking date display err ", errDate.Error())
			}
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

		// Member Card
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
			booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)
			if memberCard.PriceCode == 1 {
				// TODO: Giá riêng không theo guest style
			} else {
				// Lấy theo GuestStyle
				body.GuestStyle = memberCard.GetGuestStyle()
			}
		} else {
			if strings.TrimSpace(body.CustomerName) != "" {
				booking.CustomerName = body.CustomerName
			} else {
				response_message.BadRequest(c, "CustomerName not empty")
				return
			}
		}

		//Agency id
		if body.AgencyId > 0 {
			agency := models.Agency{}
			agency.Id = body.AgencyId
			errFindAgency := agency.FindFirst()
			if errFindAgency != nil || agency.Id == 0 {
				response_message.BadRequest(c, "agency"+errFindAgency.Error())
				return
			}

			agencyBooking := cloneToAgencyBooking(agency)

			booking.AgencyInfo = agencyBooking
			body.GuestStyle = agency.GuestStyle
			//TODO: check giá đặc biệt của agency

		}

		// Có thông tin khách hàng
		/*
			Chọn khách hàng từ agency
		*/
		if body.CustomerUid != "" {
			//check customer
			customer := models.CustomerUser{}
			customer.Uid = body.CustomerUid
			errFindCus := customer.FindFirst()
			if errFindCus != nil || customer.Uid == "" {
				response_message.BadRequest(c, "customer"+errFindCus.Error())
				return
			}

			booking.CustomerName = customer.Name
			booking.CustomerInfo = cloneToCustomerBooking(customer)
			booking.CustomerUid = body.CustomerUid
		}

		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

		// Booking Uid
		bookingUid := uuid.New()
		bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

		// Checkin Time
		checkInTime := time.Now().Unix()

		// GuestStyle
		if body.GuestStyle != "" {
			//Guest style
			golfFeeModel := models.GolfFee{
				PartnerUid: body.PartnerUid,
				CourseUid:  body.CourseUid,
				GuestStyle: body.GuestStyle,
			}
			// Lấy phí bởi Guest style với ngày tạo
			golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay()
			if errFindGF != nil {
				response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
				return
			}
			booking.GuestStyle = body.GuestStyle
			booking.GuestStyleName = golfFee.GuestStyleName

			// List Booking GolfFee
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(bUid, body, golfFee)
			booking.ListGolfFee = listBookingGolfFee

			// Current Bag Price Detail
			currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
			currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
			currentBagPriceDetail.UpdateAmount()
			booking.CurrentBagPrice = currentBagPriceDetail

			// MushPayInfo
			mushPayInfo := initBookingMushPayInfo(booking)
			booking.MushPayInfo = mushPayInfo

			// Rounds: Init First
			listRounds := initListRound(booking, bookingGolfFee, checkInTime)
			booking.Rounds = listRounds
		}

		// Check In Out
		if body.IsCheckIn {
			// Tạo booking check in luôn
			booking.BagStatus = constants.BAG_STATUS_IN
			booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
			booking.CheckInTime = checkInTime
		} else {
			// Tạo booking
			booking.BagStatus = constants.BAG_STATUS_INIT
			booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
		}

		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT

		// Update caddie
		if body.CaddieCode != "" {
			booking.CaddieId = caddie.Id
			booking.CaddieInfo = cloneToCaddieBooking(caddie)
			booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

			// Udp Note
			caddieInOutNote := model_gostarter.CaddieInOutNote{
				PartnerUid: prof.PartnerUid,
				CourseUid:  prof.CourseUid,
				BookingUid: booking.Uid,
				CaddieId:   booking.CaddieId,
				Type:       constants.STATUS_IN,
				Note:       "",
			}

			go addCaddieInOutNote(caddieInOutNote)
		}

		errC := booking.Create(bUid)

		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}

	}
	okRes(c)
}
