package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRestaurantSetting struct{}

// func (_ *CRestaurantSetting) CreateRestaurantSetting(c *gin.Context, prof models.CmsUser) {
// 	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
// 	body := request.CreateRestaurantSettingBody{}
// 	if bindErr := c.ShouldBind(&body); bindErr != nil {
// 		badRequest(c, bindErr.Error())
// 		return
// 	}
// 	now := utils.GetTimeNow()

// 	// Batch insert player score
// 	RestaurantSetting := models.RestaurantSetting{}

// 	var listPlayer []models.RestaurantSetting

// 	for _, player := range body.Players {
// 		RestaurantSetting := models.RestaurantSetting{
// 			PartnerUid:  body.PartnerUid,
// 			CourseUid:   body.CourseUid,
// 			BookingDate: body.BookingDate,
// 			FlightId:    body.FlightId,
// 			Bag:         player.Bag,
// 			Course:      body.Course,
// 			Hole:        body.Hole,
// 			HoleIndex:   body.HoleIndex,
// 			Par:         body.Par,
// 			Shots:       player.Shots,
// 			Index:       player.Index,
// 			TimeStart:   body.TimeStart,
// 			TimeEnd:     body.TimeEnd,
// 		}

// 		RestaurantSetting.ModelId.CreatedAt = now.Unix()
// 		RestaurantSetting.ModelId.UpdatedAt = now.Unix()
// 		if RestaurantSetting.ModelId.Status == "" {
// 			RestaurantSetting.ModelId.Status = constants.STATUS_ENABLE
// 		}

// 		if RestaurantSetting.IsDuplicated(db) {
// 			response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
// 			return
// 		}

// 		listPlayer = append(listPlayer, RestaurantSetting)
// 	}

// 	errC := RestaurantSetting.BatchInsert(db, listPlayer)
// 	if errC != nil {
// 		response_message.InternalServerError(c, errC.Error())
// 		return
// 	}

// 	res := map[string]interface{}{
// 		"data": listPlayer,
// 	}

// 	okResponse(c, res)
// }

func (_ *CRestaurantSetting) GetListRestaurantSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListRestaurantSettingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit: form.PageRequest.Limit,
		Page:  form.PageRequest.Page,
		// SortBy:  form.PageRequest.SortBy,
		// SortDir: form.PageRequest.SortDir,
	}

	restaurantSettingR := models.RestaurantSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := restaurantSettingR.FindList(db, page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CRestaurantSetting) UpdateRestaurantSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	RestaurantSettingIdStr := c.Param("id")
	RestaurantSettingId, err := strconv.ParseInt(RestaurantSettingIdStr, 10, 64)
	if err != nil || RestaurantSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	// validate body
	body := request.UpdateRestaurantSettingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	restaurantSetting := models.RestaurantSetting{}
	restaurantSetting.Id = RestaurantSettingId
	errF := restaurantSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// restaurantSetting.Hole = body.Hole
	// restaurantSetting.Par = body.Par
	// restaurantSetting.Shots = body.Shots
	// restaurantSetting.Index = body.Index
	// restaurantSetting.TimeStart = body.TimeStart
	// restaurantSetting.TimeEnd = body.TimeEnd

	errUdp := restaurantSetting.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, restaurantSetting)
}

func (_ *CRestaurantSetting) DeleteRestaurantSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	RestaurantSettingIdStr := c.Param("id")
	RestaurantSettingId, err := strconv.ParseInt(RestaurantSettingIdStr, 10, 64)
	if err != nil || RestaurantSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	restaurantSetting := models.RestaurantSetting{}
	restaurantSetting.Id = RestaurantSettingId
	errF := restaurantSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := restaurantSetting.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
