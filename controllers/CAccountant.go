package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	model_service "start/models/service"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CAccountant struct{}

func (item CAccountant) ImportInventory(c *gin.Context) {
	var body request.AccountantAddInventory
	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceInventory := model_service.Kiosk{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		KioskCode:  body.InventoryCode,
	}

	if errFind := serviceInventory.FindFirst(db); errFind != nil {
		response_message.BadRequestDynamicKey(c, "INVENTORY_NOT_FOUND", "")
		return
	}

	newListItem := request.ListKioskInventoryInputItemBody{}
	for _, item := range body.ListItem {
		newListItem = append(newListItem, request.KioskInventoryItemBody{
			ItemCode: item.ItemCode,
			Price:    item.Price,
			Quantity: item.Quantity,
		})
	}

	newBody := request.CreateBillBody{
		ServiceId:  serviceInventory.Id,
		Note:       body.Note,
		ListItem:   newListItem,
		OutputDate: body.OutputDate,
	}

	billcode := time.Now().Format("20060102150405")
	if errInputBill := MethodInputBill(c, nil, newBody,
		constants.KIOSK_BILL_INVENTORY_APPROVED, billcode); errInputBill != nil {
		response_message.BadRequest(c, errInputBill.Error())
		return
	}

	addItemToInventory(db, serviceInventory.Id, billcode, body.CourseUid, body.PartnerUid)
	okRes(c)
}
