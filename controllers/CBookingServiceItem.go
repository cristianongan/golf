package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CBookingServiceItem struct{}

func (_ *CBookingServiceItem) GetBookingServiceItemList(c *gin.Context, prof models.CmsUser) {
	query := request.GetBookingServiceItem{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	bookingServiceItem := model_booking.BookingServiceItem{
		GroupCode: query.GroupCode,
	}

	list, total, err := bookingServiceItem.FindList(page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}
