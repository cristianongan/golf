package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CCourseOperating struct{}

/*
Danh sách booking for caddie on course
Role: Booking đã checkin, chưa checkout và chưa out Caddies
*/
func (_ *CCourseOperating) GetListBookingCaddieOnCourse(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetBookingForCaddieOnCourseForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// TODO: filter by date

	bookingR := model_booking.Booking{
		PartnerUid:   form.PartnerUid,
		CourseUid:    form.CourseUid,
		BookingDate:  form.BookingDate,
		BuggyId:      form.BuggyId,
		CaddieId:     form.CaddieId,
		Bag:          form.Bag,
		CustomerName: form.PlayerName,
	}

	list := bookingR.FindForCaddieOnCourse(db, form.InFlight)

	okResponse(c, list)
}

/*
Add Caddie short
Chưa tạo flight
*/
func (_ *CCourseOperating) AddCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	_, errDate := time.Parse(constants.DATE_FORMAT_1, body.BookingDate)
	if errDate != nil {
		response_message.BadRequest(c, "Booking Date format invalid!")
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" || body.BookingDate == "" || body.Bag == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check can add
	errB, response := addCaddieBuggyToBooking(db, body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode, body.IsPrivateBuggy)

	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
		return
	}

	booking := response.Booking
	caddie := response.NewCaddie
	buggy := response.NewBuggy

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	if body.CaddieCode != "" {
		// Update caddie_current_status
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
		if err := caddie.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	if buggy.Code != "" {
		// Update caddie_current_status
		buggy.BuggyStatus = constants.BUGGY_CURRENT_STATUS_LOCK
		if err := buggy.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	//Update trạng thái của các old caddie
	if response.OldCaddie.Id > 0 && response.OldCaddie.Status == constants.CADDIE_CURRENT_STATUS_LOCK {
		udpCaddieOut(db, response.OldCaddie.Id)
	}

	//Update trạng thái của các old buggy
	if response.OldBuggy.Id > 0 {
		bookingR := model_booking.Booking{
			BookingDate: booking.BookingDate,
			BuggyId:     response.OldBuggy.Id,
		}
		if errBuggy := udpOutBuggy(db, &bookingR, false); errBuggy != nil {
			log.Println("AddCaddieBuggyToBooking err book udp ", errBuggy.Error())
		}
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_WAITING_LIST,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_ATTACH,
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

	okResponse(c, booking)
}

/*
Add Caddie list
Create Flight
*/
func (_ *CCourseOperating) CreateFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// TODO: validate max item in list

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	// Validate trùng cadddie
	for _, item1 := range body.ListData {
		if item1.CaddieCode != "" {
			for _, item2 := range body.ListData {
				if item1.Bag != item2.Bag && item2.CaddieCode == item1.CaddieCode {
					response_message.BadRequest(c, "Caddie chỉ được ghép cho một người ")
					return
				}
			}
		}
	}

	// Validate Buggy, Buggy ghép tối đa được 2 người trong 1 flight
	countBuggy := 0
	for index, item1 := range body.ListData {
		countBuggy = 0
		if item1.BuggyCode != "" {
			for _, item2 := range body.ListData {
				if item1.Bag != item2.Bag && item2.BuggyCode == item1.BuggyCode {
					countBuggy += 1
					body.ListData[index].BagShare = item2.Bag
					if item2.BuggyCommonCode != "" {
						body.ListData[index].BuggyCommonCode = item2.BuggyCommonCode
					} else {
						body.ListData[index].BuggyCommonCode = fmt.Sprint(utils.GetTimeNow().UnixNano())
					}
				}
			}
		}
	}

	for index, _ := range body.ListData {
		if body.ListData[index].BuggyCommonCode == "" {
			body.ListData[index].BuggyCommonCode = fmt.Sprint(utils.GetTimeNow().UnixNano())
		}
	}

	if countBuggy >= 2 {
		response_message.BadRequest(c, "Buggy ghép tối đa được 2 người trong 1 flight")
		return
	}

	// Check các bag ok hết mới tạo flight
	// Check Caddie, Buggy đang trong flight
	listError := []string{}
	listBooking := []model_booking.Booking{}
	listOldCaddie := []models.Caddie{}
	listOldBuggy := []models.Buggy{}

	listCaddie := []models.Caddie{}
	listBuggy := []models.Buggy{}
	listCaddieInOut := []model_gostarter.CaddieBuggyInOut{}
	for _, v := range body.ListData {
		errB, response := addCaddieBuggyToBooking(db, body.PartnerUid, body.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode, v.IsPrivateBuggy)

		bookingTemp := response.Booking
		caddieTemp := response.NewCaddie
		buggyTemp := response.NewBuggy

		if errB != nil {
			response_message.BadRequestFreeMessage(c, errB.Error())
			return
		}

		if *bookingTemp.LockBill {
			response_message.BadRequestFreeMessage(c, "Bag "+bookingTemp.Bag+" đã lock")
			return
		}

		if bookingTemp.Uid != "" && bookingTemp.BagStatus != constants.BAG_STATUS_WAITING {
			response_message.BadRequestFreeMessage(c, fmt.Sprintln("Bag", bookingTemp.Bag, bookingTemp.BagStatus))
			return
		}

		if response.OldCaddie.Id > 0 {
			listOldCaddie = append(listOldCaddie, response.OldCaddie)
		}

		if response.OldBuggy.Id > 0 {
			listOldBuggy = append(listOldBuggy, response.OldBuggy)
		}

		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:      bookingTemp.PartnerUid,
			CourseUid:       bookingTemp.CourseUid,
			BookingUid:      bookingTemp.Uid,
			BuggyCommonCode: v.BuggyCommonCode,
			Bag:             bookingTemp.Bag,
			BookingDate:     body.BookingDate,
		}

		if caddieTemp.Id > 0 {
			if errB == nil {
				// Update caddie_current_status
				if caddieTemp.CurrentRound == 0 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
				} else if caddieTemp.CurrentRound == 1 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE_R2
				} else if caddieTemp.CurrentRound == 2 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE_R3
				}
				caddieTemp.CurrentRound = caddieTemp.CurrentRound + 1

				caddieBuggyInNote.CaddieId = caddieTemp.Id
				caddieBuggyInNote.CaddieCode = caddieTemp.Code
				caddieBuggyInNote.CaddieType = constants.STATUS_IN
				caddieBuggyInNote.Hole = bookingTemp.Hole
				listCaddie = append(listCaddie, caddieTemp)
			}
		}

		if buggyTemp.Id > 0 {
			bookingTemp.IsPrivateBuggy = setBoolForCursor(v.IsPrivateBuggy)

			buggyTemp.BuggyStatus = constants.BUGGY_CURRENT_STATUS_IN_COURSE
			caddieBuggyInNote.IsPrivateBuggy = setBoolForCursor(v.IsPrivateBuggy)
			caddieBuggyInNote.BuggyId = buggyTemp.Id
			caddieBuggyInNote.BuggyCode = buggyTemp.Code
			caddieBuggyInNote.BuggyType = constants.STATUS_IN
			caddieBuggyInNote.BagShareBuggy = v.BagShare
			caddieBuggyInNote.HoleBuggy = bookingTemp.Hole

			listBuggy = append(listBuggy, buggyTemp)
		}

		if errB != nil {
			listError = append(listError, errB.Error())
		}

		listCaddieInOut = append(listCaddieInOut, caddieBuggyInNote)
		listBooking = append(listBooking, bookingTemp)
	}

	if len(listError) > 0 {
		errRes := response_message.ErrorResponseDataV2{
			StatusCode:  400,
			ErrorDetail: listError,
		}
		badRequest(c, errRes)
		return
	}

	// Create flight
	flight := model_gostarter.Flight{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Tee:        body.Tee,
		TeeOff:     body.TeeOff,
		CourseType: body.CourseType,
	}

	hourStr, _ := utils.GetDateFromTimestampWithFormat(utils.GetTimeNow().Unix(), constants.HOUR_FORMAT)
	yearStr, _ := utils.GetDateFromTimestampWithFormat(utils.GetTimeNow().Unix(), "060102")
	flight.GroupName = yearStr + "_" + strconv.Itoa(body.Tee) + "_" + hourStr

	// Date display
	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	if errDate == nil {
		flight.DateDisplay = dateDisplay
	} else {
		log.Println("booking date display err ", errDate.Error())
	}

	errCF := flight.Create()
	if errCF != nil {
		response_message.InternalServerError(c, errCF.Error())
		return
	}

	// Udp flight for Booking
	listBookingUpdated := []model_booking.Booking{}
	for _, b := range listBooking {
		b.FlightId = flight.Id
		b.TeeOffTime = body.TeeOff
		b.BagStatus = constants.BAG_STATUS_IN_COURSE
		errUdp := b.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err flight ", errUdp.Error())
		}

		listBookingUpdated = append(listBookingUpdated, b)
		// Update lại thông tin booking
		// if booking := b.GetBooking(); booking != nil && booking.Uid != b.Uid {
		// 	booking.CaddieId = b.CaddieId
		// 	booking.CaddieInfo = b.CaddieInfo
		// 	go booking.Update(db)
		// }

		//Add log
		opLog := models.OperationLog{
			PartnerUid:  body.PartnerUid,
			CourseUid:   body.CourseUid,
			UserName:    prof.UserName,
			UserUid:     prof.Uid,
			Module:      constants.OP_LOG_MODULE_GO,
			Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_WAITING_LIST,
			Action:      constants.OP_LOG_ACTION_COURSE_INFO_CREATE_FLIGHT,
			Body:        models.JsonDataLog{Data: body},
			ValueOld:    models.JsonDataLog{},
			ValueNew:    models.JsonDataLog{Data: flight},
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Bag:         b.Bag,
			BookingDate: b.BookingDate,
			BillCode:    b.BillCode,
			BookingUid:  b.Uid,
		}
		go createOperationLog(opLog)
	}

	// Tạo giá buggy cho bag
	go func() {
		for index, booking := range listBookingUpdated {
			bodyItem := body.ListData[index]

			// utils.ContainString(constants.MEMBER_BUGGY_FEE_FREE_LIST, booking.CardId) == -1 &&
			if bodyItem.BuggyCode != "" {
				round := models.Round{
					BillCode: booking.BillCode,
				}

				if errFindRound := round.LastRound(db); errFindRound != nil {
					log.Println("Round not found")
				}

				if round.Hole > 0 {
					buggyFee := getBuggyFeeSetting(body.PartnerUid, body.CourseUid, booking.GuestStyle, round.Hole)
					if bodyItem.BagShare != "" {
						addBuggyFee(booking, buggyFee.RentalFee, "Thuê xe (1/2 xe)", round.Hole)
					} else {
						if booking.IsPrivateBuggy != nil && *booking.IsPrivateBuggy == true {
							addBuggyFee(booking, buggyFee.PrivateCarFee, "Thuê riêng xe", round.Hole)
							addBuggyFee(booking, buggyFee.RentalFee, "Thuê xe (1/2 xe)", round.Hole)
						} else {
							addBuggyFee(booking, buggyFee.RentalFee, "Thuê xe (1/2 xe)", round.Hole)
							addBuggyFee(booking, buggyFee.OddCarFee, "Thuê lẻ xe", round.Hole)
						}
					}
				}
			}
			updatePriceWithServiceItem(&booking, prof)
		}
	}()

	// Update caddie status
	for _, ca := range listCaddie {
		errUdp := ca.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err udp caddie ", errUdp.Error())
		}
	}

	if len(listCaddie) > 0 {
		// Bắn socket để update xếp nốt caddie
		go func() {
			cNotification := CNotification{}
			cNotification.CreateCaddieWorkingStatusNotification("")
		}()
	}

	// Udp Caddie In Out Note
	for _, data := range listCaddieInOut {
		go addBuggyCaddieInOutNote(db, data)
	}

	// Update buggy status
	for _, buggy := range listBuggy {
		errUdp := buggy.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err udp buggy ", errUdp.Error())
		}
	}

	// Udp Old Caddie
	for _, caddie := range listOldCaddie {
		udpCaddieOut(db, caddie.Id)
	}

	//Update trạng thái của các old buggy
	for _, buggy := range listOldBuggy {
		//Update trạng thái của các old buggy
		if buggy.Id > 0 {
			bookingR := model_booking.Booking{
				BookingDate: dateDisplay,
				BuggyId:     buggy.Id,
			}
			if errBuggy := udpOutBuggy(db, &bookingR, false); errBuggy != nil {
				log.Println("CreateFlight err book udp ", errBuggy.Error())
			}
		}
	}

	okResponse(c, flight)
}

