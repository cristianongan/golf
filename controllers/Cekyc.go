package controllers

import (
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"

	"github.com/gin-gonic/gin"
)

type Cekyc struct{}

/*
Get List member card for eKyc
*/
func (_ *Cekyc) GetListMemberForEkycList(c *gin.Context) {
	responseBaseModel := response.EkycBaseResponse{
		Code: "00",
		Desc: "Success",
	}
	body := request.EkycGetMemberCardList{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		responseBaseModel.Code = "05"
		responseBaseModel.Desc = "Data incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseUid
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseBaseModel.Code = "01"
		responseBaseModel.Desc = "Course Uid not found"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	checkCheckSum := course.ApiKey + body.PartnerUid + body.CourseUid
	token := utils.GetSHA256Hash(checkCheckSum)

	if token != body.CheckSum {
		responseBaseModel.Code = "02"
		responseBaseModel.Desc = "Checksum incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	memberCardR := models.MemberCard{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
	}

	list, errL := memberCardR.FindListForEkyc(db)
	if errL != nil {
		responseBaseModel.Code = "03"
		responseBaseModel.Desc = "Error"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	responseBaseModel.Data = list

	c.JSON(http.StatusOK, responseBaseModel)
}

/*
check booking member for eKyc
*/
func (_ *Cekyc) CheckBookingMemberForEkyc(c *gin.Context) {
	responseBaseModel := response.EkycBaseResponse{
		Code: "00",
		Desc: "Success",
	}
	body := request.EkycCheckBookingMember{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		responseBaseModel.Code = "05"
		responseBaseModel.Desc = "Data incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	if body.BookingDate == "" {
		responseBaseModel.Code = "05"
		responseBaseModel.Desc = "Data incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseUid
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseBaseModel.Code = "01"
		responseBaseModel.Desc = "Course Uid not found"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	checkCheckSum := course.ApiKey + body.PartnerUid + body.CourseUid + body.MemberUid + body.BookingDate
	token := utils.GetSHA256Hash(checkCheckSum)

	if token != body.CheckSum {
		responseBaseModel.Code = "02"
		responseBaseModel.Desc = "Checksum incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	//Find card id
	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberUid
	errFMc := memberCard.FindFirst(db)
	if errFMc != nil {
		responseBaseModel.Code = "07"
		responseBaseModel.Desc = "không tìm thấy thông tin thẻ"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
		CustomerUid: memberCard.OwnerUid,
	}

	listBook, errLb := bookingR.FindMemberBooking(db)
	if errLb != nil || len(listBook) == 0 {
		responseBaseModel.Code = "03"
		responseBaseModel.Desc = "Not find Booking"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	idx := -1

	for i, v := range listBook {
		if idx == -1 && v.BagStatus == constants.BAG_STATUS_BOOKING {
			if v.Bag == "" {
				// Chưa đủ dk check in
				idx = -2
			} else {
				idx = i
			}
		}
	}

	if idx < 0 {
		if idx == -2 {
			responseBaseModel.Code = "04"
			responseBaseModel.Desc = "Chưa đủ điều kiện check-in: thiếu Bag"
			c.JSON(http.StatusBadRequest, responseBaseModel)
			return
		}
		responseBaseModel.Code = "08"
		responseBaseModel.Desc = "Bag da check in"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	//Get full info
	resFull := getBagDetailFromBooking(db, listBook[idx])

	responseBaseModel.Data = resFull.CloneBookingEkyc()

	c.JSON(http.StatusOK, responseBaseModel)

}

/*
checkin booking member for eKyc
*/
func (_ *Cekyc) CheckInBookingMemberForEkyc(c *gin.Context) {
	responseBaseModel := response.EkycBaseResponse{
		Code: "00",
		Desc: "Success",
	}
	body := request.EkycCheckInBookingMember{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		responseBaseModel.Code = "05"
		responseBaseModel.Desc = "Data incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseUid
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseBaseModel.Code = "01"
		responseBaseModel.Desc = "Course Uid not found"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	checkCheckSum := course.ApiKey + body.PartnerUid + body.CourseUid + body.BookingUid + body.BookingDate
	token := utils.GetSHA256Hash(checkCheckSum)

	if token != body.CheckSum {
		responseBaseModel.Code = "02"
		responseBaseModel.Desc = "Checksum incorrect"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	// Find booking
	booking := model_booking.Booking{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}
	booking.Uid = body.BookingUid
	errFB := booking.FindFirst(db)
	if errFB != nil {
		responseBaseModel.Code = "03"
		responseBaseModel.Desc = "Not find Booking"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_BOOKING {
		responseBaseModel.Code = "08"
		responseBaseModel.Desc = "Bag da check in"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	oldBooking := getBagDetailFromBooking(db, booking.CloneBooking())

	booking.CmsUser = "ekyc"
	booking.CmsUserLog = getBookingCmsUserLog("ekyc", utils.GetTimeNow().Unix())
	booking.CheckInTime = utils.GetTimeNow().Unix()
	booking.BagStatus = constants.BAG_STATUS_WAITING

	errUdp := booking.Update(db)
	if errUdp != nil {
		responseBaseModel.Code = "06"
		responseBaseModel.Desc = "Update fail"
		c.JSON(http.StatusBadRequest, responseBaseModel)
		return
	}

	// Update lại round còn thiếu bag
	cRound := CRound{}
	go cRound.UpdateBag(booking, db)

	res := getBagDetailFromBooking(db, booking)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    booking.CmsUser,
		UserUid:     booking.CmsUser,
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

	// push socket
	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)

	responseBaseModel.Data = res.CloneBookingEkyc()

	c.JSON(http.StatusOK, responseBaseModel)
}
