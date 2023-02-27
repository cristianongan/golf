package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils"
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

	serviceCart := models.ServiceCart{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		ServiceType: form.Type,
	}

	list, err := serviceCart.FindReportDetailFB(db, form.Date, form.Name)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"data": list,
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

	report, _ := reportRevenue.FindReportDayEnd(db)

	res := map[string]interface{}{
		"total":  total,
		"data":   list,
		"report": report,
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

// Báo cáo DT tổng hợp với đk đã check-in
func (cBooking *CRevenueReport) GetDailyReport(c *gin.Context, prof models.CmsUser) {
	body := request.FinishBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	bookings := model_booking.BookingList{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	db, _, err := bookings.FindAllBookingList(db)
	db = db.Where("check_in_time > 0")
	// db = db.Where("check_out_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type <> 'ROUND'")
	db = db.Where("init_type <> 'MOVEFLGIHT'")

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

	if err := reportR.DeleteByBookingDate(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

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

	vcbTransfer := int64(0)
	bidvTransfer := int64(0)

	for _, item := range listTransfer {
		if item.BankType == "VCB" {
			vcbTransfer += item.Paid
		}
		if item.BankType == "BIDV" {
			bidvTransfer += item.Paid
		}
	}

	res := map[string]interface{}{
		"revenue": data,
		"players": listTransfer,
		"cards": map[string]interface{}{
			"vcb":   vcb,
			"bidv":  bidv,
			"total": data.Card,
		},
		"transfer": map[string]interface{}{
			"vcb":   vcbTransfer,
			"bidv":  bidvTransfer,
			"total": data.Transfer,
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

// Update BC DT với đk đã check-in và check-out
func (cBooking *CRevenueReport) UpdateReportRevenue(c *gin.Context, prof models.CmsUser) {
	body := request.UpdateReportBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	reportR := model_report.ReportRevenueDetail{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	if err := reportR.DeleteByBookingDate(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	bookings := model_booking.BookingList{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
	}

	db, _, err := bookings.FindAllBookingList(db)
	db = db.Where("check_in_time > 0")
	// db = db.Where("check_out_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type <> 'ROUND'")
	db = db.Where("init_type <> 'MOVEFLGIHT'")

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

func (_ *CRevenueReport) GetReportUsingBuggyInGo(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.ReportBuggyGoForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieBuggyInOut := model_gostarter.CaddieBuggyInOut{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	report, _ := caddieBuggyInOut.FindReportBuggyUsing(db, form.Month, form.Year)

	okResponse(c, report)
}

func (_ *CRevenueReport) GetReportRevenuePointOfSale(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueReportPOSForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	serviceItem := model_booking.BookingServiceItem{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		ServiceId:  form.ServiceId,
		Type:       form.Type,
	}

	list, err := serviceItem.FindReportRevenuePOS(db, form.FromDate, form.ToDate)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"data": list,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportAgencyPayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingList := model_booking.BookingList{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		FromDate:   form.FromDate,
		ToDate:     form.ToDate,
		// AgencyName: form.AgencyName,
	}

	list, _ := bookingList.FindReportAgencyPayment(db)

	okResponse(c, list)
}

func (_ *CRevenueReport) GetReportStarter(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.BookingDate == "" {
		response_message.BadRequest(c, errors.New("Chưa chọn ngày").Error())
		return
	}

	bookings := SetParamGetBookingRequest(form)

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	list, _ := bookings.FindReportStarter(db, page)

	okResponse(c, list)
}

func (_ *CRevenueReport) GetReportPayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.ReportPaymentBagStatus{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingDate := ""
	if form.BookingDate != "" {
		bookingDate = form.BookingDate
	} else {
		toDayDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		bookingDate = toDayDate
	}

	bookingList := model_booking.BookingList{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: bookingDate,
	}

	listBooking, total, _ := bookingList.FindReportPayment(db, form.PaymentStatus)

	res := map[string]interface{}{
		"data":  listBooking,
		"total": total,
	}

	okResponse(c, res)
}

func (_ *CRevenueReport) GetReportBookingPlayers(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.ReportBookingPlayers{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	reportPlayers := int64(0)
	nonPlayers := int64(0)

	bookingDate := ""
	if form.BookingDate != "" {
		bookingDate = form.BookingDate
	} else {
		toDayDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		bookingDate = toDayDate
	}

	bookingList := model_booking.BookingList{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: bookingDate,
	}

	report, _ := bookingList.ReportAllBooking(db)

	db1, _ := bookingList.FindAllLastBooking(db)
	db1.Where("customer_type <> ?", constants.CUSTOMER_TYPE_NONE_GOLF)
	db1.Count(&reportPlayers)

	db2, _ := bookingList.FindAllLastBooking(db)
	db2.Where("customer_type = ?", constants.CUSTOMER_TYPE_NONE_GOLF)
	db2.Count(&nonPlayers)

	inCompleteTotal := bookingList.CountReportPayment(db, constants.PAYMENT_IN_COMPLETE)
	completeTotal := bookingList.CountReportPayment(db, constants.PAYMENT_COMPLETE)
	mushPayTotal := bookingList.CountReportPayment(db, constants.PAYMENT_MUSH_PAY)

	res := map[string]interface{}{
		"players":             reportPlayers,
		"non_players":         nonPlayers,
		"report_detail":       report,
		"payment_complete":    completeTotal,
		"payment_in_complete": inCompleteTotal,
		"payment_mushpay":     mushPayTotal,
	}

	okResponse(c, res)
}