/*
Out Caddie
- Không remove flight để undo book
- Check flight để
*/
func (_ *CCourseOperating) OutCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.OutCaddieBody{}
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

	if booking.CaddieId == 0 {
		response_message.ErrorResponse(c, http.StatusBadRequest, "OUT_CADDIE_ERROR", "", constants.ERROR_OUT_CADDIE)
		return
	}

	caddieId := booking.CaddieId
	caddieCode := booking.CaddieInfo.Code

	udpCaddieOut(db, caddieId)

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
	booking.CaddieHoles = body.CaddieHoles
	booking.TimeOutFlight = utils.GetTimeNow().Unix()
	booking.HoleTimeOut = body.GuestHoles
	booking.BagStatus = constants.BAG_STATUS_TIMEOUT
	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Udp Note
	caddieOutNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingUid:  booking.Uid,
		CaddieId:    caddieId,
		CaddieCode:  caddieCode,
		CaddieType:  constants.STATUS_OUT,
		Hole:        body.CaddieHoles,
		Note:        body.Note,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
	}
	go addBuggyCaddieInOutNote(db, caddieOutNote)

	okResponse(c, booking)
}

/*
TODO:
Undo Out Caddie
Check caddie lúc này có đang trên sân k
*/
func (_ *CCourseOperating) UndoOutCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.OutCaddieBody{}
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

	// Upd booking

	// Udp note

}

