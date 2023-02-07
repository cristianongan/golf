package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
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
