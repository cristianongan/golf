package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
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
	now := utils.GetTimeNow().Format(constants.DATE_FORMAT_1)

	bookingBookingList := model_booking.BookingList{
		PartnerUid:            body.PartnerUid,
		CourseUid:             body.CourseUid,
		IsCheckIn:             "1",
		BookingDate:           now,
		IsGroupBillCode:       true,
		NotNoneGolfAndWalking: true,
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
		"golfer":    bookingTotal,
		"waiting":   waitingTotal,
		"in_course": inCourseTotal,
		"time_out":  timeoutTotal,
		"check_out": checkOutTotal,
	}

	okResponse(c, res)
}

func (_ *CReportDashboard) GetReportGuestOnDay(c *gin.Context, prof models.CmsUser) {
	body := request.GetReportDashboardRequestForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	date, _ := utils.GetDateFromTimestampWithFormat(utils.GetTimeNow().Unix(), constants.DATE_FORMAT_1)

	// Money Bag
	// bookingBookingList := model_booking.BookingList{
	// 	BagStatus:       constants.BAG_STATUS_BOOKING,
	// 	PartnerUid:      body.PartnerUid,
	// 	CourseUid:       body.CourseUid,
	// 	BookingDate:     date,
	// 	IsGroupBillCode: true,
	// }
	// _, bookingTotal, _ := bookingBookingList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Total Book
	bookingList := model_booking.BookingList{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: date,
		InitType:    constants.BOOKING_INIT_TYPE_BOOKING,
	}
	_, totalBooking, _ := bookingList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Total Book With Check In
	bookingCIList := model_booking.BookingList{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: date,
		InitType:    constants.BOOKING_INIT_TYPE_BOOKING,
		IsCheckIn:   "2",
	}
	_, totalBookCI, _ := bookingCIList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Total Non golf
	bookingNGList := model_booking.BookingList{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		BookingDate:  date,
		CustomerType: constants.CUSTOMER_TYPE_NONE_GOLF,
	}
	_, totalNonGolf, _ := bookingNGList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Total Visitor
	bookingVisitorList := model_booking.BookingList{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		BookingDate:  date,
		CustomerType: constants.BOOKING_CUSTOMER_TYPE_VISITOR,
	}
	_, totalVisitor, _ := bookingVisitorList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Total Driving
	bookingDrivingList := model_booking.BookingServiceItem{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Type:       constants.DRIVING_SETTING,
	}
	_, totalDriving, _ := bookingDrivingList.FindListBookingOrder(datasources.GetDatabaseWithPartner(prof.PartnerUid), date)

	// Member
	bookingMemberList := model_booking.BookingList{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		BookingDate:  date,
		CustomerType: constants.BOOKING_CUSTOMER_TYPE_MEMBER,
	}
	_, totalMember, _ := bookingMemberList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Member Guest
	bookingMemberGuestList := model_booking.BookingList{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		BookingDate:  date,
		CustomerType: constants.BOOKING_CUSTOMER_TYPE_GUEST,
	}
	_, totalMemberGuest, _ := bookingMemberGuestList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	// Agency
	bookingAgencyList := model_booking.BookingList{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: date,
		IsAgency:    "1",
	}
	_, totalAgency, _ := bookingAgencyList.FindAllBookingList(datasources.GetDatabaseWithPartner(prof.PartnerUid))

	res := map[string]interface{}{
		"money_bag":       0,
		"book":            totalBooking,
		"ci_from_booking": totalBookCI,
		"non_golf":        totalNonGolf,
		"visitor":         totalVisitor,
		"driving":         totalDriving,
		"member":          totalMember,
		"member_guest":    totalMemberGuest,
		"agency":          totalAgency,
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

	// Get list top 10 member
	booking := model_booking.Booking{
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
	}

	list, _ := booking.FindTopMember(db, body.TypeMember, body.TypeDate, body.Date)

	res := map[string]interface{}{
		"data": list,
	}

	okResponse(c, res)
}

func (_ *CReportDashboard) GetReportRevenueFromBooking(c *gin.Context, prof models.CmsUser) {
	body := request.GetReportRevenueDashboardRequestForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	now := utils.GetTimeNow().Format(constants.MONTH_FORMAT)

	monthNow, _ := time.Parse(constants.MONTH_FORMAT, now)

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	listData := make([]map[string]interface{}, 12)

	for month := 0; month < 12; month++ {
		date := monthNow.AddDate(0, -month, 0).Format(constants.MONTH_FORMAT)

		// Add infor to response
		listData[month] = map[string]interface{}{
			"month":        date,
			"agency":       0,
			"member":       0,
			"guest_member": 0,
			"visitor":      0,
			"other":        0,
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

	res := response.PageResponse{
		Total: 12,
		Data:  listData,
	}

	c.JSON(200, res)
}
