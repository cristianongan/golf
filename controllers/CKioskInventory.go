package controllers

import (
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CKioskInventory struct{}

func (_ CKioskInventory) GetKioskInventory(c *gin.Context, prof models.CmsUser) {
	var form request.GetInOutItems
	if err := c.ShouldBind(&form); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	inputItems := kiosk_inventory.InventoryItem{}
	inputItems.KioskCode = form.KioskCode
	inputItems.PartnerUid = form.PartnerUid
	inputItems.CourseUid = form.CourseUid
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
