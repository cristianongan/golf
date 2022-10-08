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
	"time"

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
	log.Println("CreateSinglePayment checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("CreateSinglePayment checkSum ", checkSum)

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

	toDayDate, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

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
		singlePayment.BagInfo = bagInfo
		singlePayment.TotalPaid = body.Amount
		singlePayment.Note = body.Note
		singlePayment.Cashiers = prof.UserName
		singlePayment.PaymentDate = toDayDate

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
	}

	errC := singlePaymentItem.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

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

			singlePayment.TotalPaid = totalPaid
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
	log.Println("GetListPayment checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("GetListPayment checkSum ", checkSum)

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
	}

	list, total, err := paymentR.FindList(db, page, body.PlayerName)
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
	log.Println("UpdateSinglePayment checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("UpdateSinglePayment checkSum ", checkSum)

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
	log.Println("GetListSinglePaymentDetail checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("GetListSinglePaymentDetail checkSum ", checkSum)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	paymentR := model_payment.SinglePaymentItem{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		BillCode:   body.BillCode,
		Bag:        body.Bag,
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
	log.Println("UpdateSinglePayment checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("UpdateSinglePayment checkSum ", checkSum)

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

	okRes(c)
}
