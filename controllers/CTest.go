package controllers

import (
	"encoding/json"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/socket"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
	"github.com/twharmon/slices"
)

type CTest struct{}

func (_ *CTest) CreateRevenueDetail(c *gin.Context, prof models.CmsUser) {

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = form.BookingDate

	db1, _, _ := bookings.FindAllBookingList(db)

	var list []model_booking.Booking
	db1.Find(&list)

	for _, item := range list {
		mushPay := model_booking.BookingMushPay{}

		listRoundGolfFee := []models.Round{}
		hole := 0
		fbFee := int64(0)
		rentalFee := int64(0)
		buggyFee := int64(0)
		practiceBallFee := int64(0)
		proshopFee := int64(0)
		otherFee := int64(0)

		roundToFindList := models.Round{BillCode: item.BillCode}
		listRoundOfCurrentBag, _ := roundToFindList.FindAll(db)

		for _, round := range listRoundOfCurrentBag {
			listRoundGolfFee = append(listRoundGolfFee, round)
		}

		hole = slices.Reduce(listRoundGolfFee, func(prev int, item models.Round) int {
			return prev + item.Hole
		})

		bookingCaddieFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
			return prev + item.CaddieFee
		})

		bookingBuggyFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
			return prev + item.BuggyFee
		})

		bookingGreenFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
			return prev + item.GreenFee
		})

		totalGolfFeeOfSubBag := bookingCaddieFee + bookingBuggyFee + bookingGreenFee
		mushPay.TotalGolfFee = totalGolfFeeOfSubBag

		// SubBag

		// Sub Service Item của current Bag
		// Get item for current Bag
		// update lại lấy service items mới
		item.FindServiceItems(db)
		for _, v := range item.ListServiceItems {
			if v.BillCode == item.BillCode {
				if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT || v.Type == constants.MINI_B_SETTING || v.Type == constants.MINI_R_SETTING {
					fbFee += v.Amount
				} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RENTAL || v.Type == constants.DRIVING_SETTING {
					if v.ItemCode == "R1_3" {
						practiceBallFee += v.Amount
					} else {
						rentalFee += v.Amount
					}
				} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP {
					proshopFee += v.Amount
				} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
					otherFee += v.Amount
				} else if v.Type == constants.BUGGY_SETTING {
					buggyFee += v.Amount
				}
			}
		}

		RSinglePaymentItem := model_payment.SinglePaymentItem{
			Bag:         item.Bag,
			PartnerUid:  item.PartnerUid,
			CourseUid:   item.CourseUid,
			BookingDate: item.BookingDate,
		}

		list, _ := RSinglePaymentItem.FindAll(db)

		cashList := []model_payment.SinglePaymentItem{}
		debtList := []model_payment.SinglePaymentItem{}
		cardList := []model_payment.SinglePaymentItem{}

		for _, item := range list {
			if item.PaymentType == constants.PAYMENT_TYPE_CASH {
				cashList = append(cashList, item)
			} else if item.PaymentType == constants.PAYMENT_STATUS_DEBT {
				debtList = append(debtList, item)
			} else {
				cardList = append(cardList, item)
			}
		}

		cashTotal := slices.Reduce(cashList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
			return prev + item.Paid
		})

		debtTotal := slices.Reduce(debtList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
			return prev + item.Paid
		})

		cardTotal := slices.Reduce(cardList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
			return prev + item.Paid
		})

		m := model_report.ReportRevenueDetail{
			PartnerUid:     item.PartnerUid,
			CourseUid:      item.CourseUid,
			BillNo:         "",
			Bag:            item.Bag,
			GuestStyle:     item.GuestStyle,
			GuestStyleName: item.GuestStyleName,
			BookingDate:    item.BookingDate,
			CustomerId:     item.CustomerUid,
			MembershipNo:   item.CardId,
			CustomerType:   item.CustomerType,
			Hole:           hole,
			GreenFee:       bookingGreenFee,
			CaddieFee:      bookingCaddieFee,
			FBFee:          fbFee,
			RentalFee:      rentalFee,
			BuggyFee:       buggyFee,
			ProshopFee:     proshopFee,
			PraticeBallFee: practiceBallFee,
			OtherFee:       otherFee,
			MushPay:        item.MushPayInfo.MushPay,
			Cash:           cashTotal,
			Debit:          debtTotal,
			Card:           cardTotal,
		}

		m.Create(db)
	}

	okRes(c)
}

func (cBooking *CTest) Test(c *gin.Context, prof models.CmsUser) {
	// db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// form := request.GetListBookingForm{}
	// if bindErr := c.ShouldBind(&form); bindErr != nil {
	// 	response_message.BadRequest(c, bindErr.Error())
	// 	return
	// }

	// if form.Bag == "" {
	// 	response_message.BadRequest(c, errors.New("Bag invalid").Error())
	// 	return
	// }

	// booking := model_booking.Booking{}
	// booking.PartnerUid = form.PartnerUid
	// booking.CourseUid = form.CourseUid
	// booking.Bag = form.Bag

	// if form.BookingDate != "" {
	// 	booking.BookingDate = form.BookingDate
	// } else {
	// 	toDayDate, errD := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	// 	if errD != nil {
	// 		response_message.InternalServerError(c, errD.Error())
	// 		return
	// 	}
	// 	booking.BookingDate = toDayDate
	// }

	// errF := booking.FindFirst(db)
	// if errF != nil {
	// 	response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
	// 	return
	// }

	// booking.UpdateMushPay(db)
	// booking.Update(db)

	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_CADDIE_WORKING_STATUS_UPDATE,
		"title": "",
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	// socket.HubBroadcastSocket = socket.NewHub()
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}
