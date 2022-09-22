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

	booking := cBooking.CreateBookingCommon(body, c, prof)
	if booking == nil {
		return
	}

	okResponse(c, booking)
}

func (cBooking CBooking) CreateBookingCommon(body request.CreateBookingBody, c *gin.Context, prof models.CmsUser) *model_booking.Booking {

	// validate caddie_code
	var caddie models.Caddie
	var err error
	if body.CaddieCode != "" {
		caddie, err = cBooking.validateCaddie(prof.CourseUid, body.CaddieCode)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return nil
		}
	}

	// validate trường hợp đóng tee 1
	teeList := []string{constants.TEE_TYPE_1, constants.TEE_TYPE_1A, constants.TEE_TYPE_1B, constants.TEE_TYPE_1C}
	if utils.Contains(teeList, body.TeeType) {
		cBookingSetting := CBookingSetting{}
		if errors := cBookingSetting.ValidateClose1ST(body.BookingDate, body.PartnerUid, body.CourseUid); errors != nil {
			response_message.InternalServerError(c, errors.Error())
			return nil
		}
	}

	// check trạng thái Tee Time
	if body.TeeTime != "" {
		teeTime := models.LockTeeTime{}
		teeTime.TeeTime = body.TeeTime
		teeTime.TeeType = body.TeeType
		teeTime.CourseUid = body.CourseUid
		teeTime.PartnerUid = body.PartnerUid
		teeTime.DateTime = body.BookingDate
		errFind := teeTime.FindFirst()
		if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_LOCKED) {
			response_message.BadRequest(c, "Tee Time đã bị khóa")
			return nil
		}
	}

	//check Booking Source with date time rule
	if body.BookingSourceId != "" {
		bookingSourceId, err := strconv.ParseInt(body.BookingSourceId, 10, 64)
		if err != nil {
			response_message.BadRequest(c, "BookingSource không tồn tại")
		}
		bookingSource := model_booking.BookingSource{}
		bookingSource.Id = bookingSourceId
		errorTime := bookingSource.ValidateTimeRuleInBookingSource(body.BookingDate, body.TeePath)
		if errorTime != nil {
			response_message.BadRequest(c, errorTime.Error())
			return nil
		}
	}

	if !body.IsCheckIn {
		teePartList := []string{"MORNING", "NOON", "NIGHT"}

		if !checkStringInArray(teePartList, body.TeePath) {
			response_message.BadRequest(c, "Tee Part not in (MORNING, NOON, NIGHT)")
			return nil
		}
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
		HoleBooking:       body.Hole,
		BookingRestaurant: body.BookingRestaurant,
		BookingRetal:      body.BookingRetal,
		BookingCode:       body.BookingCode,
		CourseType:        body.CourseType,
	}

	// Check Guest of member, check member có còn slot đi cùng không
	var memberCard models.MemberCard
	if body.MemberUidOfGuest != "" && body.GuestStyle != "" {
		var errCheckMember error
		customerName := ""
		errCheckMember, memberCard, customerName = handleCheckMemberCardOfGuest(body.MemberUidOfGuest, body.GuestStyle)
		if errCheckMember != nil {
			response_message.InternalServerError(c, errCheckMember.Error())
			return nil
		} else {
			booking.MemberUidOfGuest = body.MemberUidOfGuest
			booking.MemberNameOfGuest = customerName
		}
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
			return nil
		}
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return nil
	}

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

	// Checkin Time
	checkInTime := time.Now().Unix()

	// Member Card
	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return nil
		}

		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return nil
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner()
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return nil
		}

		booking.MemberCardUid = body.MemberCardUid
		booking.CardId = memberCard.CardId
		booking.CustomerName = owner.Name
		booking.CustomerUid = owner.Uid
		booking.CustomerType = owner.Type
		booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)

		if memberCard.PriceCode == 1 && memberCard.IsValidTimePrecial() {
			// Check member card với giá riêng và time được áp dụng
			param := request.GolfFeeGuestyleParam{
				Uid:          bUid,
				Rate:         course.RateGolfFee,
				Bag:          body.Bag,
				CustomerName: body.CustomerName,
				Hole:         body.Hole,
				CaddieFee:    memberCard.CaddieFee,
				BuggyFee:     memberCard.BuggyFee,
				GreenFee:     memberCard.GreenFee,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
			initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, checkInTime)
			if body.GuestStyle != "" {
				body.GuestStyle = ""
			}
		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle()
		}
	} else {
		if strings.TrimSpace(body.CustomerName) != "" {
			booking.CustomerName = body.CustomerName
		} else {
			response_message.BadRequest(c, "CustomerName not empty")
			return nil
		}
	}

	//Agency id
	if body.AgencyId > 0 {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return nil
		}

		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst()
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return nil
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyInfo = agencyBooking
		booking.AgencyId = body.AgencyId

		agencySpecialPrice := models.AgencySpecialPrice{
			AgencyId: agency.Id,
		}
		errFSP := agencySpecialPrice.FindFirst()
		if errFSP == nil && agencySpecialPrice.Id > 0 {
			// Tính lại giá
			// List Booking GolfFee
			// TODO:  Check số lượt dc config ở agency
			param := request.GolfFeeGuestyleParam{
				Uid:          bUid,
				Rate:         course.RateGolfFee,
				Bag:          body.Bag,
				CustomerName: body.CustomerName,
				Hole:         body.Hole,
				CaddieFee:    agencySpecialPrice.CaddieFee,
				BuggyFee:     agencySpecialPrice.BuggyFee,
				GreenFee:     agencySpecialPrice.GreenFee,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
			initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, checkInTime)

		} else {
			body.GuestStyle = agency.GuestStyle
		}
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
			return nil
		}

		booking.CustomerName = customer.Name
		booking.CustomerType = customer.Type
		booking.CustomerInfo = cloneToCustomerBooking(customer)
		booking.CustomerUid = body.CustomerUid
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

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
			return nil
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		// List Booking GolfFee
		param := request.GolfFeeGuestyleParam{
			Uid:          bUid,
			Bag:          body.Bag,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
		}

		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
		initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, checkInTime)
	}

	// Check In Out
	if body.IsCheckIn {
		// Tạo booking check in luôn
		booking.BagStatus = constants.BAG_STATUS_WAITING
		booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
		booking.CheckInTime = checkInTime
	} else {
		// Tạo booking
		booking.BagStatus = constants.BAG_STATUS_BOOKING
		booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
	}

	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT

	// Update caddie
	if body.CaddieCode != "" {
		cBooking.UpdateBookingCaddieCommon(body.PartnerUid, body.CourseUid, &booking, caddie)
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

	if body.BookingCode == "" {
		bookingCode := utils.HashCodeUuid(bookingUid.String())
		booking.BookingCode = bookingCode
	}
	booking.BillCode = utils.HashCodeUuid(bookingUid.String())

	errC := booking.Create(bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return nil
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(memberCard)
	}

	if body.TeeTime != "" {
		cLockTeeTime := CLockTeeTime{}
		lockTurn := request.CreateLockTurn{
			BookingDate: body.BookingDate,
			CourseUid:   body.CourseUid,
			PartnerUid:  body.PartnerUid,
			TeeTime:     body.TeeTime,
			TeeType:     body.TeeType,
		}
		cLockTeeTime.LockTurn(lockTurn, c, prof)
	}

	if body.IsCheckIn && booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
	}

	return &booking
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

	res := getBagDetailFromBooking(booking)

	okResponse(c, res)
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

	list, total, err := bookingR.FindList(page, form.From, form.To, form.AgencyType)
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
	bookings.HasCaddieInOut = form.HasCaddieInOut
	bookings.FlightId = form.FlightId
	bookings.TeeType = form.TeeType
	bookings.IsCheckIn = form.IsCheckIn

	db, total, err := bookings.FindBookingListWithSelect(page)

	if form.HasCaddieInOut != "" {
		db = db.Preload("CaddieInOut")
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

	db, total, err := bookings.FindBookingListWithSelect(page)
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
	list, total, err := booking.FindListServiceItems(param, page)

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

	if body.GuestStyle != "" {
		booking.GuestStyle = body.GuestStyle
	}

	//Upd Main Pay for Sub
	if body.MainBagPay != nil {
		booking.MainBagPay = body.MainBagPay
	}

	if body.LockerNo == "" {
		booking.LockerNo = body.LockerNo
		go createLocker(booking)
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
		booking.CustomerType = owner.Type
		booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)
		if memberCard.PriceCode == 1 && memberCard.IsValidTimePrecial() {
			course := models.Course{}
			course.Uid = body.CourseUid
			errCourse := course.FindFirst()
			if errCourse != nil {
				response_message.BadRequest(c, errCourse.Error())
				return
			}
			param := request.GolfFeeGuestyleParam{
				Uid:          booking.Uid,
				Rate:         course.RateGolfFee,
				Bag:          booking.Bag,
				CustomerName: body.CustomerName,
				CaddieFee:    memberCard.CaddieFee,
				BuggyFee:     memberCard.BuggyFee,
				GreenFee:     memberCard.GreenFee,
				Hole:         booking.Hole,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
			initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
			if body.GuestStyle != "" {
				body.GuestStyle = ""
			}
		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle()
		}
	} else {
		if body.CustomerName != "" {
			booking.CustomerName = body.CustomerName
		}
	}

	//Agency id
	if body.AgencyId > 0 {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return
		}

		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst()
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyInfo = agencyBooking
		booking.AgencyId = body.AgencyId

		agencySpecialPrice := models.AgencySpecialPrice{
			AgencyId: agency.Id,
		}
		errFSP := agencySpecialPrice.FindFirst()
		if errFSP == nil && agencySpecialPrice.Id > 0 {
			// Tính lại giá
			// List Booking GolfFee
			// TODO:  Check số lượt dc config ở agency
			param := request.GolfFeeGuestyleParam{
				Uid:          booking.Uid,
				Rate:         course.RateGolfFee,
				Bag:          booking.Bag,
				CustomerName: body.CustomerName,
				Hole:         booking.Hole,
				CaddieFee:    agencySpecialPrice.CaddieFee,
				BuggyFee:     agencySpecialPrice.BuggyFee,
				GreenFee:     agencySpecialPrice.GreenFee,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
			initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)

		} else {
			body.GuestStyle = agency.GuestStyle
		}
	}
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
		param := request.GolfFeeGuestyleParam{
			Uid:          booking.Uid,
			Bag:          body.Bag,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
		}
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
		initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
	}
	//Find Booking Code
	list, _ := booking.FindListWithBookingCode()
	if len(list) == 1 {
		booking.CustomerBookingName = booking.CustomerName
		booking.CustomerBookingPhone = booking.CustomerInfo.Phone
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
		cBooking.UpdateBookingCaddieCommon(body.PartnerUid, body.CourseUid, &booking, caddie)
	}

	// Tính lại giá
	updatePriceWithServiceItem(booking, prof)

	// Get lai booking
	bookLast := model_booking.Booking{}
	bookLast.Uid = booking.Uid
	bookLast.FindFirst()

	res := getBagDetailFromBooking(bookLast)

	okResponse(c, res)
}

