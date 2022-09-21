package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CKioskOutputInventory struct{}

func (item CKioskOutputInventory) CreateOutputBill(c *gin.Context, prof models.CmsUser) {
	var body request.CreateOutputBillBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	billcode := time.Now().Format("20060102150405")
	if errOutputBill := item.MethodOutputBill(c, prof, body,
		constants.KIOSK_BILL_INVENTORY_TRANSFER, billcode); errOutputBill != nil {
		response_message.BadRequest(c, errOutputBill.Error())
		return
	}

	// Tạo import đơn
	cKioskInputInventory := CKioskInputInventory{}
	bodyInputBill := request.CreateBillBody{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		ServiceId:   body.SourceId,
		ServiceName: body.SourceName,
		SourceId:    body.ServiceId,
		SourceName:  body.ServiceName,
		ListItem:    body.ListItem,
		Note:        body.Note,
		UserExport:  body.UserExport,
		OutputDate:  body.OutputDate,
	}

	if errInputBill := cKioskInputInventory.MethodInputBill(c, prof,
		bodyInputBill, constants.KIOSK_BILL_INVENTORY_PENDING, billcode); errInputBill != nil {
		response_message.BadRequest(c, errInputBill.Error())
		return
	}

	okRes(c)
}

func (item CKioskOutputInventory) MethodOutputBill(c *gin.Context, prof models.CmsUser, body request.CreateOutputBillBody, billtype string, billcode string) error {
	inventoryStatus := kiosk_inventory.OutputInventoryBill{}
	inventoryStatus.PartnerUid = body.PartnerUid
	inventoryStatus.CourseUid = body.CourseUid
	inventoryStatus.ServiceId = body.ServiceId
	inventoryStatus.BillStatus = billtype
	inventoryStatus.Code = billcode
	inventoryStatus.UserUpdate = body.UserExport
	inventoryStatus.ServiceName = body.ServiceName
	inventoryStatus.ServiceImportId = body.SourceId
	inventoryStatus.OutputDate = body.OutputDate
	inventoryStatus.Note = body.Note
	inventoryStatus.Bag = body.Bag
	inventoryStatus.CustomerName = body.CustomerName

	kiosk := model_service.Kiosk{}
	kiosk.Id = body.SourceId

	if errFind := kiosk.FindFirst(); errFind != nil {
		return errors.New("SourceId" + strconv.Itoa(int(body.SourceId)) + "not found")
	}
	inventoryStatus.ServiceImportName = kiosk.KioskName

	quantity := 0

	for _, data := range body.ListItem {

		// check lượng hàng trong kho có đủ để xuất không
		itemInInventory := kiosk_inventory.InventoryItem{
			ServiceId:  body.ServiceId,
			Code:       data.ItemCode,
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
		}

		if errInventory := itemInInventory.FindFirst(); errInventory != nil {
			message := data.ItemCode + " không tìm thấy "
			return errors.New(message)
		}

		if data.Quantity > itemInInventory.Quantity {
			message := data.ItemCode + " vượt quá số lượng trong kho "
			return errors.New(message)
		}

		goodsService := model_service.GroupServices{
			GroupCode: data.GroupCode,
		}

		errFindGoodsService := goodsService.FindFirst()
		if errFindGoodsService != nil {
			return errFindGoodsService
		}

		var itemType = ""

		if goodsService.Type == constants.GROUP_FB {
			fb := model_service.FoodBeverage{
				FBCode: data.ItemCode,
			}

			if err := fb.FindFirst(); err == nil {
				itemType = fb.Type
			}
		}

		if errFindGoodsService := goodsService.FindFirst(); errFindGoodsService != nil {
			return errFindGoodsService
		}

		if itemType == "" {
			itemType = goodsService.Type
		}

		outputItem := kiosk_inventory.InventoryOutputItem{}
		outputItem.Code = billcode
		outputItem.PartnerUid = body.PartnerUid
		outputItem.CourseUid = body.CourseUid
		outputItem.Quantity = data.Quantity
		outputItem.ItemCode = data.ItemCode
		outputItem.ServiceId = body.ServiceId
		outputItem.ServiceName = body.ServiceName
		outputItem.Amount = int64(data.Price) * data.Quantity

		outputItem.ItemInfo = kiosk_inventory.ItemInfo{
			Price:     data.Price,
			ItemName:  data.ItemName,
			GroupName: goodsService.GroupName,
			GroupType: itemType,
			GroupCode: data.GroupCode,
			Unit:      data.Unit,
		}

		outputItem.OutputDate = time.Now().Format(constants.DATE_FORMAT_1)

		if err := outputItem.Create(); err != nil {
			return err
		}

		itemInInventory.Quantity = itemInInventory.Quantity - data.Quantity
		if errUpd := itemInInventory.Update(); errUpd != nil {
			return errUpd
		}

		quantity += int(data.Quantity)
	}

	inventoryStatus.Quantity = int64(quantity)
	err := inventoryStatus.Create()
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return err
	}

	return nil
}

func (_ CKioskOutputInventory) removeItemFromInventory(serviceId int64, code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryOutputItem{}
	item.Code = code
	list, _, _ := item.FindAllList()

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			ServiceId:  serviceId,
			Code:       data.ItemCode,
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
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

func (_ CKioskOutputInventory) returnItemToInventory(serviceId int64, code string, courseUid string, partnerUid string) error {
	// Get danh sách item của bill
	item := kiosk_inventory.InventoryInputItem{}
	item.Code = code
	list, _, _ := item.FindAllList()

	for _, data := range list {
		item := kiosk_inventory.InventoryItem{
			ServiceId:  serviceId,
			Code:       data.ItemCode,
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
		}

		if err := item.FindFirst(); err != nil {
			item.Quantity = item.Quantity + data.Quantity
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
	outputItems.ServiceId = form.ServiceId
	outputItems.PartnerUid = form.PartnerUid
	outputItems.CourseUid = form.CourseUid
	outputItems.Code = form.BillCode
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
	outputItems.ServiceId = form.ServiceId
	outputItems.PartnerUid = form.PartnerUid
	outputItems.CourseUid = form.CourseUid
	outputItems.ItemCode = form.ItemCode
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

func (_ CKioskOutputInventory) GetOutputItemsForStatistic(c *gin.Context, prof models.CmsUser) {
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
	outputItems.ServiceId = form.ServiceId
	outputItems.PartnerUid = form.PartnerUid
	outputItems.CourseUid = form.CourseUid
	outputItems.ItemCode = form.ItemCode

	var fromDateInt int64 = 0
	var toDateInt int64 = 0

	if form.FromDate != "" {
		fromDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.FromDate)
	}

	if form.ToDate != "" {
		toDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.ToDate)
	}

	list, total, err := outputItems.FindListForStatistic(page, fromDateInt, toDateInt)

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
