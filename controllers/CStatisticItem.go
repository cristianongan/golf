package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	kiosk_inventory "start/models/kiosk-inventory"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CStatisticItem struct{}

func (_ CStatisticItem) AddItemToStatistic() {
	now := time.Now().Format(constants.DATE_FORMAT_1)

	outputInventory := kiosk_inventory.InventoryOutputItem{
		OutputDate: now,
	}
	outputList, _ := outputInventory.FindStatistic()

	inputInventory := kiosk_inventory.InventoryInputItem{
		InputDate: now,
	}
	inputList, _ := inputInventory.FindStatistic()

	commonItemCode := []kiosk_inventory.StatisticItem{}

	itemList := []kiosk_inventory.StatisticItem{}

	if !(len(outputList) == 0) || (len(inputList) == 0) {
		for _, output := range outputList {
			for _, input := range inputList {
				if output.ItemCode == input.ItemCode &&
					output.PartnerUid == input.PartnerUid &&
					output.CourseUid == input.CourseUid {
					itemList = append(itemList, kiosk_inventory.StatisticItem{
						PartnerUid: input.PartnerUid,
						CourseUid:  input.CourseUid,
						ItemCode:   input.ItemCode,
						Import:     input.Total,
						Export:     output.Total,
					})
					commonItemCode = append(commonItemCode, kiosk_inventory.StatisticItem{
						PartnerUid: input.PartnerUid,
						CourseUid:  input.CourseUid,
						ItemCode:   input.ItemCode,
					})
				}
			}
		}
	}

	if len(commonItemCode) == 0 {
		for _, output := range outputList {
			itemList = append(itemList, kiosk_inventory.StatisticItem{
				PartnerUid: output.PartnerUid,
				CourseUid:  output.CourseUid,
				ItemCode:   output.ItemCode,
				Import:     0,
				Export:     output.Total,
			})
		}
	} else {
		for _, output := range outputList {
			check := false
			for _, common := range commonItemCode {
				if output.ItemCode == common.ItemCode &&
					output.PartnerUid == common.PartnerUid &&
					output.CourseUid == common.CourseUid {
					check = true
				}
			}
			if !check {
				itemList = append(itemList, kiosk_inventory.StatisticItem{
					PartnerUid: output.PartnerUid,
					CourseUid:  output.CourseUid,
					ItemCode:   output.ItemCode,
					Import:     0,
					Export:     output.Total,
				})
			}
		}
	}

	if len(commonItemCode) == 0 {
		for _, input := range inputList {
			itemList = append(itemList, kiosk_inventory.StatisticItem{
				PartnerUid: input.PartnerUid,
				CourseUid:  input.CourseUid,
				ItemCode:   input.ItemCode,
				Import:     input.Total,
				Export:     0,
			})
		}
	} else {
		for _, input := range inputList {
			check := false
			for _, common := range commonItemCode {
				if input.ItemCode == common.ItemCode &&
					input.PartnerUid == common.PartnerUid &&
					input.CourseUid == common.CourseUid {
					check = true
				}
			}
			if !check {
				itemList = append(itemList, kiosk_inventory.StatisticItem{
					PartnerUid: input.PartnerUid,
					CourseUid:  input.CourseUid,
					ItemCode:   input.ItemCode,
					Import:     0,
					Export:     input.Total,
				})
			}
		}
	}

	for _, data := range itemList {
		data.Create()
	}
}

func (_ CStatisticItem) GetStatistic(c *gin.Context, prof models.CmsUser) {
	var form request.GetItems

	if err := c.BindJSON(&form); err != nil {
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
		ItemCode: form.ItemCode,
	}

	outputList, total, _ := outputInventory.FindList(page)

	res := map[string]interface{}{
		"total": total,
		"data":  outputList,
	}

	okResponse(c, res)
}
