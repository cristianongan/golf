package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CBagAttachCaddie struct{}

func (_ *CBagAttachCaddie) GetListAttachCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListAttachCaddieForm{}
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

	caddieAttach := model_gostarter.BagAttachCaddie{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		Bag:         form.Search,
		BookingDate: form.BookingDate,
	}

	list, total, err := caddieAttach.FindList(db, page)
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

// Create attach caddie
func (_ *CBagAttachCaddie) CreateAttachCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate body
	body := request.CreateBagAttachCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// validate bag
	bookingBag := model_booking.Booking{}

	bookingBag.PartnerUid = body.PartnerUid
	bookingBag.CourseUid = body.CourseUid
	bookingBag.Bag = body.Bag
	bookingBag.BookingDate = body.BookingDate

	isDuplicated, errDupli := bookingBag.IsDuplicated(db, false, true)
	if isDuplicated {
		if errDupli != nil {
			response_message.InternalServerErrorWithKey(c, errDupli.Error(), "DUPLICATE_BAG")
			return
		}
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	// validate caddie
	if body.CaddieCode != "" {
		caddie := models.Caddie{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			Code:       body.CaddieCode,
		}
		errFC := caddie.FindFirst(db)
		if errFC != nil {
			response_message.BadRequestFreeMessage(c, "Caddie not found")
			return
		}

		if caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			response_message.BadRequestFreeMessage(c, "Caddie"+caddie.Code+"đang bị LOCK")
			return
		} else {
			if errCaddie := checkCaddieReadyForApp(caddie); errCaddie != "" {
				response_message.BadRequestFreeMessage(c, errCaddie)
				return
			}
		}

		// Update caddie_current_status
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
		if err := caddie.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// Create attach caddie
	caddieAttach := model_gostarter.BagAttachCaddie{}

	caddieAttach.PartnerUid = body.PartnerUid
	caddieAttach.CourseUid = body.CourseUid
	caddieAttach.BookingDate = body.BookingDate
	caddieAttach.Bag = body.Bag
	caddieAttach.CaddieCode = body.CaddieCode
	caddieAttach.LockerNo = body.LockerNo

	// validate booking
	if body.BookingUid != "" {
		booking := model_booking.Booking{}

		booking.Uid = body.BookingUid

		if err := booking.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Booking "+err.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
			response_message.BadRequestFreeMessage(c, "Bag status invalid")
			return
		}

		caddieAttach.BookingUid = body.BookingUid
		caddieAttach.CustomerName = body.CustomerName
		caddieAttach.BagStatus = constants.BAG_ATTACH_CADDIE_BOOKING

		go UpdateNewBooking(db, &booking, caddieAttach)
	} else {
		caddieAttach.BagStatus = constants.BAG_ATTACH_CADDIE_READY
	}

	errC := caddieAttach.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, caddieAttach)
}

