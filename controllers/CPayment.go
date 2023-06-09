package controllers

import (
	"encoding/json"
	"log"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	model_payment "start/models/payment"

	"github.com/gin-gonic/gin"
)

type CPayment struct{}

/*
create single payment and
*/
func (_ *CPayment) CreateSinglePayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateSinglePaymentBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check booking
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	booking.BillCode = body.BillCode

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Check check_sum
	amountStr := strconv.FormatInt(body.Amount, 10)
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PaymentType + "|" + body.BillCode + "|" + amountStr + "|" + body.BookingUid + "|" + body.DateStr
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	bagInfo := model_payment.PaymentBagInfo{}
	bagByte, errM := json.Marshal(booking)
	if errM != nil {
		log.Println("CreateSinglePayment errM", errM.Error())
	}
	errUM := json.Unmarshal(bagByte, &bagInfo)
	if errUM != nil {
		log.Println("CreateSinglePayment errUM", errUM.Error())
	}

	toDayDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	// Check single Payment
	singlePayment := model_payment.SinglePayment{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BillCode:   booking.BillCode,
	}
	singlePayment.Status = constants.STATUS_ENABLE
	isAdd := true
	errFind := singlePayment.FindFirst(db)

	if errFind != nil {
		// Chưa có thì tạo
		singlePayment.Bag = booking.Bag
		singlePayment.BookingUid = booking.Uid
		singlePayment.BookingDate = booking.BookingDate
		singlePayment.BookingCode = booking.BookingCode
		singlePayment.BagInfo = bagInfo
		singlePayment.TotalPaid = body.Amount
		singlePayment.Note = body.Note
		singlePayment.Cashiers = prof.UserName
		singlePayment.PaymentDate = toDayDate

		//Find prepaid from booking
		if booking.BookingCode != "" {
			bookOTA := model_booking.BookingOta{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				BookingCode: booking.BookingCode,
			}
			errFindBO := bookOTA.FindFirst(db)
			if errFindBO == nil {
				// singlePayment.PrepaidFromBooking = int64(bookOTA.NumBook) * (bookOTA.CaddieFee + bookOTA.BuggyFee + bookOTA.GreenFee)
			}
		}

		// Update payment status
		singlePayment.UpdatePaymentStatus(booking.BagStatus, db)

		errC := singlePayment.Create(db)

		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	} else {
		isAdd = false
	}

	// Tạo payment single item
	singlePaymentItem := model_payment.SinglePaymentItem{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingUid:  booking.Uid,
		BillCode:    booking.BillCode,
		Bag:         booking.Bag,
		Paid:        body.Amount,
		PaymentType: body.PaymentType,
		Note:        body.Note,
		PaymentUid:  singlePayment.Uid,
		Cashiers:    prof.UserName,
		BookingDate: booking.BookingDate,
		BankType:    body.BankType,
	}

	errC := singlePaymentItem.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_PAYMENT_SINGLE,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: singlePaymentItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}
	go createOperationLog(opLog)

	if !isAdd {
		// Find Total
		payItemR := model_payment.SinglePaymentItem{
			PartnerUid: booking.PartnerUid,
			BillCode:   booking.BillCode,
		}
		payItemR.Status = constants.STATUS_ENABLE

		listPir, errList := payItemR.FindAll(db)

		if errList == nil {
			totalPaid := int64(0)

			for _, v := range listPir {
				totalPaid = totalPaid + v.Paid
			}

			//Update other info
			singlePayment.BagInfo = bagInfo
			singlePayment.Note = body.Note
			singlePayment.Cashiers = prof.UserName
			singlePayment.PaymentDate = toDayDate
			singlePayment.TotalPaid = totalPaid
			singlePayment.UpdatePaymentStatus(booking.BagStatus, db)
			errUdp := singlePayment.Update(db)

			if errUdp != nil {
				response_message.InternalServerError(c, errUdp.Error())
				return
			}
		} else {
			log.Println("CreateSinglePayment errList", errList.Error())
		}
	}

	okRes(c)
}

