package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CRevenueReport struct{}

func (_ *CRevenueReport) GetReportRevenueFoodBeverage(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueReportFBForm{}
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

	serviceCart := models.ServiceCart{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	list, total, err := serviceCart.FindReport(db, page, form.FromDate, form.ToDate, form.TypeService)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)

}

func (_ *CRevenueReport) GetReportRevenueDetailFBBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueReportFBForm{}
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

	serviceCart := models.ServiceCart{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	list, total, err := serviceCart.FindReportDetailFBBag(db, page, form.FromDate, form.ToDate, form.TypeService)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportRevenueDetailFB(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueReportDetailFBForm{}
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

	serviceCart := models.ServiceCart{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		ServiceType: form.Service,
	}

	list, total, err := serviceCart.FindReportDetailFB(db, page, form.FromDate, form.ToDate, form.Name)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetBookingReportRevenueDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueBookingReportDetail{}
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

	reportRevenue := model_report.ReportRevenueDetailList{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GuestStyle: form.GuestStyle,
		FromDate:   form.FromDate,
		ToDate:     form.ToDate,
	}

	list, total, _ := reportRevenue.FindBookingRevenueList(db, page)

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportCashierAudit(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueBookingReportDetail{}
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

	reportRevenue := model_report.ReportRevenueDetailList{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GuestStyle: form.GuestStyle,
		FromDate:   form.FromDate,
		ToDate:     form.ToDate,
	}

	list, total, _ := reportRevenue.FindBookingRevenueList(db, page)

	newList := []model_report.ResReportCashierAudit{}

	for _, item := range list {
		newList = append(newList, model_report.ResReportCashierAudit{
			PartnerUid: item.PartnerUid,
			CourseUid:  item.CourseUid,
			Bag:        item.Bag,
			TransTime:  item.CreatedAt,
			Cash:       item.Cash,
			Card:       item.Card,
			Voucher:    item.Voucher,
			Debit:      item.Debit,
		})
	}

	res := response.PageResponse{
		Total: total,
		Data:  newList,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
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

	bookingList := model_booking.BookingList{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		FromDate:    form.FromDate,
		ToDate:      form.ToDate,
		BuggyCode:   form.BuggyCode,
		BookingDate: form.BookingDate,
	}

	list, total, _ := bookingList.FindListBookingWithBuggy(db, page)

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportGolfFeeService(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueBookingReportDetail{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	reportRevenue := model_report.ReportRevenueDetailList{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Year:       form.Year,
		Month:      form.Month,
	}

	res := reportRevenue.FindGolfFeeRevenue(db)

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportBuggyForGuestStyle(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.ReportBuggyForGuestStyleForm{}
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

	report := model_booking.Booking{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	list, total, _ := report.FindReportBuggyForGuestStyle(db, page, form.Month, form.Year)

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportSalePOS(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.ReportSalePOSForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	report := model_booking.BookingServiceItem{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Type:       form.Type,
	}

	list, _ := report.FindReportSalePOS(db, form.Date)

	res := response.PageResponse{
		Data: list,
	}

	okResponse(c, res)
}

func (cBooking *CRevenueReport) GetDailyReport(c *gin.Context, prof models.CmsUser) {
	body := request.FinishBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	bookings := model_booking.BookingList{
		BookingDate: body.BookingDate,
	}

	db, _, err := bookings.FindAllBookingList(db)
	db = db.Where("check_in_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type <> 'ROUND'")

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	reportR := model_report.ReportRevenueDetail{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	reportR.DeleteByBookingDate()

	for _, booking := range list {
		updatePriceForRevenue(booking, body.BillNo)
	}

	repotR := model_report.ReportRevenueDetail{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	data, err := repotR.FindReportDayEnd(db)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	singlePaymentItemR1 := model_payment.SinglePaymentItem{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	db1 := datasources.GetDatabaseWithPartner(body.PartnerUid)
	listTransfer, _ := singlePaymentItemR1.FindAllTransfer(db1)
	listCards, _ := singlePaymentItemR1.FindAllCards(datasources.GetDatabaseWithPartner(body.PartnerUid))

	vcb := int64(0)
	bidv := int64(0)

	for _, item := range listCards {
		if item.BankType == "VCB" {
			vcb += item.Paid
		}
		if item.BankType == "BIDV" {
			bidv += item.Paid
		}
	}

	res := map[string]interface{}{
		"revenue": data,
		"players": listTransfer,
		"cards": map[string]interface{}{
			"vcb":  vcb,
			"bidv": bidv,
		},
	}

	okResponse(c, res)
}

func (cBooking *CRevenueReport) GetBagDailyReport(c *gin.Context, prof models.CmsUser) {
	form := request.ReportBagDaily{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(form.PartnerUid)

	repotR := model_report.ReportRevenueDetail{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	list, total, _ := repotR.FindList(db, page)

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}
	okResponse(c, res)
}

func (cBooking *CRevenueReport) UpdateReportRevenue(c *gin.Context, prof models.CmsUser) {
	body := request.FinishBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	bookings := model_booking.BookingList{
		BookingDate: body.BookingDate,
	}

	db, _, err := bookings.FindAllBookingList(db)
	db = db.Where("check_in_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type IN (?)", []string{constants.BOOKING_INIT_TYPE_BOOKING, constants.BOOKING_INIT_TYPE_CHECKIN})

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	for _, booking := range list {
		updatePriceForRevenue(booking, body.BillNo)
	}

	okRes(c)
}
