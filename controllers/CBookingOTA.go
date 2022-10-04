package controllers

import (
	"start/controllers/request"
	"start/controllers/response"

	"github.com/gin-gonic/gin"
)

type CBookingOTA struct{}

/*
Booking OTA
*/
func (cBooking *CBookingOTA) CreateBookingOTA(c *gin.Context) {
	body := request.CreateBookingOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	res := response.BookingOTARes{
		BookID: "121212",
	}

	okResponse(c, res)
}
