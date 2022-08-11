package controllers

import (
	"start/models"

	"github.com/gin-gonic/gin"
)

type CCaddieFee struct{}

func (_ *CCaddieFee) CronCreateCaddieFee(c *gin.Context, prof models.CmsUser) {

	// partnerRequest := models.Partner{}
	// partnerRequest.Uid = "CHI-LINH"
	// partnerErrFind := partnerRequest.FindFirst()
	// if partnerErrFind != nil {
	// 	response_message.BadRequest(c, partnerErrFind.Error())
	// 	return
	// }

	// courseRequest := models.Course{}
	// courseRequest.Uid = "CHI-LINH-01"
	// errFind := courseRequest.FindFirst()
	// if errFind != nil {
	// 	response_message.BadRequest(c, errFind.Error())
	// 	return
	// }

	// bookingRequest := model_booking.Booking{}
	// bookingRequest.CourseUid = "CHI-LINH"
	// bookingRequest.PartnerUid = "CHI-LINH-01"
	// bookingRequest.BookingDate = ""
	// listBooking, errExist := bookingRequest.FindAll(time.Now().Format("02/06/2006"))

	// if errExist == nil {
	// 	response_message.BadRequest(c, "Caddie Id existed in course")
	// 	return
	// }

	// for _, b := range listBooking {
	// 	// b.FlightId = flight.Id
	// 	// errUdp := b.Update()
	// 	// if errUdp != nil {
	// 	// 	log.Println("CreateFlight err flight ", errUdp.Error())
	// 	// }
	// }

	// Caddie.GroupId, _ = strconv.ParseInt(body.Group, 10, 8)

	// err := Caddie.Create()
	// if err != nil {
	// 	response_message.InternalServerError(c, err.Error())
	// 	return
	// }
	// c.JSON(200, Caddie)
}
