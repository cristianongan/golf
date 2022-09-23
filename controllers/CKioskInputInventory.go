package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CKioskInputInventory struct{}

func (item CKioskInputInventory) CreateManualInputBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateBillBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	billcode := time.Now().Format("20060102150405")
	if errInputBill := item.MethodInputBill(c, prof, body,
		constants.KIOSK_BILL_INVENTORY_ACCEPT, billcode); errInputBill != nil {
		response_message.BadRequest(c, errInputBill.Error())
		return
	}

	item.addItemToInventory(db, body.ServiceId, billcode, body.CourseUid, body.PartnerUid)

	okRes(c)
}

func (item CKioskInputInventory) MethodInputBill(c *gin.Context, prof models.CmsUser, body request.CreateBillBody, billtype string, billcode string) error {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	inventoryStatus := kiosk_inventory.InputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.Code = billcode
	inventoryStatus.ServiceId = body.ServiceId
	inventoryStatus.ServiceName = body.ServiceName
	inventoryStatus.BillStatus = billtype
	inventoryStatus.UserUpdate = prof.UserName
	inventoryStatus.UserExport = body.UserExport
	inventoryStatus.ServiceExportId = body.SourceId
	inventoryStatus.ServiceExportName = body.SourceName
	inventoryStatus.Note = body.Note
	inventoryStatus.OutputDate = body.OutputDate

	quantity := 0

	for _, data := range body.ListItem {
		inputItem := kiosk_inventory.InventoryInputItem{}
		inputItem.Code = billcode
		inputItem.PartnerUid = body.PartnerUid
		inputItem.CourseUid = body.CourseUid
		inputItem.Quantity = data.Quantity
		inputItem.ItemCode = data.ItemCode
		inputItem.ServiceId = body.ServiceId
		inputItem.ServiceName = body.ServiceName

		goodsService := model_service.GroupServices{
			GroupCode: data.GroupCode,
		}

		if errSer := validateItemCodeInService(db, goodsService.Type, data.ItemCode); errSer != nil {
			return errSer
		}

		if errFindGoodsService := goodsService.FindFirst(db); errFindGoodsService != nil {
			return errFindGoodsService
		}

		var itemType = ""

		if goodsService.Type == constants.GROUP_FB {
			fb := model_service.FoodBeverage{
				FBCode: data.ItemCode,
			}

			if err := fb.FindFirst(db); err == nil {
				itemType = fb.Type
			}
		}

		if itemType == "" {
			itemType = goodsService.Type
		}

		inputItem.ItemInfo = kiosk_inventory.ItemInfo{
			Price:     data.Price,
			ItemName:  data.ItemName,
			GroupName: goodsService.GroupName,
			GroupType: itemType,
			GroupCode: data.GroupCode,
			Unit:      data.Unit,
		}

		inputItem.InputDate = time.Now().Format(constants.DATE_FORMAT_1)

		if err := inputItem.Create(db); err != nil {
			return err
		}

		quantity += int(data.Quantity)
	}

	inventoryStatus.Quantity = int64(quantity)
	err := inventoryStatus.Create(db)
	if err != nil {
		return err
	}
	return nil
}

func (item CKioskInputInventory) AcceptInputBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	if errInventoryStatus := inventoryStatus.FindFirst(db); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}

	if inventoryStatus.BillStatus != constants.KIOSK_BILL_INVENTORY_PENDING {
		response_message.BadRequest(c, body.Code+" đã "+inventoryStatus.BillStatus)
		return
	}
	// Thêm ds item vào Inventory
	item.addItemToInventory(db, body.ServiceId, body.Code, body.CourseUid, body.PartnerUid)

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_ACCEPT
	inventoryStatus.UserUpdate = prof.UserName
	inventoryStatus.InputDate = time.Now().Unix()
	if err := inventoryStatus.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okResponse(c, inventoryStatus)
}

func (_ CKioskInputInventory) addItemToInventory(db *gorm.DB, serviceId int64, code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryInputItem{}
	item.Code = code
	list, _, _ := item.FindAllList(db)

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			ServiceId:  serviceId,
			Code:       data.ItemCode,
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
		}

		if err := item.FindFirst(db); err != nil {
			item.ItemInfo = data.ItemInfo
			item.Quantity = data.Quantity
			if errCre := item.Create(db); errCre != nil {
				return errCre
			}
		} else {
			item.Quantity = item.Quantity + data.Quantity
			if errUpd := item.Update(db); errUpd != nil {
				return errUpd
			}
		}
	}
	return nil
}

func (_ CKioskInputInventory) ReturnInputItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	if errInventoryStatus := inventoryStatus.FindFirst(db); errInventoryStatus != nil {
		response_message.BadRequest(c, "")
		return
	}

	if inventoryStatus.BillStatus != constants.KIOSK_BILL_INVENTORY_PENDING {
		response_message.BadRequest(c, body.Code+" đã "+inventoryStatus.BillStatus)
		return
	}

	inventoryStatus.BillStatus = constants.KIOSK_BILL_INVENTORY_RETURN
	inventoryStatus.UserUpdate = prof.UserName
	if err := inventoryStatus.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Trả lại hàng cho Inventory
	cKioskOutputInventory := CKioskOutputInventory{}
	cKioskOutputInventory.returnItemToInventory(db, inventoryStatus.ServiceExportId, body.Code, body.CourseUid, body.PartnerUid)
	okResponse(c, inventoryStatus)
}

func (_ CKioskInputInventory) GetInputItems(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	list, total, err := inputItems.FindList(db, page, form.Type)

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

func (_ CKioskInputInventory) GetInputItemsForStatis(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	var fromDateInt int64 = 0
	var toDateInt int64 = 0

	if form.FromDate != "" {
		fromDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.FromDate)
	}

	if form.ToDate != "" {
		toDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.ToDate)
	}

	list, total, err := inputItems.FindListForStatistic(db, page, fromDateInt, toDateInt)

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	inputItems.Code = form.BillCode
	list, total, err := inputItems.FindList(db, page, form.BillStatus)

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
