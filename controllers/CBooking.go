package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils"
	"start/utils/response_message"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/twharmon/slices"
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

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	booking, _ := cBooking.CreateBookingCommon(body, c, prof)
	if booking == nil {
		// Khi booking lỗi thì remove index đã lưu trước đó trong redis
		go removeRowIndexRedis(model_booking.Booking{
			PartnerUid:  body.PartnerUid,
			CourseUid:   body.CourseUid,
			BookingDate: body.BookingDate,
			TeeType:     body.TeeType,
			TeeTime:     body.TeeTime,
			CourseType:  body.CourseType,
			RowIndex:    body.RowIndex,
		})
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: getBagDetailFromBooking(db, *booking)},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if body.IsCheckIn {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN

		cNotification := CNotification{}
		go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_ADD, booking)

	} else {
		opLog.Function = constants.OP_LOG_FUNCTION_BOOKING

		listBook := []model_booking.Booking{}
		listBook = append(listBook, *booking)
		// Send sms
		go genQRCodeListBook(listBook)
	}

	go createOperationLog(opLog)

	// Bắn socket để client update ui
	cNotification := CNotification{}
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, booking)
	// go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_ADD, booking)
	okResponse(c, booking)
}

func (cBooking CBooking) CreateBookingCommon(body request.CreateBookingBody, c *gin.Context, prof models.CmsUser) (*model_booking.Booking, error) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate caddie_code

	_, errDate := time.Parse(constants.DATE_FORMAT_1, body.BookingDate)
	if errDate != nil {
		response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
		return nil, errDate
	}

	if checkBookingAtOTAPosition(body) && !body.BookFromOTA {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", "Booking Online đang khóa tại tee time này!", constants.ERROR_BOOKING_OTA_LOCK)
		return nil, nil
	}

	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(body.BookingDate, body.CourseUid, body.TeeTime, body.TeeType+body.CourseType)

	rowIndexsRedisStr := ""

	addRejectedHandler := func(_ *datasources.Locker) error {
		rowIndexsRedisStr, _ = datasources.GetCache(teeTimeRowIndexRedis)
		return nil
	}

	keyRedisLockTee := fmt.Sprintf("redisLock_%s", teeTimeRowIndexRedis)

	errLock := datasources.Lock(datasources.LockOption{
		Key:     keyRedisLockTee,
		Ttl:     1 * time.Second,
		Handler: addRejectedHandler,
	})
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)

	if errLock != nil {
		return nil, errLock
	}

	if len(rowIndexsRedis) < constants.SLOT_TEE_TIME {
		if body.RowIndex == nil {
			rowIndex := generateRowIndex(rowIndexsRedis)
			body.RowIndex = &rowIndex
		}

		rowIndexsRedis = append(rowIndexsRedis, *body.RowIndex)
		rowIndexsRaw, _ := rowIndexsRedis.Value()
		errRedis := datasources.SetCache(teeTimeRowIndexRedis, rowIndexsRaw, 0)
		if errRedis != nil {
			log.Println("CreateBookingCommon errRedis", errRedis)
		}
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
		Hole:               body.HoleBooking,
		HoleBooking:        body.HoleBooking,
		BookingRestaurant:  body.BookingRestaurant,
		BookingRetal:       body.BookingRetal,
		BookingCode:        body.BookingCode,
		CourseType:         body.CourseType,
		NoteOfBooking:      body.NoteOfBooking,
		BookingCodePartner: body.BookingCodePartner,
		BookingSourceId:    body.BookingSourceId,
		AgencyPaidAll:      body.AgencyPaidAll,
	}

	// Check Guest of member, check member có còn slot đi cùng không
	var memberCard models.MemberCard
	guestStyle := ""

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

		if memberCard.Status == constants.STATUS_DISABLE {
			response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
			return nil, errors.New("Member Card Inactive")
		}
	}

	// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

	if body.Bag != "" {
		booking.Bag = body.Bag
	}

	if body.BookingDate != "" {
		// bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, body.BookingDate)
		// nowStr, _ := utils.GetLocalTimeFromTimeStamp("", constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())
		// nowUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, nowStr)

		// if bookingDateInt < nowUnix {
		// 	response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
		// 	return nil, errors.New(constants.BOOKING_DATE_NOT_VALID)
		// }
		booking.BookingDate = body.BookingDate
	} else {
		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		if errDate == nil {
			booking.BookingDate = dateDisplay
		} else {
			log.Println("booking date display err ", errDate.Error())
		}
	}

	if body.IsCheckIn {
		booking.Hole = body.Hole
		booking.HoleRound = body.Hole

		if body.CaddieCheckIn != nil {
			if errUpd := updateCaddieCheckIn(c, &booking, body.CaddieCheckIn); errUpd != nil {
				response_message.BadRequestFreeMessage(c, errUpd.Error())
				return nil, errUpd
			}
		}
	}

	//Check duplicated
	isDuplicated, _ := booking.IsDuplicated(db, true, true)
	if isDuplicated {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return nil, errors.New(constants.API_ERR_DUPLICATED_RECORD)
	}

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())
	booking.BillCode = utils.HashCodeUuid(bookingUid.String())

	// Checkin Time
	checkInTime := utils.GetTimeNow().Unix()

	// Member Card
	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get config course
		memberCardBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:    body.PartnerUid,
			CourseUid:     body.CourseUid,
			AgencyId:      body.AgencyId,
			BUid:          bUid,
			CustomerName:  body.CustomerName,
			Hole:          booking.Hole,
			MemberCardUid: body.MemberCardUid,
		}

		memberCard := models.MemberCard{}
		if errUpdate := cBooking.updateMemberCardToBooking(c, db, &booking, &memberCard, memberCardBody); errUpdate != nil {
			return nil, errUpdate
		}
		guestStyle = memberCard.GetGuestStyle(db)
	} else {
		booking.CustomerName = body.CustomerName
	}

	//Agency id
	if body.AgencyId > 0 {
		agencyBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   body.PartnerUid,
			CourseUid:    body.CourseUid,
			AgencyId:     body.AgencyId,
			BUid:         bUid,
			CustomerName: body.CustomerName,
			Hole:         booking.Hole,
		}
		agency := models.Agency{}
		if errAgency := cBooking.updateAgencyForBooking(db, &booking, &agency, agencyBody); errAgency != nil {
			response_message.BadRequest(c, errAgency.Error())
			return nil, errAgency
		}
		guestStyle = agency.GuestStyle
	}

	// Có thông tin khách hàng
	/*
		Chọn khách hàng từ agency
	*/
	if body.CustomerBookingEmail != nil && *body.CustomerBookingEmail != "" {
		booking.CustomerBookingEmail = *body.CustomerBookingEmail
	}

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
		// booking.CustomerType = customer.Type
		booking.CustomerInfo = cloneToCustomerBooking(customer)
		booking.CustomerUid = body.CustomerUid
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
	if body.GuestStyle != "" && guestStyle != body.GuestStyle {
		guestStyle = body.GuestStyle
	}

	// GuestStyle
	if guestStyle != "" {

		guestBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   body.PartnerUid,
			CourseUid:    body.CourseUid,
			BUid:         bUid,
			CustomerName: body.CustomerName,
			Hole:         booking.Hole,
		}

		if errUpdGs := cBooking.updateGuestStyleToBooking(c, guestStyle, db, &booking, guestBody); errUpdGs != nil {
			return nil, errUpdGs
		}
	}

	// Check In Out
	if body.IsCheckIn {
		// Tạo booking check in luôn
		booking.BagStatus = constants.BAG_STATUS_WAITING
		booking.LastBookingStatus = constants.BAG_STATUS_WAITING
		booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
		booking.CheckInTime = checkInTime
	} else {
		// Tạo booking
		booking.BagStatus = constants.BAG_STATUS_BOOKING
		booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
	}

	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT

	// Update caddie
	if body.CaddieCode != nil && *body.CaddieCode != "" {
		caddieList := models.CaddieList{}
		caddieList.CourseUid = body.CourseUid
		caddieList.CaddieCode = *body.CaddieCode
		caddieNew, err := caddieList.FindFirst(db)
		if err != nil {
			response_message.BadRequestFreeMessage(c, "Caddie "+err.Error())
			return nil, err
		}

		if caddieNew.ContractStatus == constants.CADDIE_CONTRACT_STATUS_TERMINATION {
			response_message.BadRequestFreeMessage(c, "Caddie termination")
			return nil, err
		}

		// check caddie booking
		cCaddie := CCaddie{}
		listCaddieWorkingByBookingDate := cCaddie.GetCaddieWorkingByDate(body.PartnerUid, body.CourseUid, body.BookingDate)
		if utils.ContainString(listCaddieWorkingByBookingDate, caddieNew.Code) == -1 {
			response_message.BadRequestFreeMessage(c, "Caddie "+caddieNew.Code+" không có lịch làm việc!")
		} else {
			booking.CaddieBooking = caddieNew.Code
			booking.HasBookCaddie = true

			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_RECEPTION,
				Function:    constants.OP_LOG_FUNCTION_BOOKING,
				Action:      constants.OP_LOG_ACTION_BOOK_CADDIE,
				Body:        models.JsonDataLog{Data: body},
				ValueOld:    models.JsonDataLog{},
				ValueNew:    models.JsonDataLog{Data: booking.CaddieBooking},
				Path:        c.Request.URL.Path,
				Method:      c.Request.Method,
				Bag:         booking.Bag,
				BookingDate: booking.BookingDate,
				BillCode:    booking.BillCode,
				BookingUid:  bUid,
			}
			go createOperationLog(opLog)
		}
	}

	if body.CustomerName != "" {
		booking.CustomerName = body.CustomerName
	}

	if body.LockerNo != "" {
		booking.LockerNo = body.LockerNo
		if body.IsCheckIn {
			go createLocker(db, booking)
		}
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

	if body.IsPrivateBuggy != nil {
		booking.IsPrivateBuggy = body.IsPrivateBuggy
	}

	errC := booking.Create(db, bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return nil, errC
	}

	// HaiCV: update sang hàm mới
	// go genQrCodeForBooking(&booking)

	if body.Bag != "" {
		opLog := models.OperationLog{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_RECEPTION,
			Function:    constants.OP_LOG_FUNCTION_BOOKING,
			Action:      constants.OP_LOG_ACTION_INPUT_BAG_BOOKING,
			Body:        models.JsonDataLog{Data: body},
			ValueOld:    models.JsonDataLog{},
			ValueNew:    models.JsonDataLog{Data: booking.Bag},
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
			BillCode:    booking.BillCode,
			BookingUid:  booking.Uid,
		}
		go createOperationLog(opLog)
	}

	if body.MemberUidOfGuest != "" && guestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(db, memberCard)
	}

	if body.TeeTime != "" && len(rowIndexsRedis) >= 3 {
		cLockTeeTime := CLockTeeTime{}
		lockTurn := request.CreateLockTurn{
			BookingDate: body.BookingDate,
			CourseUid:   body.CourseUid,
			PartnerUid:  body.PartnerUid,
			TeeTime:     body.TeeTime,
			TeeType:     body.TeeType,
			CourseType:  body.CourseType,
		}
		go cLockTeeTime.LockTurn(lockTurn, body.HoleBooking, c, prof)
	}

	if body.IsCheckIn && booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.CardId, booking.PartnerUid, booking.CourseUid)
	}

	// Create booking payment
	if booking.AgencyId > 0 {
		if body.FeeInfo != nil {
			go handleAgencyPaid(booking, *body.FeeInfo)
		}
	}

	if booking.AgencyId > 0 && booking.MemberCardUid == "" {
		go handleAgencyPayment(db, booking)
		// Tạo thêm single payment cho bag

	} else {
		if booking.BagStatus == constants.BAG_STATUS_WAITING {
			// checkin mới tạo payment
			go handleSinglePayment(db, booking)
		}
	}

	// Nếu booking từ waiting
	if body.BookingWaitingId > 0 {
		go deleteBookingWaiting(db, body.BookingWaitingId)
	}

	return &booking, nil
}

