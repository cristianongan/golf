package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CCourseOperating struct{}

/*
 Danh sách booking for caddie on course
 Role: Booking đã checkin, chưa checkout và chưa out Caddies
*/
func (_ *CCourseOperating) GetListBookingCaddieOnCourse(c *gin.Context, prof models.CmsUser) {
	form := request.GetBookingForCaddieOnCourseForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
	}

	list := bookingR.FindForCaddieOnCourse()

	okResponse(c, list)
}

/*
	Add Caddie short
	Chưa tạo flight
*/
func (_ *CCourseOperating) AddCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
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
	errB, booking, _, _ := addCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode)
	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
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
	Add Caddie list
	Create Flight
*/
func (_ *CCourseOperating) CreateFlight(c *gin.Context, prof models.CmsUser) {
	body := request.CreateFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	// Check các bag ok hết mới tạo flight
	// TODO:
	// Check Caddie, Buggy đang trong flight
	listError := []error{}
	listBooking := []model_booking.Booking{}
	listCaddie := []models.Caddie{}
	listBuggy := []models.Buggy{}
	for _, v := range body.ListData {
		errB, bookingTemp, caddieTemp, buggyTemp := addCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode)
		if errB == nil {
			listBooking = append(listBooking, bookingTemp)
			listCaddie = append(listCaddie, caddieTemp)
			listBuggy = append(listBuggy, buggyTemp)
		} else {
			listError = append(listError, errB)
		}
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
		errUdp := b.Update()
		if errUdp != nil {
			log.Println("CreateFlight err flight ", errUdp.Error())
		}
	}

	// Update caddie status
	for _, ca := range listCaddie {
		ca.IsInCourse = true
		errUdp := ca.Update()
		if errUdp != nil {
			log.Println("CreateFlight err udp caddie ", errUdp.Error())
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
	body := request.OutCaddieBody{}
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

	caddieId := booking.CaddieId

	errOut := udpOutCaddieBooking(booking)
	if errOut != nil {
		response_message.InternalServerError(c, errOut.Error())
		return
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CaddieHoles = body.CaddieHoles
	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errOut.Error())
		return
	}

	// Udp Note
	caddieInOutNote := model_gostarter.CaddieInOutNote{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   caddieId,
		Type:       constants.STATUS_OUT,
		Note:       body.Note,
	}
	go addCaddieInOutNote(caddieInOutNote)

	okResponse(c, booking)
}

/*
	TODO:
	Undo Out Caddie
	Check caddie lúc này có đang trên sân k
*/
func (_ *CCourseOperating) UndoOutCaddie(c *gin.Context, prof models.CmsUser) {
	body := request.OutCaddieBody{}
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

	// Upd booking

	// Udp note

}

/*
	Out All Caddie In a Flight
	Lấy tất cả các booking - bag trong Flight
*/
func (_ *CCourseOperating) OutAllInFlight(c *gin.Context, prof models.CmsUser) {
	body := request.OutAllFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	//Get list bookings trong Flight
	bookingR := model_booking.Booking{
		FlightId: body.FlightId,
	}
	bookings, err := bookingR.FindListInFlight()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//Udp các booking
	for _, booking := range bookings {
		errOut := udpOutCaddieBooking(booking)
		if errOut == nil {
			booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
			booking.CaddieHoles = body.CaddieHoles
			errUdp := booking.Update()
			if errUdp != nil {
				log.Println("OutAllFlight err book udp ", errUdp.Error())
			}
		} else {
			log.Println("OutAllFlight err out caddie ", errOut.Error())
		}
	}

	okRes(c)
}

/*
	Need more caddie
	Đổi Caddie
	Out caddie cũ và gán Caddie mới cho Bag
*/
func (_ *CCourseOperating) NeedMoreCaddie(c *gin.Context, prof models.CmsUser) {
	body := request.NeedMoreCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}
	// Get Booking detail
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
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
	errFC := caddieNew.FindFirst()
	if errFC != nil {
		response_message.BadRequest(c, errFC.Error())
		return
	}

	// Caddie đang trên sân rồi
	if caddieNew.IsInCourse {
		response_message.BadRequest(c, errors.New("Caddie new is in course").Error())
		return
	}

	// Out Caddie cũ
	udpCaddieOut(booking.CaddieId)

	// Gán Caddie mới
	booking.CaddieId = caddieNew.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
	booking.CaddieHoles = body.CaddieHoles
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
	Delete Attach caddie
	- Trường hợp khách đã ghép Flight  (Đã gán caddie vs Buggy) --> Delete Attach Caddie sẽ out khách ra khỏi filght và xóa caddie và Buggy đã gán.
	(Khách không bị cho vào danh sách out mà trở về trạng thái trước khi ghép)
	- Trường hợp chưa ghép Flight (Đã gán Caddie và Buugy) --> Delete Attach Caddie sẽ xóa caddie và buggy đã gán với khách
*/
func (_ *CCourseOperating) DeleteAttachCaddie(c *gin.Context, prof models.CmsUser) {
	body := request.OutCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check booking
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	caddieId := booking.CaddieId

	// out caddie
	udpCaddieOut(caddieId)

	//
	//Caddie
	booking.CaddieId = 0
	booking.CaddieInfo = cloneToCaddieBooking(models.Caddie{})
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_INIT
	booking.CaddieHoles = 0

	//Buggy
	booking.BuggyId = 0
	booking.BuggyInfo = cloneToBuggyBooking(models.Buggy{})

	//Flight
	if booking.FlightId > 0 {
		booking.FlightId = 0
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Udp Note
	caddieInOutNote := model_gostarter.CaddieInOutNote{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   caddieId,
		Type:       constants.STATUS_DELETE,
		Note:       body.Note,
	}
	go addCaddieInOutNote(caddieInOutNote)

	okResponse(c, booking)
}
