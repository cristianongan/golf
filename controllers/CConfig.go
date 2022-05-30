package controllers

import (
	"start/constants"
	"start/models"

	"github.com/gin-gonic/gin"
)

type CConfig struct{}

// Get booking Config của ngày
func (_ *CConfig) GetConfig(c *gin.Context, prof models.CmsUser) {
	// form := request.GetListBookingSettingForm{}
	// if bindErr := c.ShouldBind(&form); bindErr != nil {
	// 	response_message.BadRequest(c, bindErr.Error())
	// 	return
	// }

	resp := map[string]interface{}{
		"nationality": constants.NATIONAL_LIST,
	}

	okResponse(c, resp)

}