func (_ CBooking) updateAgencyForBooking(
	db *gorm.DB, booking *model_booking.Booking, agency *models.Agency,
	body request.UpdateAgencyOrMemberCardToBooking) error {
	// Get config course
	course := models.Course{}
	course.Uid = body.CourseUid
	errCourse := course.FindFirst()
	if errCourse != nil {
		return errCourse
	}

	agency.Id = body.AgencyId
	errFindAgency := agency.FindFirst(db)
	if errFindAgency != nil || agency.Id == 0 {
		return errFindAgency
	}

	agencyBooking := cloneToAgencyBooking(*agency)
	booking.AgencyInfo = agencyBooking
	booking.AgencyId = body.AgencyId
	// booking.CustomerType = agency.Type

	if booking.MemberCardUid == "" {
		// Nếu có cả member card thì ưu tiên giá member card
		agencySpecialPriceR := models.AgencySpecialPrice{
			AgencyId:   agency.Id,
			CourseUid:  booking.CourseUid,
			PartnerUid: booking.PartnerUid,
		}
		// Tính lại giá riêng nếu thoả mãn các dk time
		agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnTime(db)
		if errFSP == nil && agencySpecialPrice.Id > 0 {
			// Tính lại giá riêng nếu thoả mãn các dk time,
			// List Booking GolfFee
			param := request.GolfFeeGuestyleParam{
				Uid:          body.BUid,
				Rate:         course.RateGolfFee,
				Bag:          body.Bag,
				CustomerName: body.CustomerName,
				Hole:         body.Hole,
				CaddieFee:    agencySpecialPrice.CaddieFee,
				BuggyFee:     agencySpecialPrice.BuggyFee,
				GreenFee:     agencySpecialPrice.GreenFee,
			}
			listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
			initPriceForBooking(db, booking, listBookingGolfFee, bookingGolfFee)
			initListRound(db, *booking, bookingGolfFee)

			booking.SeparatePrice = true
		}
	}

	return nil
}

