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
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

	booking, _ := cBooking.CreateBookingCommon(body, c, prof)
	if booking == nil {
		return
	}

	okResponse(c, booking)
}

func (cBooking CBooking) CreateBookingCommon(body request.CreateBookingBody, c *gin.Context, prof models.CmsUser) (*model_booking.Booking, error) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate caddie_code
	var caddie models.Caddie
	var err error
	if body.CaddieCode != "" {
		caddie, err = cBooking.validateCaddie(db, prof.CourseUid, body.CaddieCode)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return nil, err
		}
	}

	// validate trường hợp đóng tee 1
	teeList := []string{constants.TEE_TYPE_1, constants.TEE_TYPE_1A, constants.TEE_TYPE_1B, constants.TEE_TYPE_1C}
	if utils.Contains(teeList, body.TeeType) {
		cBookingSetting := CBookingSetting{}
		if errors := cBookingSetting.ValidateClose1ST(db, body.BookingDate, body.PartnerUid, body.CourseUid); errors != nil {
			response_message.InternalServerError(c, errors.Error())
			return nil, errors
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
		errFind := teeTime.FindFirst(db)
		if errFind == nil && (teeTime.TeeTimeStatus == constants.TEE_TIME_LOCKED) {
			response_message.BadRequest(c, "Tee Time đã bị khóa")
			return nil, errFind
		}
	}

	//check Booking Source with date time rule
	if body.BookingSourceId != "" {
		//TODO: check lại khi rãnh
		// bookingSource := model_booking.BookingSource{
		// 	PartnerUid: prof.PartnerUid,
		// 	CourseUid:  prof.CourseUid,
		// }
		// bookingSource.BookingSourceId = body.BookingSourceId

		// errorTime := bookingSource.ValidateTimeRuleInBookingSource(db, body.BookingDate, body.TeePath)
		// if errorTime != nil {
		// 	log.Println("", errorTime.Error())
		// 	response_message.BadRequest(c, errorTime.Error())
		// 	return nil, errorTime
		// }
	}

	if !body.IsCheckIn {
		teePartList := []string{"MORNING", "NOON", "NIGHT"}

		if !checkStringInArray(teePartList, body.TeePath) {
			response_message.BadRequest(c, "Tee Part not in (MORNING, NOON, NIGHT)")
			return nil, errors.New("Tee Part not in (MORNING, NOON, NIGHT)")
		}
	}

	booking := model_booking.Booking{
		PartnerUid:         body.PartnerUid,
		CourseUid:          body.CourseUid,
		TeeType:            body.TeeType,
		TeePath:            body.TeePath,
		TeeTime:            body.TeeTime,
		TeeOffTime:         body.TeeTime,
		TurnTime:           body.TurnTime,
		RowIndex:           body.RowIndex,
		CmsUser:            prof.UserName,
		Hole:               body.Hole,
		HoleBooking:        body.Hole,
		BookingRestaurant:  body.BookingRestaurant,
		BookingRetal:       body.BookingRetal,
		BookingCode:        body.BookingCode,
		CourseType:         body.CourseType,
		NoteOfBooking:      body.NoteOfBooking,
		BookingCodePartner: body.BookingCodePartner,
		BookingSourceId:    body.BookingSourceId,
	}

	// Check Guest of member, check member có còn slot đi cùng không
	var memberCard models.MemberCard
	if body.MemberUidOfGuest != "" && body.GuestStyle != "" {
		var errCheckMember error
		customerName := ""
		errCheckMember, memberCard, customerName = handleCheckMemberCardOfGuest(db, body.MemberUidOfGuest, body.GuestStyle)
		if errCheckMember != nil {
			response_message.InternalServerError(c, errCheckMember.Error())
			return nil, errCheckMember
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
	isDuplicated, _ := booking.IsDuplicated(db, true, true)
	if isDuplicated {
		// if errDupli != nil {
		// 	response_message.DuplicateRecord(c, errDupli.Error())
		// 	return nil, errDupli
		// }
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return nil, errors.New(constants.API_ERR_DUPLICATED_RECORD)
	}

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())
	booking.BillCode = utils.HashCodeUuid(bookingUid.String())

	// Checkin Time
	checkInTime := time.Now().Unix()

	// Member Card
	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst(db)
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return nil, errCourse
		}

		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return nil, errFind
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner(db)
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return nil, errOwner
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
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, checkInTime)
			booking.SeparatePrice = true
			body.GuestStyle = memberCard.GetGuestStyle(db)
		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle(db)
		}
	} else {
		booking.CustomerName = body.CustomerName
	}

	//Agency id
	if body.AgencyId > 0 {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst(db)
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return nil, errCourse
		}

		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst(db)
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return nil, errFindAgency
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyInfo = agencyBooking
		booking.AgencyId = body.AgencyId

		if booking.MemberCardUid == "" {
			// Nếu có cả member card thì ưu tiên giá member card
			agencySpecialPriceR := models.AgencySpecialPrice{
				AgencyId: agency.Id,
			}
			// Tính lại giá riêng nếu thoả mãn các dk time
			agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
			if errFSP == nil && agencySpecialPrice.Id > 0 {
				// Tính lại giá riêng nếu thoả mãn các dk time,
				// List Booking GolfFee
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
				initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, checkInTime)
				booking.SeparatePrice = true
				body.GuestStyle = agency.GuestStyle
			} else {
				body.GuestStyle = agency.GuestStyle
			}
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
		errFindCus := customer.FindFirst(db)
		if errFindCus != nil || customer.Uid == "" {
			response_message.BadRequest(c, "customer"+errFindCus.Error())
			return nil, errFindCus
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

		if errGS := golfFeeModel.FindFirst(db); errGS != nil {
			response_message.InternalServerError(c, "guest style not found ")
			return nil, errGS
		}

		booking.CustomerType = golfFeeModel.CustomerType

		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			return nil, errFindGF
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		if !booking.SeparatePrice {
			// List Booking GolfFee
			param := request.GolfFeeGuestyleParam{
				Uid:          bUid,
				Bag:          body.Bag,
				CustomerName: body.CustomerName,
				Hole:         body.Hole,
			}

			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, checkInTime)
		}
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
		cBooking.UpdateBookingCaddieCommon(db, body.PartnerUid, body.CourseUid, &booking, caddie)
	}

	if body.LockerNo != "" {
		booking.LockerNo = body.LockerNo
		go createLocker(db, booking)
	}

	if body.ReportNo != "" {
		booking.ReportNo = body.ReportNo
	}

	if body.CustomerIdentify != "" && booking.CustomerInfo.Uid == "" {
		customer := models.CustomerUser{}
		customer.Identify = body.CustomerIdentify
		customer.Phone = body.CustomerBookingPhone
		customer.Nationality = body.Nationality
		booking.CustomerInfo = cloneToCustomerBooking(customer)
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
	} else {
		booking.BookingCode = body.BookingCode
	}

	errC := booking.Create(db, bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return nil, errC
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(db, memberCard)
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

	// Create booking payment
	if body.AgencyId > 0 && booking.MemberCardUid == "" {
		go createAgencyPayment(db, booking)
	} else {
		go createSinglePayment(db, booking)
	}

	return &booking, nil
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

	okResponse(c, booking)
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

	errF := booking.FindFirstWithJoin(db)
	if errF != nil {
		// response_message.InternalServerError(c, errF.Error())
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	res := getBagDetailFromBooking(db, booking)

	okResponse(c, res)
}

/*
Get Round Bag trong ngày
*/
func (_ *CBooking) GetRoundOfBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.GolfBag == "" {
		response_message.BadRequest(c, errors.New("Bag invalid").Error())
		return
	}

	booking := model_booking.BookingList{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.GolfBag = form.GolfBag
	booking.BookingDate = form.BookingDate

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

	db, total, err := booking.FindAllBookingList(db)

	db = db.Order("created_at asc")
	db = db.Preload("CaddieBuggyInOut")

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}
	okResponse(c, res)
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
	bookings.CaddieCode = form.CaddieCode
	bookings.HasBookCaddie = form.HasBookCaddie
	bookings.CustomerName = form.PlayerName
	bookings.HasCaddieInOut = form.HasCaddieInOut
	bookings.FlightId = form.FlightId
	bookings.TeeType = form.TeeType
	bookings.IsCheckIn = form.IsCheckIn
	bookings.GuestStyleName = form.GuestStyleName
	bookings.PlayerOrBag = form.PlayerOrBag

	db, total, err := bookings.FindBookingListWithSelect(db, page)

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

	db, total, err := bookings.FindBookingListWithSelect(db, page)
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

func (_ CBooking) validateCaddie(db *gorm.DB, courseUid string, caddieCode string) (models.Caddie, error) {
	caddieList := models.CaddieList{}
	caddieList.CourseUid = courseUid
	caddieList.CaddieCode = caddieCode
	caddieNew, err := caddieList.FindFirst(db)

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

	body := request.UpdateBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate caddie_code
	var caddie models.Caddie
	var err error
	if body.CaddieCode != "" {
		caddie, err = cBooking.validateCaddie(db, prof.CourseUid, body.CaddieCode)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	if body.CourseType != "" {
		booking.CourseType = body.CourseType
	}

	if body.GuestStyle != "" {
		booking.GuestStyle = body.GuestStyle
	}

	//Upd Main Pay for Sub
	if body.MainBagPay != nil {
		booking.MainBagPay = body.MainBagPay
	}

	if body.LockerNo != "" {
		booking.LockerNo = body.LockerNo
		go createLocker(db, booking)
	}

	if body.ReportNo != "" {
		booking.ReportNo = body.ReportNo
	}

	if body.CustomerBookingName != "" {
		booking.CustomerBookingName = body.CustomerBookingName
	}

	if body.CustomerBookingPhone != "" {
		booking.CustomerBookingPhone = body.CustomerBookingPhone
	}

	if body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId {
		booking.SeparatePrice = false
	}

	if body.MemberCardUid != "" && (body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId) {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner(db)
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
			errCourse := course.FindFirst(db)
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
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
			booking.SeparatePrice = true
			body.GuestStyle = memberCard.GetGuestStyle(db)
		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle(db)
		}
	} else if body.MemberCardUid == "" {
		// Update member card
		booking.MemberCardUid = ""
		booking.CardId = ""
		booking.CustomerUid = ""
		booking.CustomerType = ""
		booking.CustomerInfo = model_booking.CustomerInfo{}

		if body.CustomerName != "" {
			booking.CustomerName = body.CustomerName
		}
	}

	//Agency id
	if body.AgencyId > 0 && body.AgencyId != booking.AgencyId {
		// Get config course
		course := models.Course{}
		course.Uid = body.CourseUid
		errCourse := course.FindFirst(db)
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return
		}

		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst(db)
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyInfo = agencyBooking
		booking.AgencyId = body.AgencyId

		if booking.MemberCardUid == "" {
			// Nếu có cả member card thì ưu tiên giá member card
			agencySpecialPriceR := models.AgencySpecialPrice{
				AgencyId: agency.Id,
			}
			// Tính lại giá riêng nếu thoả mãn các dk time
			agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
			if errFSP == nil && agencySpecialPrice.Id > 0 {
				// Tính lại giá
				// List Booking GolfFee
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
				initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
				booking.SeparatePrice = true
				body.GuestStyle = agency.GuestStyle
			} else {
				body.GuestStyle = agency.GuestStyle
			}
		}
	}
	// GuestStyle
	if body.GuestStyle != "" {
		//Update Agency
		if body.AgencyId == 0 {
			booking.AgencyInfo = model_booking.BookingAgency{}
			booking.AgencyId = 0
		}

		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			GuestStyle: body.GuestStyle,
		}

		if errGS := golfFeeModel.FindFirst(db); errGS != nil {
			response_message.InternalServerError(c, "guest style not found ")
			return
		}

		booking.CustomerType = golfFeeModel.CustomerType

		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			return
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		if !booking.SeparatePrice {
			// List Booking GolfFee
			param := request.GolfFeeGuestyleParam{
				Uid:          booking.Uid,
				Bag:          body.Bag,
				CustomerName: body.CustomerName,
				Hole:         body.Hole,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
		}
	}
	//Find Booking Code
	list, _ := booking.FindListWithBookingCode(db)
	if len(list) == 1 {
		booking.CustomerBookingName = booking.CustomerName
		booking.CustomerBookingPhone = booking.CustomerInfo.Phone
	}

	// Booking Note
	if body.NoteOfBag != "" && body.NoteOfBag != booking.NoteOfBag {
		booking.NoteOfBag = body.NoteOfBag
		go createBagsNoteNoteOfBag(db, booking)
	}

	if body.NoteOfBooking != "" && body.NoteOfBooking != booking.NoteOfBooking {
		booking.NoteOfBooking = body.NoteOfBooking
		go createBagsNoteNoteOfBooking(db, booking)
	}

	if body.NoteOfGo != "" {
		booking.NoteOfGo = body.NoteOfGo
	}

	// Update caddie
	if body.CaddieCode != "" {
		cBooking.UpdateBookingCaddieCommon(db, body.PartnerUid, body.CourseUid, &booking, caddie)
	}

	// Update các thông tin khác trước
	errUdpBook := booking.Update(db)
	if errUdpBook != nil {
		response_message.InternalServerError(c, errUdpBook.Error())
		return
	}

	// udp ok -> Tính lại giá
	updatePriceWithServiceItem(booking, prof)

	// Get lai booking mới nhất trong DB
	bookLast := model_booking.Booking{}
	bookLast.Uid = booking.Uid
	bookLast.FindFirst(db)

	res := getBagDetailFromBooking(db, bookLast)

	okResponse(c, res)
}

