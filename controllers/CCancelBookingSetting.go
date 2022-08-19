package controllers

import (
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CCancelBookingSetting struct{}

func (item *CCancelBookingSetting) CreateCancelBookingSetting(c *gin.Context, prof models.CmsUser) {
	var bodyCollection model_booking.ListCancelBookingSetting
	if bindErr := c.BindJSON(&bodyCollection); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	list := model_booking.ListCancelBookingSetting{}
	uniqueNumber := time.Now().Unix()

	for _, body := range bodyCollection {
		if bind1Err := validatePartnerAndCourse(body.PartnerUid, body.CourseUid); bind1Err != nil {
			response_message.BadRequest(c, bind1Err.Error())
			return
		}

		cancelBookingSetting := model_booking.CancelBookingSetting{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			PeopleFrom: body.PeopleFrom,
			PeopleTo:   body.PeopleTo,
			Time:       body.Time,
			Type:       uniqueNumber,
		}

		cancelBookingSetting.Status = body.Status

		err := cancelBookingSetting.Create()
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
		list = append(list, cancelBookingSetting)
	}

	c.JSON(200, list)
}

func (_ *CCancelBookingSetting) DeleteCancelBookingSetting(c *gin.Context, prof models.CmsUser) {
	idRequest := c.Param("id")
	cancelBookingSettingIdIncrement, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	cancelBookingSetting := model_booking.CancelBookingSetting{}
	cancelBookingSetting.Type = cancelBookingSettingIdIncrement
	list, _, errF := cancelBookingSetting.FindList()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	for _, data := range list {
		err := data.Delete()
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return

		}
	}
	okRes(c)
}

func (_ CCancelBookingSetting) GetCancelBookingSetting(c *gin.Context, prof models.CmsUser) {
	var query model_booking.CancelBookingSetting
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cancelBookingSetting := model_booking.CancelBookingSetting{}
	cancelBookingSetting.PartnerUid = query.PartnerUid
	cancelBookingSetting.CourseUid = query.CourseUid

	list, total, err := cancelBookingSetting.FindList()

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
func (_ *CCancelBookingSetting) UpdateCancelBookingSetting(c *gin.Context, prof models.CmsUser) {
	idStr := c.Param("id")
	caddieId, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body model_booking.CancelBookingSetting
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	cancelBookingRequest := model_booking.CancelBookingSetting{}
	cancelBookingRequest.Id = caddieId

	errF := cancelBookingRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.PeopleFrom > 0 {
		cancelBookingRequest.PeopleFrom = body.PeopleFrom
	}
	if body.PeopleTo > 0 {
		cancelBookingRequest.PeopleTo = body.PeopleTo
	}
	if body.Time > 0 {
		cancelBookingRequest.Time = body.Time
	}

	err := cancelBookingRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
