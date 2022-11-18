package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	model_booking "start/models/booking"
	model_payment "start/models/payment"

	"gorm.io/gorm"
)

/*
Create Single Payment
*/
func handleSinglePayment(db *gorm.DB, booking model_booking.Booking, agencyPaid int64) {
	bagInfo := model_payment.PaymentBagInfo{}
	bagByte, errM := json.Marshal(booking)
	if errM != nil {
		log.Println("CreateSinglePayment errM", errM.Error())
	}
	errUM := json.Unmarshal(bagByte, &bagInfo)
	if errUM != nil {
		log.Println("CreateSinglePayment errUM", errUM.Error())
	}

	singlePayment := model_payment.SinglePayment{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BillCode:   booking.BillCode,
	}
	singlePayment.Status = constants.STATUS_ENABLE

	errFind := singlePayment.FindFirst(db)
	if errFind != nil {
		// Chưa có thì tạo
		singlePayment.Bag = booking.Bag
		singlePayment.BookingUid = booking.Uid
		singlePayment.BillCode = booking.BillCode
		singlePayment.BookingDate = booking.BookingDate
		singlePayment.BookingCode = booking.BookingCode
		singlePayment.BagInfo = bagInfo
		singlePayment.PaymentDate = booking.BookingDate

		if booking.AgencyId > 0 && booking.MemberCardUid == "" {
			// Agency
			singlePayment.Type = constants.PAYMENT_CATE_TYPE_AGENCY
			singlePayment.AgencyPaid = agencyPaid
		} else {
			// Single
			singlePayment.Type = constants.PAYMENT_CATE_TYPE_SINGLE
		}

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
			log.Println("handleSinglePayment errC", errC.Error())
			return
		}
	} else {
		singlePayment.Bag = booking.Bag
		singlePayment.BookingCode = booking.BookingCode
		singlePayment.PaymentDate = booking.BookingDate
		singlePayment.BagInfo = bagInfo
		singlePayment.UpdatePaymentStatus(booking.BagStatus, db)
		errUdp := singlePayment.Update(db)
		if errUdp != nil {
			log.Println("handleSinglePayment errUdp", errUdp.Error())
		}
	}
}

/*
Create Agency Payment
*/
func handleAgencyPayment(db *gorm.DB, booking model_booking.Booking) {
	agencyInfo := model_payment.PaymentAgencyInfo{
		Name:           booking.AgencyInfo.Name,
		GuestStyle:     booking.GuestStyle,
		GuestStyleName: booking.GuestStyleName,
	}
	agencyInfo.Id = booking.AgencyId

	agencyPayment := model_payment.AgencyPayment{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingCode: booking.BookingCode,
	}
	agencyPayment.Status = constants.STATUS_ENABLE

	errFind := agencyPayment.FindFirst(db)
	if errFind != nil {
		// Chưa có thì tạo
		agencyPayment.BookingDate = booking.BookingDate
		agencyPayment.BookingCode = booking.BookingCode
		agencyPayment.AgencyInfo = agencyInfo
		agencyPayment.AgencyId = booking.AgencyId
		agencyPayment.TotalPaid = 0
		agencyPayment.Note = ""
		agencyPayment.Cashiers = ""
		agencyPayment.PaymentDate = ""

		//Find prepaid from booking
		if booking.BookingCode != "" {
			agencyPayment.UpdatePlayBookInfo(db, booking)
		}

		// Update total Amount
		agencyPayment.UpdateTotalAmount(db, false)
		// Update payment status
		errC := agencyPayment.Create(db)

		if errC != nil {
			log.Println("handleSinglePayment errC", errC.Error())
			return
		}
	} else {
		agencyPayment.BookingCode = booking.BookingCode
		agencyPayment.AgencyInfo = agencyInfo
		agencyPayment.AgencyId = booking.AgencyId
		agencyPayment.UpdatePlayBookInfo(db, booking)
		agencyPayment.UpdateTotalAmount(db, false)
		errUdp := agencyPayment.Update(db)
		if errUdp != nil {
			log.Println("handleSinglePayment errUdp", errUdp.Error())
		}
	}
}

