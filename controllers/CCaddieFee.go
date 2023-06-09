package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCaddieFee struct{}

func (_ *CCaddieFee) GetDetalListCaddieFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetDetailListCaddieFee{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieFee := models.CaddieFee{}
	caddieFee.CourseUid = query.CourseUid
	caddieFee.PartnerUid = query.PartnerUid
	caddieFee.CaddieCode = query.CaddieCode

	list, total, err := caddieFee.FindAll(db, query.Month)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (_ *CCaddieFee) GetListCaddieFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetListCaddieFee{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	caddieFee := models.CaddieFee{}
	caddieFee.CourseUid = query.CourseUid
	caddieFee.PartnerUid = query.PartnerUid
	caddieFee.CaddieCode = query.CaddieCode
	caddieFee.CaddieName = query.CaddieName

	list, total, err := caddieFee.FindAllGroupBy(db, page, query.Month)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}
