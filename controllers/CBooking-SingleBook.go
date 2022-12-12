package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/*
Cancel Booking
- check chưa check-in mới cancel dc
*/
func (_ *CBooking) CancelBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CancelBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, "Booking Uid not empty")
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_BOOKING {
		response_message.InternalServerError(c, "This booking did check in")
		return
	}
	// Kiểm tra xem đủ điều kiện cancel booking không
	// cancelBookingSetting := model_booking.CancelBookingSetting{}
	// if err := cancelBookingSetting.ValidateBookingCancel(db, booking); err != nil {
	// 	response_message.InternalServerError(c, err.Error())
	// 	return
	// }

	booking.BagStatus = constants.BAG_STATUS_CANCEL
	booking.CancelNote = body.Note
	booking.CancelBookingTime = time.Now().Unix()
	booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

	errUdp := booking.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
Moving Booking
- check chưa check-in mới moving dc
*/
func (_ *CBooking) MovingBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.MovingBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if len(body.BookUidList) == 0 {
		response_message.BadRequest(c, "Booking invalid empty")
		return
	}

	if len(body.BookUidList) > 4 {
		response_message.BadRequest(c, "The number of Bookings cannot exceed 4")
		return
	}

	for _, BookingUid := range body.BookUidList {
		if BookingUid == "" {
			response_message.BadRequest(c, "Booking Uid not empty")
			return
		}

		booking := model_booking.Booking{}
		booking.Uid = BookingUid
		errF := booking.FindFirst(db)
		if errF != nil {
			response_message.InternalServerError(c, errF.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
			response_message.InternalServerError(c, booking.Uid+" did check in")
			return
		}
		if body.TeeTime != "" {
			booking.TeeTime = body.TeeTime
		}
		if body.TeeType != "" {
			booking.TeeType = body.TeeType
		}
		if body.BookingDate != "" {
			booking.BookingDate = body.BookingDate
		}
		if body.CourseType != "" {
			booking.CourseType = body.CourseType
		}

		//Check duplicated
		isDuplicated, errDupli := booking.IsDuplicated(db, true, false)
		if isDuplicated {
			if errDupli != nil {
				response_message.DuplicateRecord(c, errDupli.Error())
				return
			}
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		if body.Hole != 0 {
			booking.Hole = body.Hole
		}

		errUdp := booking.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}

	okRes(c)
}
func (cBooking *CBooking) CreateBookingTee(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for index := range bodyRequest.BookingList {
		bodyRequest.BookingList[index].BookingCode = bookingCode
		bodyRequest.BookingList[index].BookingTeeTime = true
	}

	listBooking, err := cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
	if err != nil {
		return
	}

	// khi book restaurant enable thì auto tạo 1 book reservation trong restaurant
	if len(bodyRequest.BookingList) > 0 {
		item := bodyRequest.BookingList[0]
		if item.BookingRestaurant.Enable {
			db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
			go addServiceCart(db, len(bodyRequest.BookingList), item.PartnerUid, item.CourseUid, item.CustomerBookingName, item.CustomerBookingPhone, prof.FullName)
		}
	}

	okResponse(c, listBooking)
}

func (cBooking *CBooking) CreateCopyBooking(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	for indexTarget, target := range bodyRequest.BookingList {
		if !bodyRequest.BookingList[indexTarget].BookMark {
			bookingCode := utils.HashCodeUuid(uuid.New().String())
			bodyRequest.BookingList[indexTarget].BookingCode = bookingCode
			bodyRequest.BookingList[indexTarget].BookMark = true

			if target.BookingCode != "" {
				for index, data := range bodyRequest.BookingList {
					if data.BookingCode == target.BookingCode {
						bodyRequest.BookingList[index].BookingCode = bookingCode
						bodyRequest.BookingList[index].BookMark = true
					}
				}
			}
		}
	}
	listBooking, _ := cBooking.CreateBatch(bodyRequest.BookingList, c, prof)
	okResponse(c, listBooking)
}

func (_ *CBooking) CancelAllBooking(c *gin.Context, prof models.CmsUser) {
	db1 := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.CancelAllBookingBody{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.BookingList{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
		TeeTime:     form.TeeTime,
		BookingCode: form.BookingCode,
	}

	db, _, err := bookingR.FindAllBookingList(db1)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Find(&list)

	for _, booking := range list {
		if booking.BagStatus != constants.BAG_STATUS_BOOKING {
			response_message.InternalServerError(c, "Booking:"+booking.BookingDate+" did check in")
			return
		}

		booking.BagStatus = constants.BAG_STATUS_CANCEL
		booking.CancelNote = form.Reason
		booking.CancelBookingTime = time.Now().Unix()
		booking.CmsUserLog = getBookingCmsUserLog(prof.UserName, time.Now().Unix())

		errUdp := booking.Update(db1)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}
	okRes(c)
}

func (cBooking CBooking) CreateBatch(bookingList request.ListCreateBookingBody, c *gin.Context, prof models.CmsUser) ([]model_booking.Booking, error) {
	list := []model_booking.Booking{}
	for _, body := range bookingList {
		booking, errCreate := cBooking.CreateBookingCommon(body, c, prof)
		if errCreate != nil {
			return list, errCreate
		}

		if booking != nil {
			list = append(list, *booking)
		}
	}
	return list, nil
}
