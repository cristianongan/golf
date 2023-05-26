package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRestaurantSetting struct{}

func (_ *CRestaurantSetting) CreateRestaurantSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateRestaurantSettingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	now := utils.GetTimeNow()

	// Batch insert player score
	resSettingR := models.RestaurantSetting{}

	var listSetting []models.RestaurantSetting

	for _, id := range body.ServiceIds {
		restaurantSetting := models.RestaurantSetting{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			ServiceId:  id,
			Name:       body.Name,
			Type:       body.Type,
		}

		if body.Status != "" {
			restaurantSetting.ModelId.Status = body.Status
		} else {
			restaurantSetting.ModelId.Status = constants.STATUS_ENABLE
		}

		if body.Type == constants.RESTAURANT_SETTING_TYPE_MINUTE {
			restaurantSetting.Time = body.Time
		} else {
			restaurantSetting.NumberTables = body.NumberTables
			restaurantSetting.PeopleInTable = body.PeopleInTable
			restaurantSetting.Symbol = body.Symbol
			restaurantSetting.TableFrom = body.TableFrom
			restaurantSetting.DataTables = body.DataTables
		}

		restaurantSetting.ModelId.CreatedAt = now.Unix()
		restaurantSetting.ModelId.UpdatedAt = now.Unix()

		listSetting = append(listSetting, restaurantSetting)
	}

	errC := resSettingR.BatchInsert(db, listSetting)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	res := map[string]interface{}{
		"data": listSetting,
	}

	okResponse(c, res)
}

func (_ *CRestaurantSetting) GetListRestaurantSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListRestaurantSettingForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
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

	if body.Status != "" {
		restaurantSetting.ModelId.Status = body.Status
	} else {
		restaurantSetting.ModelId.Status = constants.STATUS_ENABLE
	}

	if restaurantSetting.Type == constants.RESTAURANT_SETTING_TYPE_MINUTE {
		restaurantSetting.Name = body.Name
		restaurantSetting.Time = body.Time
	} else {
		restaurantSetting.Name = body.Name
		restaurantSetting.NumberTables = body.NumberTables
		restaurantSetting.PeopleInTable = body.PeopleInTable
		restaurantSetting.Symbol = body.Symbol
		restaurantSetting.TableFrom = body.TableFrom
		restaurantSetting.DataTables = body.DataTables
	}

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
