package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
)

type CKioskInventory struct{}

func (_ CKioskInventory) InputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	inputItem := kiosk_inventory.InventoryInputItem{}
	inputItem.Code = body.Code
	inputItem.PartnerUid = body.PartnerUid
	inputItem.CourseUid = body.CourseUid
	inputItem.Quantity = body.Quantity
	inputItem.KioskCode = body.KioskCode
	inputItem.ItemCode = body.ItemCode
	inputItem.Source = body.Source
	inputItem.Note = body.Note
	inputItem.ReviewUserUid = body.ReviewUserUid
	inputItem.InputDate = datatypes.Date(time.Now())

	if err := inputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	inventoryStatus := kiosk_inventory.InventoryBill{}
	inventoryStatus.Code = body.Code
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_PENDING
		inventoryStatus.Create()
	}

	okRes(c)
}

func (_ CKioskInventory) OutputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryOutputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	inputItem := kiosk_inventory.InventoryOutputItem{}
	inputItem.Code = body.Code
	inputItem.PartnerUid = body.PartnerUid
	inputItem.CourseUid = body.CourseUid
	inputItem.Quantity = body.Quantity
	inputItem.ItemCode = body.ItemCode

	if err := inputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	inventoryStatus := kiosk_inventory.InventoryBill{}
	inventoryStatus.Code = body.Code
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_SELL
		inventoryStatus.CourseUid = body.Code
		inventoryStatus.Create()
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

	// item.Name = body.Name
	item.Quantity = 0

	if err := item.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskInventory) CreateInputBill(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInsertBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.InventoryBill{}
	inventoryStatus.Code = body.Code
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}

	if inventoryStatus.BillStatus != constants.KIOSK_BILL_INVENTORY_PENDING {
		response_message.BadRequest(c, body.Code+" đã "+inventoryStatus.BillStatus)
		return
	}
	// Thêm ds item vào Inventory
	addItemToInventory(body.Code, body.CourseUid, body.PartnerUid)

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_ACCEPT
	if err := inventoryStatus.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// TODO Giảm ds item trong Inventory

	okRes(c)
}

func addItemToInventory(code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryInputItem{}
	item.Code = code
	list, _, _ := item.FindList()

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			KioskCode:  data.KioskCode,
			Code:       data.ItemCode,
			InputCode:  data.Code,
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
		}

		if err := item.FindFirst(); err != nil {
			item.Quantity = data.Quantity
			if errCre := item.Create(); errCre != nil {
				return errCre
			}
		} else {
			item.Quantity = item.Quantity + data.Quantity
			if errUpd := item.Update(); errUpd != nil {
				return errUpd
			}
		}
	}
	return nil
}

func removeItemFromInventory(code string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryOutputItem{}
	item.Code = code
	list, _, _ := item.FindList()

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			Code:      data.ItemCode,
			InputCode: data.Code,
		}

		if err := item.FindFirst(); err != nil {
			item.Quantity = item.Quantity - data.Quantity
			if errUpd := item.Update(); errUpd != nil {
				return errUpd
			}
		}
	}
	return nil
}

func (_ CKioskInventory) ReturnInputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInsertBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.InventoryBill{}
	inventoryStatus.Code = body.Code
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}
	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_RETURN
	if err := inventoryStatus.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}
	okRes(c)
}
