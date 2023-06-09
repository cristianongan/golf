package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CReport struct{}

func (_ *CReport) GetListReportMainBagSubBagToDay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingSettingGroupForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" || form.CourseUid == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	if dateDisplay == "" {
		response_message.InternalServerError(c, "date error")
		return
	}

	//Find have MainBag List
	mainBagR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: dateDisplay,
	}
	listBook, _ := mainBagR.FindListForReportForMainBagSubBag(db)
	listHaveMainBags := []model_booking.BookingForReportMainBagSubBags{}
	listHaveSubBags := []model_booking.BookingForReportMainBagSubBags{}

	for _, v := range listBook {
		if v.MainBags != nil && len(v.MainBags) > 0 {
			listHaveMainBags = append(listHaveMainBags, v)
		} else {
			if v.SubBags != nil && len(v.SubBags) > 0 {
				listHaveSubBags = append(listHaveSubBags, v)
			}
		}
	}

	totalSubBag := 0
	totalMyCost := int64(0)
	totalToBePaid := int64(0)

	listReportMainBagResponse := []response.ReportMainBagResponse{}

	for _, v := range listHaveSubBags {
		mainB := response.ReportMainBagResponse{}
		mainB.PartnerUid = v.PartnerUid
		mainB.CourseUid = v.CourseUid
		mainB.Uid = v.Uid
		mainB.Bag = v.Bag
		mainB.BagStatus = v.BagStatus
		mainB.BookingDate = v.BookingDate
		mainB.CheckOutTime = v.CheckOutTime
		mainB.MyCost = v.CurrentBagPrice.Amount
		mainB.ToBePaid = v.MushPayInfo.MushPay

		if mainB.SubBag == nil {
			mainB.SubBag = response.ListReportSubBagResponse{}
		}

		for _, v1 := range listHaveMainBags {
			if v1.MainBags != nil && len(v1.MainBags) > 0 {
				if v.Bag == v1.MainBags[0].GolfBag {

					subB := response.ReportSubBagResponse{}
					subB.PartnerUid = v1.PartnerUid
					subB.CourseUid = v1.CourseUid
					subB.Uid = v1.Uid
					subB.Bag = v1.Bag
					subB.BagStatus = v1.BagStatus
					subB.BookingDate = v1.BookingDate
					subB.CheckOutTime = v1.CheckOutTime
					subB.MyCost = v1.CurrentBagPrice.Amount
					subB.ToBePaid = v1.MushPayInfo.MushPay

					totalSubBag += 1
					totalMyCost = totalMyCost + subB.MyCost
					totalToBePaid = totalToBePaid + subB.ToBePaid

					mainB.SubBag = append(mainB.SubBag, subB)
				}
			}
		}

		listReportMainBagResponse = append(listReportMainBagResponse, mainB)

	}

	resp := map[string]interface{}{
		"total_sub_bag":    totalSubBag,
		"total_my_cost":    totalMyCost,
		"total_to_be_paid": totalToBePaid,
		"data":             listReportMainBagResponse,
	}

	okResponse(c, resp)
}
