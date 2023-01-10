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
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils"
	"start/utils/response_message"
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

	if checkBookingOTA(body) && !body.BookFromOTA {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", "Booking Online đang khóa tại tee time này!", constants.ERROR_BOOKING_OTA_LOCK)
		return nil, nil
	}

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
	// teeList := []string{constants.TEE_TYPE_1, constants.TEE_TYPE_1A, constants.TEE_TYPE_1B, constants.TEE_TYPE_1C}
	// if utils.Contains(teeList, body.TeeType) {
	// 	cBookingSetting := CBookingSetting{}
	// 	if errors := cBookingSetting.ValidateClose1ST(db, body.BookingDate, body.PartnerUid, body.CourseUid); errors != nil {
	// 		response_message.InternalServerError(c, errors.Error())
	// 		return nil, errors
	// 	}
	// }

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
		AgencyPaidAll:      body.AgencyPaidAll,
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

		if memberCard.Status == constants.STATUS_DISABLE {
			response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
			return nil, nil
		}
	}

	// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

	if body.Bag != "" {
		booking.Bag = body.Bag
	}

	if body.BookingDate != "" {
		bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, body.BookingDate)
		nowStr, _ := utils.GetLocalTimeFromTimeStamp("", constants.DATE_FORMAT_1, time.Now().Unix())
		nowUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, nowStr)

		if bookingDateInt < nowUnix {
			response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
			return nil, errors.New(constants.BOOKING_DATE_NOT_VALID)
		}
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
		errCourse := course.FindFirst()
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

		if memberCard.Status == constants.STATUS_DISABLE {
			response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
			return nil, nil
		}

		if memberCard.AnnualType == constants.ANNUAL_TYPE_SLEEP {
			response_message.BadRequestDynamicKey(c, "ANNUAL_TYPE_SLEEP_NOT_CHECKIN", "")
			return nil, nil
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner(db)
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return nil, errOwner
		}

		// Get Member Card Type
		memberCardType := models.MemberCardType{}
		memberCardType.Id = memberCard.McTypeId
		errMCTypeFind := memberCardType.FindFirst(db)
		if errMCTypeFind == nil && memberCardType.AnnualType == constants.ANNUAL_TYPE_LIMITED {
			// Validate số lượt chơi còn lại của memeber
			reportCustomer := model_report.ReportCustomerPlay{
				CustomerUid: owner.Uid,
			}

			if errF := reportCustomer.FindFirst(); errF == nil {
				playCountRemain := memberCard.AdjustPlayCount - reportCustomer.TotalPlayCount
				if playCountRemain <= 0 {
					response_message.ErrorResponse(c, http.StatusBadRequest, "PLAY_COUNT_INVALID", "", constants.ERROR_PLAY_COUNT_INVALID)
					return nil, errF
				}
			}
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
			initListRound(db, booking, bookingGolfFee)

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
		errCourse := course.FindFirst()
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
		booking.CustomerType = agency.Type

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
				initListRound(db, booking, bookingGolfFee)

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

		if booking.CustomerType == "" {
			booking.CustomerType = golfFeeModel.CustomerType
		}

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
			initListRound(db, booking, bookingGolfFee)
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

	if body.CustomerName != "" {
		booking.CustomerName = body.CustomerName
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

	if body.IsPrivateBuggy != nil {
		booking.IsPrivateBuggy = body.IsPrivateBuggy
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
		teeType := fmt.Sprint(body.TeeType, body.CourseType)
		lockTurn := request.CreateLockTurn{
			BookingDate: body.BookingDate,
			CourseUid:   body.CourseUid,
			PartnerUid:  body.PartnerUid,
			TeeTime:     body.TeeTime,
			TeeType:     teeType,
		}
		cLockTeeTime.LockTurn(lockTurn, c, prof)
	}

	if body.IsCheckIn && booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.CardId, booking.PartnerUid, booking.CourseUid)
	}

	// Create booking payment
	if booking.AgencyId > 0 {
		go handleAgencyPaid(booking, body.FeeInfo)
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

	// if !body.BookFromOTA {
	// 	go updateSlotTeeTimeWithLock(booking)
	// }

	go func() {
		// Bắn socket để client update ui
		if !body.BookFromOTA {
			cNotification := CNotification{}
			cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, booking)
		}
	}()

	return &booking, nil
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

