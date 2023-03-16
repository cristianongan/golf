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
	"time"

	"github.com/bsm/redislock"
	"gorm.io/gorm"
)

/*
Create Single Payment
*/
func handleSinglePayment(db *gorm.DB, booking model_booking.Booking) {
	redisKey := utils.GetRedisKeySinglePaymentCreated(booking.PartnerUid, booking.CourseUid, booking.BillCode)
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 10*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[PAYMENT] handleSinglePayment Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

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
	redisKey := utils.GetRedisKeyAgencyPaymentCreated(booking.PartnerUid, booking.CourseUid, booking.BookingCode)
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 10*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[PAYMENT] handleAgencyPayment Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

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

		updatePriceWithServiceItem(&booking, models.CmsUser{})

		updateBookingAgencyPaymentForAllFee(booking)
		handleSinglePayment(db, booking)
		//Upd lại số tiền thanh toán của agency
		handleAgencyPayment(db, booking)
		return
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
	if feeInfo.PrivateCarFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.PrivateCarFee,
			Name: constants.THUE_RIENG_XE,
			Type: constants.BOOKING_AGENCY_PRIVATE_CAR_FEE,
			Hole: booking.Hole,
		})
	}
	if feeInfo.BuggyFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.BuggyFee,
			Name: constants.THUE_NUA_XE,
			Type: constants.BOOKING_AGENCY_BUGGY_FEE,
			Hole: booking.Hole,
		})
	}
	if feeInfo.CaddieFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.CaddieFee,
			Name: constants.BOOKING_CADDIE_NAME,
			Type: constants.BOOKING_AGENCY_BOOKING_CADDIE_FEE,
			Hole: booking.Hole,
		})
	}
	if feeInfo.OddCarFee > 0 {
		bookingAgencyPayment.FeeData = append(bookingAgencyPayment.FeeData, utils.BookingAgencyPayForBagData{
			Fee:  feeInfo.OddCarFee,
			Name: constants.THUE_LE_XE,
			Type: constants.BOOKING_AGENCY_BUGGY_ODD_FEE,
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

	// go func() {
	// create bag payment
	// Ghi nhận só tiền agency thanh toán cho bag đó
	booking.AgencyPrePaid = bookingAgencyPayment.FeeData

	// update giá cho bag(main or sub nếu có)
	updatePriceWithServiceItem(&booking, models.CmsUser{})
	// }()
}

func addBuggyFee(booking model_booking.Booking, fee int64, name string, hole int) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	serviceItem := model_booking.BookingServiceItem{
		BillCode:   booking.BillCode,
		PlayerName: booking.CustomerName,
		BookingUid: booking.Uid,
	}
	serviceItem.PartnerUid = booking.PartnerUid
	serviceItem.CourseUid = booking.CourseUid
	serviceItem.Name = name
	serviceItem.UnitPrice = fee
	serviceItem.Quality = 1
	serviceItem.Amount = fee
	serviceItem.Hole = hole
	serviceItem.Type = constants.GOLF_SERVICE_RENTAL
	serviceItem.ServiceType = constants.BUGGY_SETTING
	serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION
	serviceItem.Create(db)
}

func addCaddieBookingFee(booking model_booking.Booking, fee int64, name string, hole int) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	serviceItem := model_booking.BookingServiceItem{
		BillCode:   booking.BillCode,
		PlayerName: booking.CustomerName,
		BookingUid: booking.Uid,
	}
	serviceItem.PartnerUid = booking.PartnerUid
	serviceItem.CourseUid = booking.CourseUid
	serviceItem.Name = name
	serviceItem.UnitPrice = fee
	serviceItem.Quality = 1
	serviceItem.Amount = fee
	serviceItem.Hole = hole
	serviceItem.Type = constants.GOLF_SERVICE_RENTAL
	serviceItem.ServiceType = constants.CADDIE_SETTING
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

/*
  Xoa single payment -> udp status = delete
*/
func deleteSinglePayment(pUid, cUid, billCode, bookUid string, agencyId int64, bookingCode string) {
	db := datasources.GetDatabaseWithPartner(pUid)
	// Xoa payment
	singlePayment := model_payment.SinglePayment{
		PartnerUid: pUid,
		CourseUid:  cUid,
		BillCode:   billCode,
	}

	errFP := singlePayment.FindFirst(db)
	if errFP == nil {
		payDel := model_payment.SinglePaymentDel{
			SinglePayment: singlePayment,
		}
		errUdpP := singlePayment.Delete(db)
		log.Println("[PAYMENT] deleteSinglePayment uid", singlePayment.Bag, singlePayment.BookingDate, singlePayment.BookingUid)
		if errUdpP != nil {
			log.Println("[PAYMENT] deleteSinglePayment errUdpP", errUdpP)
		} else {
			go createSinglePaymentDel(payDel, db)
		}
	}

	//Update lại Agency Payment
	if agencyId > 0 {
		booking := model_booking.Booking{}
		booking.Uid = bookUid
		errFB := booking.FindFirst(db)
		if errFB != nil {
			log.Println("[PAYMENT] deleteSinglePayment err find booking", errFB.Error())
			agencyPayment := model_payment.AgencyPayment{
				PartnerUid:  pUid,
				CourseUid:   cUid,
				BookingCode: bookingCode,
			}
			agencyPayment.Status = constants.STATUS_ENABLE

			errFind := agencyPayment.FindFirst(db)
			if errFind == nil {
				agencyPayment.UpdateTotalAmount(db, true)
			}
			return
		}
		// Agency payment
		if booking.CheckAgencyPaidAll() {
			updateBookingAgencyPaymentForAllFee(booking)
		}
		handleAgencyPayment(db, booking)
	}
}

/*
 Tạo các single payment để trace khi bị xoá
*/
func createSinglePaymentDel(singlePaymentDel model_payment.SinglePaymentDel, db *gorm.DB) {
	err := singlePaymentDel.Create(db)
	if err != nil {
		log.Println("[PAYMENT] createSinglePaymentDel err", err.Error())
	}
}
