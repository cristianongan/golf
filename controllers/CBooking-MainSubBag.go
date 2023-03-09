package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

/*
Add Sub bag to Booking
*/
func (_ *CBooking) AddSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.AddSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
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

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag đã Check Out")
		return
	}

	if booking.SubBags == nil {
		booking.SubBags = utils.ListSubBag{}
	}

	if booking.MainBagPay == nil {
		booking.MainBagPay = initMainBagForPay()
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
			err1 := subBooking.FindFirst(db)
			if err1 == nil {
				//Subbag
				subBag := utils.BookingSubBag{
					BookingUid:  v.BookingUid,
					GolfBag:     subBooking.Bag,
					PlayerName:  subBooking.CustomerName,
					BillCode:    subBooking.BillCode,
					BookingCode: subBooking.BookingCode,
					CmsUser:     subBooking.CmsUser,
					CmsUserLog:  subBooking.CmsUserLog,
				}
				booking.SubBags = append(booking.SubBags, subBag)
			} else {
				log.Println("AddSubBagToBooking err1", err1.Error())
			}
		}
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Tính lại giá
	// Cập nhật Main bag cho subbag
	err := updateMainBagForSubBag(db, booking)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	bookRes := model_booking.Booking{}
	bookRes.Uid = booking.Uid
	errFRes := bookRes.FindFirst(db)
	if errFRes != nil {
		response_message.InternalServerError(c, errFRes.Error())
		return
	}

	// Update payment info
	if bookRes.AgencyId > 0 && bookRes.MemberCardUid == "" {
		// go handleAgencyPayment(db, bookRes)
	} else {
		go handleSinglePayment(db, bookRes)
	}

	//Đánh dấu các round(của sub bag) đã được trả bởi main bag
	go bookMarkRoundPaidByMainBag(booking, db)
	res := getBagDetailFromBooking(db, bookRes)

	okResponse(c, res)
}

/*
Edit Sub bag to Booking
*/
func (_ *CBooking) EditSubBagToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.EditSubBagToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
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
		errFSB := subBooking.FindFirst(db)

		if errFSB != nil {
			log.Println("EditSubBagToBooking errFSB", errF.Error())
		}

		if v.IsOut == true {
			//remove di
			// Remove main bag
			subBooking.MainBags = utils.ListSubBag{}
			errSBUdp := subBooking.Update(db)
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
			listServiceItems := model_booking.ListBookingServiceItems{}
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
				errSBUdp := subBooking.Update(db)
				if errSBUdp != nil {
					log.Println("EditSubBagToBooking errSBUdp", errSBUdp.Error())
				}
			}
		}
	}

	if isUpdPrice {
		booking.UpdateMushPay(db)
	}

	booking.CmsUser = prof.UserName
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	errUdp := booking.Update(db)

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	res := getBagDetailFromBooking(db, booking)
	okResponse(c, res)
}

/*
Danh sách booking cho Add sub bag
*/
func (_ *CBooking) GetListBookingForAddSubBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	if errDate == nil {
		bookingR.BookingDate = dateDisplay
	} else {
		log.Println("GetListBookingForAddSubBag booking date display err ", errDate.Error())
	}

	list, errF := bookingR.FindListForSubBag(db)
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
		if v.Bag != form.Bag && len(v.MainBags) == 0 && len(v.SubBags) == 0 {
			listResponse = append(listResponse, v)
		}
	}

	okResponse(c, listResponse)
}

func (_ *CBooking) GetSubBagDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("uid")
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.PartnerUid != prof.PartnerUid || booking.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
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
		errFind := bookingTemp.FindFirst(db)
		if errFind != nil {
			log.Println("GetListSubBagDetail err", errFind.Error())
		}
		list = append(list, bookingTemp)
	}

	okResponse(c, list)
}

/*
Change To Main Bag
*/
func (cBooking *CBooking) ChangeToMainBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Body request
	body := request.BookingBaseBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.Booking{}
	bookingR.Uid = body.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.BadRequestDynamicKey(c, "BOOKING_NOT_FOUND", "")
		return
	}

	list, _ := booking.FindMainBag(db)

	if len(list) == 0 {
		response_message.BadRequestDynamicKey(c, "MAIN_BAG_NOT_FOUND", "")
		return
	}

	mainBag := list[0]
	subBags := utils.ListSubBag{}

	for _, sub := range mainBag.SubBags {
		if sub.GolfBag != booking.Bag {
			subBags = append(subBags, sub)
		}
	}

	listTempGF1 := model_booking.ListBookingGolfFee{}
	for _, v := range mainBag.ListGolfFee {
		if v.BookingUid != booking.Uid {
			listTempGF1 = append(listTempGF1, v)
		}
	}
	mainBag.ListGolfFee = listTempGF1
	mainBag.SubBags = subBags
	mainBag.UpdateMushPay(db)
	if errUpdateMainBag := mainBag.Update(db); errUpdateMainBag != nil {
		response_message.BadRequestDynamicKey(c, "UPDATE_BOOKING_ERROR", "")
		return
	}
	// update lại payment
	go handlePayment(db, mainBag)

	booking.MainBags = utils.ListSubBag{}
	booking.UpdateMushPay(db)
	if errUpdateSubBag := booking.Update(db); errUpdateSubBag != nil {
		response_message.BadRequestDynamicKey(c, "UPDATE_BOOKING_ERROR", "")
		return
	}
	// update lại payment
	go handlePayment(db, booking)

	go func() {
		// Nếu case có Round, MoveFlight
		mainBag.UpdateSubBagForBooking(db)

		cRound := CRound{}
		cRound.ResetRoundPaidByMain(booking.BillCode, db)
	}()

	okResponse(c, booking)
}