/*
Get list single payment
*/
func (_ *CPayment) GetListSinglePayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetListSinglePaymentBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PartnerUid + "|" + body.CourseUid + "|" + body.PaymentDate + "|" + body.Bag + "|" + body.PlayerName + "|" + body.PaymentStatus
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	page := models.Page{
		Limit:   body.PageRequest.Limit,
		Page:    body.PageRequest.Page,
		SortBy:  body.PageRequest.SortBy,
		SortDir: body.PageRequest.SortDir,
	}

	paymentR := model_payment.SinglePayment{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		PaymentDate:   body.PaymentDate,
		Bag:           body.Bag,
		PaymentStatus: body.PaymentStatus,
		Type:          constants.PAYMENT_CATE_TYPE_SINGLE,
	}

	list, total, err := paymentR.FindListWithJoin(db, page, body.PlayerName)
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

/*
Update single payment item
*/
func (_ *CPayment) UpdateSinglePaymentItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UpdateSinglePaymentItemBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BookingUid + "|" + body.SinglePaymentItemUid + "|" + body.DateStr
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	// payment
	paymentItem := model_payment.SinglePaymentItem{}
	paymentItem.Uid = body.SinglePaymentItemUid
	errF := paymentItem.FindFirst(db)

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	paymentItem.Note = body.Note
	errUdp := paymentItem.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  paymentItem.PartnerUid,
		CourseUid:   paymentItem.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_PAYMENT_SINGLE,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: paymentItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         paymentItem.Bag,
		BookingDate: paymentItem.BookingDate,
		BillCode:    paymentItem.BillCode,
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okRes(c)
}

/*
Get list payment detail for bag
*/
func (_ *CPayment) GetListSinglePaymentDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetListSinglePaymentDetailBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BillCode + "|" + body.Bag
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	singlPaymentR := model_payment.SinglePayment{
		BillCode: body.BillCode,
		Bag:      body.Bag,
	}

	if errSinglePaymentR := singlPaymentR.FindFirst(db); errSinglePaymentR != nil {
		response_message.InternalServerError(c, errSinglePaymentR.Error())
		return
	}

	paymentR := model_payment.SinglePaymentItem{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		BillCode:   body.BillCode,
		Bag:        body.Bag,
	}
	paymentR.Status = constants.STATUS_ENABLE

	list, err := paymentR.FindAll(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total":     len(list),
		"data":      list,
		"e_invoice": singlPaymentR.IsEInvoice,
	}

	okResponse(c, res)
}

/*
Xoá payment item
*/
func (_ *CPayment) DeleteSinglePaymentItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.DeleteSinglePaymentDetailBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BillCode + "|" + body.Bag
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	paymentItem := model_payment.SinglePaymentItem{}
	paymentItem.Uid = body.SinglePaymentItemUid

	errFindPaymentItem := paymentItem.FindFirst(db)
	if errFindPaymentItem != nil {
		response_message.InternalServerError(c, errFindPaymentItem.Error())
		return
	}

	paymentItem.Status = constants.STATUS_DELETE
	errUdpItem := paymentItem.Update(db)
	if errUdpItem != nil {
		log.Println("DeleteSinglePaymentItem errUdpItem ", errUdpItem.Error())
	}

	//find single payment
	singlePayment := model_payment.SinglePayment{}
	singlePayment.Uid = paymentItem.PaymentUid
	errFS := singlePayment.FindFirst(db)
	if errFS == nil {
		singlePayment.UpdateTotalPaid(db)
	} else {
		log.Println("DeleteSinglePaymentItem errFS", errFS.Error())
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  paymentItem.PartnerUid,
		CourseUid:   paymentItem.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_PAYMENT_SINGLE,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: paymentItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         paymentItem.Bag,
		BookingDate: paymentItem.BookingDate,
		BillCode:    paymentItem.BillCode,
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okRes(c)
}

