package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type CKioskInputInventory struct{}

func (_ CKioskInputInventory) CreateInputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Tạo BillCode
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.KioskCode = body.KioskCode
	inventoryStatus.Code = body.Code
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		inventoryStatus.Source = body.Source
		inventoryStatus.KioskName = body.KioskName
		inventoryStatus.UserUpdate = prof.UserName
		inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_PENDING
		inventoryStatus.Create()
	}

	inputItem := kiosk_inventory.InventoryInputItem{}
	inputItem.Code = body.Code
	inputItem.PartnerUid = body.PartnerUid
	inputItem.CourseUid = body.CourseUid
	inputItem.Quantity = body.Quantity
	inputItem.ItemCode = body.ItemCode
	inputItem.Source = body.Source
	inputItem.Note = body.Note
	inputItem.ReviewUserUid = prof.UserName
	inputItem.KioskCode = body.KioskCode
	inputItem.KioskName = body.KioskName

	goodsService := model_service.GroupServices{
		GroupCode: body.GoodsCode,
	}

	errFindGoodsService := goodsService.FindFirst()
	if errFindGoodsService != nil {
		response_message.BadRequest(c, errFindGoodsService.Error())
		return
	}

	inputItem.ItemInfo = kiosk_inventory.ItemInfo{
		Price:     body.Price,
		ItemName:  body.ItemName,
		GroupName: goodsService.GroupName,
		GroupType: goodsService.Type,
		GroupCode: body.GoodsCode,
	}

	inputItem.InputDate = datatypes.Date(time.Now())

	if err := inputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (item CKioskOutputInventory) CreateInputBill(c *gin.Context, prof models.CmsUser) {
	var body request.CreateKioskInventoryBillBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.KioskCode = body.KioskCode
	inventoryStatus.KioskName = body.KioskName
	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_PENDING
	inventoryStatus.CourseUid = body.Code
	inventoryStatus.UserUpdate = prof.UserName
	inventoryStatus.Source = body.Source

	err := inventoryStatus.Create()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okResponse(c, inventoryStatus)
}

func (item CKioskInputInventory) AcceptInputBill(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInsertBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.Code = body.Code
	inventoryStatus.KioskCode = body.KioskCode
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}

	if inventoryStatus.BillStatus != constants.KIOSK_BILL_INVENTORY_PENDING {
		response_message.BadRequest(c, body.Code+" đã "+inventoryStatus.BillStatus)
		return
	}
	// Thêm ds item vào Inventory
	item.addItemToInventory(body.Code, body.CourseUid, body.PartnerUid)

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_ACCEPT
	if err := inventoryStatus.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// TODO Giảm ds item trong Inventory

	okRes(c)
}

func (_ CKioskInputInventory) addItemToInventory(code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryInputItem{}
	item.Code = code
	list, _, _ := item.FindAllList()

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			KioskCode:  data.KioskCode,
			Code:       data.ItemCode,
			InputCode:  data.Code,
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			ItemInfo:   data.ItemInfo,
		}

		item.ItemInfo = data.ItemInfo

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

// func removeItemFromInventory(code string) error {
// 	// Get danh sách item của bill
// 	item := kiosk_inventory.InventoryOutputItem{}
// 	item.Code = code
// 	list, _, _ := item.FindAllList()

// 	for _, data := range list {
// 		item := kiosk_inventory.InventoryItem{
// 			KioskCode: data.KioskCode,
// 			KioskType: data.KioskType,
// 			Code:      data.ItemCode,
// 			InputCode: data.Code,
// 		}

// 		if err := item.FindFirst(); err != nil {
// 			item.Quantity = item.Quantity - data.Quantity
// 			if errUpd := item.Update(); errUpd != nil {
// 				return errUpd
// 			}
// 		}
// 	}
// 	return nil
// }

func (_ CKioskInputInventory) ReturnInputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInsertBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.Code = body.Code
	inventoryStatus.KioskCode = body.KioskCode
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}

	if inventoryStatus.BillStatus != constants.KIOSK_BILL_INVENTORY_PENDING {
		response_message.BadRequest(c, body.Code+" đã "+inventoryStatus.BillStatus)
		return
	}

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_RETURN
	if err := inventoryStatus.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}
	okRes(c)
}

func (_ CKioskInputInventory) GetInputItems(c *gin.Context, prof models.CmsUser) {
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

	inputItems := kiosk_inventory.InventoryInputItem{}
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

func (_ CKioskInputInventory) GetInputBills(c *gin.Context, prof models.CmsUser) {
	var form request.GetBill
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

	inputItems := kiosk_inventory.InputInventoryBill{}
	inputItems.BillStatus = form.BillStatus
	inputItems.KioskCode = form.KioskCode
	inputItems.PartnerUid = form.PartnerUid
	inputItems.CourseUid = form.CourseUid
	list, total, err := inputItems.FindList(page, form.BillStatus)

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
