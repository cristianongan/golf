package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CKioskInventory struct{}

func (_ CKioskInventory) GetKioskInventory(c *gin.Context, prof models.CmsUser) {
	var form request.GetInOutItems
	if err := c.ShouldBind(&form); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	inputItems := kiosk_inventory.InventoryItem{}
	inputItems.ServiceId = form.ServiceId
	inputItems.PartnerUid = form.PartnerUid
	inputItems.CourseUid = form.CourseUid
	inputItems.Code = form.ItemCode
	list, total, err := inputItems.FindList(page)

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

func (_ CKioskInventory) AddItemToInventory(c *gin.Context, prof models.CmsUser) {
	var data request.AddItemToInventoryBody
	if err := c.ShouldBind(&data); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	list := []kiosk_inventory.InventoryItem{}

	for _, body := range data.ListItem {
		groupCode := ""
		var price = 0.0
		var unit = ""
		var name = ""

		if body.ServiceType == constants.GROUP_PROSHOP {
			proshop := model_service.Proshop{
				ProShopId: body.ItemCode,
			}

			if err := proshop.FindFirst(); err == nil {
				groupCode = proshop.GroupCode
				price = proshop.Price
				unit = proshop.Unit
				name = proshop.Name
			} else {
				response_message.InternalServerError(c, errors.New(body.ItemCode+"không tìm thấy").Error())
				return
			}
		} else if body.ServiceType == constants.GROUP_FB {
			fb := model_service.FoodBeverage{
				FBCode: body.ItemCode,
			}

			if err := fb.FindFirst(); err == nil {
				groupCode = fb.GroupCode
				price = fb.Price
				unit = fb.Unit
				name = fb.Name
			} else {
				response_message.InternalServerError(c, errors.New(body.ItemCode+"không tìm thấy").Error())
				return
			}
		} else if body.ServiceType == constants.GROUP_RENTAL {
			rental := model_service.Rental{
				RentalId: body.ItemCode,
			}

			if err := rental.FindFirst(); err == nil {
				groupCode = rental.GroupCode
				price = rental.Price
				unit = rental.Unit
				name = rental.Name
			} else {
				response_message.InternalServerError(c, errors.New(body.ItemCode+" không tìm thấy ").Error())
				return
			}
		}

		goodsService := model_service.GroupServices{
			GroupCode: groupCode,
		}

		errFindGoodsService := goodsService.FindFirst()
		if errFindGoodsService != nil {
			return
		}

		itemInfo := kiosk_inventory.ItemInfo{
			Price:     price,
			ItemName:  name,
			GroupName: goodsService.GroupName,
			GroupType: goodsService.Type,
			GroupCode: groupCode,
			Unit:      unit,
		}

		item := kiosk_inventory.InventoryItem{
			ServiceId:  data.ServiceId,
			Code:       body.ItemCode,
			PartnerUid: data.PartnerUid,
			CourseUid:  data.CourseUid,
		}

		if err := item.FindFirst(); err != nil {
			item.ItemInfo = itemInfo
			item.Quantity = body.Quantity
			if errCre := item.Create(); errCre != nil {
				return
			}
		} else {
			item.Quantity = item.Quantity + body.Quantity
			if errUpd := item.Update(); errUpd != nil {
				return
			}
		}
		list = append(list, item)
	}

	okResponse(c, list)
}
