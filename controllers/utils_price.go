package controllers

import (
	"encoding/json"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	"start/utils"
	"strings"

	model_report "start/models/report"

	"github.com/twharmon/slices"
	"gorm.io/gorm"
)

// Check Fee Data để lưu vào DB
func formatGolfFee(feeText string) string {
	feeTextFormat0 := strings.TrimSpace(feeText)
	feeTextFormat1 := strings.ReplaceAll(feeTextFormat0, " ", "")
	feeTextFormat2 := strings.ReplaceAll(feeTextFormat1, ",", "")
	feeTextFormatLast := strings.ReplaceAll(feeTextFormat2, ".", "")

	if strings.Contains(feeTextFormatLast, constants.FEE_SEPARATE_CHAR) {
		return feeTextFormatLast
	}
	list1 := strings.Split(feeTextFormatLast, constants.FEE_SEPARATE_CHAR)
	if len(list1) == 0 {
		return feeTextFormat2
	}
	if len(list1) == 1 {
		return list1[0]
	}
	return strings.Join(list1, constants.FEE_SEPARATE_CHAR)
}

/*
	  Tính golf fee cho tạo đơn có guest style
		Là phần tử đầu của list golfFee
*/
func getInitListGolfFeeForBooking(param request.GolfFeeGuestyleParam, golfFee models.GolfFee) (model_booking.ListBookingGolfFee, model_booking.BookingGolfFee) {
	listBookingGolfFee := model_booking.ListBookingGolfFee{}
	bookingGolfFee := model_booking.BookingGolfFee{}
	bookingGolfFee.BookingUid = param.Uid
	bookingGolfFee.Bag = param.Bag
	bookingGolfFee.PlayerName = param.CustomerName
	bookingGolfFee.RoundIndex = 0

	bookingGolfFee.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, param.Hole)
	bookingGolfFee.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, param.Hole)
	bookingGolfFee.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, param.Hole)

	listBookingGolfFee = append(listBookingGolfFee, bookingGolfFee)
	return listBookingGolfFee, bookingGolfFee
}

/*
Tính golf fee cho đơn thqay đổi hố
*/
func getInitGolfFeeForChangeHole(db *gorm.DB, body request.ChangeBookingHole, golfFee models.GolfFee) model_booking.BookingGolfFee {
	holePriceFormula := models.HolePriceFormula{}
	holePriceFormula.Hole = body.Hole
	err := holePriceFormula.FindFirst(db)
	if err != nil {
		log.Println("find hole price err", err.Error())
	}

	bookingGolfFee := model_booking.BookingGolfFee{}

	bookingGolfFee.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, body.Hole)
	bookingGolfFee.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, body.Hole)
	bookingGolfFee.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, body.Hole)

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_SELF && holePriceFormula.StopBySelf != "" {
		bookingGolfFee.CaddieFee = utils.GetFeeWidthHolePrice(golfFee.CaddieFee, body.Hole, holePriceFormula.StopBySelf)
		bookingGolfFee.BuggyFee = utils.GetFeeWidthHolePrice(golfFee.BuggyFee, body.Hole, holePriceFormula.StopBySelf)
		bookingGolfFee.GreenFee = utils.GetFeeWidthHolePrice(golfFee.GreenFee, body.Hole, holePriceFormula.StopBySelf)
	}

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_RAIN && holePriceFormula.StopByRain != "" {
		bookingGolfFee.CaddieFee = utils.GetFeeWidthHolePrice(golfFee.CaddieFee, body.Hole, holePriceFormula.StopByRain)
		bookingGolfFee.BuggyFee = utils.GetFeeWidthHolePrice(golfFee.BuggyFee, body.Hole, holePriceFormula.StopByRain)
		bookingGolfFee.GreenFee = utils.GetFeeWidthHolePrice(golfFee.GreenFee, body.Hole, holePriceFormula.StopByRain)
	}

	return bookingGolfFee
}

