package controllers

import (
	"fmt"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	if booking.BagStatus == constants.BAG_STATUS_BOOKING {
		// Kiểm tra xem đủ điều kiện cancel booking không
		// cancelBookingSetting := model_booking.CancelBookingSetting{}
		// if err := cancelBookingSetting.ValidateBookingCancel(db, booking); err != nil {
		// 	response_message.InternalServerError(c, err.Error())
		// 	return
		// }
		// Old Booking
		oldBooking := booking

		booking.BagStatus = constants.BAG_STATUS_CANCEL
		booking.CancelNote = body.Note
		booking.CancelBookingTime = utils.GetTimeNow().Unix()
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

		bookingDel := booking.CloneBookingDel()

		errCancel := booking.Delete(db)
		if errCancel != nil {
			response_message.InternalServerError(c, errCancel.Error())
			return
		} else {
			errCreateBDel := bookingDel.Create(db)
			if errCreateBDel != nil {
				log.Println("CancelBooking err", errCreateBDel.Error())
			}
		}

		go func() {
			removeRowIndexRedis(booking)
			// updateSlotTeeTimeWithLock(booking)
			if booking.TeeTime != "" {
				// Lấy số slot đã book
				teeType := fmt.Sprint(booking.TeeType, booking.CourseType)
				teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(booking.BookingDate, booking.CourseUid, booking.TeeTime, teeType)
				rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
				rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
				if len(rowIndexsRedis) < 3 {
					unlockTurnTime(db, booking)
				}
			}
		}()

		//Add log
		opLog := models.OperationLog{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_RECEPTION,
			Function:    constants.OP_LOG_FUNCTION_BOOKING,
			Action:      constants.OP_LOG_ACTION_CANCEL,
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

	bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, body.BookingDate)
	nowStr, _ := utils.GetLocalTimeFromTimeStamp("", constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())
	nowUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, nowStr)

	if bookingDateInt < nowUnix {
		response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
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

	listBookingReadyMoved := []model_booking.Booking{}
	cloneListBooking := []model_booking.Booking{}

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

		// if booking.BagStatus != constants.BAG_STATUS_BOOKING {
		// 	response_message.InternalServerError(c, booking.Uid+" did check in")
		// 	return
		// }
		cloneListBooking = append(cloneListBooking, booking)

		teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(body.BookingDate, booking.CourseUid, body.TeeTime, body.TeeType+body.CourseType)
		rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
		rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)

		if len(rowIndexsRedis) < constants.SLOT_TEE_TIME {
			rowIndex := generateRowIndex(rowIndexsRedis)
			booking.RowIndex = &rowIndex
			rowIndexsRedis = append(rowIndexsRedis, rowIndex)
			rowIndexsRaw, _ := rowIndexsRedis.Value()
			errRedis := datasources.SetCache(teeTimeRowIndexRedis, rowIndexsRaw, 0)
			if errRedis != nil {
				log.Println("CreateBookingCommon errRedis", errRedis)
			}
		} else {
			response_message.BadRequest(c, body.TeeTime+" "+" is Full")
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
		if body.TurnTime != "" {
			booking.TurnTime = body.TurnTime
		}
		if body.TeePath != "" {
			booking.TeePath = body.TeePath
		}

		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(db, true, false)
		if isDuplicated {
			if errDupli != nil {
				// Có lỗi update lại redis
				removeRowIndexRedis(booking)
				response_message.DuplicateRecord(c, errDupli.Error())
				return
			}
			// Có lỗi update lại redis
			removeRowIndexRedis(booking)
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		if body.Hole != 0 {
			booking.Hole = body.Hole
		}

		listBookingReadyMoved = append(listBookingReadyMoved, booking)
	}

	for index, booking := range listBookingReadyMoved {
		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}

		//Add log
		opLog := models.OperationLog{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_RECEPTION,
			Function:    constants.OP_LOG_FUNCTION_BOOKING,
			Action:      constants.OP_LOG_ACTION_MOVE,
			Body:        models.JsonDataLog{Data: body},
			ValueOld:    models.JsonDataLog{Data: cloneListBooking[index]},
			ValueNew:    models.JsonDataLog{Data: booking},
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
			BillCode:    booking.BillCode,
			BookingUid:  booking.Uid,
		}
		go createOperationLog(opLog)
		// go updateSlotTeeTimeWithLock(booking)
	}

	go func() {
		for _, booking := range cloneListBooking {
			removeRowIndexRedis(booking)
			// updateSlotTeeTimeWithLock(booking)
		}
	}()
	okRes(c)
}

