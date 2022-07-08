package controllers

import (
	"github.com/go-playground/validator/v10"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
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

	// TODO: filter by date

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
	errB, booking, caddie, _ := addCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode)
	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
		return
	}

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Update caddie_current_status
	caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
	if err := caddie.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Udp Note
	caddieInOutNote := model_gostarter.CaddieInOutNote{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		Type:       constants.STATUS_IN,
		Note:       "",
	}

	go addCaddieInOutNote(caddieInOutNote)

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

	// TODO: validate max item in list

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	// Check các bag ok hết mới tạo flight
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

			// Update caddie_current_status
			caddieTemp.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
			caddieTemp.CurrentRound = caddieTemp.CurrentRound + 1
			if err := caddieTemp.Update(); err != nil {
				response_message.InternalServerError(c, err.Error())
				return
			}

			// Udp Note
			caddieInOutNote := model_gostarter.CaddieInOutNote{
				PartnerUid: prof.PartnerUid,
				CourseUid:  prof.CourseUid,
				BookingUid: bookingTemp.Uid,
				CaddieId:   bookingTemp.CaddieId,
				Type:       constants.STATUS_IN,
				Note:       "",
			}

			go addCaddieInOutNote(caddieInOutNote)
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
		//ca.IsInCourse = true
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

	errOut := udpOutCaddieBooking(&booking)
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
		errOut := udpOutCaddieBooking(&booking)
		if errOut == nil {
			booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
			booking.CaddieHoles = body.CaddieHoles
			errUdp := booking.Update()
			if errUdp != nil {
				log.Println("OutAllFlight err book udp ", errUdp.Error())
			}

			// update caddie in out note
			caddieInOutNote := model_gostarter.CaddieInOutNote{
				PartnerUid: booking.PartnerUid,
				CourseUid:  booking.CourseUid,
				BookingUid: booking.Uid,
				CaddieId:   booking.CaddieId,
				Type:       constants.STATUS_OUT,
				Hole:       body.CaddieHoles,
				Note:       body.Note,
			}

			go addCaddieInOutNote(caddieInOutNote)
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

	// TODO: validate current_status

	// TODO: validate caddie_holes

	// Out Caddie cũ
	udpCaddieOut(booking.CaddieId)

	// Gán Caddie mới
	booking.CaddieId = caddieNew.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
	booking.CaddieHoles = booking.Hole - body.CaddieHoles
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	caddieInOutNote := model_gostarter.CaddieInOutNote{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BookingUid: booking.Uid,
		CaddieId:   booking.CaddieId,
		Type:       constants.STATUS_OUT,
		Hole:       body.CaddieHoles,
		Note:       body.Note,
	}

	go addCaddieInOutNote(caddieInOutNote)

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

/*
	Get data for starting sheet display
*/
func (_ *CCourseOperating) GetStartingSheet(c *gin.Context, prof models.CmsUser) {
	form := request.GetStartingSheetForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.BookingDate == "" {
		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
		if errDate == nil {
			form.BookingDate = dateDisplay
		} else {
			log.Println("GetStartingSheet display err ", errDate.Error())
		}
	}

	//Get Flight Data
	flightR := model_gostarter.Flight{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		DateDisplay: form.BookingDate,
	}

	listFlight, _ := flightR.FindListAll()

	//Get Booking data
	bookingR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
	}

	listBooking := bookingR.FindForFlightAll()

	res := map[string]interface{}{
		"flight_data":  listFlight,
		"booking_data": listBooking,
	}

	okResponse(c, res)
}

func (_ CCourseOperating) validateBooking(bookindUid string) (model_booking.Booking, error) {
	booking := model_booking.Booking{}
	booking.Uid = bookindUid
	if err := booking.FindFirst(); err != nil {
		return booking, err
	}

	return booking, nil
}

func (_ CCourseOperating) validateCaddie(courseUid string, caddieCode string) (models.Caddie, error) {
	caddieList := models.CaddieList{}
	caddieList.CourseUid = courseUid
	caddieList.CaddieCode = caddieCode
	caddieList.WorkingStatus = constants.CADDIE_WORKING_STATUS_ACTIVE
	caddieList.InCurrentStatus = []string{constants.CADDIE_CURRENT_STATUS_READY, constants.CADDIE_CURRENT_STATUS_FINISH}
	caddieNew, err := caddieList.FindFirst()

	if err != nil {
		return caddieNew, err
	}

	return caddieNew, nil
}

func (cCourseOperating CCourseOperating) ChangeCaddie(c *gin.Context, prof models.CmsUser) {
	body := request.ChangeCaddieBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate caddie_code
	caddieNew, err := cCourseOperating.validateCaddie(prof.CourseUid, body.CaddieCode)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// TODO: validate current_status

	if err := udpCaddieOut(booking.CaddieId); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// set new caddie
	booking.CaddieId = caddieNew.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddieNew)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (cCourseOperating CCourseOperating) ChangeBuggy(c *gin.Context, prof models.CmsUser) {
	body := request.ChangeBuggyBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	booking, err := cCourseOperating.validateBooking(body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate buggy_code
	buggyNew := models.Buggy{}
	buggyNew.CourseUid = prof.CourseUid
	buggyNew.Code = body.BuggyCode
	if err := buggyNew.FindFirst(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// TODO: validate current_status

	if err := udpBuggyOut(booking.BuggyId); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// set new buggy
	booking.BuggyId = buggyNew.Id
	booking.BuggyInfo = cloneToBuggyBooking(buggyNew)
	//booking.BuggyStatus
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (cCourseOperating CCourseOperating) EditHolesOfCaddie(c *gin.Context, prof models.CmsUser) {
	body := request.EditHolesOfCaddiesBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// validate caddie_code
	caddie, err := cCourseOperating.validateCaddie(prof.CourseUid, body.CaddieCode)
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

	if err := booking.Update(); err != nil {
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

	return flight, nil
}

func (cCourseOperating CCourseOperating) AddBagToFlight(c *gin.Context, prof models.CmsUser) {
	body := request.AddBagToFlightBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(body.BookingUid)
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

	// TODO: validate max list

	booking.FlightId = body.FlightId

	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Update caddie_current_status
	caddie := models.Caddie{}
	caddie.Id = booking.CaddieId
	if err := caddie.FindFirst(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
	caddie.CurrentRound = caddie.CurrentRound + 1
	if err := caddie.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (_ CCourseOperating) GetFlight(c *gin.Context, prof models.CmsUser) {
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

	bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)

	flights := model_gostarter.FlightList{}

	flights.BookingDate = bookingDate.Format("02/01/2006")

	list, total, err := flights.FindFlightList(page)

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
	body := request.MoveBagToFlightBody{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate booking_uid
	booking, err := cCourseOperating.validateBooking(body.BookingUid)
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

	// TODO: validate max list

	booking.FlightId = body.FlightId

	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}

func (cCourseOperating *CCourseOperating) Checkout(c *gin.Context, prof models.CmsUser) {
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
	booking, err := cCourseOperating.validateBooking(body.BookingUid)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if booking.Bag != body.GolfBag {
		response_message.InternalServerError(c, "Booking uid and golf bag do not match")
		return
	}

	if booking.CustomerUid != body.CustomerUid {
		response_message.InternalServerError(c, "Booking uid and customer uid do not match")
		return
	}

	booking.BagStatus = constants.BAG_STATUS_OUT
	if err := booking.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, booking)
}