/*
Theo giá đặc biệt, k theo GuestStyle
*/
func getInitListGolfFeeWithOutGuestStyleForBooking(param request.GolfFeeGuestyleParam) (model_booking.ListBookingGolfFee, model_booking.BookingGolfFee) {
	listBookingGolfFee := model_booking.ListBookingGolfFee{}
	bookingGolfFee := model_booking.BookingGolfFee{}
	bookingGolfFee.BookingUid = param.Uid
	bookingGolfFee.Bag = param.Bag
	bookingGolfFee.PlayerName = param.CustomerName
	bookingGolfFee.RoundIndex = 0

	bookingGolfFee.CaddieFee = utils.CalculateFeeByHole(param.Hole, param.CaddieFee, param.Rate)
	bookingGolfFee.BuggyFee = utils.CalculateFeeByHole(param.Hole, param.BuggyFee, param.Rate)
	bookingGolfFee.GreenFee = utils.CalculateFeeByHole(param.Hole, param.GreenFee, param.Rate)

	listBookingGolfFee = append(listBookingGolfFee, bookingGolfFee)
	return listBookingGolfFee, bookingGolfFee
}

/*
Update fee when action round
*/
func updateListGolfFeeWithRound(db *gorm.DB, round *models.Round, booking model_booking.Booking, guestStyle string, hole int) {
	// Check giá guest style
	if guestStyle != "" {
		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			GuestStyle: guestStyle,
		}
		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay(db)
		if errFindGF != nil {
			log.Println("golf fee err " + errFindGF.Error())
			return
		}

		// Update fee in round
		round.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, hole)
		round.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, hole)
		round.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, hole)
	} else {
		// Get config course
		course := models.Course{}
		course.Uid = booking.CourseUid
		errCourse := course.FindFirst()
		if errCourse != nil {
			log.Println("course config err " + errCourse.Error())
			return
		}
		// Lấy giá đặc biệt của member card
		if booking.MemberCardUid != "" {
			// Get Member Card
			memberCard := models.MemberCard{}
			memberCard.Uid = booking.MemberCardUid
			errFind := memberCard.FindFirst(db)
			if errFind != nil {
				log.Println("member card err " + errCourse.Error())
				return
			}

			if memberCard.PriceCode == 1 {
				// Update fee in round
				round.BuggyFee = utils.CalculateFeeByHole(hole, memberCard.BuggyFee, course.RateGolfFee)
				round.CaddieFee = utils.CalculateFeeByHole(hole, memberCard.CaddieFee, course.RateGolfFee)
				round.GreenFee = utils.CalculateFeeByHole(hole, memberCard.GreenFee, course.RateGolfFee)
			}
		}

		// Lấy giá đặc biệt của member card
		if booking.AgencyId > 0 {
			agency := models.Agency{}
			agency.Id = booking.AgencyId
			errFindAgency := agency.FindFirst(db)
			if errFindAgency != nil || agency.Id == 0 {
				log.Println("agency err " + errCourse.Error())
				return
			}

			agencySpecialPrice := models.AgencySpecialPrice{
				AgencyId:   agency.Id,
				CourseUid:  agency.CourseUid,
				PartnerUid: agency.PartnerUid,
			}
			errFSP := agencySpecialPrice.FindFirst(db)
			if errFSP == nil && agencySpecialPrice.Id > 0 {
				// Update fee in round
				round.BuggyFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.BuggyFee, course.RateGolfFee)
				round.CaddieFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.CaddieFee, course.RateGolfFee)
				round.GreenFee = utils.CalculateFeeByHole(hole, agencySpecialPrice.GreenFee, course.RateGolfFee)
			}
		}
	}

}

