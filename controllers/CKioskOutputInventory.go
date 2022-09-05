package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CKioskOutputInventory struct{}

func (_ CKioskOutputInventory) CreateOutputItem(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryOutputItemBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}
	inventoryStatus := kiosk_inventory.OutputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.CourseUid = body.Code
	inventoryStatus.KioskCode = body.KioskCode
	if errInventoryStatus := inventoryStatus.FindFirst(); errInventoryStatus != nil {
		inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_SELL
		inventoryStatus.UserUpdate = prof.UserName
		inventoryStatus.KioskName = body.KioskName
		inventoryStatus.Source = body.Source
		inventoryStatus.Create()
	}

	inputItem := kiosk_inventory.InventoryOutputItem{}
	inputItem.Code = body.Code
	inputItem.KioskCode = body.KioskCode
	inputItem.KioskName = body.KioskName
	inputItem.PartnerUid = body.PartnerUid
	inputItem.CourseUid = body.CourseUid
	inputItem.Quantity = body.Quantity
	inputItem.ItemCode = body.ItemCode

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

	if err := inputItem.Create(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (item CKioskOutputInventory) CreateOutputBill(c *gin.Context, prof models.CmsUser) {
	var body request.CreateKioskInventoryBillBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}
	inventoryStatus := kiosk_inventory.OutputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.KioskCode = body.KioskCode
	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_SELL
	inventoryStatus.CourseUid = body.Code
	inventoryStatus.UserUpdate = prof.UserName
	inventoryStatus.KioskName = body.KioskName
	inventoryStatus.Source = body.Source

	err := inventoryStatus.Create()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okResponse(c, inventoryStatus)
}

func (item CKioskOutputInventory) TransferOutputBill(c *gin.Context, prof models.CmsUser) {
	var body request.KioskInventoryInsertBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	// Update trạng thái kiosk inventory
	inventoryStatus := kiosk_inventory.OutputInventoryBill{}
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
	item.removeItemToInventory(body.Code, body.CourseUid, body.PartnerUid)

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_ACCEPT
	if err := inventoryStatus.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// TODO Giảm ds item trong Inventory

	okRes(c)
}

func (_ CKioskOutputInventory) removeItemToInventory(code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryOutputItem{}
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

		if err := item.FindFirst(); err != nil {
			item.Quantity = item.Quantity - data.Quantity
			if errUpd := item.Update(); errUpd != nil {
				return errUpd
			}
		}
	}
	return nil
}

func (_ CKioskOutputInventory) GetOutputBills(c *gin.Context, prof models.CmsUser) {
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

	outputItems := kiosk_inventory.OutputInventoryBill{}
	outputItems.BillStatus = form.BillStatus
	outputItems.KioskCode = form.KioskCode
	outputItems.PartnerUid = form.PartnerUid
	outputItems.CourseUid = form.CourseUid
	list, total, err := outputItems.FindList(page, form.BillStatus)

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
func (_ CKioskOutputInventory) GetOutputItems(c *gin.Context, prof models.CmsUser) {
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

	outputItems := kiosk_inventory.InventoryOutputItem{}
	outputItems.KioskCode = form.KioskCode
	outputItems.PartnerUid = form.PartnerUid
	outputItems.CourseUid = form.CourseUid
	list, total, err := outputItems.FindList(page)

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