/*
Update booking caddie when create booking or update
*/
func (_ *CBooking) UpdateBookingCaddieCommon(db *gorm.DB, PartnerUid string, CourseUid string, booking *model_booking.Booking, caddie models.Caddie) {
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
	if errCad := caddie.Update(db); errCad != nil {
		log.Println("err udp caddie", errCad.Error())
	}

}

/*
Check in
*/
func (_ *CBooking) CheckIn(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.CheckInBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Check Guest of member, check member có còn slot đi cùng không
	var memberCard models.MemberCard
	if body.MemberUidOfGuest != "" && body.GuestStyle != "" {
		var errCheckMember error
		customerName := ""
		errCheckMember, memberCard, customerName = handleCheckMemberCardOfGuest(db, body.MemberUidOfGuest, body.GuestStyle)
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
		isDuplicated, errDupli := booking.IsDuplicated(db, false, true)
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

	if body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId {
		booking.SeparatePrice = false
	}

	if body.MemberCardUid != "" && (body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId) {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner(db)
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
			course.Uid = booking.CourseUid
			errCourse := course.FindFirst(db)
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
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
			booking.SeparatePrice = true
			body.GuestStyle = memberCard.GetGuestStyle(db)
		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle(db)
		}
	} else if body.MemberCardUid == "" {
		// Update member card
		booking.MemberCardUid = ""
		booking.CardId = ""
		booking.CustomerUid = ""
		booking.CustomerType = ""
		booking.CustomerInfo = model_booking.CustomerInfo{}

		if body.CustomerName != "" {
			booking.CustomerName = body.CustomerName
		}
	}

	//Agency id
	if body.AgencyId > 0 && body.AgencyId != booking.AgencyId {
		// Get config course
		course := models.Course{}
		course.Uid = booking.CourseUid
		errCourse := course.FindFirst(db)
		if errCourse != nil {
			response_message.BadRequest(c, errCourse.Error())
			return
		}

		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst(db)
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)
		booking.AgencyInfo = agencyBooking
		booking.AgencyId = body.AgencyId

		if booking.MemberCardUid == "" {
			// Nếu có cả member card thì ưu tiên giá member card
			agencySpecialPriceR := models.AgencySpecialPrice{
				AgencyId: agency.Id,
			}
			// Tính lại giá riêng nếu thoả mãn các dk time
			agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
			if errFSP == nil && agencySpecialPrice.Id > 0 {
				// Tính lại giá
				// List Booking GolfFee
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
				initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
				booking.SeparatePrice = true
				body.GuestStyle = agency.GuestStyle
			} else {
				body.GuestStyle = agency.GuestStyle
			}
		}
	}

	if body.GuestStyle != "" {
		//Update Agency
		if body.AgencyId == 0 {
			booking.AgencyInfo = model_booking.BookingAgency{}
			booking.AgencyId = 0
		}

		//Update giá đặc biệt nếu có

		// Tính giá
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			GuestStyle: body.GuestStyle,
		}

		if errGS := golfFeeModel.FindFirst(db); errGS != nil {
			response_message.InternalServerError(c, "guest style not found ")
			return
		}

		booking.CustomerType = golfFeeModel.CustomerType

		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFind := golfFeeModel.GetGuestStyleOnDay(db)
		if errFind != nil {
			response_message.InternalServerError(c, "golf fee err "+errFind.Error())
			return
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		if !booking.SeparatePrice {
			// List Booking GolfFee
			param := request.GolfFeeGuestyleParam{
				Uid:          booking.Uid,
				Bag:          booking.Bag,
				CustomerName: booking.CustomerName,
				Hole:         booking.Hole,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(param, golfFee)
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, checkInTime)
		}

	}

	if body.Locker != "" {
		booking.LockerNo = body.Locker
		go createLocker(db, booking)
	}

	if body.TeeType != "" {
		booking.TeeType = body.TeeType
	}

	if body.CustomerName != "" {
		booking.CustomerName = body.CustomerName
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CheckInTime = time.Now().Unix()
	booking.BagStatus = constants.BAG_STATUS_WAITING
	booking.CourseType = body.CourseType

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(db, memberCard)
	}

	if booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
	}

	// Create payment info
	if body.AgencyId > 0 && booking.MemberCardUid == "" {
		go createAgencyPayment(db, booking)
	} else {
		go createSinglePayment(db, booking)
	}

	res := getBagDetailFromBooking(db, booking)

	okResponse(c, res)
}