/*
	Booking Init and Update

init price
init Golf Fee
init MushPay
init Rounds
*/
func initPriceForBooking(db *gorm.DB, booking *model_booking.Booking, listBookingGolfFee model_booking.ListBookingGolfFee, bookingGolfFee model_booking.BookingGolfFee) {
	if booking == nil {
		log.Println("initPriceForBooking err booking nil")
		return
	}
	var bookingTemp model_booking.Booking
	bookingTempByte, err0 := json.Marshal(booking)
	if err0 != nil {
		log.Println("initPriceForBooking err0", err0.Error())
	}
	err1 := json.Unmarshal(bookingTempByte, &bookingTemp)
	if err1 != nil {
		log.Println("initPriceForBooking err1", err1.Error())
	}

	booking.ListGolfFee = listBookingGolfFee
	bookingTemp.ListGolfFee = listBookingGolfFee

	// Current Bag Price Detail
	if booking.AgencyId > 0 {
		bookingPayment := model_payment.BookingAgencyPayment{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			BookingUid: booking.Uid,
			AgencyId:   booking.AgencyId,
		}

		list, _ := bookingPayment.FindAll(db)

		if booking.BookingCode != "" && len(list) > 0 {
			return
		}
	}

	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	currentBagPriceDetail.UpdateAmount()

	booking.CurrentBagPrice = currentBagPriceDetail
	bookingTemp.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(bookingTemp)

	booking.MushPayInfo = mushPayInfo
	bookingTemp.MushPayInfo = mushPayInfo

	currencyPaidGet := models.CurrencyPaid{
		Currency: "usd",
	}
	if err := currencyPaidGet.FindFirst(); err == nil {
		booking.CurrentBagPrice.AmountUsd = mushPayInfo.MushPay / currencyPaidGet.Rate
	}
}

func initUpdatePriceBookingForChanegHole(booking *model_booking.Booking, bookingGolfFee model_booking.BookingGolfFee) {
	if booking == nil {
		log.Println("initPriceForBooking err booking nil")
		return
	}
	var bookingTemp model_booking.Booking
	bookingTempByte, err0 := json.Marshal(booking)
	if err0 != nil {
		log.Println("initPriceForBooking err0", err0.Error())
	}
	err1 := json.Unmarshal(bookingTempByte, &bookingTemp)
	if err1 != nil {
		log.Println("initPriceForBooking err1", err1.Error())
	}

	// update last golffee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].GreenFee = bookingGolfFee.GreenFee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].CaddieFee = bookingGolfFee.CaddieFee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].BuggyFee = bookingGolfFee.BuggyFee
	bookingTemp.ListGolfFee = booking.ListGolfFee

	// Current Bag Price Detail
	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	currentBagPriceDetail.UpdateAmount()

	booking.CurrentBagPrice = currentBagPriceDetail
	bookingTemp.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(bookingTemp)

	booking.MushPayInfo = mushPayInfo
}

/*
Init Booking MushPayInfo
*/
func initBookingMushPayInfo(booking model_booking.Booking) model_booking.BookingMushPay {
	mushPayInfo := model_booking.BookingMushPay{}
	mushPayInfo.TotalGolfFee = booking.GetTotalGolfFee()
	mushPayInfo.TotalServiceItem = booking.GetTotalServicesFee()
	mushPayInfo.MushPay = mushPayInfo.TotalGolfFee + mushPayInfo.TotalServiceItem
	return mushPayInfo
}