/*
Out All Caddie In a Flight
Lấy tất cả các booking - bag trong Flight
*/
func (_ *CCourseOperating) OutAllInFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.OutAllFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	//Get list bookings trong Flight
	bookingR := model_booking.Booking{
		FlightId: body.FlightId,
	}
	bookings, err := bookingR.FindListInFlight(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//Udp các booking
	timeOutFlight := utils.GetTimeNow().Unix()
	caddieList := []string{}
	partnerUid := ""
	courseUid := ""

	for _, booking := range bookings {
		partnerUid = booking.PartnerUid
		courseUid = booking.CourseUid

		oldBooking := booking

		if booking.BagStatus != constants.BAG_STATUS_TIMEOUT &&
			booking.BagStatus != constants.BAG_STATUS_CHECK_OUT {

			if *booking.LockBill {
				response_message.BadRequestFreeMessage(c, "Bag "+booking.Bag+" đã lock")
				return
			}

			udpCaddieOut(db, booking.CaddieId)
			if errBuggy := udpOutBuggy(db, &booking, false); errBuggy != nil {
				log.Println("OutAllFlight err book udp ", errBuggy.Error())
			}

			booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
			booking.CaddieHoles = body.CaddieHoles
			booking.TimeOutFlight = timeOutFlight
			booking.HoleTimeOut = body.GuestHoles
			booking.BagStatus = constants.BAG_STATUS_TIMEOUT
			errUdp := booking.Update(db)
			if errUdp != nil {
				log.Println("OutAllFlight err book udp ", errUdp.Error())
			}

			// Update lại giá của Round theo số hố
			cRound := CRound{}
			go cRound.UpdateListFeePriceInBookingAndRound(c, db, booking, booking.GuestStyle, body.GuestHoles)

			// Update giờ chơi nếu khách là member
			if booking.MemberCardUid != "" {
				go updateReportTotalHourPlayCountForCustomerUser(booking, booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
			}

			// update caddie in out note
			caddieOutNote := model_gostarter.CaddieBuggyInOut{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				BookingUid:  booking.Uid,
				Note:        body.Note,
				Bag:         booking.Bag,
				BookingDate: booking.BookingDate,
			}

			if booking.CaddieId > 0 {
				caddieOutNote.CaddieId = booking.CaddieId
				caddieOutNote.CaddieCode = booking.CaddieInfo.Code
				caddieOutNote.CaddieType = constants.STATUS_OUT
				caddieOutNote.Hole = body.CaddieHoles
				caddieList = append(caddieList, booking.CaddieInfo.Code)
			}

			if booking.BuggyId > 0 {
				caddieOutNote.BuggyId = booking.BuggyId
				caddieOutNote.BuggyCode = booking.BuggyInfo.Code
				caddieOutNote.BuggyType = constants.STATUS_OUT
				caddieOutNote.HoleBuggy = body.BuggyHoles
			}

			go addBuggyCaddieInOutNote(db, caddieOutNote)
			go updateCaddieOutSlot(partnerUid, courseUid, []string{booking.CaddieInfo.Code})
			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_GO,
				Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
				Action:      constants.OP_LOG_ACTION_COURSE_INFO_OUT_ALL_FLIGHT,
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
	}

	go func() {
		cNotification := CNotification{}
		cNotification.CreateCaddieWorkingStatusNotification("")
	}()

	okRes(c)
}

/*
Simple Out Caddie In a Flight
*/
func (cCourseOperating *CCourseOperating) SimpleOutFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.SimpleOutFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	//Get booking trong Flight
	bookingR := model_booking.Booking{
		FlightId: body.FlightId,
		Bag:      body.Bag,
	}
	bookingResponse, err := bookingR.FindListInFlight(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if len(bookingResponse) == 0 {
		response_message.BadRequest(c, "Bag Not Found")
		return
	}

	bookingFirst := bookingResponse[0]

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(db, bookingFirst.Uid)
	if err != nil {
		response_message.BadRequestFreeMessage(c, err.Error())
		return
	}

	oldBooking := booking

	if booking.BagStatus != constants.BAG_STATUS_IN_COURSE {
		response_message.BadRequestDynamicKey(c, "BAG_NOT_IN_COURSE", "")
		return
	}

	udpCaddieOut(db, booking.CaddieId)
	if errBuggy := udpOutBuggy(db, &booking, false); errBuggy != nil {
		log.Println("SimpleOutFlight err book udp ", errBuggy.Error())
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
	booking.CaddieHoles = body.CaddieHoles
	booking.HoleTimeOut = body.GuestHoles
	booking.TimeOutFlight = utils.GetTimeNow().Unix()
	booking.BagStatus = constants.BAG_STATUS_TIMEOUT
	errUdp := booking.Update(db)
	if errUdp != nil {
		log.Println("SimpleOutFlight err book udp ", errUdp.Error())
	}

	// Update lại giá của Round theo số hố
	cRound := CRound{}
	go cRound.UpdateListFeePriceInBookingAndRound(c, db, booking, booking.GuestStyle, body.GuestHoles)

	// Update giờ chơi nếu khách là member
	if booking.MemberCardUid != "" {
		go updateReportTotalHourPlayCountForCustomerUser(booking, booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
	}

	// update caddie in out note
	caddieOutNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingUid:  booking.Uid,
		Note:        body.Note,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
	}

	if booking.CaddieId > 0 {
		caddieOutNote.CaddieId = booking.CaddieId
		caddieOutNote.CaddieCode = booking.CaddieInfo.Code
		caddieOutNote.CaddieType = constants.STATUS_OUT
		caddieOutNote.Hole = body.CaddieHoles
	}

	if booking.BuggyId > 0 {
		caddieOutNote.BuggyId = booking.BuggyId
		caddieOutNote.BuggyCode = booking.BuggyInfo.Code
		caddieOutNote.BuggyType = constants.STATUS_OUT
		caddieOutNote.HoleBuggy = body.BuggyHoles
	}

	go addBuggyCaddieInOutNote(db, caddieOutNote)

	if booking.CaddieId > 0 {
		// Update node caddie
		caddieList := []string{booking.CaddieInfo.Code}
		updateCaddieOutSlot(booking.PartnerUid, booking.CourseUid, caddieList)
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_SIMPLE_OUT_FLIGHT,
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

	okRes(c)
}

/*
Need more caddie
Đổi Caddie
Out caddie cũ và gán Caddie mới cho Bag
*/
func (cCourseOperating *CCourseOperating) NeedMoreCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.NeedMoreCaddieBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if booking.CaddieInfo.Code == body.CaddieCode {
		response_message.BadRequestFreeMessage(c, "Bag "+"đã ghép với "+body.CaddieCode)
		return
	}

	// validate caddie_code
	oldCaddie := booking.CaddieInfo

	if oldCaddie.Id > 0 {
		udpCaddieOut(db, oldCaddie.Id)

		// Udp Note
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:  prof.PartnerUid,
			CourseUid:   prof.CourseUid,
			BookingUid:  booking.Uid,
			CaddieId:    oldCaddie.Id,
			CaddieCode:  oldCaddie.Code,
			Hole:        body.CaddieHoles,
			CaddieType:  constants.STATUS_OUT,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
		}

		// Update node caddie
		caddieList := []string{oldCaddie.Code}
		updateCaddieOutSlot(booking.PartnerUid, booking.CourseUid, caddieList)

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	if body.CaddieCode != "" {
		caddieNew, err := cCourseOperating.validateCaddie(db, prof.CourseUid, body.CaddieCode)

		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		if caddieNew.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			if booking.CaddieId != caddieNew.Id {
				response_message.InternalServerError(c, errors.New(caddieNew.Code+" đang bị LOCK").Error())
				return
			}
		} else {
			if errCaddie := checkCaddieReady(booking, caddieNew); errCaddie != nil {
				response_message.InternalServerError(c, errCaddie.Error())
				return
			}
		}

		// set new caddie
		booking.CaddieId = caddieNew.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

		if err := booking.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Update caddie_current_status
		if booking.FlightId != 0 {
			caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
			caddieNew.CurrentRound = caddieNew.CurrentRound + 1
		} else {
			//HAICV: disable bug
			// caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
		}

		if err := caddieNew.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Udp new caddie in
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:  prof.PartnerUid,
			CourseUid:   prof.CourseUid,
			BookingUid:  booking.Uid,
			CaddieId:    caddieNew.Id,
			CaddieCode:  caddieNew.Code,
			CaddieType:  constants.STATUS_IN,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
		}

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	okResponse(c, booking)
}

/*
Delete Attach caddie
- Trường hợp khách đã ghép Flight  (Đã gán caddie vs Buggy) --> Delete Attach Caddie sẽ out khách ra khỏi filght và xóa caddie và Buggy đã gán.
(Khách không bị cho vào danh sách out mà trở về trạng thái trước khi ghép)
- Trường hợp chưa ghép Flight (Đã gán Caddie và Buugy) --> Delete Attach Caddie sẽ xóa caddie và buggy đã gán với khách
*/
func (cCourseOperating *CCourseOperating) DeleteAttach(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.DeleteAttachBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check booking
	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	olgBooking := booking.CloneBooking()

	caddieId := booking.CaddieId

	if body.IsOutCaddie != nil && *body.IsOutCaddie == true {
		// out caddie

		//Caddie
		booking.CaddieId = 0
		booking.CaddieInfo = cloneToCaddieBooking(models.Caddie{})
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT
		booking.CaddieHoles = 0
	}

	if body.IsOutBuggy != nil && *body.IsOutBuggy == true {
		//Buggy

		if err := udpOutBuggy(db, &booking, false); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		booking.BuggyId = 0
		booking.BuggyInfo = cloneToBuggyBooking(models.Buggy{})
	}

	//Flight
	if booking.FlightId > 0 {
		booking.FlightId = 0
	}

	booking.TeeOffTime = ""
	booking.BagStatus = constants.BAG_STATUS_WAITING
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// auto delete buggy fee
	deleteBuggyFee(booking)
	udpCaddieOut(db, caddieId)

	caddie := models.Caddie{}
	caddie.Id = caddieId
	if errC := caddie.FindFirst(db); errC == nil {
		updateCaddieOutSlot(booking.PartnerUid, booking.CourseUid, []string{caddie.Code})
	}

	// ADD LOG
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_DELETE_ATTACH_FLIGHT,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: olgBooking},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	go createOperationLog(opLog)

	okResponse(c, booking)
}

/*
Get data for starting sheet display
*/
func (_ *CCourseOperating) GetStartingSheet(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetStartingSheetForm{}
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

	//Get Booking data
	bookingR := model_booking.Booking{
		PartnerUid:   form.PartnerUid,
		CourseUid:    form.CourseUid,
		BookingDate:  form.BookingDate,
		CustomerName: form.CustomerName,
	}

	listBooking := bookingR.FindForFlightAll(db, form.CaddieCode, form.CaddieName, form.NumberPeopleInFlight, page)

	okResponse(c, listBooking)
}

func (_ CCourseOperating) validateBooking(db *gorm.DB, bookindUid string) (model_booking.Booking, error) {
	bookingR := model_booking.Booking{}
	bookingR.Uid = bookindUid
	booking, err := bookingR.FindFirstByUId(db)
	if err != nil {
		return booking, err
	}

	if *booking.LockBill {
		return booking, errors.New("Bag " + booking.Bag + " đã lock")
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		return booking, errors.New("Bag " + booking.Bag + " đã check out!")
	}

	return booking, nil
}

func (_ CCourseOperating) validateCaddie(db *gorm.DB, courseUid string, caddieCode string) (models.Caddie, error) {
	caddieList := models.CaddieList{}
	caddieList.CourseUid = courseUid
	caddieList.CaddieCode = caddieCode
	caddieList.WorkingStatus = constants.CADDIE_WORKING_STATUS_ACTIVE
	caddieNew, err := caddieList.FindFirst(db)

	if err != nil {
		return caddieNew, err
	}

	return caddieNew, nil
}

func (cCourseOperating CCourseOperating) ChangeCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.ChangeCaddieBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate caddie_code
	oldCaddie := booking.CaddieInfo

	if oldCaddie.Id > 0 {
		udpCaddieOut(db, oldCaddie.Id)

		// Udp Note
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:  prof.PartnerUid,
			CourseUid:   prof.CourseUid,
			BookingUid:  booking.Uid,
			CaddieId:    oldCaddie.Id,
			CaddieCode:  oldCaddie.Code,
			Hole:        body.CaddieHoles,
			CaddieType:  constants.STATUS_OUT,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
		}

		// Update node caddie
		caddieList := []string{oldCaddie.Code}
		updateCaddieOutSlot(booking.PartnerUid, booking.CourseUid, caddieList)

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	if body.CaddieCode != "" {
		caddieNew, err := cCourseOperating.validateCaddie(db, prof.CourseUid, body.CaddieCode)

		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		if caddieNew.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			if booking.CaddieId != caddieNew.Id {
				response_message.InternalServerError(c, errors.New(caddieNew.Code+" đang bị LOCK").Error())
				return
			}
		} else {
			if errCaddie := checkCaddieReady(booking, caddieNew); errCaddie != nil {
				response_message.InternalServerError(c, errCaddie.Error())
				return
			}
		}

		// set new caddie
		booking.CaddieId = caddieNew.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

		if err := booking.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Update caddie_current_status
		caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
		caddieNew.CurrentRound = caddieNew.CurrentRound + 1

		if err := caddieNew.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Udp new caddie in
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:  prof.PartnerUid,
			CourseUid:   prof.CourseUid,
			BookingUid:  booking.Uid,
			CaddieId:    caddieNew.Id,
			CaddieCode:  caddieNew.Code,
			CaddieType:  constants.STATUS_IN,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
		}

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	// ADD LOG
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_CHANGE_CADDIE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldCaddie},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	go createOperationLog(opLog)

	okResponse(c, booking)
}