func (cBooking *CBooking) CreateBookingTee(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	isAddMore := false

	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for index, body := range bodyRequest.BookingList {
		if body.BookingCode == "" {
			bodyRequest.BookingList[index].BookingCode = bookingCode
			bodyRequest.BookingList[index].BookingTeeTime = true
		} else {
			bodyRequest.BookingList[index].BookingCode = body.BookingCode
			isAddMore = true
		}
	}

	listBooking, err := cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
	if err != nil {
		return
	}

	// khi book restaurant enable thì auto tạo 1 book reservation trong restaurant
	if len(bodyRequest.BookingList) > 0 {
		item := bodyRequest.BookingList[0]
		if item.BookingRestaurant.Enable {
			db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
			go addServiceCart(db, len(bodyRequest.BookingList), item.PartnerUid, item.CourseUid, item.CustomerBookingName, item.CustomerBookingPhone, item.BookingDate, prof.FullName)
		}
	}

	// Log add mỏe booking
	if isAddMore {
		go cBooking.AddLogCreateBatchBooking(c, prof, bodyRequest, listBooking)
	}

	// Bắn socket để client update ui
	go func() {
		cNotification := CNotification{}
		cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, listBooking)
	}()

	//Send Sms
	go genQRCodeListBook(listBooking)

	okResponse(c, listBooking)
}

