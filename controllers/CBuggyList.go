package controllers

import (
	"fmt"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBuggyList struct{}

func (_ *CBuggyList) GetBuggyList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	bookings.CaddieCode = query.CaddieCode
	bookings.GolfBag = query.GolfBag
	bookings.IsTimeOut = query.IsTimeOut

	bookings.IsFlight = strconv.FormatInt(1, 10)
	bookings.HasBuggy = strconv.FormatInt(1, 10)

	// add course_uid
	bookings.CourseUid = prof.CourseUid

	db, total, err := bookings.FindBookingListWithSelect(db, page)

	var list []response.BuggyListResponse
	db.Find(&list)

	result := make(map[int64]map[string][]response.BuggyListResponse)

	for _, booking := range list {
		if result[booking.FlightId] == nil {
			result[booking.FlightId] = make(map[string][]response.BuggyListResponse)
		}
		fmt.Println("[DEBUG]", booking.Uid)
		result[booking.FlightId][booking.BuggyId] = append(result[booking.FlightId][booking.BuggyId], booking)
	}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  result,
	}

	c.JSON(200, res)
}