func (cCourseOperating CCourseOperating) ChangeBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.ChangeBuggyBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	hole, errCast := strconv.Atoi(body.Hole)

	if errCast != nil {
		response_message.BadRequestFreeMessage(c, "Cast Error!")
		return
	}

	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	oldBuggy := booking.BuggyInfo

	if booking.BuggyInfo.Code == body.BuggyCode {
		response_message.BadRequestFreeMessage(c, "Bag đang dùng buggy "+body.BuggyCode)
		return
	}

	// validate buggy_code
	buggyNew := models.Buggy{}
	buggyNew.CourseUid = prof.CourseUid
	buggyNew.Code = body.BuggyCode
	if err := buggyNew.FindFirst(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if errBuggy := checkBuggyReady(db, buggyNew, booking, body.IsPrivateBuggy, true); errBuggy != nil {

		response_message.InternalServerError(c, errBuggy.Error())
		return
	}

	if booking.BuggyId > 0 {
		if err := udpOutBuggy(db, &booking, false); err != nil {
			log.Println(err.Error())
		}

		go updateBagShareEmptyWhenChangeBuggy(booking)

		// Udp Note
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid:  prof.PartnerUid,
			CourseUid:   prof.CourseUid,
			BookingUid:  booking.Uid,
			BuggyId:     booking.BuggyId,
			BuggyCode:   booking.BuggyInfo.Code,
			BuggyType:   constants.STATUS_OUT,
			Bag:         booking.Bag,
			BookingDate: booking.BookingDate,
			HoleBuggy:   hole,
		}

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	// set new buggy
	booking.BuggyId = buggyNew.Id
	booking.BuggyInfo = cloneToBuggyBooking(buggyNew)
	booking.IsPrivateBuggy = setBoolForCursor(body.IsPrivateBuggy)
	//booking.BuggyStatus
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	buggyNew.BuggyStatus = constants.BUGGY_CURRENT_STATUS_IN_COURSE

	if err := buggyNew.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	go updateBagShareWhenChangeBuggy(booking, hole, body.IsPrivateBuggy)

	// ADD LOG
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_CHANGE_BUGGY,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldBuggy},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	go createOperationLog(opLog)

	okResponse(c, booking)
}