/*
Add Sub bag to Booking
*/
func (_ *CBooking) AddSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.AddSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
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

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag đã Check Out")
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
			err1 := subBooking.FindFirst(db)
			if err1 == nil {
				//Subbag
				subBag := utils.BookingSubBag{
					BookingUid: v.BookingUid,
					GolfBag:    subBooking.Bag,
					PlayerName: subBooking.CustomerName,
					BillCode:   subBooking.BillCode,
					BagStatus:  subBooking.BagStatus,
				}
				booking.SubBags = append(booking.SubBags, subBag)
			} else {
				log.Println("AddSubBagToBooking err1", err1.Error())
			}
		}
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Tính lại giá
	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(db, booking)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	bookRes := model_booking.Booking{}
	bookRes.Uid = booking.Uid
	errFRes := bookRes.FindFirst(db)
	if errFRes != nil {
		response_message.InternalServerError(c, errFRes.Error())
		return
	}

	// Update payment info
	if bookRes.AgencyId > 0 && bookRes.MemberCardUid == "" {
		// go createAgencyPayment(db, bookRes)
	} else {
		go createSinglePayment(db, bookRes)
	}

	res := getBagDetailFromBooking(db, bookRes)

	okResponse(c, res)
}

/*
Edit Sub bag to Booking
*/
func (_ *CBooking) EditSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.EditSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
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
		errFSB := subBooking.FindFirst(db)

		if errFSB != nil {
			log.Println("EditSubBagToBooking errFSB", errF.Error())
		}

		if v.IsOut == true {
			//remove di
			// Remove main bag
			subBooking.MainBags = utils.ListSubBag{}
			errSBUdp := subBooking.Update(db)
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
				errSBUdp := subBooking.Update(db)
				if errSBUdp != nil {
					log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
				}
			}
		}
	}

	if isUpdPrice {
		booking.UpdateMushPay(db)
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update(db)

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}