///  -------------- Agency Payment -------------
/*
 */
func (_ *CPayment) GetListAgencyPayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetListAgencyPaymentBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PartnerUid + "|" + body.CourseUid + "|" + body.PaymentDate + "|" + body.AgencyName + "|" + body.PaymentStatus
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	page := models.Page{
		Limit:   body.PageRequest.Limit,
		Page:    body.PageRequest.Page,
		SortBy:  body.PageRequest.SortBy,
		SortDir: body.PageRequest.SortDir,
	}

	paymentR := model_payment.AgencyPayment{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		PaymentDate: body.PaymentDate,
	}

	list, total, err := paymentR.FindList(db, page)
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

/*
Thanh toán cho agency
*/
func (_ *CPayment) CreateAgencyPaymentItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateAgencyPaymentItemBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	amountStr := strconv.FormatInt(body.Amount, 10)
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PaymentType + "|" + body.AgencyPaymentUid + "|" + amountStr + "|" + body.BookingCode + "|" + body.DateStr
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	// Find agency
	// Find agencyPayment
	agencyPayment := model_payment.AgencyPayment{}
	agencyPayment.Uid = body.AgencyPaymentUid
	errF := agencyPayment.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	agencyPaymentItem := model_payment.AgencyPaymentItem{
		PartnerUid:  agencyPayment.PartnerUid,
		CourseUid:   agencyPayment.CourseUid,
		PaymentUid:  body.AgencyPaymentUid,
		PaymentType: body.PaymentType,
		Paid:        body.Amount,
		BookingCode: agencyPayment.BookingCode,
		Cashiers:    prof.UserName,
		BookingDate: agencyPayment.BookingDate,
		Note:        body.Note,
	}

	errC := agencyPaymentItem.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//Update agency payment
	agencyPayment.UpdateTotalPaid(db)

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  agencyPaymentItem.PartnerUid,
		CourseUid:   agencyPaymentItem.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_PAYMENT_AGENCY,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: agencyPaymentItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         "",
		BookingDate: agencyPaymentItem.BookingDate,
		BillCode:    "",
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okRes(c)
}

/*
Xoá agency payment item
*/
func (_ *CPayment) DeleteAgencyPaymentItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.DeleteAgencyPaymentDetailBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BookingCode + "|" + body.AgencyPaymentItemUid
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	paymentItem := model_payment.AgencyPaymentItem{}
	paymentItem.Uid = body.AgencyPaymentItemUid

	errFindPaymentItem := paymentItem.FindFirst(db)
	if errFindPaymentItem != nil {
		response_message.InternalServerError(c, errFindPaymentItem.Error())
		return
	}

	paymentItem.Status = constants.STATUS_DELETE
	errUdpItem := paymentItem.Update(db)
	if errUdpItem != nil {
		log.Println("DeleteSinglePaymentItem errUdpItem ", errUdpItem.Error())
	}

	// find agency payment
	agencyPayment := model_payment.AgencyPayment{}
	agencyPayment.Uid = paymentItem.PaymentUid
	errFAP := agencyPayment.FindFirst(db)
	if errFAP == nil {
		agencyPayment.UpdateTotalPaid(db)
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  paymentItem.PartnerUid,
		CourseUid:   paymentItem.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_PAYMENT_AGENCY,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: paymentItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         "",
		BookingDate: paymentItem.BookingDate,
		BillCode:    "",
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ *CPayment) GetListAgencyPaymentItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetListAgencyPaymentItemBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PartnerUid + "|" + body.CourseUid + "|" + body.BookingCode + "|" + body.PaymentUid
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	paymentR := model_payment.AgencyPaymentItem{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		PaymentUid: body.PaymentUid,
	}

	list, err := paymentR.FindAll(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": len(list),
		"data":  list,
	}

	okResponse(c, res)
}

