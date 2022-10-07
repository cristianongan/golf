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

func (_ *CPayment) UpdateSinglePayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UpdateSinglePaymentBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	checkSumMessage := config.GetPaymentSecretKey() + "|" + body.BookingUid + "|" + body.PaymentUid + "|" + body.DateStr
	log.Println("UpdateSinglePayment checkSumMessage ", checkSumMessage)
	checkSum := utils.GetSHA256Hash(checkSumMessage)
	log.Println("UpdateSinglePayment checkSum ", checkSum)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	// payment
	payment := model_payment.SinglePayment{}
	payment.Uid = body.PaymentUid
	errF := payment.FindFirst(db)

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	payment.Note = body.Note
	errUdp := payment.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okRes(c)
}