func (_ CBooking) updateMemberCardToBooking(c *gin.Context,
	db *gorm.DB, booking *model_booking.Booking, memberCard *models.MemberCard,
	body request.UpdateAgencyOrMemberCardToBooking) error {
	course := models.Course{}
	course.Uid = body.CourseUid
	errCourse := course.FindFirst()
	if errCourse != nil {
		response_message.BadRequest(c, errCourse.Error())
		return errCourse
	}

	// Get Member Card
	memberCard.Uid = body.MemberCardUid
	errFind := memberCard.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return errFind
	}

	if memberCard.Status == constants.STATUS_DISABLE {
		response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
		return errors.New("Error!")
	}

	if memberCard.AnnualType == constants.ANNUAL_TYPE_SLEEP {
		response_message.BadRequestDynamicKey(c, "ANNUAL_TYPE_SLEEP_NOT_CHECKIN", "")
		return errors.New("Error!")
	}

	// Get Owner
	owner, errOwner := memberCard.GetOwner(db)
	if errOwner != nil {
		response_message.BadRequest(c, errOwner.Error())
		return errOwner
	}

	// Get Member Card Type
	memberCardType := models.MemberCardType{}
	memberCardType.Id = memberCard.McTypeId
	errMCTypeFind := memberCardType.FindFirst(db)
	if errMCTypeFind == nil && memberCard.AnnualType == constants.ANNUAL_TYPE_LIMITED {
		// Validate số lượt chơi còn lại của memeber
		reportCustomer := model_report.ReportCustomerPlay{
			CustomerUid: owner.Uid,
		}

		if errF := reportCustomer.FindFirst(); errF == nil {
			playCountRemain := memberCard.AdjustPlayCount - reportCustomer.TotalPlayCount
			if playCountRemain <= 0 {
				response_message.ErrorResponse(c, http.StatusBadRequest, "PLAY_COUNT_INVALID", "", constants.ERROR_PLAY_COUNT_INVALID)
				return errF
			}
		}
	}

	booking.MemberCardUid = body.MemberCardUid
	booking.CardId = memberCard.CardId
	booking.CustomerName = owner.Name
	booking.CustomerUid = owner.Uid
	// booking.CustomerType = owner.Type
	booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)

	if memberCard.PriceCode == 1 && memberCard.IsValidTimePrecial() {
		// Check member card với giá riêng và time được áp dụng
		param := request.GolfFeeGuestyleParam{
			Uid:          body.BUid,
			Rate:         course.RateGolfFee,
			Bag:          body.Bag,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
			CaddieFee:    memberCard.CaddieFee,
			BuggyFee:     memberCard.BuggyFee,
			GreenFee:     memberCard.GreenFee,
		}
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeWithOutGuestStyleForBooking(param)
		initPriceForBooking(db, booking, listBookingGolfFee, bookingGolfFee)
		initListRound(db, *booking, bookingGolfFee)

		booking.SeparatePrice = true
	}
	return nil
}

func (_ CBooking) updateGuestStyleToBooking(c *gin.Context, guestStyle string,
	db *gorm.DB, booking *model_booking.Booking,
	body request.UpdateAgencyOrMemberCardToBooking) error {
	//Guest style
	golfFeeModel := models.GolfFee{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		GuestStyle: guestStyle,
	}

	if errGS := golfFeeModel.FindFirst(db); errGS != nil {
		response_message.InternalServerError(c, "guest style not found ")
	}

	// Lấy phí bởi Guest style với ngày tạo
	golfFee := models.GolfFee{}
	var errFindGF error

	if booking.BookingDate != "" {
		timeDate, _ := time.Parse(constants.DATE_FORMAT_1, booking.BookingDate)
		golfFee, errFindGF = golfFeeModel.GetGuestStyleOnTime(db, timeDate)
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
		}
	} else {
		golfFee, errFindGF = golfFeeModel.GetGuestStyleOnDay(db)
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
		}
	}

	booking.CustomerType = golfFee.CustomerType
	booking.GuestStyle = guestStyle
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
		initPriceForBooking(db, booking, listBookingGolfFee, bookingGolfFee)
		initListRound(db, *booking, bookingGolfFee)
	}
	return nil
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

