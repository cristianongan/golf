package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

		booking.BagStatus = constants.BAG_STATUS_CANCEL
		booking.CancelNote = body.Note
		booking.CancelBookingTime = time.Now().Unix()
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}

		go func() {
			removeRowIndexRedis(booking)
			// updateSlotTeeTimeWithLock(booking)
			if booking.TeeTime != "" {
				unlockTurnTime(db, booking)
			}
		}()
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
	nowStr, _ := utils.GetLocalTimeFromTimeStamp("", constants.DATE_FORMAT_1, time.Now().Unix())
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

	for _, booking := range listBookingReadyMoved {
		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
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

	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for index := range bodyRequest.BookingList {
		bodyRequest.BookingList[index].BookingCode = bookingCode
		bodyRequest.BookingList[index].BookingTeeTime = true
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

	// Bắn socket để client update ui
	cNotification := CNotification{}
	cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, listBooking)
	okResponse(c, listBooking)
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
	listBooking, _ := cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
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

	db, _, err := bookingR.FindAllBookingList(db1)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	for _, booking := range list {
		if booking.BagStatus == constants.BAG_STATUS_BOOKING {
			booking.BagStatus = constants.BAG_STATUS_CANCEL
			booking.CancelNote = form.Reason
			booking.CancelBookingTime = time.Now().Unix()
			booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

			errUdp := booking.Update(db1)
			if errUdp != nil {
				response_message.InternalServerError(c, errUdp.Error())
				return
			}
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
	}
	return list, nil
}
