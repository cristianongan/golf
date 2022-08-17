package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	"start/utils/response_message"
	"time"
)

type CKioskInventory struct{}

func (_ CKioskInventory) InputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	item := kiosk_inventory.InventoryItem{}
	item.Code = body.Code

	if err := item.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	inputItem := kiosk_inventory.InventoryInputItem{}
	inputItem.Code = item.Code
	inputItem.Quantity = body.Quantity
	inputItem.InputDate = datatypes.Date(time.Now())

	if err := inputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	item.Quantity = item.Quantity + inputItem.Quantity
	if err := item.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskInventory) OutputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryOutputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	item := kiosk_inventory.InventoryItem{}
	item.Code = body.Code

	if err := item.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	outputItem := kiosk_inventory.InventoryOutputItem{}
	outputItem.Code = item.Code
	outputItem.Quantity = body.Quantity
	outputItem.OutputDate = datatypes.Date(time.Now())

	if err := outputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	item.Quantity = item.Quantity - outputItem.Quantity
	if err := item.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskInventory) CreateItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryCreateItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	item := kiosk_inventory.InventoryItem{}
	item.Code = body.Code

	if err := item.FindFirst(); err == nil {
		response_message.BadRequest(c, errors.New("item is exist").Error())
		return
	}

	item.Name = body.Name
	item.Quantity = 0

	if err := item.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}
