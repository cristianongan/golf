package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	models_agency_booking "start/models/agency-booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CBookingTransaction struct{}

/*
Bắt đầu transaction mới, lock tee và lưu thông tin booking
*/
func (_ *CBookingTransaction) CreateBookingTransaction(c *gin.Context, prof models.CmsUser) {
	body := request.AgencyBookingTransactionDTO{}

	if err := c.ShouldBind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	bookingList := body.BookingList

	if len(bookingList) <= 0 {
		response_message.BadRequest(c, "")
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	transaction := models_agency_booking.AgencyBookingTransaction{
		// TransactionId:     transactionId,
		CourseUid:            prof.CourseUid,
		PartnerUid:           prof.PartnerUid,
		AgencyId:             body.AgencyId,
		PaymentStatus:        "",
		CustomerPhoneNumber:  body.CustomerPhoneNumber,
		CustomerEmail:        body.CustomerEmail,
		CustomerName:         body.CustomerName,
		BookingAmount:        body.BookingAmount,
		ServiceAmount:        body.ServiceAmount,
		PaymentDueDate:       0,
		PlayerNote:           body.PlayerNote,
		AgentNote:            body.PlayerNote,
		PaymentType:          body.PaymentType,
		Company:              body.Company,
		CompanyAddress:       body.CompanyAddress,
		CompanyTax:           body.CompanyTax,
		ReceiptEmail:         body.ReceiptEmail,
		TransactionStatus:    constants.AGENCY_BOOKING_TRANSACTION_INIT,
		BookingRequestStatus: constants.AGENCY_REQUEST_BOOKING_LOCK_TEE,
	}

	// lock tee time
	for _, item := range bookingList {
		errLock := lockTee(item.TeeType, 1, item.BookingDate, item.TeeTime, item.CourseUid, "lock from agency")
		if errLock != nil {
			response_message.BadRequest(c, errLock.Error())
			return
		}
	}

	for _, item := range bookingList {
		item.TransactionId = transaction.TransactionId
	}

	// create temp
	errBatch := bookingList[0].CreateBatch(bookingList, db)

	if errBatch != nil {
		response_message.BadRequest(c, errBatch.Error())
		return
	}

	// create transaction
	errCreate := transaction.Create(db)

	if errCreate != nil {
		response_message.BadRequest(c, errCreate.Error())
		return
	}

	okRes(c)
}

/*
Cập nhật 1 transaction
*/
func (_ *CBookingTransaction) UpdateBookingTransaction(c *gin.Context, prof models.CmsUser) {
	body := request.AgencyBookingTransactionDTO{}

	if err := c.ShouldBind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if body.TransactionId == "" {
		response_message.BadRequest(c, "")
		return
	}

	transaction := models_agency_booking.AgencyBookingTransaction{
		TransactionId: body.TransactionId,
	}

	bookingList := body.BookingList

	if len(bookingList) <= 0 {
		response_message.BadRequest(c, "")
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	if findErr := transaction.FindFirst(db); findErr != nil {
		response_message.BadRequest(c, findErr.Error())
		return
	}

	// kiểm tra trạng thái init & wait
	if transaction.TransactionStatus != constants.AGENCY_BOOKING_TRANSACTION_INIT ||
		transaction.PaymentStatus != constants.AGENCY_PAYMENT_WAIT {
		response_message.BadRequest(c, "")
		return
	}
	transaction = models_agency_booking.AgencyBookingTransaction{
		TransactionStatus:   constants.AGENCY_BOOKING_TRANSACTION_INIT,
		CourseUid:           prof.CourseUid,
		PartnerUid:          prof.PartnerUid,
		AgencyId:            body.AgencyId,
		PaymentStatus:       "",
		CustomerPhoneNumber: body.CustomerPhoneNumber,
		CustomerEmail:       body.CustomerEmail,
		CustomerName:        body.CustomerName,
		BookingAmount:       body.BookingAmount,
		ServiceAmount:       body.ServiceAmount,
		PaymentDueDate:      0,
		PlayerNote:          body.PlayerNote,
		AgentNote:           body.PlayerNote,
		PaymentType:         body.PaymentType,
		Company:             body.Company,
		CompanyAddress:      body.CompanyAddress,
		CompanyTax:          body.CompanyTax,
		ReceiptEmail:        body.ReceiptEmail,
	}

	if updErr := transaction.Update(db); updErr != nil {
		response_message.BadRequest(c, updErr.Error())
		return
	}

	delBooking := models_agency_booking.AgencyBookingInfo{
		TransactionId: transaction.TransactionId,
	}

	// del old list
	errDel := delBooking.DeleteBatch(db)
	if errDel != nil {
		response_message.BadRequest(c, errDel.Error())
		return
	}

	// create new list
	for _, item := range bookingList {
		item.TransactionId = transaction.TransactionId
	}

	bookingList[0].CreateBatch(bookingList, db)

	okRes(c)
}

func (_ *CBookingTransaction) FindBookingTransactionList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetAgencyTransactionRequest{}
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

	var fromDate int64
	var toDate int64

	if utils.IsDateValue(form.FromDate) && utils.IsDateValue(form.ToDate) {
		fromDate = utils.GetStartDayByTimeStamp(utils.GetTimeStampFromLocationTime(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, form.FromDate), constants.LOCATION_DEFAULT)
		toDate = utils.GetEndDayByTimeStamp(utils.GetTimeStampFromLocationTime(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, form.ToDate), constants.LOCATION_DEFAULT)
	} else {
		now := time.Now().Unix()

		fromDate = utils.GetStartDayByTimeStamp(now, constants.LOCATION_DEFAULT)
		toDate = utils.GetEndDayByTimeStamp(now, constants.LOCATION_DEFAULT)
	}

	model := models_agency_booking.AgencyBookingTransaction{}

	result, total, errFind := model.FindList(fromDate, toDate, page, db)

	if errFind != nil {
		response_message.InternalServerError(c, errFind.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  result,
	}

	c.JSON(200, res)

}

func (_ *CBookingTransaction) GetBookingTransaction(c *gin.Context, prof models.CmsUser) {
	transactionId := c.Param("id")

	if transactionId == "" {
		response_message.BadRequest(c, "transaction id is required")
		return
	}

	transaction := models_agency_booking.AgencyBookingTransaction{
		TransactionId: transactionId,
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	if errFind := transaction.FindFirst(db); errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	hisModel := models_agency_booking.AgencyBookingTransactionHis{
		TransactionId: transactionId,
	}

	histories, errHis := hisModel.FindList(db)

	if errHis != nil {
		response_message.BadRequest(c, errHis.Error())
		return
	}

	response := map[string]interface{}{
		"histories": histories,
		"detail":    transaction,
	}

	okResponse(c, response)
}
