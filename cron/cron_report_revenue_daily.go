package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils"
	"strings"

	"github.com/twharmon/slices"
)

func runReportRevenueDailyJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemReportRevenue(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runReportRevenueDaily()
}

func runReportRevenueDaily() {

	bookingDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	bookings := model_booking.BookingList{
		BookingDate: bookingDate,
	}

	db1 := datasources.GetDatabase()
	db, _, err := bookings.FindAllBookingList(db1)
	db = db.Where("check_in_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type <> 'ROUND'")

	if err != nil {
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	reportR := model_report.ReportRevenueDetail{
		PartnerUid:  "CHI-LINH",
		CourseUid:   "CHI-LINH-01",
		BookingDate: bookingDate,
	}

	reportR.DeleteByBookingDate()

	for _, booking := range list {
		updatePriceForRevenue(booking, "")
	}
}

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
		checkBuggy := strings.Contains(v.Name, "xe")

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
						rentalFee += v.Amount
					}
				}
			} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP {
				proshopFee += v.Amount
			} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
				otherFee += v.Amount
			}

			if checkBuggy {
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

	bookingR := model_booking.Booking{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		BookingDate: item.BookingDate,
		Bag:         item.Bag,
	}

	bookingR.FindFirst(db)

	m := model_report.ReportRevenueDetail{
		PartnerUid:       item.PartnerUid,
		CourseUid:        item.CourseUid,
		BillNo:           billNo,
		Bag:              item.Bag,
		GuestStyle:       item.GuestStyle,
		GuestStyleName:   item.GuestStyleName,
		BookingDate:      item.BookingDate,
		CustomerId:       item.CustomerUid,
		MembershipNo:     item.CardId,
		CustomerType:     item.CustomerType,
		Hole:             hole,
		Paid:             item.GetAgencyPaid(),
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
	}

	m.Create(db)
}
