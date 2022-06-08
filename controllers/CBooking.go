package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CBooking struct{}

/// --------- Booking ----------
/*
 Tạo Booking từ TeeSheet
*/
func (_ *CBooking) CreateBooking(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		TeeType:    body.TeeType,
		TeePath:    body.TeePath,
		TeeTime:    body.TeeTime,
		TeeOffTime: body.TeeTime,
		TurnTime:   body.TurnTime,
		RowIndex:   body.RowIndex,
		CmsUser:    prof.UserName,
		Hole:       body.Hole,
	}

	// TODO: check kho tea time trong ngày đó còn trống mới cho đặt

	if body.Bag != "" {
		booking.Bag = body.Bag
	}

	if body.BookingDate != "" {
		booking.BookingDate = body.BookingDate
	} else {
		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
		if errDate == nil {
			booking.BookingDate = dateDisplay
		} else {
			log.Println("booking date display err ", errDate.Error())
		}
	}

	//Check duplicated
	isDuplicated, errDupli := booking.IsDuplicated(true, true)
	if isDuplicated {
		if errDupli != nil {
			response_message.DuplicateRecord(c, errDupli.Error())
			return
		}
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	// Member Card
	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner()
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return
		}

		booking.MemberCardUid = body.MemberCardUid
		booking.CardId = memberCard.CardId
		booking.CustomerName = owner.Name
		booking.CustomerUid = owner.Uid
		booking.CustomerInfo = convertToCustomerSqlIntoBooking(owner)
		if memberCard.PriceCode == 1 {
			// TODO: Giá riêng không theo guest style

		} else {
			// Lấy theo GuestStyle
			body.GuestStyle = memberCard.GetGuestStyle()
		}
	} else {
		booking.CustomerName = body.CustomerName
	}

	//Agency id
	if body.AgencyId > 0 {
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst()
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, "agency"+errFindAgency.Error())
			return
		}

		agencyBooking := cloneToAgencyBooking(agency)

		booking.AgencyInfo = agencyBooking
		body.GuestStyle = agency.GuestStyle
		//TODO: check giá đặc biệt của agency

	}

	// Có thông tin khách hàng
	/*
		Chọn khách hàng từ agency
	*/
	if body.CustomerUid != "" {
		//check customer
		customer := models.CustomerUser{}
		customer.Uid = body.CustomerUid
		errFindCus := customer.FindFirst()
		if errFindCus != nil || customer.Uid == "" {
			response_message.BadRequest(c, "customer"+errFindCus.Error())
			return
		}

		booking.CustomerName = customer.Name
		booking.CustomerInfo = cloneToCustomerBooking(customer)
		booking.CustomerUid = body.CustomerUid
	}

	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())

	// Checkin Time
	checkInTime := time.Now().Unix()

	// GuestStyle
	if body.GuestStyle != "" {
		//Guest style
		golfFeeModel := models.GolfFee{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			GuestStyle: body.GuestStyle,
		}
		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFindGF := golfFeeModel.GetGuestStyleOnDay()
		if errFindGF != nil {
			response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			return
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		// List Booking GolfFee
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(bUid, body, golfFee)
		booking.ListGolfFee = listBookingGolfFee

		// Current Bag Price Detail
		currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
		currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
		currentBagPriceDetail.UpdateAmount()
		booking.CurrentBagPrice = currentBagPriceDetail

		// MushPayInfo
		mushPayInfo := initBookingMushPayInfo(booking)
		booking.MushPayInfo = mushPayInfo

		// Rounds: Init First
		listRounds := initListRound(booking, bookingGolfFee, checkInTime)
		booking.Rounds = listRounds
	}

	// Check In Out
	if body.IsCheckIn {
		// Tạo booking check in luôn
		booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_IN
		booking.InitType = constants.BOOKING_INIT_TYPE_CHECKIN
		booking.CheckInTime = checkInTime
	} else {
		// Tạo booking
		booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_INIT
		booking.InitType = constants.BOOKING_INIT_TYPE_BOOKING
	}

	errC := booking.Create(bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Get booking Detail With Uid
*/
func (_ *CBooking) GetBookingDetail(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Get booking by Bag
 Get Booking Bag trong ngày
*/
func (_ *CBooking) GetBookingByBag(c *gin.Context, prof models.CmsUser) {
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

	if form.BookingDate != "" {
		booking.BookingDate = form.BookingDate
	} else {
		toDayDate, errD := utils.GetBookingDateFromTimestamp(time.Now().Unix())
		if errD != nil {
			response_message.InternalServerError(c, errD.Error())
			return
		}
		booking.BookingDate = toDayDate
	}

	errF := booking.FindFirst()
	if errF != nil {
		// response_message.InternalServerError(c, errF.Error())
		response_message.InternalServerErrorWithKey(c, errF.Error(), "BAG_NOT_FOUND")
		return
	}

	okResponse(c, booking)
}

/*
 Danh sách booking
*/
func (_ *CBooking) GetListBooking(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	bookingR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
	}
	list, total, err := bookingR.FindList(page, form.From, form.To)
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

/*
 Cập nhật booking
 Thêm Service item
*/
func (_ *CBooking) UpdateBooking(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.Booking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Không udp lại Bag
	// if body.Bag != "" {
	// 	booking.Bag = body.Bag
	// }
	if body.ListServiceItems != nil {
		countItems := len(body.ListServiceItems)
		if countItems > 0 {
			for i := 0; i < countItems; i++ {
				body.ListServiceItems[i].BookingUid = booking.Uid
				body.ListServiceItems[i].PlayerName = booking.CustomerName
				body.ListServiceItems[i].Bag = booking.Bag
			}
		}
	}

	if body.GuestStyle != "" {
		booking.GuestStyle = body.GuestStyle
	}

	//Upd Main Pay for Sub
	if body.MainBagNoPay != nil {
		booking.MainBagNoPay = body.MainBagNoPay
	}

	if body.LockerNo == "" {
		booking.LockerNo = body.LockerNo
	}

	if body.ReportNo == "" {
		booking.ReportNo = body.ReportNo
	}

	//Update service items
	booking.ListServiceItems = body.ListServiceItems

	// Tính lại giá
	booking.UpdatePriceDetailCurrentBag()
	booking.UpdateMushPay()

	// Nếu có MainBag thì udp lại giá cho MainBag
	// Update Lại GolfFee, Service items
	if booking.MainBags != nil && len(booking.MainBags) > 0 {
		// Chỉ có 1 main bags
		errBookingMainBag := booking.UpdateBookingMainBag()
		if errBookingMainBag != nil {
			response_message.BadRequest(c, errBookingMainBag.Error())
			return
		}
	}

	// Booking Note
	if body.NoteOfBag != "" && body.NoteOfBag != booking.NoteOfBag {
		booking.NoteOfBag = body.NoteOfBag
		go createBagsNoteNoteOfBag(booking)
	}

	if body.NoteOfBooking != "" && body.NoteOfBooking != booking.NoteOfBooking {
		booking.NoteOfBooking = body.NoteOfBooking
		go createBagsNoteNoteOfBooking(booking)
	}

	// Udp Log Tracking
	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
  Check in
*/
func (_ *CBooking) CheckIn(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.CheckInBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	booking.Hole = body.Hole

	if body.Bag != "" {
		booking.Bag = body.Bag
		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(false, true)
		if isDuplicated {
			if errDupli != nil {
				response_message.DuplicateRecord(c, errDupli.Error())
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		// Cập nhật lại info Bag
		booking.UpdateBagGolfFee()
	}

	if body.Locker != "" {
		booking.LockerNo = body.Locker
	}

	if body.Hole > 0 {
		booking.Hole = body.Hole
	}

	if body.GuestStyle != "" {
		// Tính giá
		golfFeeModel := models.GolfFee{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			GuestStyle: body.GuestStyle,
		}
		// Lấy phí bởi Guest style với ngày tạo
		golfFee, errFind := golfFeeModel.GetGuestStyleOnDay()
		if errFind != nil {
			response_message.InternalServerError(c, "golf fee err "+errFind.Error())
			return
		}
		booking.GuestStyle = body.GuestStyle
		booking.GuestStyleName = golfFee.GuestStyleName

		// List Booking GolfFee
		bodyCreate := request.CreateBookingBody{
			Hole:         booking.Hole,
			CustomerName: booking.CustomerName,
			Bag:          booking.Bag,
		}
		listBookingGolfFee, bookingGolfFee := getInitListGolfFeeForBooking(booking.Uid, bodyCreate, golfFee)
		booking.ListGolfFee = listBookingGolfFee

		// Current Bag Price Detail
		currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
		currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
		currentBagPriceDetail.UpdateAmount()
		booking.CurrentBagPrice = currentBagPriceDetail

		// MushPayInfo
		mushPayInfo := initBookingMushPayInfo(booking)
		booking.MushPayInfo = mushPayInfo

		// Rounds: Init First
		checkInTime := time.Now().Unix()
		listRounds := initListRound(booking, bookingGolfFee, checkInTime)
		booking.Rounds = listRounds
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())
	booking.CheckInTime = time.Now().Unix()
	booking.CheckInOutStatus = constants.CHECK_IN_OUT_STATUS_IN

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
  Check out: c
*/
func (_ *CBooking) CheckOut(c *gin.Context, prof models.CmsUser) {
}

/*
 Add Sub bag to Booking
*/
func (_ *CBooking) AddSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.SubBags == nil {
		response_message.BadRequest(c, "Subbags invalid nil")
		return
	}

	if len(body.SubBags) == 0 {
		response_message.BadRequest(c, "Subbags invalid empty")
		return
	}

	if booking.SubBags == nil {
		booking.SubBags = utils.ListSubBag{}
	}
	if booking.ListServiceItems == nil {
		booking.ListServiceItems = utils.ListBookingServiceItems{}
	}

	// Check lại SubBag
	// Có thể udp thêm vào hoặc remove đi
	// Check exits
	for _, v := range body.SubBags {
		if checkCheckSubBagDupli(v.BookingUid, booking) {
			// Có rồi không thêm nữa
			log.Println("AddSubBagToBooking dupli book", v.BookingUid)
		} else {
			subBooking := model_booking.Booking{}
			subBooking.Uid = v.BookingUid
			err1 := subBooking.FindFirst()
			if err1 == nil {
				//Subbag
				subBag := utils.BookingSubBag{
					BookingUid: v.BookingUid,
					GolfBag:    subBooking.Bag,
					PlayerName: subBooking.CustomerName,
				}
				booking.SubBags = append(booking.SubBags, subBag)

				//Udp List GolfFee
				subBagGolfFee := subBooking.GetCurrentBagGolfFee()
				if booking.ListGolfFee == nil {
					booking.ListGolfFee = model_booking.ListBookingGolfFee{}
				}
				booking.ListGolfFee = append(booking.ListGolfFee, subBagGolfFee)

				//Udp lại Sub service items
				if subBooking.ListServiceItems != nil {
					booking.ListServiceItems = append(booking.ListServiceItems, subBooking.ListServiceItems...)
				}
			} else {
				log.Println("AddSubBagToBooking err1", err1.Error())
			}
		}
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	// Tính lại giá
	booking.UpdateMushPay()

	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(body, booking.Bag, booking.CustomerName)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Edit Sub bag to Booking
*/
func (_ *CBooking) EditSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.EditSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.SubBags == nil {
		response_message.BadRequest(c, "Subbags invalid nil")
		return
	}

	if len(body.SubBags) == 0 {
		response_message.BadRequest(c, "subbag empty")
		return
	}

	// Khi có xoá note thì Udp lại giá
	isUpdPrice := false

	for i, v := range body.SubBags {
		// Get Booking Detail
		subBooking := model_booking.Booking{}
		subBooking.Uid = v.BookingUid
		errFSB := subBooking.FindFirst()

		if errFSB != nil {
			log.Println("EditSubBagToBooking errFSB", errF.Error())
		}

		if v.IsOut == true {
			//remove di
			// Remove main bag
			subBooking.MainBags = utils.ListSubBag{}
			errSBUdp := subBooking.Update()
			if errSBUdp != nil {
				log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
			}
			isUpdPrice = true
			// Udp sub bag for booking
			// remove sub bag
			subBags := utils.ListSubBag{}
			for _, v1 := range booking.SubBags {
				if v1.BookingUid != v.BookingUid {
					subBags = append(subBags, v1)
				}
			}
			booking.SubBags = subBags
			// remove list golf fee
			listGolfFees := model_booking.ListBookingGolfFee{}
			for _, v1 := range booking.ListGolfFee {
				if v1.BookingUid != v.BookingUid {
					listGolfFees = append(listGolfFees, v1)
				}
			}
			booking.ListGolfFee = listGolfFees
			// remove list service items
			listServiceItems := utils.ListBookingServiceItems{}
			for _, v1 := range booking.ListServiceItems {
				if v1.BookingUid != v.BookingUid {
					listServiceItems = append(listServiceItems, v1)
				}
			}
			booking.ListServiceItems = listServiceItems
		} else {
			isCanUdp := false

			if subBooking.SubBagNote != v.SubBagNote {
				subBooking.SubBagNote = v.SubBagNote
				isCanUdp = true
			}
			if subBooking.CustomerName != v.PlayerName {
				subBooking.CustomerName = v.PlayerName
				if i < len(booking.SubBags) {
					booking.SubBags[i].PlayerName = v.PlayerName
				}
				isCanUdp = true
			}

			if isCanUdp {
				errSBUdp := subBooking.Update()
				if errSBUdp != nil {
					log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
				}
			}
		}
	}

	if isUpdPrice {
		booking.UpdateMushPay()
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Booking Rounds: Vòng Round
 Thêm Round cho Booking
*/
func (_ *CBooking) AddRound(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddRoundBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, errors.New("Uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Handle Round Logic
	// Tính lại giá

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
 Danh sách booking cho Add sub bag
*/
func (_ *CBooking) GetListBookingForAddSubBag(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" || form.CourseUid == "" || form.Bag == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid:       form.PartnerUid,
		CourseUid:        form.CourseUid,
		CheckInOutStatus: constants.CHECK_IN_OUT_STATUS_IN,
	}
	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	if errDate == nil {
		bookingR.BookingDate = dateDisplay
	} else {
		log.Println("GetListBookingForAddSubBag booking date display err ", errDate.Error())
	}

	list, errF := bookingR.FindListForSubBag()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	listResponse := []model_booking.BookingForSubBag{}

	if len(list) == 0 {
		okResponse(c, listResponse)
		return
	}

	for _, v := range list {
		if v.Bag != form.Bag {
			listResponse = append(listResponse, v)
		}
	}

	okResponse(c, listResponse)
}

func (_ *CBooking) GetSubBagDetail(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	list := []model_booking.Booking{}

	if booking.SubBags == nil || len(booking.SubBags) == 0 {
		okResponse(c, list)
		return
	}

	for _, v := range booking.SubBags {
		bookingTemp := model_booking.Booking{}
		bookingTemp.Uid = v.BookingUid
		errFind := bookingTemp.FindFirst()
		if errFind != nil {
			log.Println("GetListSubBagDetail err", errFind.Error())
		}
		list = append(list, bookingTemp)
	}

	okResponse(c, list)
}

/*
	Other Paid
*/
func (_ *CBooking) AddOtherPaid(c *gin.Context, prof models.CmsUser) {
	// Body request
	body := request.AddOtherPaidBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, errors.New("Uid not valid").Error())
		return
	}

	if body.OtherPaids == nil || len(body.OtherPaids) == 0 {
		response_message.BadRequest(c, errors.New("other paid empty").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// list service items
	// Remove cái cũ
	listServiceItems := utils.ListBookingServiceItems{}
	for _, v := range booking.ListServiceItems {
		if v.Type != constants.BOOKING_OTHER_FEE {
			listServiceItems = append(listServiceItems, v)
		}
	}

	// add cái mới
	for _, v := range body.OtherPaids {
		serviceItem := utils.BookingServiceItem{
			Type:       constants.BOOKING_OTHER_FEE,
			Amount:     v.Amount,
			Name:       v.Reason,
			PlayerName: booking.CustomerName,
			Bag:        booking.Bag,
		}
		listServiceItems = append(listServiceItems, serviceItem)
	}

	booking.ListServiceItems = listServiceItems
	booking.UpdateMushPay()

	booking.OtherPaids = body.OtherPaids

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}
