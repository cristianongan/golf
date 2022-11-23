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
	"time"

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

func getInitListGolfFeeForAddRound(booking *model_booking.Booking, golfFee models.GolfFee, hole int) {
	bookingGolfFee := booking.ListGolfFee[0]

	bookingGolfFee.BookingUid = booking.Uid
	bookingGolfFee.CaddieFee += utils.GetFeeFromListFee(golfFee.CaddieFee, hole)
	bookingGolfFee.BuggyFee += utils.GetFeeFromListFee(golfFee.BuggyFee, hole)
	bookingGolfFee.GreenFee += utils.GetFeeFromListFee(golfFee.GreenFee, hole)

	booking.ListGolfFee[0] = bookingGolfFee
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

func getInitListGolfFeeWithOutGuestStyleForAddRound(booking *model_booking.Booking, rate string, caddieFee, buggyFee, greenFee int64, hole int) {
	bookingGolfFee := booking.ListGolfFee[0]

	bookingGolfFee.BookingUid = booking.Uid
	bookingGolfFee.CaddieFee += utils.CalculateFeeByHole(hole, caddieFee, rate)
	bookingGolfFee.BuggyFee += utils.CalculateFeeByHole(hole, buggyFee, rate)
	bookingGolfFee.GreenFee += utils.CalculateFeeByHole(hole, greenFee, rate)

	booking.ListGolfFee[0] = bookingGolfFee
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
				AgencyId: agency.Id,
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

func updateGolfFeeInBooking(booking *model_booking.Booking, db *gorm.DB) {
	roundToFindList := models.Round{BillCode: booking.BillCode}
	listRound, _ := roundToFindList.FindAll(db)

	bookingCaddieFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.CaddieFee
	})

	bookingBuggyFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.BuggyFee
	})

	bookingGreenFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.GreenFee
	})

	bookingGolfFee := booking.ListGolfFee[0]
	bookingGolfFee.BookingUid = booking.Uid
	bookingGolfFee.CaddieFee = bookingCaddieFee
	bookingGolfFee.BuggyFee = bookingBuggyFee
	bookingGolfFee.GreenFee = bookingGreenFee
	booking.ListGolfFee[0] = bookingGolfFee
	booking.UpdatePriceDetailCurrentBag(db)
	booking.UpdateMushPay(db)
	booking.Update(db)

	if len(booking.MainBags) > 0 {
		// Get data main bag
		bookingMain := model_booking.Booking{
			CourseUid:   booking.CourseUid,
			PartnerUid:  booking.PartnerUid,
			Bag:         booking.MainBags[0].GolfBag,
			BookingDate: booking.BookingDate,
		}
		if err := bookingMain.FindFirst(db); err != nil {
			return
		}

		round1 := models.Round{}
		round2 := models.Round{}

		for _, round := range listRound {
			if round.Index == 1 {
				round1 = round
			}
			if round.Index == 2 {
				round2 = round
			}
		}

		updateGolfFeeOfMainBag := func(buggyFee, caddieFee, greenFee int64) {
			for i, v2 := range bookingMain.ListGolfFee {
				if v2.Bag == booking.Bag {
					bookingMain.ListGolfFee[i].BookingUid = booking.Uid
					bookingMain.ListGolfFee[i].BuggyFee = buggyFee
					bookingMain.ListGolfFee[i].CaddieFee = caddieFee
					bookingMain.ListGolfFee[i].GreenFee = greenFee

					break
				}
			}
			for i, v2 := range bookingMain.SubBags {
				if v2.GolfBag == booking.Bag {
					bookingMain.SubBags[i].BookingUid = booking.Uid

					break
				}
			}
			// Update mush pay, current bag
			var totalGolfFeeOfBookingMain int64 = 0

			for _, v3 := range bookingMain.ListGolfFee {
				totalGolfFeeOfBookingMain += v3.BuggyFee + v3.CaddieFee + v3.GreenFee
			}

			bookingMain.MushPayInfo.TotalGolfFee = totalGolfFeeOfBookingMain
			bookingMain.MushPayInfo.MushPay = bookingMain.MushPayInfo.TotalServiceItem + totalGolfFeeOfBookingMain

			errUpdateBooking := bookingMain.Update(db)

			if errUpdateBooking != nil {
				log.Println("UpdateGolfFeeInBooking Error")
			}
		}

		checkIsFirstRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
		checkIsNextRound := utils.ContainString(bookingMain.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)
		totalGolfFeeOfSubBag := bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
		golfFeeMustPayOfSubbag := totalGolfFeeOfSubBag
		if checkIsFirstRound > -1 && checkIsNextRound > -1 {
			buggyFee := round1.BuggyFee + round2.BuggyFee
			caddieFee := round1.CaddieFee + round2.CaddieFee
			greenFee := round1.GreenFee + round2.GreenFee
			updateGolfFeeOfMainBag(buggyFee, caddieFee, greenFee)

			//update lại giá của booking(sub bag)
			golfFeeMustPayOfSubbag = totalGolfFeeOfSubBag - buggyFee - caddieFee - greenFee
		} else if checkIsFirstRound > -1 {
			updateGolfFeeOfMainBag(round1.BuggyFee, round1.CaddieFee, round1.GreenFee)

			//update lại giá của booking(sub bag)
			golfFeeMustPayOfSubbag = totalGolfFeeOfSubBag - round1.BuggyFee - round1.CaddieFee - round1.GreenFee
		} else if checkIsNextRound > -1 {
			updateGolfFeeOfMainBag(round2.BuggyFee, round2.CaddieFee, round2.GreenFee)

			//update lại giá của booking(sub bag)
			golfFeeMustPayOfSubbag = totalGolfFeeOfSubBag - round2.BuggyFee - round2.CaddieFee - round2.GreenFee
		}

		booking.MushPayInfo.TotalGolfFee = golfFeeMustPayOfSubbag
		booking.MushPayInfo.MushPay = booking.MushPayInfo.TotalServiceItem + golfFeeMustPayOfSubbag
		booking.Update(db)
	}
}

