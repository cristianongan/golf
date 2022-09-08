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
)

type CKioskInputInventory struct{}

func (item CKioskInputInventory) CreateInputBill(c *gin.Context, prof models.CmsUser) {
	var body request.CreateInputBillBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.Code = body.BillCode
	inventoryStatus.ServiceId = body.ServiceId
	inventoryStatus.ServiceName = body.ServiceName
	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_PENDING
	inventoryStatus.UserUpdate = prof.UserName
	inventoryStatus.ServiceExportId = body.ServiceExportId
	inventoryStatus.ServiceExportName = body.ServiceExportName
	inventoryStatus.Note = body.Note
	inventoryStatus.InputDate = time.Now().Unix()

	quantity := 0

	for _, data := range body.ListItem {
		inputItem := kiosk_inventory.InventoryInputItem{}
		inputItem.Code = body.BillCode
		inputItem.PartnerUid = body.PartnerUid
		inputItem.CourseUid = body.CourseUid
		inputItem.Quantity = data.Quantity
		inputItem.ItemCode = data.ItemCode
		inputItem.UserUpdate = prof.UserName
		inputItem.ServiceId = body.ServiceId
		inputItem.ServiceName = body.ServiceName
		inputItem.UserUpdate = data.UserUpdate

		goodsService := model_service.GroupServices{
			GroupCode: data.GroupCode,
		}

		errFindGoodsService := goodsService.FindFirst()
		if errFindGoodsService != nil {
			response_message.BadRequest(c, errFindGoodsService.Error())
			return
		}

		inputItem.ItemInfo = kiosk_inventory.ItemInfo{
			Price:     data.Price,
			ItemName:  data.ItemName,
			GroupName: goodsService.GroupName,
			GroupType: goodsService.Type,
			GroupCode: data.GroupCode,
			Unit:      data.Unit,
		}

		inputItem.InputDate = time.Now().Format(constants.DATE_FORMAT_1)

		if err := inputItem.Create(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		quantity += int(data.Quantity)
	}

	inventoryStatus.Quantity = int64(quantity)
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
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.Code = body.Code
	inventoryStatus.ServiceId = body.ServiceId
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
			ServiceId:  data.ServiceId,
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
	inventoryStatus.ServiceId = body.ServiceId
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
	inputItems.ServiceId = form.ServiceId
	inputItems.PartnerUid = form.PartnerUid
	inputItems.CourseUid = form.CourseUid
	inputItems.ItemCode = form.ItemCode
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
	inputItems.ServiceId = form.ServiceId
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
