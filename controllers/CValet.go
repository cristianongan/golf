package controllers

import (
	"errors"
	"fmt"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CValet struct{}

/*
Add Caddie short
Chưa tạo flight
*/
func (_ *CValet) AddBagCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	dataBody := request.ValetAddListBagCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&dataBody); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	for _, body := range dataBody.Data {
		// Check can add
		errB, response := AddCaddieBuggyToBooking(db, body.PartnerUid, body.CourseUid, body.BookingUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode, body.IsPrivateBuggy)

		if errB != nil {
			response_message.InternalServerError(c, errB.Error())
			return
		}

		booking := response.Booking
		caddie := response.NewCaddie
		buggy := response.NewBuggy

		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}

		if body.CaddieCode != "" {
			// Update caddie_current_status
			// caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_LOCK
			if err := caddie.Update(db); err != nil {
				response_message.InternalServerError(c, err.Error())
				return
			}
		}

		if buggy.Code != "" {
			// Update caddie_current_status
			buggy.BuggyStatus = constants.BUGGY_CURRENT_STATUS_LOCK
			if err := buggy.Update(db); err != nil {
				response_message.InternalServerError(c, err.Error())
				return
			}
		}

		//Update trạng thái của các old caddie
		if response.OldCaddie.Id > 0 {
			udpCaddieOut(db, response.OldCaddie.Id)
		}

		//Update trạng thái của các old buggy
		if response.OldBuggy.Id > 0 {
			bookingR := model_booking.Booking{
				BookingDate: booking.BookingDate,
				BuggyId:     response.OldBuggy.Id,
			}
			if errBuggy := udpOutBuggy(db, &bookingR, false); errBuggy != nil {
				log.Println("AddCaddieBuggyToBooking err book udp ", errBuggy.Error())
			}
		}
	}

	okRes(c)
}
func AddCaddieBuggyToBooking(db *gorm.DB, partnerUid, courseUid, bookingUid, bookingDate, bag, caddieCode, buggyCode string, isPrivateBuggy bool) (error, response.AddCaddieBuggyToBookingRes) {
	// Get booking
	booking := model_booking.Booking{}
	booking.Uid = bookingUid

	errBooking := booking.FindFirst(db)
	if errBooking != nil {
		return errBooking, response.AddCaddieBuggyToBookingRes{}
	}

	//Check nếu booking chưa đc ghép bag
	if booking.Bag == "" {
		bookingValidateBag := model_booking.Booking{
			PartnerUid:  partnerUid,
			CourseUid:   courseUid,
			BookingDate: bookingDate,
			Bag:         bag,
		}

		err := bookingValidateBag.FindFirst(db)
		if err == nil {
			errTitle := fmt.Sprintln(bag, "đã được ghép")
			return errors.New(errTitle), response.AddCaddieBuggyToBookingRes{}
		}

		booking.Bag = bag
	}

	response := response.AddCaddieBuggyToBookingRes{}

	//get old caddie
	if booking.CaddieId > 0 {
		oldCaddie := models.Caddie{}
		oldCaddie.Id = booking.CaddieId
		if errFC := oldCaddie.FindFirst(db); errFC == nil {
			response.OldCaddie = oldCaddie
		}
	}

	//get old buggy
	if booking.BuggyId > 0 {
		oldBuggy := models.Buggy{}
		oldBuggy.Id = booking.BuggyId
		if errFC := oldBuggy.FindFirst(db); errFC == nil {
			response.OldBuggy = oldBuggy
		}
	}

	if !(*booking.ShowCaddieBuggy) {
		booking.ResetCaddieBuggy()
	}

	//Check caddie
	var caddie models.Caddie
	if caddieCode != "" {
		caddie = models.Caddie{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       caddieCode,
		}
		errFC := caddie.FindFirst(db)
		if errFC != nil {
			return errFC, response
		}

		if caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			if booking.CaddieId != caddie.Id {
				errTitle := fmt.Sprintln("Caddie", caddie.Code, "đang bị LOCK")
				return errors.New(errTitle), response
			}
		} else {
			if errCaddie := checkCaddieReady(booking, caddie); errCaddie != nil {
				return errCaddie, response
			}
		}

		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

		if response.OldCaddie.Id == caddie.Id {
			response.OldCaddie = models.Caddie{}
		}
	}

	//Check buggy
	var buggy models.Buggy
	if buggyCode != "" {
		buggy = models.Buggy{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       buggyCode,
		}

		errFB := buggy.FindFirst(db)
		if errFB != nil {
			return errFB, response
		}

		if err := checkBuggyReady(db, buggy, booking, isPrivateBuggy, false); err != nil {
			return err, response
		}

		booking.BuggyId = buggy.Id
		booking.IsPrivateBuggy = setBoolForCursor(isPrivateBuggy)
		booking.BuggyInfo = cloneToBuggyBooking(buggy)

		if response.OldBuggy.Id == buggy.Id {
			response.OldBuggy = models.Buggy{}
		}
	}

	booking.ShowCaddieBuggy = setBoolForCursor(true)
	response.NewCaddie = caddie
	response.NewBuggy = buggy
	response.Booking = booking
	return nil, response
}
