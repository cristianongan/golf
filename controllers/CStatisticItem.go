package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CStatisticItem struct{}

func (_ CStatisticItem) AddItemToStatistic(db *gorm.DB) {
	now := time.Now().Format(constants.DATE_FORMAT_1)
	yesterday := time.Now().AddDate(0, 0, -1).Format(constants.DATE_FORMAT_1)

	outputInventory := kiosk_inventory.InventoryOutputItem{
		OutputDate: now,
		ServiceId:  1,
	}
	outputList, _ := outputInventory.FindStatistic(db)

	inputInventory := kiosk_inventory.InventoryInputItem{
		InputDate: now,
		ServiceId: 1,
	}
	inputList, _ := inputInventory.FindStatistic(db)

	commonItemCode := []kiosk_inventory.StatisticItem{}

	itemList := []kiosk_inventory.StatisticItem{}

	if !(len(outputList) == 0) || (len(inputList) == 0) {
		for _, output := range outputList {
			for _, input := range inputList {
				if output.ItemCode == input.ItemCode &&
					output.PartnerUid == input.PartnerUid &&
					output.CourseUid == input.CourseUid &&
					output.ServiceId == input.ServiceId {

					newItem := InitStatisticItem(input, yesterday, output.Total, input.Total, now)

					itemList = append(itemList, newItem)

					commonItemCode = append(commonItemCode, kiosk_inventory.StatisticItem{
						PartnerUid: input.PartnerUid,
						CourseUid:  input.CourseUid,
						ItemCode:   input.ItemCode,
						ServiceId:  input.ServiceId,
					})
				}
			}
		}
	}

	for _, output := range outputList {
		check := false
		for _, common := range commonItemCode {
			if output.ItemCode == common.ItemCode &&
				output.PartnerUid == common.PartnerUid &&
				output.CourseUid == common.CourseUid &&
				output.ServiceId == common.ServiceId {
				check = true
			}
		}
		if !check {
			newItem := InitStatisticItem(output, yesterday, output.Total, 0, now)
			itemList = append(itemList, newItem)
		}
	}

	for _, input := range inputList {
		check := false
		for _, common := range commonItemCode {
			if input.ItemCode == common.ItemCode &&
				input.PartnerUid == common.PartnerUid &&
				input.CourseUid == common.CourseUid &&
				input.ServiceId == common.ServiceId {
				check = true
			}
		}
		if !check {
			newItem := InitStatisticItem(input, yesterday, 0, input.Total, now)
			itemList = append(itemList, newItem)
		}
	}

	for _, data := range itemList {
		data.Create()
	}
}

func (_ CStatisticItem) TestAddItemToStatistic(c *gin.Context, prof models.CmsUser) {
	// now := time.Now().Format(constants.DATE_FORMAT_1)
	now := "22/10/2022"
	yesterday := time.Now().AddDate(0, 0, -1).Format(constants.DATE_FORMAT_1)

	outputInventory := kiosk_inventory.InventoryOutputItem{
		OutputDate: now,
		ServiceId:  1,
	}
	outputList, _ := outputInventory.FindStatistic(datasources.GetDatabase())

	inputInventory := kiosk_inventory.InventoryInputItem{
		InputDate: now,
		ServiceId: 1,
	}
	inputList, _ := inputInventory.FindStatistic(datasources.GetDatabase())

	commonItemCode := []kiosk_inventory.StatisticItem{}

	itemList := []kiosk_inventory.StatisticItem{}

	if len(outputList) != 0 && len(inputList) != 0 {
		for _, output := range outputList {
			for _, input := range inputList {
				if output.ItemCode == input.ItemCode &&
					output.PartnerUid == input.PartnerUid &&
					output.CourseUid == input.CourseUid &&
					output.ServiceId == input.ServiceId {

					newItem := InitStatisticItem(input, yesterday, output.Total, input.Total, now)

					itemList = append(itemList, newItem)

					commonItemCode = append(commonItemCode, kiosk_inventory.StatisticItem{
						PartnerUid: input.PartnerUid,
						CourseUid:  input.CourseUid,
						ItemCode:   input.ItemCode,
						ServiceId:  input.ServiceId,
					})
				}
			}
		}
	}

	for _, output := range outputList {
		check := false
		for _, common := range commonItemCode {
			if output.ItemCode == common.ItemCode &&
				output.PartnerUid == common.PartnerUid &&
				output.CourseUid == common.CourseUid &&
				output.ServiceId == common.ServiceId {
				check = true
			}
		}
		if !check {
			newItem := InitStatisticItem(output, yesterday, output.Total, 0, now)
			itemList = append(itemList, newItem)
		}
	}

	for _, input := range inputList {
		check := false
		for _, common := range commonItemCode {
			if input.ItemCode == common.ItemCode &&
				input.PartnerUid == common.PartnerUid &&
				input.CourseUid == common.CourseUid &&
				input.ServiceId == common.ServiceId {
				check = true
			}
		}
		if !check {
			newItem := InitStatisticItem(input, yesterday, 0, input.Total, now)
			itemList = append(itemList, newItem)
		}
	}

	for _, data := range itemList {
		data.Create()
	}
	okRes(c)
}

func InitStatisticItem(item kiosk_inventory.OutputStatisticItem, yesterday string, outputTotal int64, inputTotal int64, now string) kiosk_inventory.StatisticItem {
	var endingInventory int64 = 0

	yesterdayItem := kiosk_inventory.StatisticItem{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		ItemCode:   item.ItemCode,
		ServiceId:  item.ServiceId,
		Time:       yesterday,
	}

	if errFind := yesterdayItem.FindFirst(); errFind == nil {
		endingInventory = yesterdayItem.Total
	}

	totalNow := endingInventory + inputTotal - outputTotal

	newItem := kiosk_inventory.StatisticItem{
		PartnerUid:      item.PartnerUid,
		CourseUid:       item.CourseUid,
		ItemCode:        item.ItemCode,
		ServiceId:       item.ServiceId,
		Import:          inputTotal,
		Export:          outputTotal,
		EndingInventory: yesterdayItem.Total,
		Total:           totalNow,
		Time:            now,
	}
	return newItem
}

func (_ CStatisticItem) GetItemStatisticDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var form request.GetItems

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

	outputInventory := kiosk_inventory.StatisticItem{
		ItemCode:   form.ItemCode,
		ServiceId:  form.ServiceId,
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	inventoryItem := kiosk_inventory.InventoryItem{
		Code:       form.ItemCode,
		ServiceId:  form.ServiceId,
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if errInventoryItem := inventoryItem.FindFirst(db); errInventoryItem != nil {
		response_message.BadRequest(c, errInventoryItem.Error())
		return
	}

	var fromDateInt int64 = 0
	var toDateInt int64 = 0

	if form.FromDate != "" {
		fromDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.FromDate)
	}

	if form.ToDate != "" {
		toDateInt = utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, form.ToDate)
	}

	outputList, total, _ := outputInventory.FindList(page, fromDateInt, toDateInt)

	res := map[string]interface{}{
		"total":     total,
		"item-info": inventoryItem.ItemInfo,
		"quantity":  inventoryItem.Quantity,
		"data":      outputList,
	}

	okResponse(c, res)
}