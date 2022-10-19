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

	if body.PartnerUid == "" || body.CourseUid == "" || body.BookingDate == "" || body.Bag == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check can add
	errB, booking, caddie, buggy := addCaddieBuggyToBooking(db, body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode, body.IsPrivateBuggy)

	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
		return
	}

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
				}
			}
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
	listCaddie := []models.Caddie{}
	listBuggy := []models.Buggy{}
	listCaddieInOut := []model_gostarter.CaddieBuggyInOut{}
	for _, v := range body.ListData {
		errB, bookingTemp, caddieTemp, buggyTemp := addCaddieBuggyToBooking(db, body.PartnerUid, body.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode, v.IsPrivateBuggy)

		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid: bookingTemp.PartnerUid,
			CourseUid:  bookingTemp.CourseUid,
			BookingUid: bookingTemp.Uid,
		}

		if caddieTemp.Id > 0 {
			if errB == nil {
				// Update caddie_current_status
				caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
				caddieTemp.CurrentRound = caddieTemp.CurrentRound + 1

				caddieBuggyInNote.CaddieId = caddieTemp.Id
				caddieBuggyInNote.CaddieCode = caddieTemp.Code
				caddieBuggyInNote.CaddieType = constants.STATUS_IN
				listCaddie = append(listCaddie, caddieTemp)
			}
		}

		if buggyTemp.Id > 0 {
			bookingTemp.IsPrivateBuggy = newTrue(v.IsPrivateBuggy)

			buggyTemp.BuggyStatus = constants.BUGGY_CURRENT_STATUS_IN_COURSE
			caddieBuggyInNote.IsPrivateBuggy = newTrue(v.IsPrivateBuggy)
			caddieBuggyInNote.BuggyId = buggyTemp.Id
			caddieBuggyInNote.BuggyCode = buggyTemp.Code
			caddieBuggyInNote.BuggyType = constants.STATUS_IN
			caddieBuggyInNote.BagShareBuggy = v.BagShare

			listBuggy = append(listBuggy, buggyTemp)
		}

		if errB != nil {
			listError = append(listError, errB.Error())
		}

		bookingTemp.CourseType = body.CourseType
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
	}

	hourStr, _ := utils.GetDateFromTimestampWithFormat(time.Now().Unix(), constants.HOUR_FORMAT)
	yearStr, _ := utils.GetDateFromTimestampWithFormat(time.Now().Unix(), "060102")
	flight.GroupName = yearStr + "_" + strconv.Itoa(body.Tee) + "_" + hourStr

	// Date display
	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
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
	for _, b := range listBooking {
		b.FlightId = flight.Id
		b.TeeOffTime = body.TeeOff
		b.BagStatus = constants.BAG_STATUS_IN_COURSE
		errUdp := b.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err flight ", errUdp.Error())
		}
	}

	// Update caddie status
	for _, ca := range listCaddie {
		//ca.IsInCourse = true
		errUdp := ca.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err udp caddie ", errUdp.Error())
		}
	}

	// Udp Caddie In Out Note
	for _, data := range listCaddieInOut {
		go addBuggyCaddieInOutNote(db, data)
	}

	// Udp Caddie In Out Note
	for _, buggy := range listBuggy {
		//ca.IsInCourse = true
		errUdp := buggy.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err udp buggy ", errUdp.Error())
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

	caddieId := booking.CaddieId
	caddieCode := booking.CaddieInfo.Code

	errOut := udpOutCaddieBooking(db, &booking)
	if errOut != nil {
		response_message.InternalServerError(c, errOut.Error())
		return
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CaddieHoles = body.CaddieHoles
	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errOut.Error())
		return
	}

	// Udp Note
	caddieOutNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   caddieId,
		CaddieCode: caddieCode,
		CaddieType: constants.STATUS_OUT,
		Hole:       booking.CaddieHoles,
		Note:       body.Note,
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
	timeOutFlight := time.Now().Unix()
	for _, booking := range bookings {
		if booking.BagStatus != constants.BAG_STATUS_TIMEOUT &&
			booking.BagStatus != constants.BAG_STATUS_CHECK_OUT {
			errOut := udpOutCaddieBooking(db, &booking)
			if errBuggy := udpOutBuggy(db, &booking, false); errBuggy != nil {
				log.Println("OutAllFlight err book udp ", errBuggy.Error())
			}
			if errOut == nil {
				booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
				booking.CaddieHoles = body.CaddieHoles
				booking.TimeOutFlight = timeOutFlight
				booking.HoleTimeOut = body.GuestHoles
				booking.BagStatus = constants.BAG_STATUS_TIMEOUT
				errUdp := booking.Update(db)
				if errUdp != nil {
					log.Println("OutAllFlight err book udp ", errUdp.Error())
				}

				// Update giờ chơi nếu khách là member
				if booking.MemberCardUid != "" {
					go updateReportTotalHourPlayCountForCustomerUser(booking, booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
				}

				// update caddie in out note
				caddieOutNote := model_gostarter.CaddieBuggyInOut{
					PartnerUid: booking.PartnerUid,
					CourseUid:  booking.CourseUid,
					BookingUid: booking.Uid,
					Note:       body.Note,
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
				}

				go addBuggyCaddieInOutNote(db, caddieOutNote)

			} else {
				log.Println("OutAllFlight err out caddie ", errOut.Error())
			}
		}
	}

	okRes(c)
}

