package controllers

import (
	"start/constants"
	"start/models"
	"start/utils"

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

	otherPaid := utils.ListString{}
	otherPaid = append(otherPaid, "Làm hỏng xe")
	otherPaid = append(otherPaid, "Phí thuê bể bơi")
	otherPaid = append(otherPaid, "Mất chìa khoá")
	otherPaid = append(otherPaid, "Phí làm mất Locker")

	resp := map[string]interface{}{
		"nationality":        constants.NATIONAL_LIST,
		"booking_other_paid": otherPaid,
		"units":              constants.UNIT_LIST,
	}

	okResponse(c, resp)

}
