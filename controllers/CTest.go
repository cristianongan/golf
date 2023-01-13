package controllers

import (
	"errors"
	"fmt"
	"log"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_payment "start/models/payment"
	model_report "start/models/report"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (cBooking *CTest) TestFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.Bag == "" {
		response_message.BadRequest(c, errors.New("Bag invalid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.Bag = form.Bag

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

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	booking.UpdatePriceDetailCurrentBag(db)
	booking.UpdateMushPay(db)
	booking.Update(db)
	go handlePayment(db, booking)

	// notiData := map[string]interface{}{
	// 	"type":  constants.NOTIFICATION_CADDIE_WORKING_STATUS_UPDATE,
	// 	"title": "",
	// }

	// newFsConfigBytes, _ := json.Marshal(notiData)
	// // socket.HubBroadcastSocket = socket.NewHub()
	// socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes

	// m := socket_room.Message{
	// 	Data: newFsConfigBytes,
	// 	Room: "1",
	// }
	// socket_room.Hub.Broadcast <- m
}

func (cBooking *CTest) TestFunc(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteLockRequest{}
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	course := models.Course{}
	course.Uid = query.CourseUid
	errCourse := course.FindFirst()
	if errCourse != nil {
		log.Println(errCourse)
	}

	form := request.GetListBookingSettingForm{
		CourseUid:  query.CourseUid,
		PartnerUid: query.PartnerUid,
		OnDate:     query.BookingDate,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	bookingDateTime, _ := time.Parse(constants.DATE_FORMAT_1, query.BookingDate)
	weekday := strconv.Itoa(int(bookingDateTime.Weekday() + 1))
	turnTimeH := 2
	// endTime := ""
	bookSetting := model_booking.BookingSetting{}

	for _, data := range listSettingDetail {
		// if strings.ContainsAny(data.Dow, weekday) {
		// 	turnLength = data.TurnLength
		// 	endTime = data.EndPart3
		// 	break
		// }
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	currentTeeTimeDate, _ := utils.ConvertHourToTime(query.TeeTime)
	// endTimeDate, _ := utils.ConvertHourToTime(endTime)

	teeList := []string{}

	if course.Hole == 18 {

		if query.TeeType == "1" {
			teeList = []string{"10"}
		} else {
			teeList = []string{"1"}
		}
	} else if course.Hole == 27 {

		if query.CourseType == "A" {
			teeList = []string{"1B", "1C"}
		} else if query.CourseType == "B" {
			teeList = []string{"1C", "1A"}
		} else if query.CourseType == "C" {
			teeList = []string{"1A", "1B"}
		}

	}

	if len(teeList) == 0 {
		log.Println(errors.New("Không tìm thấy sân"))
	}

	timeParts := []response.TeeTimePartOTA{
		{
			IsHideTeePart: bookSetting.IsHideTeePart1,
			StartPart:     bookSetting.StartPart1,
			EndPart:       bookSetting.EndPart1,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart2,
			StartPart:     bookSetting.StartPart2,
			EndPart:       bookSetting.EndPart2,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart3,
			StartPart:     bookSetting.StartPart3,
			EndPart:       bookSetting.EndPart3,
		},
	}

	index := 0
	teeTimeListLL := []string{}

	for _, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1

				hour := teeTimeInit.Hour()
				minute := teeTimeInit.Minute()

				hourStr_ := strconv.Itoa(hour)
				if hour < 10 {
					hourStr_ = "0" + hourStr_
				}
				minuteStr := strconv.Itoa(minute)
				if minute < 10 {
					minuteStr = "0" + minuteStr
				}

				hourStr := hourStr_ + ":" + minuteStr

				teeTimeListLL = append(teeTimeListLL, hourStr)
				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.Unix() > endTime.Unix() {
					break
				}
			}
		}
	}

	for index, _ := range teeList {

		t := currentTeeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(bookSetting.TurnLength)) * time.Duration(index+1))

		hour := t.Hour()
		minute := t.Minute()

		hourStr_ := strconv.Itoa(hour)
		if hour < 10 {
			hourStr_ = "0" + hourStr_
		}
		minuteStr := strconv.Itoa(minute)
		if minute < 10 {
			minuteStr = "0" + minuteStr
		}

		teeTime1B := hourStr_ + ":" + minuteStr

		if utils.Contains(teeTimeListLL, teeTime1B) {
			log.Println(teeTime1B)
		}
		// lockTeeTime := models.LockTeeTimeWithSlot{
		// 	PartnerUid:     query.PartnerUid,
		// 	CourseUid:      query.CourseUid,
		// 	TeeTime:        teeTime1B,
		// 	TeeTimeStatus:  "LOCKED",
		// 	DateTime:       query.BookingDate,
		// 	CurrentTeeTime: query.TeeTime,
		// 	TeeType:        data,
		// 	Type:           constants.BOOKING_CMS,
		// }

		// lockTeeTimeToRedis(lockTeeTime)
	}
}

func (cBooking *CTest) TestCaddieSlot(c *gin.Context, prof models.CmsUser) {
	query := request.RCaddieSlotExample{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieList := strings.Split(query.Caddie, ",")
	go updateCaddieOutSlot("CHI-LINH", "CHI-LINH-01", caddieList)

	okRes(c)
}

func (cBooking *CTest) TestFastCustomer(c *gin.Context, prof models.CmsUser) {
	uid := utils.HashCodeUuid(uuid.New().String())
	customerBody := request.CustomerBody{
		MaKh:   uid,
		TenKh:  "Duy Tuan",
		DiaChi: "ddddddd",
	}

	_, res := callservices.CreateCustomer(customerBody)

	okResponse(c, res)
}

func (cBooking *CTest) TestFastFee(c *gin.Context, prof models.CmsUser) {
	uid := utils.HashCodeUuid(uuid.New().String())
	billNo := fmt.Sprint(time.Now().UnixMilli())
	customerBody := request.CustomerBody{
		MaKh:   uid,
		TenKh:  "Duy Tuan",
		DiaChi: "ddddddd",
	}

	check, customer := callservices.CreateCustomer(customerBody)
	if check {
		callservices.TransferFast(constants.PAYMENT_TYPE_CASH, 100000, "", uid, customerBody.TenKh, billNo)
	}

	res := map[string]interface{}{
		"customer": customer,
	}
	okResponse(c, res)
}