func (cBooking *CBooking) AddLogCreateBatchBooking(c *gin.Context, prof models.CmsUser, body request.CreateBatchBookingBody, list []model_booking.Booking) {
	for index, booking := range list {
		opLog := models.OperationLog{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_RECEPTION,
			Function:    constants.OP_LOG_FUNCTION_BOOKING,
			Action:      constants.OP_LOG_ACTION_ADD_MORE,
			Body:        models.JsonDataLog{Data: body},
			ValueOld:    models.JsonDataLog{Data: body.BookingList[index]},
			ValueNew:    models.JsonDataLog{Data: booking},
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

// func (cBooking *CBooking) CreateBookingTee(c *gin.Context, prof models.CmsUser) {
// 	bodyRequest := request.CreateBatchBookingBody{}
// 	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
// 		badRequest(c, bindErr.Error())
// 		return
// 	}

// 	if len(bodyRequest.BookingList) == 0 {
// 		badRequest(c, "CreateBookingTee bindErr.Error()")
// 		return
// 	}

// 	bookingCode := utils.HashCodeUuid(uuid.New().String())
// 	for index := range bodyRequest.BookingList {
// 		bodyRequest.BookingList[index].BookingCode = bookingCode
// 		bodyRequest.BookingList[index].BookingTeeTime = true
// 	}

// 	list := []map[string]interface{}{}
// 	bodyErroList := request.ListCreateBookingBody{}

// 	for _, body := range bodyRequest.BookingList {
// 		bUid := body.CourseUid + "-" + utils.HashCodeUuid(uuid.New().String())
// 		booking, errCreate := cBooking.AssignBooking(body, bUid, c, prof)

// 		// Đánh dấu lại body đã xử lý,
// 		bodyErroList = append(bodyErroList, body)

// 		if errCreate != nil {
// 			for _, body := range bodyErroList {
// 				removeRowIndexRedis(model_booking.Booking{
// 					PartnerUid:  body.PartnerUid,
// 					CourseUid:   body.CourseUid,
// 					BookingDate: body.BookingDate,
// 					TeeType:     body.TeeType,
// 					TeeTime:     body.TeeTime,
// 					CourseType:  body.CourseType,
// 					RowIndex:    body.RowIndex,
// 				})
// 			}
// 			log.Println("CreateBookingTee ", errCreate.Error())
// 			return
// 		}

// 		if booking != nil {
// 			item := map[string]interface{}{
// 				"booking": booking,
// 				"uid":     bUid,
// 			}
// 			list = append(list, item)
// 		}
// 	}

// 	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
// 	listBooking := []model_booking.Booking{}
// 	for idx, item := range list {
// 		booking := item["booking"].(*model_booking.Booking)
// 		errCreate := booking.Create(db, item["uid"].(string))

// 		if errCreate != nil {
// 			response_message.InternalServerError(c, errCreate.Error())
// 			return
// 		}

// 		if booking != nil {
// 			response_message.InternalServerError(c, "CreateBookingTee booking nil")
// 			return
// 		}

// 		listBooking = append(listBooking, *booking)

// 		go func() {
// 			if booking.InitType == constants.BOOKING_INIT_TYPE_CHECKIN && booking.CustomerUid != "" {
// 				go updateReportTotalPlayCountForCustomerUser(booking.CustomerUid, booking.CardId, booking.PartnerUid, booking.CourseUid)
// 			}

// 			// Create booking payment
// 			if booking.AgencyId > 0 {
// 				go handleAgencyPaid(*booking, bodyRequest.BookingList[idx].FeeInfo)
// 			}

// 			if booking.AgencyId > 0 && booking.MemberCardUid == "" {
// 				go handleAgencyPayment(db, *booking)
// 				// Tạo thêm single payment cho bag

// 			} else {
// 				if booking.BagStatus == constants.BAG_STATUS_WAITING {
// 					// checkin mới tạo payment
// 					go handleSinglePayment(db, *booking)
// 				}
// 			}
// 		}()
// 	}

// 	bodyFirst := bodyRequest.BookingList[0]
// 	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bodyFirst.BookingDate, bodyFirst.CourseUid, bodyFirst.TeeTime, bodyFirst.TeeType+bodyFirst.CourseType)
// 	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
// 	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
// 	if bodyFirst.TeeTime != "" && len(rowIndexsRedis) >= 3 {
// 		cLockTeeTime := CLockTeeTime{}
// 		lockTurn := request.CreateLockTurn{
// 			BookingDate: bodyFirst.BookingDate,
// 			CourseUid:   bodyFirst.CourseUid,
// 			PartnerUid:  bodyFirst.PartnerUid,
// 			TeeTime:     bodyFirst.TeeTime,
// 			TeeType:     bodyFirst.TeeType,
// 			CourseType:  bodyFirst.CourseType,
// 		}
// 		go cLockTeeTime.LockTurn(lockTurn, bodyFirst.Hole, c, prof)
// 	}

// 	// khi book restaurant enable thì auto tạo 1 book reservation trong restaurant
// 	if len(bodyRequest.BookingList) > 0 {
// 		item := bodyRequest.BookingList[0]
// 		if item.BookingRestaurant.Enable {
// 			db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
// 			go addServiceCart(db, len(bodyRequest.BookingList), item.PartnerUid, item.CourseUid, item.CustomerBookingName, item.CustomerBookingPhone, item.BookingDate, prof.FullName)
// 		}
// 	}

// 	// Bắn socket để client update ui
// 	go func() {
// 		cNotification := CNotification{}
// 		cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, listBooking)
// 	}()

// 	okResponse(c, listBooking)
// }

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
	listBooking, _ := cBooking.CreateBatch(bodyRequest.BookingList, c, prof)

	// //Add Log
	// go func() {
	// 	for _, booking := range listBooking {
	// 		opLog := models.OperationLog{
	// 			PartnerUid:  booking.PartnerUid,
	// 			CourseUid:   booking.CourseUid,
	// 			UserName:    prof.UserName,
	// 			UserUid:     prof.Uid,
	// 			Module:      constants.OP_LOG_MODULE_RECEPTION,
	// 			Function:    constants.OP_LOG_FUNCTION_BOOKING,
	// 			Action:      constants.OP_LOG_ACTION_COPY,
	// 			Body:        models.JsonDataLog{Data: bodyRequest},
	// 			ValueOld:    models.JsonDataLog{},
	// 			ValueNew:    models.JsonDataLog{Data: booking},
	// 			Path:        c.Request.URL.Path,
	// 			Method:      c.Request.Method,
	// 			Bag:         booking.Bag,
	// 			BookingDate: booking.BookingDate,
	// 			BillCode:    booking.BillCode,
	// 			BookingUid:  booking.Uid,
	// 		}
	// 		go createOperationLog(opLog)
	// 	}
	// }()

	okResponse(c, listBooking)
}

func (_ *CBooking) CancelAllBooking(c *gin.Context, prof models.CmsUser) {
	db1 := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		CourseType:  form.CourseType,
		TeeType:     form.TeeType,
	}

	db, _, err := bookingR.FindAllBookingNotCancelList(db1)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	for _, booking := range list {
		if booking.BagStatus == constants.BAG_STATUS_BOOKING {
			oldBooking := booking

			booking.BagStatus = constants.BAG_STATUS_CANCEL
			booking.CancelNote = form.Reason
			booking.CancelBookingTime = utils.GetTimeNow().Unix()
			booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

			bookDel := booking.CloneBookingDel()

			errCancel := booking.Delete(db1)
			if errCancel != nil {
				response_message.InternalServerError(c, errCancel.Error())
				return
			} else {
				errCreateDel := bookDel.Create(db1)
				if errCreateDel != nil {
					log.Println("CancelAllBooking err", errCreateDel.Error())
				}
			}

			//Add log
			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_RECEPTION,
				Function:    constants.OP_LOG_FUNCTION_BOOKING,
				Action:      constants.OP_LOG_ACTION_CANCEL_ALL,
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
		}
	}

	go func() {
		for _, booking := range list {
			removeRowIndexRedis(booking)
			// updateSlotTeeTimeWithLock(booking)
		}
	}()
	okRes(c)
}

func (cBooking CBooking) CreateBatch(bookingList request.ListCreateBookingBody, c *gin.Context, prof models.CmsUser) ([]model_booking.Booking, error) {
	list := []model_booking.Booking{}
	for _, body := range bookingList {
		booking, errCreate := cBooking.CreateBookingCommon(body, c, prof)
		if errCreate != nil {
			return list, errCreate
		}

		if booking != nil {
			list = append(list, *booking)
		}

		//Add log
		opLog := models.OperationLog{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_RECEPTION,
			Function:    constants.OP_LOG_FUNCTION_BOOKING,
			Action:      constants.OP_LOG_ACTION_CREATE,
			Body:        models.JsonDataLog{Data: body},
			ValueOld:    models.JsonDataLog{},
			ValueNew:    models.JsonDataLog{Data: booking},
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
			BillCode:    booking.BillCode,
			BookingUid:  booking.Uid,
		}
		go createOperationLog(opLog)
	}
	return list, nil
}

// func (cBooking CBooking) AssignBooking(body request.CreateBookingBody, bUid string, c *gin.Context, prof models.CmsUser) (*model_booking.Booking, error) {
// 	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
// 	// validate caddie_code

// 	_, errDate := time.Parse(constants.DATE_FORMAT_1, body.BookingDate)
// 	if errDate != nil {
// 		response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
// 		return nil, errDate
// 	}

// 	if checkBookingAtOTAPosition(body) && !body.BookFromOTA {
// 		response_message.ErrorResponse(c, http.StatusBadRequest, "", "Booking Online đang khóa tại tee time này!", constants.ERROR_BOOKING_OTA_LOCK)
// 		return nil, nil
// 	}

// 	var caddie models.Caddie
// 	var err error
// 	if body.CaddieCode != "" {
// 		caddie, err = cBooking.validateCaddie(db, prof.CourseUid, body.CaddieCode)
// 		if err != nil {
// 			response_message.InternalServerError(c, err.Error())
// 			return nil, err
// 		}

// 	}

// 	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(body.BookingDate, body.CourseUid, body.TeeTime, body.TeeType+body.CourseType)

// 	rowIndexsRedisStr := ""

// 	addRejectedHandler := func(_ *datasources.Locker) error {
// 		rowIndexsRedisStr, _ = datasources.GetCache(teeTimeRowIndexRedis)
// 		return nil
// 	}

// 	keyRedisLockTee := fmt.Sprintf("redisLock_%s", teeTimeRowIndexRedis)

// 	errLock := datasources.Lock(datasources.LockOption{
// 		Key:     keyRedisLockTee,
// 		Ttl:     1 * time.Second,
// 		Handler: addRejectedHandler,
// 	})
// 	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)

// 	if errLock != nil {
// 		return nil, errLock
// 	}

// 	if len(rowIndexsRedis) < constants.SLOT_TEE_TIME {
// 		if body.RowIndex == nil {
// 			rowIndex := generateRowIndex(rowIndexsRedis)
// 			body.RowIndex = &rowIndex
// 		}

// 		rowIndexsRedis = append(rowIndexsRedis, *body.RowIndex)
// 		rowIndexsRaw, _ := rowIndexsRedis.Value()
// 		errRedis := datasources.SetCache(teeTimeRowIndexRedis, rowIndexsRaw, 0)
// 		if errRedis != nil {
// 			log.Println("CreateBookingCommon errRedis", errRedis)
// 		}
// 	}

// 	if !body.IsCheckIn {
// 		teePartList := []string{"MORNING", "NOON", "NIGHT"}

// 		if !checkStringInArray(teePartList, body.TeePath) {
// 			response_message.BadRequest(c, "Tee Part not in (MORNING, NOON, NIGHT)")
// 			return nil, errors.New("Tee Part not in (MORNING, NOON, NIGHT)")
// 		}
// 	}

// 	booking := model_booking.Booking{
// 		PartnerUid:         body.PartnerUid,
// 		CourseUid:          body.CourseUid,
// 		TeeType:            body.TeeType,
// 		TeePath:            body.TeePath,
// 		TeeTime:            body.TeeTime,
// 		TeeOffTime:         body.TeeTime,
// 		TurnTime:           body.TurnTime,
// 		RowIndex:           body.RowIndex,
// 		CmsUser:            prof.UserName,
// 		Hole:               body.Hole,
// 		HoleBooking:        body.Hole,
// 		BookingRestaurant:  body.BookingRestaurant,
// 		BookingRetal:       body.BookingRetal,
// 		BookingCode:        body.BookingCode,
// 		CourseType:         body.CourseType,
// 		NoteOfBooking:      body.NoteOfBooking,
// 		BookingCodePartner: body.BookingCodePartner,
// 		BookingSourceId:    body.BookingSourceId,
// 		AgencyPaidAll:      body.AgencyPaidAll,
// 	}

// 	// Check Guest of member, check member có còn slot đi cùng không
// 	var memberCard models.MemberCard
// 	guestStyle := ""

// 	if body.MemberUidOfGuest != "" && body.GuestStyle != "" {
// 		var errCheckMember error
// 		customerName := ""
// 		errCheckMember, memberCard, customerName = handleCheckMemberCardOfGuest(db, body.MemberUidOfGuest, body.GuestStyle)
// 		if errCheckMember != nil {
// 			response_message.InternalServerError(c, errCheckMember.Error())
// 			return nil, errCheckMember
// 		} else {
// 			booking.MemberUidOfGuest = body.MemberUidOfGuest
// 			booking.MemberNameOfGuest = customerName
// 		}

// 		if memberCard.Status == constants.STATUS_DISABLE {
// 			response_message.BadRequestDynamicKey(c, "MEMBER_CARD_INACTIVE", "")
// 			return nil, nil
// 		}
// 	}

// 	// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

// 	if body.Bag != "" {
// 		booking.Bag = body.Bag
// 	}

// 	if body.BookingDate != "" {
// 		// bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, body.BookingDate)
// 		// nowStr, _ := utils.GetLocalTimeFromTimeStamp("", constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())
// 		// nowUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, nowStr)

// 		// if bookingDateInt < nowUnix {
// 		// 	response_message.BadRequest(c, constants.BOOKING_DATE_NOT_VALID)
// 		// 	return nil, errors.New(constants.BOOKING_DATE_NOT_VALID)
// 		// }
// 		booking.BookingDate = body.BookingDate
// 	} else {
// 		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
// 		if errDate == nil {
// 			booking.BookingDate = dateDisplay
// 		} else {
// 			log.Println("booking date display err ", errDate.Error())
// 		}
// 	}

// 	//Check duplicated
// 	isDuplicated, _ := booking.IsDuplicated(db, true, true)
// 	if isDuplicated {
// 		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
// 		return nil, errors.New(constants.API_ERR_DUPLICATED_RECORD)
// 	}

// 	// Booking Uid
// 	bookingUid := uuid.New()
// 	booking.BillCode = utils.HashCodeUuid(string(bookingUid.String()))

// 	// Checkin Time
// 	checkInTime := utils.GetTimeNow().Unix()

// 	// Member Card
// 	// Check xem booking guest hay booking member
// 	if body.MemberCardUid != "" {
// 		// Get config course
// 		memberCardBody := request.UpdateAgencyOrMemberCardToBooking{
// 			PartnerUid:    body.PartnerUid,
// 			CourseUid:     body.CourseUid,
// 			AgencyId:      body.AgencyId,
// 			BUid:          bUid,
// 			CustomerName:  body.CustomerName,
// 			Hole:          body.Hole,
// 			MemberCardUid: body.MemberCardUid,
// 		}

// 		memberCard := models.MemberCard{}
// 		if errUpdate := cBooking.updateMemberCardToBooking(c, db, &booking, &memberCard, memberCardBody); errUpdate != nil {
// 			return nil, errUpdate
// 		}
// 		guestStyle = memberCard.GetGuestStyle(db)
// 	} else {
// 		booking.CustomerName = body.CustomerName
// 	}

// 	//Agency id
// 	if body.AgencyId > 0 {
// 		agencyBody := request.UpdateAgencyOrMemberCardToBooking{
// 			PartnerUid:   body.PartnerUid,
// 			CourseUid:    body.CourseUid,
// 			AgencyId:     body.AgencyId,
// 			BUid:         bUid,
// 			CustomerName: body.CustomerName,
// 			Hole:         body.Hole,
// 		}
// 		agency := models.Agency{}
// 		if errAgency := cBooking.updateAgencyForBooking(db, &booking, &agency, agencyBody); errAgency != nil {
// 			response_message.BadRequest(c, errAgency.Error())
// 			return nil, errAgency
// 		}
// 		guestStyle = agency.GuestStyle
// 	}

// 	// Có thông tin khách hàng
// 	/*
// 		Chọn khách hàng từ agency
// 	*/
// 	if body.CustomerUid != "" {
// 		//check customer
// 		customer := models.CustomerUser{}
// 		customer.Uid = body.CustomerUid
// 		errFindCus := customer.FindFirst(db)
// 		if errFindCus != nil || customer.Uid == "" {
// 			response_message.BadRequest(c, "customer"+errFindCus.Error())
// 			return nil, errFindCus
// 		}

// 		booking.CustomerName = customer.Name
// 		booking.CustomerType = customer.Type
// 		booking.CustomerInfo = cloneToCustomerBooking(customer)
// 		booking.CustomerUid = body.CustomerUid
// 	}

// 	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

// 	// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
// 	if body.GuestStyle != "" && guestStyle != body.GuestStyle {
// 		guestStyle = body.GuestStyle
// 	}

// 	// GuestStyle
// 	if guestStyle != "" {

// 		guestBody := request.UpdateAgencyOrMemberCardToBooking{
// 			PartnerUid:   body.PartnerUid,
// 			CourseUid:    body.CourseUid,
// 			AgencyId:     body.AgencyId,
// 			BUid:         bUid,
// 			CustomerName: body.CustomerName,
// 			Hole:         body.Hole,
// 		}

// 		if errUpdGs := cBooking.updateGuestStyleToBooking(c, guestStyle, db, &booking, guestBody); errUpdGs != nil {
// 			return nil, errUpdGs
// 		}
// 	}

// 	// Check In Out
// 	if body.IsCheckIn {
// 		// Tạo booking check in luôn
// 		booking.BagStatus = constants.BAG_STATUS_WAITING
// 		booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
// 		booking.CheckInTime = checkInTime
// 	} else {
// 		// Tạo booking
// 		booking.BagStatus = constants.BAG_STATUS_BOOKING
// 		booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
// 	}

// 	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT

// 	// Update caddie
// 	if body.CaddieCode != "" {
// 		// cBooking.UpdateBookingCaddieCommon(db, body.PartnerUid, body.CourseUid, &booking, caddie)
// 	}

// 	if body.CustomerName != "" {
// 		booking.CustomerName = body.CustomerName
// 	}

// 	if body.LockerNo != "" {
// 		booking.LockerNo = body.LockerNo
// 		go createLocker(db, booking)
// 	}

// 	if body.ReportNo != "" {
// 		booking.ReportNo = body.ReportNo
// 	}

// 	if body.CustomerIdentify != "" && booking.CustomerInfo.Uid == "" {
// 		customer := models.CustomerUser{}
// 		customer.Identify = body.CustomerIdentify
// 		customer.Phone = body.CustomerBookingPhone
// 		customer.Nationality = body.Nationality
// 		booking.CustomerInfo = cloneToCustomerBooking(customer)
// 	}

// 	if body.CustomerBookingName != "" {
// 		booking.CustomerBookingName = body.CustomerBookingName
// 	} else {
// 		booking.CustomerBookingName = booking.CustomerName
// 	}

// 	if body.CustomerBookingPhone != "" {
// 		booking.CustomerBookingPhone = body.CustomerBookingPhone
// 	} else {
// 		booking.CustomerBookingPhone = booking.CustomerInfo.Phone
// 	}

// 	if body.BookingCode == "" {
// 		bookingCode := utils.HashCodeUuid(bookingUid.String())
// 		booking.BookingCode = bookingCode
// 	} else {
// 		booking.BookingCode = body.BookingCode
// 	}

// 	if body.IsPrivateBuggy != nil {
// 		booking.IsPrivateBuggy = body.IsPrivateBuggy
// 	}

// 	return &booking, nil
// }

func (cBooking *CBooking) SendInforGuest(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	body := request.SendInforGuestBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	var listBooking []model_booking.Booking

	for _, item := range body.ListBooking {
		// Update booking
		booking := model_booking.Booking{}
		booking.Uid = item.Uid

		if err := booking.FindFirst(db); err != nil {
			response_message.BadRequestFreeMessage(c, "Booking not found")
			return
		}

		booking.CustomerName = item.CustomerName
		booking.CustomerBookingEmail = item.CustomerBookingEmail
		booking.CustomerBookingPhone = item.CustomerBookingPhone

		listBooking = append(listBooking, booking)
	}

	if len(listBooking) > 0 {
		// Update list booking
		go updateListBooking(db, listBooking)

		// Send email
		if body.SendMethod == constants.SEND_INFOR_GUEST_BOTH || body.SendMethod == constants.SEND_INFOR_GUEST_EMAIL {
			go sendEmailBooking(listBooking, body.ListBooking[0].CustomerBookingEmail)
		}

		// Send sms
		if body.SendMethod == constants.SEND_INFOR_GUEST_BOTH || body.SendMethod == constants.SEND_INFOR_GUEST_SMS {
			go sendSmsBooking(listBooking, body.ListBooking[0].CustomerBookingPhone)
		}

		// Add log
		go addLogSendInforGuest(db, listBooking, prof, body.SendMethod)

	}

	okRes(c)
}

func updateListBooking(db *gorm.DB, listBooking []model_booking.Booking) {
	for _, booking := range listBooking {
		// Update booking
		if err := booking.FindFirst(db); err != nil {
			log.Println("Update list booking err", err.Error())
		}
	}
}

func addLogSendInforGuest(db *gorm.DB, listBooking []model_booking.Booking, prof models.CmsUser, method string) {
	// Booking Infor
	bookingInfor := listBooking[0]

	// Add log send infỏ guest
	logInfor := model_booking.SendInforGuest{
		PartnerUid:     prof.PartnerUid,
		CourseUid:      prof.CourseUid,
		BookingCode:    bookingInfor.BookingCode,
		BookingDate:    bookingInfor.BookingDate,
		BookingName:    bookingInfor.CustomerBookingName,
		GuestStyle:     bookingInfor.GuestStyle,
		GuestStyleName: bookingInfor.GuestStyleName,
		NumberPeople:   len(listBooking),
		SendMethod:     method,
		PhoneNumber:    bookingInfor.CustomerBookingPhone,
		Email:          bookingInfor.CustomerBookingEmail,
		CmsUser:        prof.UserName,
	}

	if err := logInfor.Create(db); err != nil {
		log.Println("Create send infor guest err", err.Error())
	}
}