/*
Lấy chi tiết số tiền agency thanh toán cho bag
*/
func (_ *CPayment) GetAgencyPayForBagDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetAgencyPayForBagDetailBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BookingCode + "|" + body.BookingUid
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	payForBag := model_payment.BookingAgencyPayment{
		BookingCode: body.BookingCode,
		BookingUid:  body.BookingUid,
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	err := payForBag.FindFirst(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if booking.CheckAgencyPaidAll() {
		serviceGolfs := model_booking.BookingServiceItem{
			BillCode: booking.BillCode,
		}
		listGolfService, _ := serviceGolfs.FindAll(db)

		hasCaddie := false
		hasBuggy := false

		agencyBuggyBookingFee := int64(0)
		agencyCaddieBookingFee := int64(0)

		for _, item := range listGolfService {
			if item.ServiceType == constants.BUGGY_SETTING {
				hasBuggy = true
			}
			if item.ServiceType == constants.CADDIE_SETTING {
				hasCaddie = true
			}
			if item.Type == constants.AGENCY_PAID_ALL_BUGGY {
				agencyBuggyBookingFee = item.Amount
			}
			if item.Type == constants.AGENCY_PAID_ALL_CADDIE {
				agencyCaddieBookingFee = item.Amount
			}
		}

		booking.FindServiceItemsOfBag(db)
		payForBag.FeeData = utils.ListBookingAgencyPayForBagData{}
		payForBag.FeeData = append(payForBag.FeeData, utils.BookingAgencyPayForBagData{
			Type: constants.BOOKING_AGENCY_GOLF_FEE,
			Fee:  booking.CurrentBagPrice.GolfFee,
			Name: "Golf Fee",
		})

		if !hasBuggy {
			payForBag.FeeData = append(payForBag.FeeData, utils.BookingAgencyPayForBagData{
				Type: constants.BOOKING_AGENCY_BUGGY_FEE,
				Fee:  agencyBuggyBookingFee,
				Name: "Thuê xe (1/2 xe)",
			})
		}

		if !hasCaddie {
			payForBag.FeeData = append(payForBag.FeeData, utils.BookingAgencyPayForBagData{
				Type: constants.BOOKING_AGENCY_BOOKING_CADDIE_FEE,
				Fee:  agencyCaddieBookingFee,
				Name: constants.BOOKING_CADDIE_NAME,
			})
		}

		for _, item := range booking.ListServiceItems {
			if !(item.Type == constants.AGENCY_PAID_ALL_BUGGY || item.Type == constants.AGENCY_PAID_ALL_CADDIE) {
				payForBag.FeeData = append(payForBag.FeeData, utils.BookingAgencyPayForBagData{
					Type: item.Type,
					Fee:  item.Amount,
					Name: item.Name,
				})
			}
		}
	}

	okResponse(c, payForBag)
}

/*
Lấy chi tiết số tiền agency thanh toán cho bag
*/
func (_ *CPayment) GetListBagOfAgency(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetBagsOfAgencyBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Checksum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.PartnerUid + "|" + body.CourseUid + "|" + body.BookingCode
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	bagPayOfAgencyR := model_payment.SinglePayment{
		BookingCode: body.BookingCode,
		PartnerUid:  body.PartnerUid,
	}

	list, err := bagPayOfAgencyR.FindAllForAgency(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, list)
}

/*
Check E-Invoice
*/
func (_ *CPayment) GetEInvoice(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetEInvoice{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	singlPaymentR := model_payment.SinglePayment{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		BillCode:   body.BillCode,
	}

	if errSinglePaymentR := singlPaymentR.FindFirst(db); errSinglePaymentR != nil {
		response_message.InternalServerError(c, errSinglePaymentR.Error())
		return
	}

	if body.EInvoice != nil && *body.EInvoice {
		singlPaymentR.IsEInvoice = 1
	} else {
		singlPaymentR.IsEInvoice = 0
	}

	errUdp := singlPaymentR.Update(db)

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okRes(c)
}