// Handle Payment
func handlePayment(db *gorm.DB, booking model_booking.Booking) {
	if booking.AgencyId > 0 && booking.MemberCardUid == "" {
		// Agency payment
		go handleAgencyPayment(db, booking)
	}
	// single payment
	go handleSinglePayment(db, booking, 0)
}

/*
 Handle agency Paid
 Xứ lý tính toán số tiền Agency đã thanh toán
*/
func handleAgencyPaid(booking model_booking.Booking, feeInfo request.AgencyFeeInfo) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	bookingAgencyPayment := model_payment.BookingAgencyPayment{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingCode: booking.BookingCode,
		AgencyId:    booking.AgencyId,
		BookingUid:  booking.Uid,
		CaddieId:    fmt.Sprint(booking.CaddieId),
	}

	// feeInfo := bodyRequest.BookingList[index].FeeInfo
	if feeInfo.GolfFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, model_payment.BookingAgencyPayForBagData{
			Fee:  feeInfo.GolfFee,
			Name: "Golf Fee",
			Type: constants.BOOKING_AGENCY_GOLF_FEE,
		})
	}
	if feeInfo.BuggyFee > 0 {
		name := ""
		if *booking.IsPrivateBuggy {
			name = "Buggy (1 xe)"
		} else {
			name = "Buggy (1/2 xe)"
		}
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, model_payment.BookingAgencyPayForBagData{
			Fee:  feeInfo.BuggyFee,
			Name: name,
			Type: constants.BOOKING_AGENCY_BUGGY_FEE,
		})
		serviceItem := model_booking.BookingServiceItem{
			BillCode:   booking.BillCode,
			PlayerName: booking.CustomerName,
			BookingUid: booking.Uid,
		}
		serviceItem.Name = name
		serviceItem.UnitPrice = feeInfo.BuggyFee
		serviceItem.Amount = feeInfo.BuggyFee
		serviceItem.Type = constants.GOLF_SERVICE_RENTAL
		serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION
		serviceItem.PaidBy = constants.PAID_BY_AGENCY
		serviceItem.Create(datasources.GetDatabaseWithPartner(booking.PartnerUid))
	}
	if feeInfo.CaddieFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, model_payment.BookingAgencyPayForBagData{
			Fee:  feeInfo.CaddieFee,
			Name: "Booking Caddie fee",
			Type: constants.BOOKING_AGENCY_BOOKING_CADDIE_FEE,
		})
		serviceItem := model_booking.BookingServiceItem{
			BillCode:   booking.BillCode,
			PlayerName: booking.CustomerName,
			BookingUid: booking.Uid,
		}
		serviceItem.Name = "Booking Caddie"
		serviceItem.UnitPrice = feeInfo.CaddieFee
		serviceItem.Amount = feeInfo.CaddieFee
		serviceItem.Type = constants.GOLF_SERVICE_RENTAL
		serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION
		serviceItem.PaidBy = constants.PAID_BY_AGENCY
		serviceItem.Create(db)
	}

	if feeInfo.BuggyFee > 0 || feeInfo.CaddieFee > 0 || feeInfo.GolfFee > 0 {
		// Ghi nhận số tiền agency thanh toán của agency
		bookingAgencyPayment.Create(db)

		go func() {
			// create bag payment
			// Ghi nhận só tiền agency thanh toán cho bag đó
			booking.UpdatePriceDetailCurrentBag(db)
			booking.UpdateMushPay(db)
			booking.Update(db)

			handleSinglePayment(db, booking, bookingAgencyPayment.GetTotalFee())
			//Upd lại số tiền thanh toán của agency
			handleAgencyPayment(db, booking)
		}()
	}
}