func updateBagShareWhenChangeBuggy(booking model_booking.Booking, holeChange int, IsPrivateBuggy bool) {
	var bagShare = ""
	buggyCommonCode := fmt.Sprint(utils.GetTimeNow().UnixNano())

	bookingR := model_booking.BookingList{
		BookingDate: booking.BookingDate,
		BuggyId:     booking.BuggyId,
		BagStatus:   constants.BAG_STATUS_IN_COURSE,
	}

	// Tìm kiếm booking đang sử dụng buggy
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	db1, _, _ := bookingR.FindAllBookingList(db)
	db1 = db1.Where("bag <> ?", booking.Bag)

	bookingList := []model_booking.Booking{}
	db1.Find(&bookingList)

	if len(bookingList) > 0 {
		bagShare = bookingList[0].Bag
		newCaddieBuggyInOut := model_gostarter.CaddieBuggyInOut{
			PartnerUid: bookingList[0].PartnerUid,
			CourseUid:  bookingList[0].CourseUid,
			BookingUid: bookingList[0].Uid,
		}

		list, total, _ := newCaddieBuggyInOut.FindOrderByDateList(db)
		if total > 0 {
			lastItem := list[0]
			lastItem.BagShareBuggy = booking.Bag
			lastItem.BuggyCommonCode = buggyCommonCode
			if err := lastItem.Update(db); err != nil {
				return
			}
		}
	}

	holeDiff := utils.RoundHole(booking.Hole - holeChange)
	// Udp Note
	caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid:      booking.PartnerUid,
		CourseUid:       booking.CourseUid,
		BookingUid:      booking.Uid,
		BuggyId:         booking.BuggyId,
		BuggyCode:       booking.BuggyInfo.Code,
		BuggyType:       constants.STATUS_IN,
		IsPrivateBuggy:  &IsPrivateBuggy,
		BagShareBuggy:   bagShare,
		BuggyCommonCode: buggyCommonCode,
		Bag:             booking.Bag,
		BookingDate:     booking.BookingDate,
		HoleBuggy:       holeDiff,
	}

	addBuggyCaddieInOutNote(db, caddieBuggyInNote)
}