func checkBookingAtOTAPosition(body request.CreateBookingBody) bool {

	// Lấy số slot đã book
	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(body.BookingDate, body.CourseUid, body.TeeTime, body.TeeType+body.CourseType)
	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
	//

	// Nếu row_index không trùng với vị trí row index của ota
	if body.RowIndex != nil && !utils.Contains(rowIndexsRedis, *body.RowIndex) {
		return false
	}

	prefixRedisKey := getKeyTeeTimeLockRedis(body.BookingDate, body.CourseUid, body.TeeTime, body.TeeType+body.CourseType)
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)

	haveLockOTA := false
	if errRedis == nil && len(listKey) > 0 {
		strData, errGet := datasources.GetCaches(listKey...)
		if errGet != nil {
			log.Println("checkBookingOTA-error", errGet.Error())
		} else {
			for _, data := range strData {
				if data != nil {
					byteData := []byte(data.(string))
					teeTime := models.LockTeeTimeWithSlot{}
					if err2 := json.Unmarshal(byteData, &teeTime); err2 == nil {
						if teeTime.Type == constants.LOCK_OTA {
							haveLockOTA = true
							break
						}
					}
				}
			}
		}
	}
	return haveLockOTA
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

	bookingR := model_booking.Booking{}
	bookingR.Uid = bookingIdStr
	bookingR.PartnerUid = prof.PartnerUid
	bookingR.CourseUid = prof.CourseUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Data old booking
	oldBooking := getBagDetailFromBooking(db, booking.CloneBooking())

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag đã CheckOut!")
		return
	}

	body := request.UpdateBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate caddie_code
	guestStyle := ""
	checkHoleChange := false

	if body.HoleBooking > 0 {
		booking.HoleBooking = body.HoleBooking
	}

	if body.Hole > 0 && body.Hole != booking.Hole {
		booking.Hole = body.Hole
		checkHoleChange = true
	}

	if body.CourseType != "" {
		booking.CourseType = body.CourseType
	}

	// if body.GuestStyle != "" {
	// 	booking.GuestStyle = body.GuestStyle
	// }
	// Upd email
	if body.CustomerBookingEmail != nil && *body.CustomerBookingEmail != "" {
		booking.CustomerBookingEmail = *body.CustomerBookingEmail
	}

	//Upd Main Pay for Sub
	isPriceChanged := false
	if body.MainBagPay != nil {
		if !reflect.DeepEqual(booking.MainBagPay, body.MainBagPay) {
			booking.MainBagPay = body.MainBagPay
			isPriceChanged = true
			go bookMarkRoundPaidByMainBag(booking, db)
		}
	}

	if body.LockerNo != nil {
		if booking.CheckInTime > 0 {
			if booking.LockerNo != "" {
				locker := models.Locker{
					PartnerUid:  booking.PartnerUid,
					CourseUid:   booking.CourseUid,
					Locker:      booking.LockerNo,
					BookingDate: booking.BookingDate,
					BookingUid:  booking.Uid,
				}

				// Find locker
				_ = locker.FindFirst(db)

				errC := locker.Delete(db)
				if errC != nil {
					log.Println("deleteLocker errC", errC.Error())
				}
			}
			booking.LockerNo = *body.LockerNo
			go createLocker(db, booking)
		} else {
			booking.LockerNo = *body.LockerNo
		}

	}

	if body.ReportNo != nil {
		booking.ReportNo = *body.ReportNo
	}

	if body.CustomerBookingName != "" {
		booking.CustomerBookingName = body.CustomerBookingName
	}

	if body.CustomerBookingPhone != "" {
		booking.CustomerBookingPhone = body.CustomerBookingPhone
	}

	if body.MemberCardUid != nil && *body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId {
		booking.SeparatePrice = false
	}

	if body.BookingCode != "" {
		booking.BookingCode = body.BookingCode
	}

	//TODO: if body.MemberCardUid != "" && (body.MemberCardUid != booking.MemberCardUid ||
	// 	body.AgencyId != booking.AgencyId) {
	if body.MemberCardUid != nil {
		if *body.MemberCardUid != "" {
			memberCardBody := request.UpdateAgencyOrMemberCardToBooking{
				PartnerUid:    body.PartnerUid,
				CourseUid:     body.CourseUid,
				AgencyId:      body.AgencyId,
				BUid:          booking.Uid,
				Bag:           booking.Bag,
				CustomerName:  body.CustomerName,
				Hole:          body.Hole,
				MemberCardUid: *body.MemberCardUid,
			}
			memberCard := models.MemberCard{}
			if errUpdate := cBooking.updateMemberCardToBooking(c, db, &booking, &memberCard, memberCardBody); errUpdate != nil {
				return
			}
			guestStyle = memberCard.GetGuestStyle(db)
		} else {
			booking.MemberCardUid = ""
			booking.CardId = ""
			booking.CustomerUid = ""
			booking.CustomerType = ""
			booking.CustomerInfo = model_booking.CustomerInfo{}

			if body.CustomerName != "" {
				booking.CustomerName = body.CustomerName
			}
		}
	}

	//Agency id
	// isAgencyChanged := false
	if body.AgencyId > 0 && body.AgencyId != booking.AgencyId {
		agencyBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   body.PartnerUid,
			CourseUid:    body.CourseUid,
			AgencyId:     body.AgencyId,
			BUid:         booking.Uid,
			Bag:          booking.Bag,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
		}
		agency := models.Agency{}
		if errAgency := cBooking.updateAgencyForBooking(db, &booking, &agency, agencyBody); errAgency != nil {
			response_message.BadRequest(c, errAgency.Error())
		}
		guestStyle = agency.GuestStyle
		// isAgencyChanged = true
	}

	// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
	if body.GuestStyle != "" && guestStyle != body.GuestStyle {
		guestStyle = body.GuestStyle
	}

	// GuestStyle
	if guestStyle != "" && booking.GuestStyle != guestStyle {
		//Guest style
		guestBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   body.PartnerUid,
			CourseUid:    body.CourseUid,
			BUid:         booking.Uid,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
			Bag:          booking.Bag,
		}

		if errUpdGs := cBooking.updateGuestStyleToBooking(c, guestStyle, db, &booking, guestBody); errUpdGs != nil {
			return
		}

		isPriceChanged = true
	}

	// Booking Note
	if body.NoteOfBag != nil {
		booking.NoteOfBag = *body.NoteOfBag
		go createBagsNoteNoteOfBag(db, booking)
	}

	if body.NoteOfBooking != nil {
		booking.NoteOfBooking = *body.NoteOfBooking
		go createBagsNoteNoteOfBooking(db, booking)
	}

	if body.NoteOfGo != nil {
		booking.NoteOfGo = *body.NoteOfGo
	}

	if body.CustomerName != "" {
		booking.CustomerName = body.CustomerName
		booking.CustomerInfo.Name = body.CustomerName
	}

	// Create booking payment
	if booking.AgencyId > 0 {

		if body.AgencyPaidAll != nil {
			booking.AgencyPaidAll = body.AgencyPaidAll
		}
	}

	// Update bag nếu có thay đổi
	if errUdpBag := updateBag(c, &oldBooking, &booking, body, prof); errUdpBag != nil {
		return
	}

	if checkHoleChange {
		updateHole(c, &booking, body.Hole)
	}

	if body.CaddieCode != nil {
		if errUpd := updateCaddieBooking(c, &oldBooking, &booking, body, prof); errUpd != nil {
			response_message.BadRequestFreeMessage(c, errUpd.Error())
			return
		}
	}

	if body.CaddieCheckIn != nil {
		if errUpd := updateCaddieCheckIn(c, &booking, body.CaddieCheckIn); errUpd != nil {
			response_message.BadRequestFreeMessage(c, errUpd.Error())
			return
		}
	}

	// Update các thông tin khác trước
	errUdpBook := booking.Update(db)
	if errUdpBook != nil {
		response_message.InternalServerError(c, errUdpBook.Error())
		return
	}

	// Create booking payment
	if booking.AgencyId > 0 {
		if body.FeeInfo != nil {
			handleAgencyPaid(booking, *body.FeeInfo)
		}
	}

	// udp ok -> Tính lại giá
	if isPriceChanged {
		updatePriceWithServiceItem(&booking, prof)
	}

	// Update lại thông tin agency cho các round, move flight
	// if isAgencyChanged {
	// 	go func() {
	// 		booking.UpdateAgencyForBooking(db)
	// 		updateAgencyInfoInPayment(booking)
	// 	}()
	// }

	// Get lai booking mới nhất trong DB
	bookLast := model_booking.Booking{}
	bookLast.Uid = booking.Uid
	bookLast.FindFirst(db)

	res := getBagDetailFromBooking(db, bookLast)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_BOOKING,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: res},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}
	go createOperationLog(opLog)

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_UPD_BOOKING_CMS, model_booking.Booking{})

	okResponse(c, res)
}