/*
Danh sách booking cho Add sub bag
*/
func (_ *CBooking) GetListBookingForAddSubBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	list, errF := bookingR.FindListForSubBag(db)
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

	list := []model_booking.Booking{}

	if booking.SubBags == nil || len(booking.SubBags) == 0 {
		okResponse(c, list)
		return
	}

	for _, v := range booking.SubBags {
		bookingTemp := model_booking.Booking{}
		bookingTemp.Uid = v.BookingUid
		errFind := bookingTemp.FindFirst(db)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	errF := booking.FindFirst(db)
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
		errF := serviceItem.FindFirst(db)
		if errF != nil {
			//Chưa có thì tạo mới
			serviceItem.Amount = v.Amount
			serviceItem.PlayerName = booking.CustomerName
			serviceItem.Bag = booking.Bag
			serviceItem.BookingUid = booking.Uid
			errC := serviceItem.Create(db)
			if errC != nil {
				log.Println("AddOtherPaid errC", errC.Error())
			}
		} else {
			// Check đã có thì udp
			if serviceItem.Amount != v.Amount {
				serviceItem.Amount = v.Amount
				errUdp := serviceItem.Update(db)
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

	errUdp := booking.Update(db)

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	res := getBagDetailFromBooking(db, booking)

	okResponse(c, res)
}

/*
Cancel Booking
- check chưa check-in mới cancel dc
*/
func (_ *CBooking) CancelBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	errF := booking.FindFirst(db)
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
	if err := cancelBookingSetting.ValidateBookingCancel(db, booking); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	booking.BagStatus = constants.BAG_STATUS_CANCEL
	booking.CancelNote = body.Note
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update(db)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		errF := booking.FindFirst(db)
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
		if body.CourseType != "" {
			booking.CourseType = body.CourseType
		}

		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(db, true, false)
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

		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}

	okRes(c)
}