/*
Update booking caddie when create booking or update
*/
func (_ *CBooking) UpdateBookingCaddieCommon(PartnerUid string, CourseUid string, booking *model_booking.Booking, caddie models.Caddie) {
	booking.CaddieId = caddie.Id

	// if booking.CheckDuplicatedCaddieInTeeTime() {
	// 	response_message.InternalServerError(c, "Caddie không được trùng trong cùng TeeTime")
	// 	return
	// }

	booking.CaddieInfo = cloneToCaddieBooking(caddie)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

	// Set has_book_caddie
	if booking.BagStatus == constants.BAG_STATUS_BOOKING {
		booking.HasBookCaddie = true
	}

	// udp trạng thái caddie sang LOCK
	caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
	if errCad := caddie.Update(); errCad != nil {
		log.Println("err addCaddieInOutNote", errCad.Error())
	}

	// Udp Note
	caddieInNote := model_gostarter.CaddieInOutNote{
		PartnerUid: PartnerUid,
		CourseUid:  CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		CaddieCode: booking.CaddieInfo.Code,
		Type:       constants.STATUS_IN,
		Note:       "",
	}

	go addCaddieInOutNote(caddieInNote)
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

	// Check Guest of member, check member có còn slot đi cùng không
	var memberCard models.MemberCard
	if body.MemberUidOfGuest != "" && body.GuestStyle != "" {
		var errCheckMember error
		customerName := ""
		errCheckMember, memberCard, customerName = handleCheckMemberCardOfGuest(body.MemberUidOfGuest, body.GuestStyle)
		if errCheckMember != nil {
			response_message.InternalServerError(c, errCheckMember.Error())
			return
		} else {
			booking.MemberUidOfGuest = body.MemberUidOfGuest
			booking.MemberNameOfGuest = customerName
		}
	}

	if body.Bag != "" {
		booking.Bag = body.Bag
		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(false, true)
		if isDuplicated {
			if errDupli != nil {
				response_message.InternalServerErrorWithKey(c, errDupli.Error(), "BAG_NOT_FOUND")
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		// Cập nhật lại info Bag
		booking.UpdateBagGolfFee()
	}

	if body.Hole > 0 {
		booking.Hole = body.Hole
	}

	checkInTime := time.Now().Unix()

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
		booking.CustomerType = golfFee.CustomerType

		// List Booking GolfFee
		param := request.GolfFeeGuestyleParam{
			Uid:          booking.Uid,
			Bag:          booking.Bag,
			CustomerName: booking.CustomerName,
			Hole:         booking.Hole,
		}
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
		initPriceForBooking(&booking, listBookingGolfFee, bookingGolfFee, checkInTime)
	}

	if body.Locker != "" {
		booking.LockerNo = body.Locker
		go createLocker(booking)
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CheckInTime = time.Now().Unix()
	booking.BagStatus = constants.BAG_STATUS_WAITING

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(memberCard)
	}

	if booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
	}

	res := getBagDetailFromBooking(booking)

	okResponse(c, res)
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

	if booking.MainBagPay == nil {
		booking.MainBagPay = initMainBagForPay()
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
					BillCode:   subBooking.BillCode,
				}
				booking.SubBags = append(booking.SubBags, subBag)
			} else {
				log.Println("AddSubBagToBooking err1", err1.Error())
			}
		}
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Tính lại giá
	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(booking)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	bookRes := model_booking.Booking{}
	bookRes.Uid = booking.Uid
	errFRes := bookRes.FindFirst()
	if errFRes != nil {
		response_message.InternalServerError(c, errFRes.Error())
		return
	}

	res := getBagDetailFromBooking(bookRes)

	okResponse(c, res)
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

	res := getBagDetailFromBooking(booking)
	okResponse(c, res)
}

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
		if v.Bag != form.Bag && len(v.MainBags) == 0 && len(v.SubBags) == 0 {
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

	// add cái mới
	for _, v := range body.OtherPaids {
		serviceItem := model_booking.BookingServiceItem{
			Type:     constants.BOOKING_OTHER_FEE,
			Name:     v.Reason,
			BillCode: booking.BillCode,
		}
		errF := serviceItem.FindFirst()
		if errF != nil {
			//Chưa có thì tạo mới
			serviceItem.Amount = v.Amount
			serviceItem.PlayerName = booking.CustomerName
			serviceItem.Bag = booking.Bag
			serviceItem.BookingUid = booking.Uid
			errC := serviceItem.Create()
			if errC != nil {
				log.Println("AddOtherPaid errC", errC.Error())
			}
		} else {
			// Check đã có thì udp
			if serviceItem.Amount != v.Amount {
				serviceItem.Amount = v.Amount
				errUdp := serviceItem.Update()
				if errUdp != nil {
					log.Println("AddOtherPaid errUdp", errUdp.Error())
				}
			}
		}
	}

	updatePriceWithServiceItem(booking, prof)

	booking.OtherPaids = body.OtherPaids

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	res := getBagDetailFromBooking(booking)

	okResponse(c, res)
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

	if booking.BagStatus != constants.BAG_STATUS_BOOKING {
		response_message.InternalServerError(c, "This booking did check in")
		return
	}
	// Kiểm tra xem đủ điều kiện cancel booking không
	cancelBookingSetting := model_booking.CancelBookingSetting{}
	if err := cancelBookingSetting.ValidateBookingCancel(booking); err != nil {
		response_message.InternalServerError(c, err.Error())
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

		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
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

	booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
	booking.CheckOutTime = time.Now().Unix()

	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// udp trạng thái caddie
	udpOutCaddieBooking(&booking)

	// delete tee time locked theo booking date
	if booking.TeeTime != "" {
		go unlockTurnTime(booking)
	}

	okResponse(c, booking)
}

func (cBooking *CBooking) CreateBookingTee(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for index := range bodyRequest.BookingList {
		bodyRequest.BookingList[index].BookingCode = bookingCode
	}

	cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
}

func (cBooking *CBooking) CreateCopyBooking(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	for indexTarget, target := range bodyRequest.BookingList {
		if !bodyRequest.BookingList[indexTarget].BookMark {
			bookingCode := utils.HashCodeUuid(uuid.New().String())
			bodyRequest.BookingList[indexTarget].BookingCode = bookingCode
			bodyRequest.BookingList[indexTarget].BookMark = true

			if target.BookingCode != "" {
				for index, data := range bodyRequest.BookingList {
					if data.BookingCode == target.BookingCode {
						bodyRequest.BookingList[index].BookingCode = bookingCode
						bodyRequest.BookingList[index].BookMark = true
					}
				}
			}
		}
	}
	cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
}

func (cBooking CBooking) CreateBatch(bookingList request.ListCreateBookingBody, c *gin.Context, prof models.CmsUser) {
	list := []model_booking.Booking{}
	for _, body := range bookingList {
		booking := cBooking.CreateBookingCommon(body, c, prof)
		if booking != nil {
			list = append(list, *booking)
		} else {
			return
		}
	}
	okResponse(c, list)
}
func (_ *CBooking) CancelAllBooking(c *gin.Context, prof models.CmsUser) {
	form := request.CancelAllBookingBody{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.BookingList{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
		TeeTime:     form.TeeTime,
		BookingCode: form.BookingCode,
	}

	db, _, err := bookingR.FindAllBookingList()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	for _, booking := range list {
		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
			response_message.InternalServerError(c, "Booking:"+booking.BookingDate+" did check in")
			return
		}

		booking.BagStatus = constants.BAG_STATUS_CANCEL
		booking.CancelNote = form.Reason
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

		errUdp := booking.Update()
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}
	okRes(c)
}

/*
Update booking fee by hole price formula
*/
func (_ *CBooking) ChangeBookingHole(c *gin.Context, prof models.CmsUser) {
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

	body := request.ChangeBookingHole{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Booking Note
	if body.NoteOfBag != "" && body.NoteOfBag != booking.NoteOfBag {
		booking.NoteOfBag = body.NoteOfBag
		go createBagsNoteNoteOfBag(booking)
	}

	// Update hole and type change hole
	booking.Hole = body.Hole
	booking.TypeChangeHole = constants.BOOKING_CHANGE_HOLE

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_RAIN {
		booking.TypeChangeHole = constants.BOOKING_STOP_BY_RAIN
	}

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_SELF {
		booking.TypeChangeHole = constants.BOOKING_STOP_BY_SELF
	}

	// List Booking GolfFee
	golfFeeModel := models.GolfFee{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		GuestStyle: booking.GuestStyle,
	}
	golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay()
	if errFindGF != nil {
		response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
		return
	}

	bookingGolfFee := getInitGolfFeeForChangeHole(body, golfFee)
	initUpdatePriceBookingForChanegHole(&booking, bookingGolfFee)

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}