// Update attach caddie
func (_ *CBagAttachCaddie) UpdateAttachCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bagACIdStr := c.Param("id")
	bagACId, err := strconv.ParseInt(bagACIdStr, 10, 64)
	if err != nil || bagACId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	// validate body
	body := request.UpdateBagAttachCaddieBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Find attach caddie
	caddieAttach := model_gostarter.BagAttachCaddie{}
	caddieAttach.Id = bagACId
	errF := caddieAttach.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// validate bag
	if body.Bag != caddieAttach.Bag {
		bookingBag := model_booking.Booking{}

		bookingBag.PartnerUid = prof.PartnerUid
		bookingBag.CourseUid = prof.CourseUid
		bookingBag.Bag = body.Bag
		bookingBag.BookingDate = body.BookingDate

		isDuplicated, errDupli := bookingBag.IsDuplicated(db, false, true)
		if isDuplicated {
			if errDupli != nil {
				response_message.InternalServerErrorWithKey(c, errDupli.Error(), "DUPLICATE_BAG")
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
	}

	// Updtae
	caddieAttach.Bag = body.Bag
	caddieAttach.BookingDate = body.BookingDate
	caddieAttach.CaddieCode = body.CaddieCode
	caddieAttach.LockerNo = body.LockerNo

	// validate caddie
	if body.CaddieCode != "" && body.CaddieCode != caddieAttach.CaddieCode {
		caddie := models.Caddie{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			Code:       body.CaddieCode,
		}
		errFC := caddie.FindFirst(db)
		if errFC != nil {
			response_message.BadRequestFreeMessage(c, "Caddie not found")
			return
		}

		if caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			response_message.BadRequestFreeMessage(c, "Caddie"+caddie.Code+"đang bị LOCK")
			return
		} else {
			if errCaddie := checkCaddieReadyForApp(caddie); errCaddie != "" {
				response_message.BadRequestFreeMessage(c, errCaddie)
				return
			}
		}

		// Update caddie_current_status
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
		if err := caddie.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		//Update caddie old
		caddieOld := models.Caddie{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			Code:       caddieAttach.CaddieCode,
		}

		if caddieOld.CurrentRound == 0 {
			caddieOld.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		} else if caddieOld.CurrentRound == 1 {
			caddieOld.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH
		} else if caddieOld.CurrentRound == 2 {
			caddieOld.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH_R2
		} else if caddieOld.CurrentRound == 3 {
			caddieOld.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH_R3
		}

		errFC = caddieOld.FindFirst(db)
		if errFC != nil {
			response_message.BadRequestFreeMessage(c, "Caddie not found")
			return
		}
	}

	// validate booking
	if body.BookingUid != "" {
		booking := model_booking.Booking{}

		booking.Uid = body.BookingUid

		if err := booking.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Booking "+err.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
			response_message.BadRequestFreeMessage(c, "Bag status invalid")
			return
		}

		if caddieAttach.BookingUid != "" && caddieAttach.BookingUid != body.BookingUid {
			go UpdateOldBooking(db, caddieAttach)
		}

		go UpdateNewBooking(db, &booking, caddieAttach)

		caddieAttach.BookingUid = body.BookingUid
		caddieAttach.CustomerName = body.CustomerName
		caddieAttach.BagStatus = constants.BAG_ATTACH_CADDIE_BOOKING
	} else {
		if caddieAttach.BookingUid != "" {
			go UpdateOldBooking(db, caddieAttach)
		}
		caddieAttach.BookingUid = ""
		caddieAttach.CustomerName = ""
		caddieAttach.BagStatus = constants.BAG_ATTACH_CADDIE_READY
	}

	errC := caddieAttach.Update(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, caddieAttach)
}

func (_ *CBagAttachCaddie) DeleteAttachCaddie(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bagACIdStr := c.Param("id")
	bagACId, err := strconv.ParseInt(bagACIdStr, 10, 64)
	if err != nil || bagACId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	// Find attach caddie
	caddieAttach := model_gostarter.BagAttachCaddie{}
	caddieAttach.Id = bagACId
	errF := caddieAttach.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := caddieAttach.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// Update thông tin booking
func UpdateNewBooking(db *gorm.DB, booking *model_booking.Booking, caddieAtt model_gostarter.BagAttachCaddie) {
	if caddieAtt.CaddieCode != "" {
		caddie := models.Caddie{
			PartnerUid: booking.PartnerUid,
			CourseUid:  booking.CourseUid,
			Code:       caddieAtt.CaddieCode,
		}
		errFC := caddie.FindFirst(db)
		if errFC != nil {
			log.Println("UpdateBooking err: ", errFC)
		}

		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
	}

	if caddieAtt.Bag != "" {
		booking.Bag = caddieAtt.Bag
	}

	if caddieAtt.LockerNo != "" {
		booking.LockerNo = caddieAtt.LockerNo
		// booking.LockerStatus = constants.LOCKER_STATUS_RETURNED
	} else {
		booking.LockerNo = ""
		// booking.LockerStatus = ""
	}

	if err := booking.Update(db); err != nil {
		log.Println("UpdateBooking err: ", err)
	}

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, booking)
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, booking)
}

func UpdateOldBooking(db *gorm.DB, caddieAtt model_gostarter.BagAttachCaddie) {
	booking := model_booking.Booking{}

	booking.Uid = caddieAtt.BookingUid
	if err := booking.FindFirst(db); err != nil {
		log.Println("FindFrist err: ", err)
	}

	// Update infor
	booking.CaddieId = 0
	booking.CaddieInfo = model_booking.BookingCaddie{}
	booking.Bag = ""
	booking.LockerNo = ""
	// booking.LockerStatus = ""

	if err := booking.Update(db); err != nil {
		log.Println("UpdateBooking err: ", err)
	}

	cNotification := CNotification{}
	go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_UPD, &booking)
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, booking)
}
