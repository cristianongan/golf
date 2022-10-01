package controllers

import (
	"encoding/json"
	"log"
	"start/config"
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

	singlePayment := model_payment.SinglePayment{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		Bag:         booking.Bag,
		BookingUid:  booking.Uid,
		BillCode:    booking.BillCode,
		BookingDate: booking.BookingDate,
		BagInfo:     bagInfo,
		PaymentType: body.PaymentType,
		TotalPaid:   body.Amount,
		Note:        body.Note,
		Cashiers:    prof.UserName,
		PaymentDate: toDayDate,
	}

	errC := singlePayment.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
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
