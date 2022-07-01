package controllers

import (
	"github.com/gin-gonic/gin"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
)

type CBuggyList struct{}

func (_ *CBuggyList) GetBuggyList(c *gin.Context, prof models.CmsUser) {
	query := request.GetBuggyList{}
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

	bookings := model_booking.BookingList{}

	bookings.FromDate = query.FromDate
	bookings.ToDate = query.ToDate
	bookings.BuggyCode = query.BuggyCode

	// add course_uid
	bookings.CourseUid = prof.CourseUid

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.BuggyListResponse
	db.Find(&list)

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
