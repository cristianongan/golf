package controllers

import (
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CRevenueReport struct{}

func (_ *CRevenueReport) GetReportRevenueFoodBeverage(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.RevenueReportFBForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	serviceCart := models.ServiceCart{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	list, total, err := serviceCart.FindReport(db, page, form.FromDate, form.ToDate, form.TypeService)
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

func (_ *CRevenueReport) GetReportRevenueDetailFBBag(c *gin.Context, prof models.CmsUser) {}

func (_ *CRevenueReport) GetReportRevenueDetailFB(c *gin.Context, prof models.CmsUser) {}