func updateBagShareEmptyWhenChangeBuggy(booking model_booking.Booking) {
	bookingR := model_booking.BookingList{
		BookingDate: booking.BookingDate,
		BuggyId:     booking.BuggyId,
		BagStatus:   constants.BAG_STATUS_IN_COURSE,
	}

	// Tìm kiếm booking đang sử dụng buggy
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	db1, _, _ := bookingR.FindAllBookingList(db)
	db1 = db1.Where("bag <> ?", booking.Bag)

	bookingList := []model_booking.Booking{}
	db1.Find(&bookingList)

	if len(bookingList) > 0 {
		newCaddieBuggyInOut := model_gostarter.CaddieBuggyInOut{
			PartnerUid: bookingList[0].PartnerUid,
			CourseUid:  bookingList[0].CourseUid,
			BookingUid: bookingList[0].Uid,
		}

		list, total, _ := newCaddieBuggyInOut.FindOrderByDateList(db)
		if total > 0 {
			lastItem := list[0]
			lastItem.BagShareBuggy = ""
			if err := lastItem.Update(db); err != nil {
				log.Println("updateBagShareEmptyWhenChangeBuggy update error")
			}
		}
	}
}

func (cCourseOperating CCourseOperating) EditHolesOfCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.EditHolesOfCaddiesBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate caddie_code
	caddie, err := cCourseOperating.validateCaddie(db, prof.CourseUid, body.CaddieCode)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if caddie.Id != booking.CaddieId {
		response_message.InternalServerError(c, "Booking uid and caddie code do not match")
		return
	}

	booking.CaddieHoles = body.Hole

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (_ CCourseOperating) validateFlight(courseUid string, flightId int64) (model_gostarter.Flight, error) {
	flight := model_gostarter.Flight{}
	flight.CourseUid = courseUid
	flight.Id = flightId
	if err := flight.FindFirst(); err != nil {
		return flight, err
	}

	// dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	// bookingList := model_booking.BookingList{
	// 	FlightId:    flightId,
	// 	BagStatus:   constants.BAG_STATUS_IN_COURSE,
	// 	PartnerUid:  flight.PartnerUid,
	// 	CourseUid:   flight.CourseUid,
	// 	BookingDate: dateDisplay,
	// }

	// db := datasources.GetDatabaseWithPartner(flight.PartnerUid)
	// bookingList.FindAllBookingList(db)
	// _, total, _ := bookingList.FindAllBookingList(db)

	// course := models.Course{}
	// course.Uid = courseUid
	// if errCourse := course.FindFirst(db); errCourse != nil {
	// 	return flight, errors.New("Id not valid")
	// }

	// if total >= int64(course.MaxPeopleInFlight) {
	// 	return flight, errors.New("Flight đã đủ người")
	// }

	return flight, nil
}

