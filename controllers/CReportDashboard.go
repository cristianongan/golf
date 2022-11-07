package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strconv"
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

func (_ *CReportDashboard) GetReportTop10Member(c *gin.Context, prof models.CmsUser) {
	body := request.GetReportTop10MemberForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	// Validate date
	var date string
	if body.TypeDate == constants.TOP_MEMBER_DATE_TYPE_MONTH {
		date, _ = utils.GetDateFromTimestampWithFormat(time.Now().Unix(), constants.MONTH_FORMAT)
	} else if body.TypeDate == constants.TOP_MEMBER_DATE_TYPE_WEEK {
		_, week := time.Now().ISOWeek()
		date = strconv.Itoa(week)
	} else if body.TypeDate == constants.TOP_MEMBER_DATE_TYPE_DAY {
		date, _ = utils.GetDateFromTimestampWithFormat(time.Now().Unix(), constants.DATE_FORMAT_1)
	}

	// Get list top 10 member
	booking := model_booking.Booking{
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
	}

	list, _ := booking.FindTopMember(db, body.TypeMember, body.TypeDate, date)

	res := map[string]interface{}{
		"data": list,
	}

	okResponse(c, res)
}

func (_ *CReportDashboard) GetReportRevenueFromBooking(c *gin.Context, prof models.CmsUser) {
	body := request.GetReportDashboardRequestForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	listData := make([]map[string]interface{}, 12)

	for month := 0; month < 12; month++ {
		// Add infor to response
		listData[month] = map[string]interface{}{
			"month":        month + 1,
			"agency":       0,
			"member":       0,
			"guest_member": 0,
			"visitor":      0,
			"other":        0,
		}

		var date string
		year, _ := utils.GetDateFromTimestampWithFormat(time.Now().Unix(), constants.YEAR_FORMAT)

		if month < 9 {
			date = year + "-0" + strconv.Itoa(month+1)
		} else {
			date = year + "-" + strconv.Itoa(month+1)
		}

		booking := model_booking.Booking{}
		booking.PartnerUid = body.PartnerUid
		booking.CourseUid = body.CourseUid

		// report customer type agency
		listAgency, _ := booking.ReportBookingRevenue(db, constants.BOOKING_CUSTOMER_TYPE_AGENCY, date)
		if len(listAgency) > 0 {
			listData[month]["agency"] = listAgency[0]["revenue"]
		}

		// report customer type member
		listMember, _ := booking.ReportBookingRevenue(db, constants.BOOKING_CUSTOMER_TYPE_MEMBER, date)
		if len(listMember) > 0 {
			listData[month]["member"] = listMember[0]["revenue"]
		}

		// report customer type guest member
		listGuestMember, _ := booking.ReportBookingRevenue(db, constants.BOOKING_CUSTOMER_TYPE_GUEST, date)
		if len(listGuestMember) > 0 {
			listData[month]["guest_member"] = listGuestMember[0]["revenue"]
		}

		// report customer type visitor
		listVisitor, _ := booking.ReportBookingRevenue(db, constants.BOOKING_CUSTOMER_TYPE_VISITOR, date)
		if len(listVisitor) > 0 {
			listData[month]["visitor"] = listVisitor[0]["revenue"]
		}

		// report customer type other
		listOther, _ := booking.ReportBookingRevenue(db, "", date)
		if len(listOther) > 0 {
			listData[month]["other"] = listOther[0]["revenue"]
		}
	}

	okResponse(c, listData)
}
