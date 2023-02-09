package controllers

import (
	"errors"
	"fmt"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CTest struct{}

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
	// 	toDayDate, errD := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
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

	// go handlePayment(db, booking)
	// go handleAgencyPaid(booking, request.AgencyFeeInfo{
	// 	GolfFee:  booking.AgencyPaid[0].Fee,
	// 	BuggyFee: booking.AgencyPaid[1].Fee,
	// })

	// notiData := map[string]interface{}{
	// 	"type":  constants.NOTIFICATION_CADDIE_WORKING_STATUS_UPDATE,
	// 	"title": "",
	// }

	// newFsConfigBytes, _ := json.Marshal(notiData)
	// // socket.GetHubSocket() = socket.NewHub()
	// socket.GetHubSocket().Broadcast <- newFsConfigBytes

	// m := socket_room.Message{
	// 	Data: newFsConfigBytes,
	// 	Room: "1",
	// }
	// socket_room.Hub.Broadcast <- m
}

func (cBooking *CTest) TestFunc(c *gin.Context, prof models.CmsUser) {
	caddieList := []string{"16", "23", "30", "29", "39", "40", "51", "54", "57", "05", "02"}
	updateCaddieWorkingOnDay(caddieList, "CHI-LINH", "CHI-LINH-01", true)
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
	billNo := fmt.Sprint(utils.GetTimeNow().UnixMilli())
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
