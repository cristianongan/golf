package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// type CBookingOTA struct{}

/*
Booking OTA

Update để đồng bộ với cách lưu trong redis và database:
database mysql đang chia tee_type: 1, 10m, course_type: A,B,C
redis đang lưu teeType: 1A, 1B, 1C,...
*/
func (cBooking *CBooking) CreateBookingOTA(c *gin.Context) {
	dataRes := response.BookingOTARes{}

	body := request.CreateBookingOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = bindErr.Error()
		c.JSON(500, dataRes)
		return
	}

	if body.NumBook <= 0 {
		body.NumBook = 1
	}

	// Find course
	course := models.Course{}
	course.Uid = body.CourseCode
	errFCourse := course.FindFirstHaveKey()
	if errFCourse != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Not found course"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	// Check token
	checkToken := course.ApiKey + body.CourseCode + body.DateStr + body.TeeOffStr + body.BookingCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"
		dataRes.CourseCode = body.CourseCode

		okResponse(c, dataRes)
		return
	}

	prof := models.CmsUser{
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
		UserName:   "ota",
	}

	//convert booking date
	bookDate, errBD := utils.GetBookingTimeFrom(body.DateStr)
	if errBD != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errBD.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	if body.TeeType == "" {
		body.TeeType = "1A"
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	// Convert time
	dateConvert := body.TeeOffStr
	date, errConvert := time.Parse("15:04", dateConvert)
	if errConvert != nil {
		//
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Convert fail"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	dateTeeStrConv := date.Format(constants.HOUR_FORMAT)
	body.TeeOffStr = dateTeeStrConv

	// Check tee time status
	// Check TeeTime Index
	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bookDate, body.CourseCode, dateTeeStrConv, body.TeeType)
	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
	log.Println("CreateBookingOTA rowIndexsRedis", rowIndexsRedis)

	if len(rowIndexsRedis) == constants.SLOT_TEE_TIME {
		//
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Tee is full"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	if len(rowIndexsRedis) > 0 && len(rowIndexsRedis) < body.NumBook {
		//
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Tee khong du"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	// Check agency
	// Find Agency
	agency := models.Agency{
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
		AgencyId:   body.AgentCode,
	}
	errFA := agency.FindFirst(db)
	if errFA != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Not found agency"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	bookingOta := model_booking.BookingOta{
		PartnerUid:   prof.PartnerUid,
		CourseUid:    prof.CourseUid,
		PlayerName:   body.PlayerName,
		Contact:      body.Contact,
		Note:         body.Note,
		NumBook:      body.NumBook,
		Holes:        body.Holes,
		IsMainCourse: body.IsMainCourse,
		Tee:          body.TeeType,
		TeeOffStr:    dateTeeStrConv,

		AgentCode:          body.AgentCode,
		GuestStyle:         body.GuestStyle,
		BookingCodePartner: body.BookingCode,
		EmailConfirm:       body.EmailConfirm,

		CaddieFee: body.CaddieFee,
		BuggyFee:  body.BuggyFee,
		GreenFee:  body.GreenFee,
	}

	// Find booking source
	bookingSource := model_booking.BookingSource{
		PartnerUid: prof.PartnerUid,
		AgencyId:   agency.Id,
	}

	errFindBS := bookingSource.FindFirst(db)
	bookSourceId := ""
	if errFindBS == nil {
		bookSourceId = bookingSource.BookingSourceId
	} else {
		log.Println("CreateBookingOTA errFindBS", errFindBS.Error())
	}

	// Create booking code
	bookingCode := body.BookingCode + "_" + utils.RandomCharNumber(5) + "_" + bookSourceId
	bookingOta.BookingCode = bookingCode

	errCBO := bookingOta.Create(db)
	if errCBO != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errCBO.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	var errCreateBook error
	listBooking := []model_booking.Booking{}

	for i := 0; i < body.NumBook; i++ {
		bodyCreate := request.CreateBookingBody{
			PartnerUid:           prof.PartnerUid,
			CourseUid:            prof.CourseUid,
			BookingDate:          bookDate,
			HoleBooking:          body.Holes,
			CustomerName:         body.PlayerName,
			CustomerBookingName:  body.PlayerName,
			CustomerBookingPhone: body.Contact,
			NoteOfBooking:        body.Note,
			TeeTime:              dateTeeStrConv,
			GuestStyle:           body.GuestStyle,
			BookingOtaId:         bookingOta.Id,
			// RowIndex:             &listIndex[i],
			AgencyId:           agency.Id,
			TeePath:            "MORNING",
			BookingCodePartner: body.BookingCode,
			BookingCode:        bookingOta.BookingCode,
			BookingSourceId:    bookSourceId,
			BookFromOTA:        true,
			FeeInfo: &request.AgencyFeeInfo{
				GolfFee:  body.GreenFee + body.CaddieFee,
				BuggyFee: body.BuggyFee,
			},
		}

		bodyCreate.UpdateTeeType(body.TeeType)

		// if body.Tee == "1" {
		// 	bodyCreate.CourseType = "A"
		// 	bodyCreate.TeeType = "1"
		// }

		// if body.Tee == "10" {
		// 	bodyCreate.CourseType = "B"
		// 	bodyCreate.TeeType = "1"
		// }

		booking, errBook := cBooking.CreateBookingCommon(bodyCreate, nil, prof)
		if booking == nil {
			//error
			log.Println("CreateBookingOTA error", errBook)
		}
		if errBook != nil {
			errCreateBook = errBook

			// Khi booking lỗi thì remove index đã lưu trước đó trong redis
			go removeRowIndexRedis(model_booking.Booking{
				PartnerUid:  bodyCreate.PartnerUid,
				CourseUid:   bodyCreate.CourseUid,
				BookingDate: bodyCreate.BookingDate,
				TeeType:     bodyCreate.TeeType,
				TeeTime:     bodyCreate.TeeTime,
				CourseType:  bodyCreate.CourseType,
				RowIndex:    bodyCreate.RowIndex,
			})
		}
		if booking != nil && errBook == nil {
			listBooking = append(listBooking, *booking)
		}
	}

	if errCreateBook != nil {
		dataRes.Result.Status = 1000
		dataRes.Result.Infor = errCreateBook.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	bodyByte, _ := body.Marshal()
	_ = json.Unmarshal(bodyByte, &dataRes)

	dataRes.Result.Status = http.StatusOK

	dataRes.BookOtaID = bookingOta.BookingCode

	unlockTeeOTA(body)
	if len(listBooking) > 0 {
		go func() {
			cNotification := CNotification{}
			cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_OTA, listBooking)

			// Gen QR
			// Send sms
			genQRCodeListBook(listBooking)
		}()
	}

	okResponse(c, dataRes)
}