/*
Update lại gia của
Bag hiện tại
main bag nếu có
sub bag nếu có
*/
func updatePriceWithServiceItem(booking *model_booking.Booking, prof models.CmsUser) {

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	if prof.UserName != "" {
		booking.CmsUser = prof.UserName
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())
	}

	booking.UpdatePriceDetailCurrentBag(db)
	booking.UpdateMushPay(db)
	errUdp := booking.Update(db)
	if errUdp != nil {
		log.Println("updatePriceWithServiceItem errUdp", errUdp.Error())
	} else {
		go handlePayment(db, *booking)
	}

	if booking.MainBags != nil && len(booking.MainBags) > 0 {
		// Nếu bag có Main Bag
		mainBook := model_booking.Booking{
			CourseUid:   booking.CourseUid,
			PartnerUid:  booking.PartnerUid,
			Bag:         booking.MainBags[0].GolfBag,
			BookingDate: booking.BookingDate,
		}

		errFMB := mainBook.FindFirst(db)
		if errFMB != nil {
			log.Println("UpdateMushPay-"+booking.Bag+"-Find Main Bag", errFMB.Error())
		}

		mainBook.UpdateMushPay(db)
		errUdp := mainBook.Update(db)
		if errUdp != nil {
			log.Println("updatePriceWithServiceItem errUdp", errUdp.Error())
		} else {
			go handlePayment(db, mainBook)
		}
	}

	if booking.SubBags != nil && len(booking.SubBags) > 0 {
		// Udp lại giá sub bag mới nhất nếu có sub bag
		// Udp cho case sửa main bag pay
		for _, v := range booking.SubBags {
			subBookR := model_booking.Booking{}
			subBookR.Uid = v.BookingUid
			subBook, errFSub := subBookR.FindFirstByUId(db)
			if errFSub == nil {
				// TODO: optimal và check xử lý udp cho subbag fail
				subBook.UpdatePriceDetailCurrentBag(db)
				subBook.UpdateMushPay(db)
				errUdpSubBag := subBook.Update(db)
				if errUdpSubBag != nil {
					log.Println("updatePriceWithServiceItem errUdpSubBag", errUdpSubBag.Error())
				} else {
					go handlePayment(db, subBook)
				}
			} else {
				log.Println("updatePriceWithServiceItem errFSub", errFSub.Error())
			}
		}

		booking.UpdatePriceDetailCurrentBag(db)
		booking.UpdateMushPay(db)
		errFMB := booking.Update(db)

		if errFMB == nil {
			go handlePayment(db, *booking)
		}
	}
}

