package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/logger"
	"start/models"
	"start/utils/response_message"
	"strings"

	"github.com/gin-gonic/gin"
)

type CActivityLog struct{}

func (_ CActivityLog) GetLog(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetLogList{}
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

	var list interface{}
	var total int64
	var err error

	if strings.ToUpper(query.Action) == logger.EVENT_ACTION_UPDATE {
		activityLog := logger.UpdateActivityLogData{}
		activityLog.Category = strings.ToUpper(query.Category) + "_ACTIVITY_LOG"
		activityLog.Label = query.Code
		activityLog.Action = query.Action

		list, total, err = activityLog.FindList(db, page)
	}

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