func (cCourseOperating CCourseOperating) AddBagToFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddBagToFlightBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if body.FlightId == nil {
		response_message.BadRequestFreeMessage(c, "Not Found Flight Id")
		return
	}

	// validate flight_id
	flight, err := cCourseOperating.validateFlight(prof.CourseUid, *body.FlightId)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// TODO: validate max item in list

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	// Check các bag ok hết mới tạo flight
	// Check Caddie, Buggy đang trong flight
	listError := []string{}
	listBooking := []model_booking.Booking{}

	listOldCaddie := []models.Caddie{}
	listOldBuggy := []models.Buggy{}

	listCaddie := []models.Caddie{}
	listBuggy := []models.Buggy{}
	listCaddieInOut := []model_gostarter.CaddieBuggyInOut{}
	for _, v := range body.ListData {
		errB, response := addCaddieBuggyToBooking(db, prof.PartnerUid, prof.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode, v.IsPrivateBuggy)

		bookingTemp := response.Booking
		caddieTemp := response.NewCaddie
		buggyTemp := response.NewBuggy

		if *bookingTemp.LockBill {
			response_message.BadRequestFreeMessage(c, "Bag "+bookingTemp.Bag+" đã lock")
			return
		}

		if bookingTemp.BagStatus != constants.BAG_STATUS_WAITING {
			response_message.BadRequest(c, "BAG STATUS "+bookingTemp.BagStatus)
			return
		}

		if response.OldCaddie.Id > 0 {
			listOldCaddie = append(listOldCaddie, response.OldCaddie)
		}

		if response.OldBuggy.Id > 0 {
			listOldBuggy = append(listOldBuggy, response.OldBuggy)
		}

		if caddieTemp.Id > 0 {
			if errB == nil {
				// Update caddie_current_status
				if caddieTemp.CurrentRound == 0 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
				} else if caddieTemp.CurrentRound == 1 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE_R2
				} else if caddieTemp.CurrentRound == 2 {
					caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE_R3
				}
				caddieTemp.CurrentRound = caddieTemp.CurrentRound + 1

				listCaddie = append(listCaddie, caddieTemp)

				caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
					PartnerUid: prof.PartnerUid,
					CourseUid:  prof.CourseUid,
					BookingUid: bookingTemp.Uid,
					CaddieId:   bookingTemp.CaddieId,
					CaddieCode: bookingTemp.CaddieInfo.Code,
					CaddieType: constants.STATUS_IN,
					Note:       "",
				}
				listCaddieInOut = append(listCaddieInOut, caddieBuggyInNote)
			}
		}

		if buggyTemp.Id > 0 {
			buggyTemp.BuggyStatus = constants.BUGGY_CURRENT_STATUS_IN_COURSE
			listBuggy = append(listBuggy, buggyTemp)
		}

		if errB != nil {
			listError = append(listError, errB.Error())
		}

		listBooking = append(listBooking, bookingTemp)
	}

	if len(listError) > 0 {
		errRes := response_message.ErrorResponseDataV2{
			StatusCode:  400,
			ErrorDetail: listError,
		}
		badRequest(c, errRes)
		return
	}

	listBookingUpdated := []model_booking.Booking{}
	// Udp flight for Booking
	for _, b := range listBooking {
		b.FlightId = flight.Id
		b.BagStatus = constants.BAG_STATUS_IN_COURSE
		errUdp := b.Update(db)
		if errUdp != nil {
			log.Println("AddBagToFlight err flight ", errUdp.Error())
		}
		listBookingUpdated = append(listBookingUpdated, b)
		// Update lại thông tin booking
		if booking := b.GetBooking(); booking != nil && booking.Uid != b.Uid {
			booking.CaddieId = b.CaddieId
			booking.CaddieInfo = b.CaddieInfo

			go func() {
				booking.Update(db)

				// Add log
				opLog := models.OperationLog{
					PartnerUid:  booking.PartnerUid,
					CourseUid:   booking.CourseUid,
					UserName:    prof.UserName,
					UserUid:     prof.Uid,
					Module:      constants.OP_LOG_MODULE_GO,
					Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
					Action:      constants.OP_LOG_ACTION_COURSE_INFO_ADD_BAG_TO_FLIGHT,
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

				createOperationLog(opLog)
			}()

		}
	}

	// Tạo giá buggy cho bag
	go func() {
		for index, booking := range listBookingUpdated {
			bodyItem := body.ListData[index]
			// utils.ContainString(constants.MEMBER_BUGGY_FEE_FREE_LIST, booking.CardId) == -1 &&
			if bodyItem.BuggyCode != "" {
				round := models.Round{
					BillCode: booking.BillCode,
				}

				if errFindRound := round.LastRound(db); errFindRound != nil {
					log.Println("Round not found")
				}

				if round.Hole > 0 {
					buggyFee := getBuggyFeeSetting(booking.PartnerUid, booking.CourseUid, booking.GuestStyle, round.Hole)
					addBuggyFee(booking, buggyFee.RentalFee, "Thuê xe (1/2 xe)", round.Hole)
				}
			}
			updatePriceWithServiceItem(&booking, prof)
		}
	}()

	// Update caddie status
	for _, ca := range listCaddie {
		//ca.IsInCourse = true
		errUdp := ca.Update(db)
		if errUdp != nil {
			log.Println("AddBagToFlight err udp caddie ", errUdp.Error())
		}
	}

	if len(listCaddie) > 0 {
		// Bắn socket để update xếp nốt caddie
		go func() {
			cNotification := CNotification{}
			cNotification.CreateCaddieWorkingStatusNotification("")
		}()
	}

	// Udp Old Caddie
	for _, caddie := range listOldCaddie {
		udpCaddieOut(db, caddie.Id)
	}

	//Update trạng thái của các old buggy
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	for _, buggy := range listOldBuggy {
		//Update trạng thái của các old buggy
		if buggy.Id > 0 {
			bookingR := model_booking.Booking{
				BookingDate: dateDisplay,
				BuggyId:     buggy.Id,
			}
			if errBuggy := udpOutBuggy(db, &bookingR, false); errBuggy != nil {
				log.Println("AddBagToFlight err book udp ", errBuggy.Error())
			}
		}
	}

	okResponse(c, flight)
}

