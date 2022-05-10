package controllers

import (
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CSystem struct{}

func (_ *CSystem) GetListCategoryType(c *gin.Context, prof models.CmsUser) {

	cusTypesGet := models.CustomerType{}

	list, err := cusTypesGet.FindAll()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": len(list),
		"data":  list,
	}

	okResponse(c, res)
}