/*
Simple Out Caddie In a Flight
*/
func (_ *CCourseOperating) SimpleOutFlight(c *gin.Context, prof models.CmsUser) {
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
		response_message.BadRequest(c, err.Error())
		return
	}
	booking := bookingResponse[0]
	errOut := udpOutCaddieBooking(db, &booking)
	if errBuggy := udpOutBuggy(db, &booking, false); errBuggy != nil {
		log.Println("OutAllFlight err book udp ", errBuggy.Error())
	}

	if errOut == nil {
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
		booking.CaddieHoles = body.CaddieHoles
		booking.HoleTimeOut = body.GuestHoles
		booking.TimeOutFlight = time.Now().Unix()
		booking.BagStatus = constants.BAG_STATUS_TIMEOUT
		errUdp := booking.Update(db)
		if errUdp != nil {
			log.Println("OutAllFlight err book udp ", errUdp.Error())
		}

		// Update giờ chơi nếu khách là member
		if booking.MemberCardUid != "" {
			go updateReportTotalHourPlayCountForCustomerUser(booking, booking.CustomerUid, booking.PartnerUid, booking.CourseUid)
		}

		// update caddie in out note
		caddieOutNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			BookingUid: booking.Uid,
			Note:       body.Note,
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
		}

		go addBuggyCaddieInOutNote(db, caddieOutNote)

		if booking.TeeTime != "" {
			go unlockTurnTime(db, booking)
		}
	} else {
		log.Println("OutAllFlight err out caddie ", errOut.Error())
	}

	okRes(c)
}

/*
Need more caddie
Đổi Caddie
Out caddie cũ và gán Caddie mới cho Bag
*/
func (_ *CCourseOperating) NeedMoreCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.NeedMoreCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}
	// Get Booking detail
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Check Caddie mới
	caddieNew := models.Caddie{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		Code:       body.CaddieCode,
	}
	errFC := caddieNew.FindFirst(db)
	if errFC != nil {
		response_message.BadRequest(c, errFC.Error())
		return
	}

	// TODO: validate current_status

	// TODO: validate caddie_holes

	// Out Caddie cũ
	udpCaddieOut(db, booking.CaddieId)

	caddieOutNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		CaddieCode: booking.CaddieInfo.Code,
		CaddieType: constants.STATUS_OUT,
		Hole:       body.CaddieHoles,
		Note:       body.Note,
	}

	go addBuggyCaddieInOutNote(db, caddieOutNote)

	// Gán Caddie mới
	booking.CaddieId = caddieNew.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
	booking.CaddieHoles = booking.Hole - body.CaddieHoles
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Update caddie_current_status
	caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
	caddieNew.CurrentRound = caddieNew.CurrentRound + 1

	if err := caddieNew.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		CaddieCode: booking.CaddieInfo.Code,
		CaddieType: constants.STATUS_IN,
		Hole:       booking.Hole - body.CaddieHoles,
		Note:       body.Note,
	}

	go addBuggyCaddieInOutNote(db, caddieBuggyInNote)

	okResponse(c, booking)
}

/*
Delete Attach caddie
- Trường hợp khách đã ghép Flight  (Đã gán caddie vs Buggy) --> Delete Attach Caddie sẽ out khách ra khỏi filght và xóa caddie và Buggy đã gán.
(Khách không bị cho vào danh sách out mà trở về trạng thái trước khi ghép)
- Trường hợp chưa ghép Flight (Đã gán Caddie và Buugy) --> Delete Attach Caddie sẽ xóa caddie và buggy đã gán với khách
*/
func (_ *CCourseOperating) DeleteAttach(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.DeleteAttachBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check booking
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	caddieId := booking.CaddieId

	if body.IsOutCaddie != nil && *body.IsOutCaddie == true {
		// out caddie
		if err := udpCaddieOut(db, caddieId); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

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

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

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
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(db); err != nil {
		return booking, err
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

	if booking.CaddieId > 0 {
		if err := udpCaddieOut(db, booking.CaddieId); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Udp Note
		caddieOutNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			BookingUid: booking.Uid,
			CaddieId:   booking.CaddieId,
			CaddieCode: booking.CaddieInfo.Code,
			CaddieType: constants.STATUS_OUT,
			Note:       "",
		}

		go addBuggyCaddieInOutNote(db, caddieOutNote)
	}

	// set new caddie
	booking.CaddieId = caddieNew.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Update caddie_current_status
	if booking.FlightId != 0 {
		caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
		caddieNew.CurrentRound = caddieNew.CurrentRound + 1
	} else {
		caddieNew.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
	}

	if err := caddieNew.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Udp Note
	caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		CaddieCode: booking.CaddieInfo.Code,
		CaddieType: constants.STATUS_IN,
		Note:       "",
	}

	go addBuggyCaddieInOutNote(db, caddieBuggyInNote)

	okResponse(c, booking)
}

