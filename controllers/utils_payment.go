package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	"start/utils"

	"gorm.io/gorm"
)

/*
Create Single Payment
*/
func handleSinglePayment(db *gorm.DB, booking model_booking.Booking) {
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

			agencyPaid := model_payment.BookingAgencyPayment{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				BookingCode: booking.BookingCode,
				AgencyId:    booking.AgencyId,
				BookingUid:  booking.Uid,
			}
			if errFindAgency := agencyPaid.FindFirst(db); errFindAgency == nil {
				singlePayment.AgencyPaid = agencyPaid.GetTotalFee()
			}
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

		if booking.AgencyId > 0 && booking.MemberCardUid == "" {
			// Agency
			singlePayment.Type = constants.PAYMENT_CATE_TYPE_AGENCY
			agencyPaid := model_payment.BookingAgencyPayment{
				PartnerUid:  booking.PartnerUid,
				CourseUid:   booking.CourseUid,
				BookingCode: booking.BookingCode,
				AgencyId:    booking.AgencyId,
				BookingUid:  booking.Uid,
			}
			if errFindAgency := agencyPaid.FindFirst(db); errFindAgency == nil {
				singlePayment.AgencyPaid = agencyPaid.GetTotalFee()
			}
			// singlePayment.AgencyPaid = agencyPaid
		} else {
			// Single
			singlePayment.Type = constants.PAYMENT_CATE_TYPE_SINGLE
		}

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
		if booking.CheckAgencyPaidAll() {
			updateBookingAgencyPaymentForAllFee(booking)
		}
		handleAgencyPayment(db, booking)
	}
	// single payment
	handleSinglePayment(db, booking)
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
	}

	isUpdate := false

	if errFind := bookingAgencyPayment.FindFirst(db); errFind == nil {
		isUpdate = true
	}

	if booking.AgencyPaidAll != nil && *booking.AgencyPaidAll {

		caddieBooking := model_booking.BookingServiceItem{
			Type:     "AGENCY_PAID_ALL_CADDIE",
			Quality:  1,
			Amount:   feeInfo.CaddieFee,
			BillCode: booking.BillCode,
			Location: constants.SERVICE_ITEM_ADD_BY_MANUAL,
		}
		buggyBooking := model_booking.BookingServiceItem{
			Type:     "AGENCY_PAID_ALL_BUGGY",
			Quality:  1,
			Amount:   feeInfo.BuggyFee,
			BillCode: booking.BillCode,
			Location: constants.SERVICE_ITEM_ADD_BY_MANUAL,
		}

		caddieBooking.Create(db)
		buggyBooking.Create(db)

		booking.UpdatePriceDetailCurrentBag(db)
		booking.UpdateMushPay(db)
		booking.Update(db)

		updateBookingAgencyPaymentForAllFee(booking)
		handleSinglePayment(db, booking)
		//Upd lại số tiền thanh toán của agency
		handleAgencyPayment(db, booking)
	}

	bookingAgencyPayment.FeeData = []utils.BookingAgencyPayForBagData{}

	if feeInfo.GolfFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.GolfFee,
			Name: "Golf Fee",
			Type: constants.BOOKING_AGENCY_GOLF_FEE,
			Hole: booking.Hole,
		})
	}
	if feeInfo.BuggyFee > 0 {
		name := ""
		if *booking.IsPrivateBuggy {
			name = "Thuê riêng xe"
		} else {
			name = "Thuê xe (1/2 xe)"
		}
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.BuggyFee,
			Name: name,
			Type: constants.BOOKING_AGENCY_BUGGY_FEE,
			Hole: booking.Hole,
		})
	}
	if feeInfo.CaddieFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.CaddieFee,
			Name: "Booking Caddie",
			Type: constants.BOOKING_AGENCY_BOOKING_CADDIE_FEE,
			Hole: booking.Hole,
		})
	}

	// Ghi nhận số tiền agency thanh toán của agency
	bookingAgencyPayment.CaddieId = fmt.Sprint(booking.CaddieId)
	if isUpdate {
		bookingAgencyPayment.Update(db)
	} else {
		bookingAgencyPayment.Create(db)
	}

	go func() {
		// create bag payment
		// Ghi nhận só tiền agency thanh toán cho bag đó
		booking.AgencyPaid = bookingAgencyPayment.FeeData

		// update giá cho bag(main or sub nếu có)
		updatePriceWithServiceItem(booking, models.CmsUser{})

		handleSinglePayment(db, booking)
		//Upd lại số tiền thanh toán của agency
		handleAgencyPayment(db, booking)
	}()
}

func addBuggyFee(booking model_booking.Booking, fee int64, name string) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	serviceItem := model_booking.BookingServiceItem{
		BillCode:   booking.BillCode,
		PlayerName: booking.CustomerName,
		BookingUid: booking.Uid,
	}
	serviceItem.Name = name
	serviceItem.UnitPrice = fee
	serviceItem.Quality = 1
	serviceItem.Amount = fee
	serviceItem.Type = constants.GOLF_SERVICE_RENTAL
	serviceItem.ServiceType = constants.BUGGY_SETTING
	serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION
	serviceItem.Create(db)
}

func updateBookingAgencyPaymentForAllFee(booking model_booking.Booking) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	bookingAgencyPayment := model_payment.BookingAgencyPayment{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingCode: booking.BookingCode,
		AgencyId:    booking.AgencyId,
		BookingUid:  booking.Uid,
	}
	if errFind := bookingAgencyPayment.FindFirst(db); errFind != nil {
		bookingAgencyPayment.FeeData = utils.ListBookingAgencyPayForBagData{}
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Type: constants.BOOKING_AGENCY_PAID_ALL,
			Fee:  booking.GetAgencyPaid(),
		})
		bookingAgencyPayment.Create(db)
	} else {
		bookingAgencyPayment.FeeData = utils.ListBookingAgencyPayForBagData{}
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Type: constants.BOOKING_AGENCY_PAID_ALL,
			Fee:  booking.GetAgencyPaid(),
		})
		bookingAgencyPayment.Update(db)
	}
}

func updateSinglePaymentOfSubBag(mainBag model_booking.Booking, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(mainBag.PartnerUid)
	for _, subBooking := range mainBag.SubBags {
		bookingR := model_booking.Booking{
			Model: models.Model{Uid: subBooking.BookingUid},
		}

		booking, errF := bookingR.FindFirstByUId(db)
		if errF == nil {
			totalPaid := booking.CurrentBagPrice.MainBagPaid

			singlePayment := model_payment.SinglePayment{
				PartnerUid: booking.PartnerUid,
				CourseUid:  booking.CourseUid,
				BillCode:   booking.BillCode,
			}
			singlePayment.Status = constants.STATUS_ENABLE
			if errFind := singlePayment.FindFirst(db); errFind == nil {
				singlePaymentItem := model_payment.SinglePaymentItem{
					PartnerUid:  booking.PartnerUid,
					CourseUid:   booking.CourseUid,
					BookingUid:  booking.Uid,
					BillCode:    booking.BillCode,
					Bag:         booking.Bag,
					Paid:        totalPaid,
					PaymentType: constants.PAYMENT_STATUS_PREPAID,
					PaymentUid:  singlePayment.Uid,
					Cashiers:    prof.UserName,
					BookingDate: booking.BookingDate,
				}
				if errC := singlePaymentItem.Create(db); errC == nil {
					singlePayment.UpdateTotalPaid(db)
					handleSinglePayment(db, booking)
				}
			}
		}
	}
}