func checkBookingOTA(body request.CreateBookingBody) bool {
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

	if body.Hole > 0 {
		booking.HoleBooking = body.Hole
		booking.Hole = body.Hole
	}

	if body.CourseType != "" {
		booking.CourseType = body.CourseType
	}

	if body.GuestStyle != "" {
		booking.GuestStyle = body.GuestStyle
	}

	//Upd Main Pay for Sub
	isMainBagPayChanged := false
	if body.MainBagPay != nil {
		if !reflect.DeepEqual(booking.MainBagPay, body.MainBagPay) {
			booking.MainBagPay = body.MainBagPay
			isMainBagPayChanged = true
			go bookMarkRoundPaidByMainBag(booking, db)
		}
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
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
			initListRound(db, booking, bookingGolfFee)

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
		errCourse := course.FindFirst()
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
				initListRound(db, booking, bookingGolfFee)

				booking.SeparatePrice = true
				body.GuestStyle = agency.GuestStyle
			} else {
				body.GuestStyle = agency.GuestStyle
			}
		}
	}
	// GuestStyle
	if body.GuestStyle != "" && booking.GuestStyle != body.GuestStyle {
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
			initListRound(db, booking, bookingGolfFee)
		}
	}
	//Find Booking Code
	// list, _ := booking.FindListWithBookingCode(db)
	// if len(list) == 1 {
	// 	booking.CustomerBookingName = booking.CustomerName
	// 	booking.CustomerBookingPhone = booking.CustomerInfo.Phone
	// }

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
	if body.CaddieCode != "" && booking.CaddieInfo.Code != body.CaddieCode {
		cBooking.UpdateBookingCaddieCommon(db, body.PartnerUid, body.CourseUid, &booking, caddie)
	} else {
		if booking.CaddieId > 0 && body.CaddieCode == "" {
			booking.CaddieId = 0
			booking.CaddieInfo = model_booking.BookingCaddie{}
			booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT
			booking.HasBookCaddie = false
		}
	}

	// Create booking payment
	if booking.AgencyId > 0 {
		if validateAgencyFeeBeforUpdate(booking, body.FeeInfo) {
			go handleAgencyPaid(booking, body.FeeInfo)
		}
	}

	// Update các thông tin khác trước
	errUdpBook := booking.Update(db)
	if errUdpBook != nil {
		response_message.InternalServerError(c, errUdpBook.Error())
		return
	}

	// udp ok -> Tính lại giá
	if isMainBagPayChanged {
		updatePriceWithServiceItem(booking, prof)
	}

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
	// caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
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

	checkInTime := time.Now().Unix()

	if body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId {
		booking.SeparatePrice = false
	}

	if body.MemberCardUid != "" && (body.MemberCardUid != booking.MemberCardUid ||
		body.AgencyId != booking.AgencyId || body.Hole != booking.Hole) {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		if memberCard.Status == constants.STATUS_DISABLE {
			response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
			return
		}

		if memberCard.AnnualType == constants.ANNUAL_TYPE_SLEEP {
			response_message.BadRequestDynamicKey(c, "ANNUAL_TYPE_SLEEP_NOT_CHECKIN", "")
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner(db)
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return
		}

		// Get Member Card Type
		memberCardType := models.MemberCardType{}
		memberCardType.Id = memberCard.McTypeId
		errMCTypeFind := memberCardType.FindFirst(db)
		if errMCTypeFind == nil && memberCardType.AnnualType == constants.ANNUAL_TYPE_LIMITED {
			// Validate số lượt chơi còn lại của memeber
			reportCustomer := model_report.ReportCustomerPlay{
				CustomerUid: owner.Uid,
			}

			if errF := reportCustomer.FindFirst(); errF == nil {
				playCountRemain := memberCard.AdjustPlayCount - reportCustomer.TotalPlayCount
				if playCountRemain <= 0 {
					response_message.ErrorResponse(c, http.StatusBadRequest, "PLAY_COUNT_INVALID", "", constants.ERROR_PLAY_COUNT_INVALID)
					return
				}
			}
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
			initPriceForBooking(db, &booking, listBookingGolfFee, bookingGolfFee, booking.CheckInTime)
			initListRound(db, booking, bookingGolfFee)
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
	if body.AgencyId > 0 && (body.AgencyId != booking.AgencyId || body.Hole != booking.Hole) {
		// Get config course
		course := models.Course{}
		course.Uid = booking.CourseUid
		errCourse := course.FindFirst()
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
				initListRound(db, booking, bookingGolfFee)
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
			initListRound(db, booking, bookingGolfFee)
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

	// Create booking payment
	if booking.AgencyId > 0 {
		if validateAgencyFeeBeforUpdate(booking, body.FeeInfo) {
			go handleAgencyPaid(booking, body.FeeInfo)
		}
	}

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	if body.MemberUidOfGuest != "" && body.GuestStyle != "" && memberCard.Uid != "" {
		go updateMemberCard(db, memberCard)
	}

	if booking.CustomerUid != "" {
		go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.CardId, booking.PartnerUid, booking.CourseUid)
	}

	// Create payment info
	go handlePayment(db, booking)

	// Update lại round còn thiếu bag
	cRound := CRound{}
	go cRound.UpdateBag(booking, db)

	res := getBagDetailFromBooking(db, booking)

	okResponse(c, res)
}

func validateAgencyFeeBeforUpdate(booking model_booking.Booking, feeInfo request.AgencyFeeInfo) bool {
	if len(booking.AgencyPaid) > 0 && feeInfo.GolfFee > 0 && booking.AgencyPaid[0].Fee != feeInfo.GolfFee {
		return true
	}
	if len(booking.AgencyPaid) > 1 && feeInfo.BuggyFee > 0 && booking.AgencyPaid[1].Fee != feeInfo.BuggyFee {
		return true
	}
	if len(booking.AgencyPaid) > 2 && feeInfo.CaddieFee > 0 && booking.AgencyPaid[2].Fee != feeInfo.CaddieFee {
		return true
	}
	return false
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

	// delete tee time locked theo booking date
	// if booking.TeeTime != "" {
	// 	go unlockTurnTime(db, booking)
	// }

	okResponse(c, booking)
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
						if serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH || serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE {
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

func (cBooking *CBooking) FinishBill(c *gin.Context, prof models.CmsUser) {
	body := request.FinishBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	today, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

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

	go updatePriceForRevenue(db, booking, body.BillNo)
	okRes(c)
}

func (cBooking *CBooking) LockBill(c *gin.Context, prof models.CmsUser) {
	body := request.LockBill{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	today, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

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

	booking.LockBill = setBoolForCursor(*body.LockBill)
	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
