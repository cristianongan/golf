package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CAccountant struct{}

func (item CAccountant) ImportInventory(c *gin.Context) {
	var body request.CreateBillBody
	db := datasources.GetDatabaseWithPartner(body.PartnerUid)
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceInventory := model_service.Kiosk{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		ModelId:    models.ModelId{Id: body.ServiceId},
	}

	if errFind := serviceInventory.FindFirst(db); errFind != nil {
		response_message.BadRequestDynamicKey(c, "INVENTORY_NOT_FOUND", "")
		return
	}

	billcode := time.Now().Format("20060102150405")
	if errInputBill := MethodInputBill(c, nil, body,
		constants.KIOSK_BILL_INVENTORY_APPROVED, billcode); errInputBill != nil {
		response_message.BadRequest(c, errInputBill.Error())
		return
	}

	addItemToInventory(db, body.ServiceId, billcode, body.CourseUid, body.PartnerUid)
	okRes(c)
}