func updateCaddieCheckIn(c *gin.Context, booking *model_booking.Booking, caddie *string) error {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	if caddie != nil {
		oldCaddie := booking.CaddieInfo
		if *caddie != "" {
			if *caddie != booking.CaddieInfo.Code {

				caddieList := models.CaddieList{}
				caddieList.CourseUid = booking.CourseUid
				caddieList.CaddieCode = *caddie
				caddieNew, err := caddieList.FindFirst(db)

				if err != nil {
					return errors.New("Caddie Not Found!")
				}

				if caddieNew.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
					if booking.CaddieId != caddieNew.Id {
						return errors.New("Caddie " + caddieNew.Code + " đang bị LOCK")
					}
				} else {
					if errCaddie := checkCaddieReady(*booking, caddieNew); errCaddie != nil {
						if strings.Contains(caddieNew.CurrentStatus, "IN_COURSE") {
							return errors.New("Caddie " + caddieNew.Code + " đang IN COURSE")
						}
						return errors.New("Caddie " + caddieNew.Code + " đang " + caddieNew.CurrentStatus)
					}
				}

				if caddieNew.ContractStatus == constants.CADDIE_CONTRACT_STATUS_TERMINATION {
					return errors.New("Caddie termination!")
				}

				cCaddie := CCaddie{}
				listCaddieWorkingByBookingDate := cCaddie.GetCaddieWorkingByDate(booking.PartnerUid, booking.CourseUid, booking.BookingDate)
				if utils.ContainString(listCaddieWorkingByBookingDate, caddieNew.Code) == -1 {
					return errors.New("Caddie " + caddieNew.Code + " không có lịch làm việc!")
				}

				booking.CaddieId = caddieNew.Id
				booking.CaddieInfo = cloneToCaddieBooking(caddieNew)

				//Update lại trạng thái caddie mới
				caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
				_ = caddieNew.Update(db)

			}
		} else {
			booking.CaddieId = 0
			booking.CaddieInfo = model_booking.BookingCaddie{}
		}

		// Out Caddie, nếu caddie trong in course
		go func() {
			if oldCaddie.Id > 0 {
				caddie := models.Caddie{}
				caddie.Id = oldCaddie.Id
				if err := caddie.FindFirst(db); err == nil {
					udpCaddieOut(db, oldCaddie.Id)
				}

			}
		}()
	}
	return nil
}

func updateCaddieBooking(c *gin.Context, oldBooking *model_booking.BagDetail, booking *model_booking.Booking, body request.UpdateBooking, prof models.CmsUser) error {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	if body.CaddieCode != nil {

		if *body.CaddieCode != "" {
			if *body.CaddieCode != booking.CaddieBooking {
				caddieList := models.CaddieList{}
				caddieList.CourseUid = booking.CourseUid
				caddieList.CaddieCode = *body.CaddieCode
				caddieNew, err := caddieList.FindFirst(db)

				if err != nil {
					return errors.New("Caddie Not Found!")
				}

				if caddieNew.ContractStatus == constants.CADDIE_CONTRACT_STATUS_TERMINATION {
					return errors.New("Caddie termination!")
				}

				// check caddie booking
				cCaddie := CCaddie{}
				listCaddieWorkingByBookingDate := cCaddie.GetCaddieWorkingByDate(body.PartnerUid, body.CourseUid, booking.BookingDate)
				if utils.ContainString(listCaddieWorkingByBookingDate, caddieNew.Code) == -1 {
					return errors.New("Caddie " + caddieNew.Code + " không có lịch làm việc!")
				}

				booking.CaddieBooking = caddieNew.Code
			}
		} else {
			booking.CaddieBooking = ""
		}

		if booking.CheckInTime == 0 && oldBooking.CaddieBooking != *body.CaddieCode {
			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_RECEPTION,
				Function:    constants.OP_LOG_FUNCTION_BOOKING,
				Action:      constants.OP_LOG_ACTION_UPD_BOOK_CADDIE,
				Body:        models.JsonDataLog{Data: body},
				ValueOld:    models.JsonDataLog{Data: oldBooking.CaddieBooking},
				ValueNew:    models.JsonDataLog{Data: booking.CaddieBooking},
				Path:        c.Request.URL.Path,
				Method:      c.Request.Method,
				Bag:         booking.Bag,
				BookingDate: booking.BookingDate,
				BillCode:    booking.BillCode,
				BookingUid:  booking.Uid,
			}
			go createOperationLog(opLog)
		}
	}
	return nil
}

func updateHole(c *gin.Context, booking *model_booking.Booking, hole int) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	round := models.Round{
		BillCode: booking.BillCode,
	}

	if errFindRound := round.LastRound(db); errFindRound != nil {
		log.Println("Round not found")
	}

	cRound := CRound{}
	cRound.UpdateListFeePriceInRound(c, db, booking, booking.GuestStyle, &round, hole)
}

func updateBag(c *gin.Context, oldBooking *model_booking.BagDetail, booking *model_booking.Booking, body request.UpdateBooking, prof models.CmsUser) error {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)

	if body.Bag != nil {
		// data old
		if *body.Bag != "" {
			if booking.Bag != *body.Bag {
				booking.Bag = *body.Bag
				//Check duplicated
				isDuplicated, errDupli := booking.IsDuplicated(db, false, true)
				if isDuplicated {
					if errDupli != nil {
						response_message.InternalServerErrorWithKey(c, errDupli.Error(), "DUPLICATE_BAG")
						return errDupli
					}
					response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
					return errors.New("Update Bag Failed!")
				}

				if len(booking.MainBags) > 0 {
					response_message.BadRequestFreeMessage(c, "Update Bag Failed!")
					return errors.New("Update Bag Failed!")
				}

				if len(booking.SubBags) > 0 {
					response_message.BadRequestFreeMessage(c, "Update Bag Failed!")
					return errors.New("Update Bag Failed!")
				}

				bookingServiceItemsR := model_booking.BookingServiceItem{
					PartnerUid: booking.PartnerUid,
					CourseUid:  booking.CourseUid,
					BillCode:   booking.BillCode,
				}
				list, _ := bookingServiceItemsR.FindAll(db)

				hasUpdateBag := true
				listItem := []model_booking.BookingServiceItem{}
				for _, item := range list {
					if item.ServiceType == constants.BUGGY_SETTING || item.ServiceType == constants.CADDIE_SETTING {
						listItem = append(listItem, item)
					} else {
						hasUpdateBag = false
					}
				}
				if !hasUpdateBag {
					response_message.BadRequestFreeMessage(c, "Update Bag Failed!")
					return errors.New("Update Bag Failed!")
				}

				// Cập nhật lại info Bag
				booking.UpdateBagGolfFee()
				booking.UpdateRoundForBooking(db)

				go func() {
					for _, item := range listItem {
						item.Bag = booking.Bag
						item.Update(db)
					}
					roundR := models.Round{
						BillCode: booking.BillCode,
					}
					listRound, _ := roundR.FindAll(db)
					for _, round := range listRound {
						round.Bag = booking.Bag
						round.Update(db)
					}
				}()
			}
		} else {
			if booking.CheckInTime == 0 {
				booking.Bag = *body.Bag
			}
		}

		if booking.CheckInTime == 0 && oldBooking.Bag != *body.Bag {
			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_RECEPTION,
				Function:    constants.OP_LOG_FUNCTION_BOOKING,
				Action:      constants.OP_LOG_ACTION_INPUT_BAG_BOOKING,
				Body:        models.JsonDataLog{Data: body},
				ValueOld:    models.JsonDataLog{Data: oldBooking.Bag},
				ValueNew:    models.JsonDataLog{Data: booking.Bag},
				Path:        c.Request.URL.Path,
				Method:      c.Request.Method,
				Bag:         booking.Bag,
				BookingDate: booking.BookingDate,
				BillCode:    booking.BillCode,
				BookingUid:  booking.Uid,
			}
			go createOperationLog(opLog)
		}
	}
	return nil
}