func (_ CBooking) validateBooking(db *gorm.DB, bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(db); err != nil {
		return booking, err
	}

	return booking, nil
}

func (cBooking *CBooking) Checkout(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	booking, err := cBooking.validateBooking(db, body.BookingUid)
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

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// udp trạng thái caddie
	errCd := udpCaddieOut(db, booking.CaddieId)
	if errCd != nil {
		response_message.InternalServerError(c, errCd.Error())
		return
	}

	// delete tee time locked theo booking date
	if booking.TeeTime != "" {
		go unlockTurnTime(db, booking)
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
		booking, _ := cBooking.CreateBookingCommon(body, c, prof)
		if booking != nil {
			list = append(list, *booking)
		} else {
			return
		}
	}
	okResponse(c, list)
}
func (_ *CBooking) CancelAllBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	db, _, err := bookingR.FindAllBookingList(db)
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

		errUdp := booking.Update(db)
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

	body := request.ChangeBookingHole{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Booking Note
	if body.NoteOfBag != "" && body.NoteOfBag != booking.NoteOfBag {
		booking.NoteOfBag = body.NoteOfBag
		go createBagsNoteNoteOfBag(db, booking)
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
	golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
	if errFindGF != nil {
		response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
		return
	}

	bookingGolfFee := getInitGolfFeeForChangeHole(db, body, golfFee)
	initUpdatePriceBookingForChanegHole(&booking, bookingGolfFee)

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
	Check bag có được checkout hay không
*/
func (cBooking *CBooking) CheckBagCanCheckout(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CheckBagCanCheckoutBody{}
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//Find Bag
	bag := model_booking.Booking{
		BookingDate: body.BookingDate,
		Bag:         body.GolfBag,
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
	}

	errFF := bag.FindFirst(db)
	if errFF != nil {
		response_message.InternalServerError(c, errFF.Error())
		return
	}

	isCanCheckOut := false
	errMessage := "ok"

	if bag.BagStatus == constants.BAG_STATUS_TIMEOUT || bag.BagStatus == constants.BAG_STATUS_WAITING {
		isCanCheckOut = true

		//Check sub bag
		if bag.SubBags != nil && len(bag.SubBags) > 0 {
			for _, v := range bag.SubBags {
				subBag := model_booking.Booking{}
				subBag.Uid = v.BookingUid
				errF := subBag.FindFirst(db)

				if errF == nil {
					if bag.BagStatus == constants.BAG_STATUS_CHECK_OUT || bag.BagStatus == constants.BAG_STATUS_CANCEL {
					} else {
						errMessage = "Sub-bag chưa check checkout"
						isCanCheckOut = false
						break
					}
				}
			}
		}

		// Check service items
		// Find bag detail
		if isCanCheckOut {
			// Check tiep service items
			bagDetail := getBagDetailFromBooking(db, bag)
			if bagDetail.ListServiceItems != nil && len(bagDetail.ListServiceItems) > 0 {
				for _, v1 := range bagDetail.ListServiceItems {
					serviceCart := models.ServiceCart{}
					serviceCart.Id = v1.ServiceBill

					errSC := serviceCart.FindFirst(db)
					if errSC != nil {
						log.Println("FindFristServiceCart errSC", errSC.Error())
						return
					}

					// Check trong MainBag có trả mới add
					if v1.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
						// ok
					} else {
						if serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT || serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE {
							// ok
						} else {
							if v1.BillCode != bag.BillCode {
								errMessage = "Dich vụ của sub-bag chưa đủ điều kiện được checkout"
							} else {
								errMessage = "Dich vụ của bag chưa đủ điều kiện được checkout"
							}

							isCanCheckOut = false
							break
						}
					}
				}
			}
		}
	} else {
		isCanCheckOut = false
		errMessage = "Trạng thái bag không được checkout"
	}

	res := map[string]interface{}{
		"is_can_check_out": isCanCheckOut,
		"message":          errMessage,
	}
	c.JSON(200, res)
}