/*
Theo TeeType format: 1A, 1B,...
*/
func unlockTeeOTA(body request.CreateBookingOTABody) {
	// get các teetime đang bị khóa ở redis
	bookDate, _ := utils.GetBookingTimeFrom(body.DateStr)
	lockTeeTime := models.LockTeeTime{
		CourseUid: body.CourseCode,
		TeeTime:   body.TeeOffStr,
		DateTime:  bookDate,
	}

	// if body.Tee == "1" {
	// 	lockTeeTime.TeeType = "1A"
	// }

	// if body.Tee == "10" {
	// 	lockTeeTime.TeeType = "1B"
	// }

	listTeeTimeLockRedis := getTeeTimeLockRedis(body.CourseCode, bookDate, lockTeeTime.TeeType)
	for _, teeTimeLockRedis := range listTeeTimeLockRedis {
		if teeTimeLockRedis.TeeTime == body.TeeOffStr && teeTimeLockRedis.DateTime == bookDate &&
			teeTimeLockRedis.CourseUid == body.CourseCode && teeTimeLockRedis.TeeType == lockTeeTime.TeeType {

			teeTimeRedisKey := ""
			// if body.Tee == "1" {
			// 	teeTimeRedisKey = getKeyTeeTimeLockRedis(bookDate, body.CourseCode, body.TeeOffStr, "1A")
			// }

			// if body.Tee == "10" {
			// 	teeTimeRedisKey = getKeyTeeTimeLockRedis(bookDate, body.CourseCode, body.TeeOffStr, "1B")
			// }

			teeTimeRedisKey = getKeyTeeTimeLockRedis(bookDate, body.CourseCode, body.TeeOffStr, body.TeeType)

			err := datasources.DelCacheByKey(teeTimeRedisKey)

			if err != nil {
				log.Println("unlockTee err", err.Error())
			}

			// Bắn socket để client update ui
			// cNotification := CNotification{}
			// cNotification.PushNotificationLockTee(constants.NOTIFICATION_UNLOCK_TEE)
			break
		}
	}
}

/*
Cancel Booking OTA
*/
func (cBooking *CBooking) CancelBookingOTA(c *gin.Context) {
	dataRes := response.CancelBookOTARes{}

	body := request.CancelBookOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = bindErr.Error()
		c.JSON(500, dataRes)
		return
	}

	// Find course
	course := models.Course{}
	course.Uid = body.CourseCode
	errFCourse := course.FindFirstHaveKey()
	if errFCourse != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Not found course"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	// Check token
	checkToken := course.ApiKey + body.CourseCode + body.AgentCode + body.BookingCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"

		okResponse(c, dataRes)
		return
	}

	prof := models.CmsUser{
		PartnerUid: course.PartnerUid,
		CourseUid:  body.CourseCode,
		UserName:   "ota",
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	//Get payment
	bookingOta := model_booking.BookingOta{
		BookingCodePartner: body.BookingCode,
		CourseUid:          body.CourseCode,
	}

	errFindBO := bookingOta.FindFirst(db)
	if errFindBO == nil {
		if body.DeleteBook {
			errDel := bookingOta.Delete(db)
			if errDel != nil {
				log.Println("CancelBookingOTA errDel", errDel.Error())
			}
		} else {
			bookingOta.Status = constants.STATUS_DELETE
			errUdp := bookingOta.Update(db)
			if errUdp != nil {
				log.Println("CancelBookingOTA errUdp", errUdp.Error())
			}
		}
	}

	//Get Bag Booking
	bookR := model_booking.Booking{
		BookingCode: bookingOta.BookingCode,
		PartnerUid:  prof.PartnerUid,
	}

	listBook, errL := bookR.FindAllBookingOTA(db)
	if errL == nil {
		for _, v := range listBook {
			if body.DeleteBook {
				errDel := v.Delete(db)
				if errDel != nil {
					log.Println("CancelBookingOTA Book errDel", errDel.Error())
				} else {
					removeRowIndexRedis(v)
				}
			} else {
				v.BagStatus = constants.BAG_STATUS_CANCEL
				bookCancel := v.CloneBookingDel()
				errCancel := v.Delete(db)
				if errCancel != nil {
					log.Println("CancelBookingOTA Book errCancel", errCancel.Error())
				} else {
					errCBC := bookCancel.Create(db)
					if errCBC != nil {
						log.Println("CancelBookingOTA Book errCBC", errCBC.Error())
					}
					removeRowIndexRedis(v)
				}
			}

			cNotification := CNotification{}
			go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &v)
			go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_UPD_BOOKING_CMS, v)
		}
	}

	dataRes.Result.Status = http.StatusOK
	dataRes.BookingCode = body.BookingCode

	okResponse(c, dataRes)
}