/*
	Booking Init and Update

init price
init Golf Fee
init MushPay
init Rounds
*/
func initPriceForBooking(db *gorm.DB, booking *model_booking.Booking, listBookingGolfFee model_booking.ListBookingGolfFee, bookingGolfFee model_booking.BookingGolfFee, checkInTime int64) {
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
func updatePriceWithServiceItem(booking model_booking.Booking, prof models.CmsUser) {

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	if booking.MainBags != nil && len(booking.MainBags) > 0 {
		// Nếu bag có Main Bag
		booking.UpdatePriceForBagHaveMainBags(db)
	} else {
		if booking.SubBags != nil && len(booking.SubBags) > 0 {
			// Udp orther data
			booking.Update(db)
			// Udp lại giá sub bag mới nhất nếu có sub bag
			// Udp cho case sửa main bag pay
			for _, v := range booking.SubBags {
				subBookR := model_booking.Booking{}
				subBookR.Uid = v.BookingUid
				subBook, errFSub := subBookR.FindFirstByUId(db)
				if errFSub == nil {
					// TODO: optimal và check xử lý udp cho subbag fail
					subBook.UpdatePriceForBagHaveMainBags(db)
					errUdpSubBag := subBook.Update(db)
					if errUdpSubBag != nil {
						log.Println("updatePriceWithServiceItem errUdpSubBag", errUdpSubBag.Error())
					} else {
						handlePayment(db, subBook)
					}
				} else {
					log.Println("updatePriceWithServiceItem errFSub", errFSub.Error())
				}
			}
			// Co sub bag thì main bag dc udp ở trên rồi
			// find main bag udp lại payment
			mainBookUdp := model_booking.Booking{}
			mainBookUdp.Uid = booking.Uid
			mainBookUdp.PartnerUid = booking.PartnerUid
			errFMB := mainBookUdp.FindFirst(db)
			if errFMB == nil {
				handlePayment(db, mainBookUdp)
			}

			return
		}
		booking.UpdateMushPay(db)
		booking.UpdatePriceDetailCurrentBag(db)
	}
	errUdp := booking.Update(db)
	if errUdp != nil {
		log.Println("updatePriceWithServiceItem errUdp", errUdp.Error())
	} else {
		handlePayment(db, booking)
	}
}