/*
Check in
*/
func (cBooking *CBooking) CheckIn(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.CheckInBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	guestStyle := ""
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	oldBooking := getBagDetailFromBooking(db, booking.CloneBooking())

	//Checkin rồi thì k check in lại dc nữa
	if booking.BagStatus == constants.BAG_STATUS_WAITING && booking.CheckInTime > 0 {
		response_message.BadRequest(c, "da checkin roi")
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

	if body.Bag != "" && booking.Bag != body.Bag {
		booking.Bag = body.Bag
		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(db, false, true)
		if isDuplicated {
			if errDupli != nil {
				response_message.InternalServerErrorWithKey(c, errDupli.Error(), "DUPLICATE_BAG")
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

	if body.Locker != "" {
		// locker := models.Locker{
		// 	PartnerUid:   booking.PartnerUid,
		// 	CourseUid:    booking.CourseUid,
		// 	Locker:       body.Locker,
		// 	BookingDate:  booking.BookingDate,
		// 	LockerStatus: constants.LOCKER_STATUS_UNRETURNED,
		// }

		// // check tồn tại
		// _ = locker.FindFirst(db)

		// if locker.Id > 0 {
		// 	response_message.BadRequestFreeMessage(c, "Locker đã được mượn.")
		// 	return
		// }

		booking.LockerNo = body.Locker
		go createLocker(db, booking)
	}

	if body.MemberCardUid != nil && *body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId {
		booking.SeparatePrice = false
	}

	if body.MemberCardUid != nil {
		if *body.MemberCardUid != "" {
			memberCardBody := request.UpdateAgencyOrMemberCardToBooking{
				PartnerUid:    booking.PartnerUid,
				CourseUid:     booking.CourseUid,
				AgencyId:      body.AgencyId,
				BUid:          booking.Uid,
				Bag:           booking.Bag,
				CustomerName:  body.CustomerName,
				Hole:          body.Hole,
				MemberCardUid: *body.MemberCardUid,
			}
			memberCard := models.MemberCard{}
			if errUpdate := cBooking.updateMemberCardToBooking(c, db, &booking, &memberCard, memberCardBody); errUpdate != nil {
				return
			}
			guestStyle = memberCard.GetGuestStyle(db)
		} else {
			booking.MemberCardUid = ""
			booking.CardId = ""
			booking.CustomerUid = ""
			booking.CustomerType = ""
			booking.CustomerInfo = model_booking.CustomerInfo{}

			if body.CustomerName != "" {
				booking.CustomerName = body.CustomerName
			}
		}
	}

	//Agency id
	// isAgencyChanged := false
	if body.AgencyId > 0 && (body.AgencyId != booking.AgencyId || body.Hole != booking.Hole) {
		agencyBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   booking.PartnerUid,
			CourseUid:    booking.CourseUid,
			AgencyId:     body.AgencyId,
			BUid:         booking.Uid,
			Bag:          booking.Bag,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
		}
		agency := models.Agency{}
		if errAgency := cBooking.updateAgencyForBooking(db, &booking, &agency, agencyBody); errAgency != nil {
			response_message.BadRequest(c, errAgency.Error())
		}
		guestStyle = agency.GuestStyle
		// isAgencyChanged = true
	}

	// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
	if body.GuestStyle != "" && guestStyle != body.GuestStyle {
		guestStyle = body.GuestStyle
	}

	if guestStyle != "" {
		// Tính giá
		//Guest style
		guestBody := request.UpdateAgencyOrMemberCardToBooking{
			PartnerUid:   booking.PartnerUid,
			CourseUid:    booking.CourseUid,
			BUid:         booking.Uid,
			CustomerName: body.CustomerName,
			Hole:         body.Hole,
			Bag:          booking.Bag,
		}

		if errUpdGs := cBooking.updateGuestStyleToBooking(c, guestStyle, db, &booking, guestBody); errUpdGs != nil {
			return
		}
	}

	if body.CaddieCode != "" {
		caddieList := models.CaddieList{}
		caddieList.CourseUid = booking.CourseUid
		caddieList.CaddieCode = body.CaddieCode
		caddieNew, err := caddieList.FindFirst(db)

		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		if caddieNew.ContractStatus == constants.CADDIE_CONTRACT_STATUS_TERMINATION {
			response_message.BadRequestFreeMessage(c, "Caddie "+err.Error())
			return
		}

		// Out Caddie, nếu caddie trong in course
		if booking.CaddieInfo.Code != body.CaddieCode {
			oldCad := booking.CaddieId
			go func() {
				if oldCad > 0 {
					caddie := models.Caddie{}
					caddie.Id = oldCad
					if err := caddie.FindFirst(db); err == nil {
						udpCaddieOut(db, oldCad)
					}
				}
			}()
		}

		booking.CaddieId = caddieNew.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	} else if booking.CaddieId == 0 && booking.CaddieBooking != "" {
		caddieList := models.CaddieList{}
		caddieList.CourseUid = booking.CourseUid
		caddieList.CaddieCode = booking.CaddieBooking
		caddieNew, err := caddieList.FindFirst(db)

		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		if caddieNew.ContractStatus == constants.CADDIE_CONTRACT_STATUS_TERMINATION {
			response_message.BadRequestFreeMessage(c, "Caddie "+err.Error())
			return
		}

		booking.CaddieId = caddieNew.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	}

	if body.TeeType != "" {
		booking.TeeType = body.TeeType
	}

	//Update customer infor
	if booking.CustomerUid != "" {
		//check customer
		customer := models.CustomerUser{}
		customer.Uid = booking.CustomerUid
		errFindCus := customer.FindFirst(db)
		if errFindCus != nil || customer.Uid == "" {
			log.Print("customer" + errFindCus.Error())
			// return
		}

		booking.CustomerName = customer.Name
		// booking.CustomerType = customer.Type
		booking.CustomerInfo = cloneToCustomerBooking(customer)
	}

	if body.CustomerName != "" {
		booking.CustomerName = body.CustomerName
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
	booking.CheckInTime = utils.GetTimeNow().Unix()
	booking.BagStatus = constants.BAG_STATUS_WAITING
	booking.CourseType = body.CourseType
	booking.HoleRound = body.Hole

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Create booking payment
	if booking.AgencyId > 0 {

		if body.AgencyPaidAll != nil {
			booking.AgencyPaidAll = body.AgencyPaidAll
		}

		if body.FeeInfo != nil {
			go handleAgencyPaid(booking, *body.FeeInfo)
		}
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(db, memberCard)
	}

	if booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.CardId, booking.PartnerUid, booking.CourseUid)
	}

	go func() {
		if booking.CaddieBooking != "" {
			caddieBookingFee := getBookingCadieFeeSetting(booking.PartnerUid, booking.CourseUid, booking.GuestStyle, body.Hole)
			addCaddieBookingFee(booking, caddieBookingFee.Fee, constants.BOOKING_CADDIE_NAME, body.Hole)
			updatePriceWithServiceItem(&booking, prof)
		} else {
			// Create payment info
			handlePayment(db, booking)
		}
	}()

	// Update lại round còn thiếu bag
	cRound := CRound{}
	go cRound.UpdateBag(booking, db)

	res := getBagDetailFromBooking(db, booking)

	// Update bag attach caddie
	go cBooking.updateBagAttachCaddie(db, booking.Uid)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_CHECK_IN,
		Action:      constants.OP_LOG_ACTION_CHECK_IN,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: res},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         res.Bag,
		BookingDate: res.BookingDate,
		BillCode:    res.BillCode,
		BookingUid:  res.Uid,
	}
	go createOperationLog(opLog)

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_UPD_BOOKING_CMS, model_booking.Booking{})

	okResponse(c, res)
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

	// if body.OtherPaids == nil || len(body.OtherPaids) == 0 {
	// 	response_message.BadRequest(c, errors.New("other paid empty").Error())
	// 	return
	// }

	bookingR := model_booking.Booking{}
	bookingR.Uid = body.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag đã CheckOut!")
		return
	}

	// Xóa all trước khi add mới
	serviceItemR := model_booking.BookingServiceItem{
		Type:       constants.BOOKING_OTHER_FEE,
		BillCode:   booking.BillCode,
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
	}
	list, _ := serviceItemR.FindAll(db)
	for _, item := range list {
		item.Delete(db)
	}

	// add cái mới
	for _, v := range body.OtherPaids {
		serviceItem := model_booking.BookingServiceItem{
			Type:       constants.BOOKING_OTHER_FEE,
			Name:       v.Reason,
			BillCode:   booking.BillCode,
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
		}
		errF := serviceItem.FindFirst(db)
		if errF != nil {
			//Chưa có thì tạo mới
			serviceItem.Amount = v.Amount
			serviceItem.UnitPrice = v.Amount
			serviceItem.PlayerName = booking.CustomerName
			serviceItem.Bag = booking.Bag
			serviceItem.BookingUid = booking.Uid
			serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION
			serviceItem.Quality = 1
			errC := serviceItem.Create(db)
			if errC != nil {
				log.Println("AddOtherPaid errC", errC.Error())
			}
		} else {
			// Check đã có thì udp
			if serviceItem.Amount != v.Amount {
				serviceItem.Amount = v.Amount
				serviceItem.UnitPrice = v.Amount
				errUdp := serviceItem.Update(db)
				if errUdp != nil {
					log.Println("AddOtherPaid errUdp", errUdp.Error())
				}
			}
		}
	}

	booking.OtherPaids = body.OtherPaids

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	errUdp := booking.Update(db)

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}
	updatePriceWithServiceItem(&booking, prof)

	res := getBagDetailFromBooking(db, booking)

	okResponse(c, res)
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

	isCanCheckOut, errMessage := checkForCheckOut(booking)

	if !isCanCheckOut {
		response_message.InternalServerError(c, errMessage)
		return
	}

	oldBooking := getBagDetailFromBooking(db, booking.CloneBooking())

	booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
	booking.CheckOutTime = utils.GetTimeNow().Unix()

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//TODO: check lại logic này
	// go updateSinglePaymentOfSubBag(booking, prof)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_CHECK_IN,
		Action:      constants.OP_LOG_ACTION_CHECK_OUT,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}
	go createOperationLog(opLog)

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)

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

	isCanCheckOut, errMessage := checkForCheckOut(bag)

	res := map[string]interface{}{
		"is_can_check_out": isCanCheckOut,
		"message":          errMessage,
	}
	c.JSON(200, res)
}

