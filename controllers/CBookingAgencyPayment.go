package controllers

import (
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_payment "start/models/payment"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CBookingAgencyPayment struct{}

func (_ *CBookingAgencyPayment) GetDetailBookingAgencyPayment(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetAllBookingAgencyPayment{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingAgencyPaymentRequest := model_payment.BookingAgencyPayment{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingCode: form.BookingCode,
		BookingUid:  form.BookingUid,
		AgencyId:    form.AgencyId,
	}

	list, _ := bookingAgencyPaymentRequest.FindAll(db)

	okResponse(c, list)
}