func (_ CCourseOperating) GetFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetFlightList{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	flights := model_gostarter.FlightList{}

	if query.BookingDate != "" {
		bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)
		flights.BookingDate = bookingDate.Format("02/01/2006")
	}

	flights.CourseUid = query.CourseUid
	flights.PartnerUid = query.PartnerUid
	flights.CaddieName = query.CaddieName
	flights.CaddieCode = query.CaddieCode
	flights.PlayerName = query.PlayerName
	flights.GolfBag = query.GolfBag

	list, total, err := flights.FindFlightList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (cCourseOperating CCourseOperating) MoveBagToFlight(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.MoveBagToFlightBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate golf_bag
	if booking.Bag != body.GolfBag {
		response_message.InternalServerError(c, "Booking uid and golf bag do not match")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestFreeMessage(c, "Bag "+booking.Bag+" đã lock")
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag "+booking.Bag+" đã check out!")
		return
	}

	// validate flight_id
	_, err = cCourseOperating.validateFlight(prof.CourseUid, body.FlightId)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Chuyển booking cũ sang time out
	booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
	booking.TimeOutFlight = utils.GetTimeNow().Unix()
	booking.HoleTimeOut = int(body.HolePlayed)
	booking.HoleMoveFlight = body.HoleMoveFlight
	booking.MovedFlight = setBoolForCursor(true)
	errBookingUpd := booking.Update(db)
	if errBookingUpd != nil {
		response_message.InternalServerError(c, errBookingUpd.Error())
		return
	}

	// Tạo booking mới với flightID và bag_status in course
	bookingUid := uuid.New()
	newBooking := cloneToBooking(booking)
	newBooking.MovedFlight = setBoolForCursor(false)
	newBooking.HoleTimeOut = 0
	newBooking.HoleMoveFlight = 0
	newBooking.BagStatus = constants.BAG_STATUS_IN_COURSE
	newBooking.FlightId = body.FlightId
	newBooking.TimeOutFlight = 0
	newBooking.BuggyId = 0
	newBooking.BuggyInfo = model_booking.BookingBuggy{}
	newBooking.InitType = constants.BOOKING_INIT_MOVE_FLGIHT

	bUid := booking.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())
	errCreateBooking := newBooking.Create(db, bUid)

	if errCreateBooking != nil {
		response_message.InternalServerError(c, errCreateBooking.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE,
		Action:      constants.OP_LOG_ACTION_COURSE_INFO_MOVE_FLIGHT,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: booking},
		ValueNew:    models.JsonDataLog{Data: newBooking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	go createOperationLog(opLog)

	okResponse(c, newBooking)
}

func (cCourseOperating CCourseOperating) UndoTimeOut(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UndoTimeOutBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	//Get booking trong Flight
	bookingR := model_booking.Booking{
		FlightId: body.FlightId,
		Bag:      body.Bag,
	}

	bookingResponse, err := bookingR.FindListInFlight(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if len(bookingResponse) == 0 {
		response_message.BadRequest(c, "Bag Not Found")
		return
	}

	for _, booking := range bookingResponse {
		if *booking.LockBill {
			response_message.BadRequestFreeMessage(c, "Bag "+booking.Bag+" đã lock")
			return
		}

		if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
			response_message.BadRequestFreeMessage(c, "Bag "+booking.Bag+" đã check out!")
			return
		}
	}

	for _, booking := range bookingResponse {
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
		booking.TimeOutFlight = 0
		booking.BagStatus = constants.BAG_STATUS_IN_COURSE
		errUdp := booking.Update(db)
		if errUdp != nil {
			log.Println("SimpleOutFlight err book udp ", errUdp.Error())
		} else {
			opLog := models.OperationLog{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				UserName:    prof.UserName,
				UserUid:     prof.Uid,
				Module:      constants.OP_LOG_MODULE_GO,
				Function:    constants.OP_LOG_FUNCTION_COURSE_INFO_TIME_OUT,
				Action:      constants.OP_LOG_ACTION_COURSE_INFO_UNDO_OUT_FLIGHT,
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

	}

	okRes(c)
}
