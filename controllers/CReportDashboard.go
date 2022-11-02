package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"time"

	"github.com/gin-gonic/gin"
)

type CReportDashboard struct{}

func (_ *CReportDashboard) GetReportBookingStatusOnDay(c *gin.Context, prof models.CmsUser) {
	body := request.GetReportDashboardRequestForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	now := time.Now().Format(constants.DATE_FORMAT_1)

	bookingBookingList := model_booking.BookingList{
		BagStatus:       constants.BAG_STATUS_BOOKING,
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingDate:     now,
		IsGroupBillCode: true,
	}
	_, bookingTotal, _ := bookingBookingList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	bookingWaitingList := model_booking.BookingList{
		BagStatus:       constants.BAG_STATUS_WAITING,
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingDate:     now,
		IsGroupBillCode: true,
	}
	_, waitingTotal, _ := bookingWaitingList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	bookingInCourseList := model_booking.BookingList{
		BagStatus:       constants.BAG_STATUS_IN_COURSE,
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingDate:     now,
		IsGroupBillCode: true,
	}
	_, inCourseTotal, _ := bookingInCourseList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	bookingTimeOutList := model_booking.BookingList{
		BagStatus:       constants.BAG_STATUS_TIMEOUT,
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingDate:     now,
		IsGroupBillCode: true,
	}
	_, timeoutTotal, _ := bookingTimeOutList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	bookingCheckOutList := model_booking.BookingList{
		BagStatus:       constants.BAG_STATUS_CHECK_OUT,
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingDate:     now,
		IsGroupBillCode: true,
	}
	_, checkOutTotal, _ := bookingCheckOutList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	res := map[string]interface{}{
		"Booking":  bookingTotal,
		"Waiting":  waitingTotal,
		"InCourse": inCourseTotal,
		"TimeOut":  timeoutTotal,
		"CheckOut": checkOutTotal,
	}

	okResponse(c, res)
}