func (cCourseOperating CCourseOperating) ChangeBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.ChangeBuggyBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	booking, err := cCourseOperating.validateBooking(db, body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
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

		// Udp Note
		caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			BookingUid: booking.Uid,
			BuggyId:    booking.BuggyId,
			BuggyCode:  booking.BuggyInfo.Code,
			BuggyType:  constants.STATUS_OUT,
		}

		go addBuggyCaddieInOutNote(db, caddieBuggyInNote)
	}

	// set new buggy
	booking.BuggyId = buggyNew.Id
	booking.BuggyInfo = cloneToBuggyBooking(buggyNew)
	booking.IsPrivateBuggy = newTrue(body.IsPrivateBuggy)
	//booking.BuggyStatus
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	if err := booking.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	buggyNew.BuggyStatus = constants.BUGGY_CURRENT_STATUS_IN_COURSE

	if err := buggyNew.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Udp Note
	caddieBuggyInNote := model_gostarter.CaddieBuggyInOut{
		PartnerUid:     prof.PartnerUid,
		CourseUid:      prof.CourseUid,
		BookingUid:     booking.Uid,
		BuggyId:        booking.BuggyId,
		BuggyCode:      booking.BuggyInfo.Code,
		BuggyType:      constants.STATUS_IN,
		IsPrivateBuggy: &body.IsPrivateBuggy,
		Note:           "",
	}

	go addBuggyCaddieInOutNote(db, caddieBuggyInNote)

	okResponse(c, booking)
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

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

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

	// dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())
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

	// validate flight_id
	flight, err := cCourseOperating.validateFlight(prof.CourseUid, body.FlightId)
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
	listCaddie := []models.Caddie{}
	listBuggy := []models.Buggy{}
	listCaddieInOut := []model_gostarter.CaddieBuggyInOut{}
	for _, v := range body.ListData {
		errB, bookingTemp, caddieTemp, buggyTemp := addCaddieBuggyToBooking(db, prof.PartnerUid, prof.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode, v.IsPrivateBuggy)

		if caddieTemp.Id > 0 {
			if errB == nil {
				// Update caddie_current_status
				caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
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

	// Udp flight for Booking
	for _, b := range listBooking {
		b.FlightId = flight.Id
		b.BagStatus = constants.BAG_STATUS_IN_COURSE
		errUdp := b.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err flight ", errUdp.Error())
		}
	}

	// Update caddie status
	for _, ca := range listCaddie {
		//ca.IsInCourse = true
		errUdp := ca.Update(db)
		if errUdp != nil {
			log.Println("CreateFlight err udp caddie ", errUdp.Error())
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

	if query.PeopleNumberInFlight != nil {
		flights.PeopleNumberInFlight = query.PeopleNumberInFlight
	}

	if query.PartnerUid != "" {
		flights.PartnerUid = query.PartnerUid
	}

	if query.CourseUid != "" {
		flights.CourseUid = query.CourseUid
	}

	flights.GolfBag = query.GolfBag
	flights.CaddieName = query.CaddieName
	flights.CaddieCode = query.CaddieCode
	flights.CustomerName = query.CustomerName

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

	// validate flight_id
	_, err = cCourseOperating.validateFlight(prof.CourseUid, body.FlightId)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Chuyển booking cũ sang time out
	booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
	booking.TimeOutFlight = time.Now().Unix()
	booking.HoleTimeOut = int(body.HolePlayed)
	booking.MovedFlight = newTrue(true)
	errBookingUpd := booking.Update(db)
	if errBookingUpd != nil {
		response_message.InternalServerError(c, errBookingUpd.Error())
		return
	}

	// Tạo booking mới với flightID và bag_status in course
	bookingUid := uuid.New()
	newBooking := cloneToBooking(booking)
	newBooking.MovedFlight = newTrue(false)
	newBooking.HoleTimeOut = 0
	newBooking.HoleMoveFlight = body.HoleMoveFlight
	newBooking.BagStatus = constants.BAG_STATUS_IN_COURSE
	newBooking.FlightId = body.FlightId
	newBooking.TimeOutFlight = 0
	newBooking.BuggyId = 0
	newBooking.BuggyInfo = model_booking.BookingBuggy{}

	bUid := booking.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())
	errCreateBooking := newBooking.Create(db, bUid)

	if errCreateBooking != nil {
		response_message.InternalServerError(c, errCreateBooking.Error())
		return
	}

	okResponse(c, newBooking)
}