func (cBooking *CBooking) FinishBill(c *gin.Context, prof models.CmsUser) {
	body := request.FinishBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	today, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	booking := model_booking.Booking{
		Bag:         body.Bag,
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: today,
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	booking.FindFirst(db)

	if booking.BagStatus != constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag chưa check out!")
		return
	}

	RSinglePaymentItem := model_payment.SinglePaymentItem{
		Bag:         body.Bag,
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: today,
	}

	list, _ := RSinglePaymentItem.FindAll(db)

	cashList := []model_payment.SinglePaymentItem{}
	otherList := []model_payment.SinglePaymentItem{}

	for _, item := range list {
		if item.PaymentType == constants.PAYMENT_TYPE_CASH {
			cashList = append(cashList, item)
		} else {
			otherList = append(cashList, item)
		}
	}

	cashTotal := slices.Reduce(cashList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
		return prev + item.Paid
	})

	otherTotal := slices.Reduce(otherList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
		return prev + item.Paid
	})

	if cashTotal != 0 {
		if booking.CustomerUid == "" {
			uid := utils.HashCodeUuid(uuid.New().String())
			customerBody := request.CustomerBody{
				MaKh:      uid,
				TenKh:     booking.CustomerName,
				MaSoThue:  booking.CustomerInfo.Mst,
				DiaChi:    "ddddddd",
				Tk:        "",
				DienThoai: booking.CustomerInfo.Phone,
				Fax:       booking.CustomerInfo.Fax,
				EMail:     booking.CustomerInfo.Email,
				DoiTac:    "",
				NganHang:  "",
				TkNh:      "",
			}

			check, _ := callservices.CreateCustomer(customerBody)
			if check {
				callservices.TransferFast(constants.PAYMENT_TYPE_CASH, cashTotal, "", uid, booking.CustomerName, body.BillNo)
			}
		} else {
			callservices.TransferFast(constants.PAYMENT_TYPE_CASH, cashTotal, "", booking.CustomerUid, booking.CustomerName, body.BillNo)
		}
	}

	if otherTotal != 0 {
		// callservices.TransferFast(constants.PAYMENT_TYPE_CASH, otherTotal, "", booking.CustomerUid, booking.CustomerName)
	}

	go updatePriceForRevenue(booking, body.BillNo)
	okRes(c)
}