// Udp Revenue
func updatePriceForRevenue(item model_booking.Booking, billNo string) {
	db := datasources.GetDatabaseWithPartner(item.PartnerUid)
	mushPay := model_booking.BookingMushPay{}

	listRoundGolfFee := []models.Round{}
	hole := 0
	fbFee := int64(0)
	rentalFee := int64(0)
	buggyFee := int64(0)
	bookingCaddieFee := int64(0)
	practiceBallFee := int64(0)
	proshopFee := int64(0)
	otherFee := int64(0)
	restaurantFee := int64(0)
	minibarFee := int64(0)
	kioskFee := int64(0)
	phiPhat := int64(0)

	roundToFindList := models.Round{BillCode: item.BillCode}
	listRoundOfCurrentBag, _ := roundToFindList.FindAll(db)

	// golfFeeByAgency := int64(0)
	caddieAgenPaid := int64(0)
	greenAgenPaid := int64(0)

	for _, round := range listRoundOfCurrentBag {
		listRoundGolfFee = append(listRoundGolfFee, round)
	}

	hole = slices.Reduce(listRoundGolfFee, func(prev int, item models.Round) int {
		return prev + item.Hole
	})

	caddieFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.CaddieFee
	})

	bookingBuggyFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.BuggyFee
	})

	bookingGreenFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.GreenFee
	})

	caddieFee += caddieAgenPaid
	bookingGreenFee += greenAgenPaid

	totalGolfFee := caddieFee + bookingBuggyFee + bookingGreenFee
	mushPay.TotalGolfFee = totalGolfFee

	// SubBag

	// Sub Service Item của current Bag
	// Get item for current Bag
	// update lại lấy service items mới
	totalServiceItems := int64(0)
	item.FindServiceItemsOfBag(db)
	for _, v := range item.ListServiceItems {
		totalServiceItems += v.Amount
		checkBuggy := strings.Contains(v.Unit, "xe")

		if v.BillCode == item.BillCode {
			if v.Type == constants.MINI_B_SETTING {
				minibarFee += v.Amount
			}
			if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT || v.Type == constants.MINI_R_SETTING {
				restaurantFee += v.Amount
			}
			if v.Type == constants.KIOSK_SETTING {
				kioskFee += v.Amount
			}
			if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT || v.Type == constants.MINI_B_SETTING || v.Type == constants.MINI_R_SETTING {
				fbFee += v.Amount
			} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RENTAL || v.Type == constants.DRIVING_SETTING {
				if v.ItemCode == "R1_3" {
					practiceBallFee += v.Amount
				} else {
					if v.ServiceType != constants.BUGGY_SETTING && v.ServiceType != constants.CADDIE_SETTING && !checkBuggy {
						if v.ItemCode != constants.THAM_QUAN {
							rentalFee += v.Amount
						}
					}
				}
			} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP {
				if v.ItemCode == "FS-8" {
					phiPhat += v.Amount
				} else {
					proshopFee += v.Amount
				}
			} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
				otherFee += v.Amount
			}

			if checkBuggy || v.ItemCode == constants.THAM_QUAN {
				buggyFee += v.Amount
			}

			if v.ServiceType == constants.CADDIE_SETTING {
				bookingCaddieFee += v.Amount
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
	transferList := []model_payment.SinglePaymentItem{}

	for _, item := range list {
		if item.PaymentType == constants.PAYMENT_TYPE_CASH {
			cashList = append(cashList, item)
		} else if item.PaymentType == constants.PAYMENT_STATUS_DEBIT {
			debtList = append(debtList, item)
		} else if item.PaymentType == constants.PAYMENT_TYPE_CARDS {
			cardList = append(cardList, item)
		} else if item.PaymentType == constants.PAYMENT_TYPE_TRANSFER {
			transferList = append(transferList, item)
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

	transferTotal := slices.Reduce(transferList, func(prev int64, item model_payment.SinglePaymentItem) int64 {
		return prev + item.Paid
	})

	bookingR := model_booking.Booking{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		BookingDate: item.BookingDate,
		Bag:         item.Bag,
	}

	bookingR.FindFirst(db)

	agencyInfo := model_report.BookingAgency{}
	if item.AgencyId > 0 {
		agencyInfo = model_report.BookingAgency{
			AgencyId:    item.AgencyInfo.AgencyId,
			ShortName:   item.AgencyInfo.ShortName,
			Category:    item.AgencyInfo.Category,
			Name:        item.AgencyInfo.Name,
			BookingCode: item.NoteOfBooking,
		}
	}

	m := model_report.ReportRevenueDetail{
		PartnerUid:       item.PartnerUid,
		CourseUid:        item.CourseUid,
		BillNo:           billNo,
		Bag:              item.Bag,
		GuestStyle:       item.GuestStyle,
		GuestStyleName:   item.GuestStyleName,
		BookingDate:      item.BookingDate,
		CustomerId:       item.CustomerUid,
		CustomerName:     item.CustomerName,
		MembershipNo:     item.CardId,
		CustomerType:     item.CustomerType,
		Hole:             hole,
		Paid:             bookingR.GetAgencyPaid(),
		GreenFee:         bookingGreenFee,
		CaddieFee:        caddieFee,
		FBFee:            fbFee,
		RentalFee:        rentalFee,
		BuggyFee:         buggyFee,
		BookingCaddieFee: bookingCaddieFee,
		ProshopFee:       proshopFee,
		PraticeBallFee:   practiceBallFee,
		OtherFee:         otherFee,
		MushPay:          bookingR.MushPayInfo.MushPay,
		Total:            totalGolfFee + totalServiceItems,
		Cash:             cashTotal,
		Debit:            debtTotal,
		Card:             cardTotal,
		RestaurantFee:    restaurantFee,
		MinibarFee:       minibarFee,
		KioskFee:         kioskFee,
		PhiPhat:          phiPhat,
		Transfer:         transferTotal,
		AgencyInfo:       agencyInfo,
	}

	m.Create(db)
}