func (cBooking *CBooking) LockBill(c *gin.Context, prof models.CmsUser) {
	body := request.LockBill{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	today, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	Rbooking := model_booking.Booking{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		Bag:         body.Bag,
		BookingDate: today,
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	booking, err := Rbooking.FindFirstByUId(db)

	if err != nil {
		response_message.BadRequestDynamicKey(c, "BAG_NOT_FOUND", "")
		return
	}

	oldBooking := getBagDetailFromBooking(db, booking.CloneBooking())

	booking.LockBill = setBoolForCursor(*body.LockBill)
	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Check sub bag
	if booking.SubBags != nil && len(booking.SubBags) > 0 {
		for _, v := range booking.SubBags {
			subBag := model_booking.Booking{}
			subBag.Uid = v.BookingUid
			subBooking, errF := subBag.FindFirstByUId(db)

			if errF == nil {
				subBooking.LockBill = setBoolForCursor(*body.LockBill)
				if err := subBooking.Update(db); err != nil {
					response_message.InternalServerError(c, err.Error())
					return
				}
			}
		}
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_CHECK_IN,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if *body.LockBill == true {
		opLog.Action = constants.OP_LOG_ACTION_LOCK_BAG
	} else {
		opLog.Action = constants.OP_LOG_ACTION_UN_LOCK_BAG
	}

	go createOperationLog(opLog)

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)

	okRes(c)
}

/*
Undo check thi thif check xoá payment đi
*/
func (cBooking *CBooking) UndoCheckIn(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.CheckInBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.Booking{}
	bookingR.Uid = body.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	oldBooking := booking

	if booking.BagStatus != constants.BAG_STATUS_WAITING {
		response_message.InternalServerError(c, "Bag Status is not Waiting")
		return
	}

	if len(booking.MainBags) > 0 {
		response_message.InternalServerError(c, "Bag can not undo checkin")
		return
	}

	if len(booking.SubBags) > 0 {
		response_message.InternalServerError(c, "Bag can not undo checkin")
		return
	}

	// check đk undo checkin in list items
	if booking.BillCode != "" {
		bookingServiceItemsR := model_booking.BookingServiceItemList{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			BillCode:   booking.BillCode,
		}
		db1, _ := bookingServiceItemsR.FindAll(db)
		db1.Not("service_type = ?", constants.CADDIE_SETTING)

		list := []model_booking.BookingServiceItem{}
		db1.Find(&list)

		if len(list) > 0 {
			response_message.InternalServerError(c, "Bag can not undo checkin")
			return
		}

		// Xóa fee caddie booking
		bookingSID := model_booking.BookingServiceItem{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			BillCode:    booking.BillCode,
			ServiceType: constants.CADDIE_SETTING,
		}

		errD := bookingSID.DeleteBatch(db)
		if errD != nil {
			response_message.InternalServerError(c, errD.Error())
			return
		}

	}

	if booking.LockerNo != "" {
		locker := models.Locker{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			Locker:      booking.LockerNo,
			BookingDate: booking.BookingDate,
			BookingUid:  booking.Uid,
		}

		// Find locker
		_ = locker.FindFirst(db)

		errC := locker.Delete(db)
		if errC != nil {
			log.Println("deleteLocker errC", errC.Error())
		}
	}

	pUid := booking.PartnerUid
	cUid := booking.CourseUid
	billCode := booking.BillCode
	bUid := booking.Uid
	agencyId := booking.AgencyId
	bookingCode := booking.BookingCode

	if booking.InitType == constants.BOOKING_INIT_TYPE_CHECKIN {
		if err := booking.Delete(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	} else {
		roundR := models.Round{
			BillCode: booking.BillCode,
		}

		listRound, _ := roundR.FindAll(db)
		if len(listRound) > 1 {
			response_message.InternalServerError(c, "Bag can not undo checkin")
			return
		}

		booking.Bag = ""
		booking.CheckInTime = 0
		booking.CmsUser = prof.UserName
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
		booking.BagStatus = constants.BAG_STATUS_BOOKING
		booking.UndoCheckInTime = utils.GetTimeNow().Unix()

		// reset caddie status
		if booking.CaddieId > 0 {
			udpCaddieOut(db, booking.CaddieId)
		}

		if err := booking.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
		roundR.DeleteByBillCode(db)
	}

	// Xoa payment
	deleteSinglePayment(pUid, cUid, billCode, bUid, agencyId, bookingCode)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_CHECK_IN,
		Action:      constants.OP_LOG_ACTION_UNDO_CHECK_IN,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}
	go createOperationLog(opLog)
	// push socket
	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_UPD_BOOKING_CMS, model_booking.Booking{})

	okRes(c)
}

func (cBooking *CBooking) ResetBag(c *gin.Context, prof models.CmsUser) {
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
		toDayDate, errD := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
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

	if form.BagStatus != "" {
		booking.BagStatus = form.BagStatus
	}

	if form.Bag == "" {
		booking.Bag = form.Bag
	}

	booking.Update(db)
	okRes(c)
}

func (cBooking *CBooking) UndoCheckOut(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.UndoCheckOut{}
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
		toDayDate, errD := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
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

	oldBooking := booking

	booking.BagStatus = booking.LastBookingStatus
	booking.CheckOutTime = 0

	booking.Update(db)

	// push socket
	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_CHECK_IN,
		Action:      constants.OP_LOG_ACTION_UNDO_CHECK_OUT,
		Body:        models.JsonDataLog{Data: form},
		ValueOld:    models.JsonDataLog{Data: oldBooking},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}
	go createOperationLog(opLog)
	okRes(c)
}

func (_ CBooking) updateBagAttachCaddie(db *gorm.DB, bookindUid string) {
	caddieAttach := model_gostarter.BagAttachCaddie{}
	caddieAttach.BookingUid = bookindUid

	if err := caddieAttach.FindFirst(db); err == nil {
		caddieAttach.BagStatus = constants.BAG_ATTACH_CADDIE_WAITING

		_ = caddieAttach.Update(db)
	}
}
