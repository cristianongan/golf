package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CHolePriceFormula struct{}

func (_ *CHolePriceFormula) CreateHolePriceFormula(c *gin.Context, prof models.CmsUser) {
	body := request.CreateHolePriceFormulaBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	holePriceFormula := models.HolePriceFormula{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Hole:       body.Hole,
	}
	//Check duplicated
	if holePriceFormula.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	holePriceFormula.StopByRain = body.StopByRain
	holePriceFormula.StopBySelf = body.StopBySelf

	errC := holePriceFormula.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, holePriceFormula)
}

func (_ *CHolePriceFormula) GetListHolePriceFormula(c *gin.Context, prof models.CmsUser) {
	form := request.GetListHolePriceFormulaForm{}
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

	holePriceFormulaR := models.HolePriceFormula{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := holePriceFormulaR.FindList(page)
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

func (_ *CHolePriceFormula) UpdateHolePriceFormula(c *gin.Context, prof models.CmsUser) {
	holePriceFormulaIdStr := c.Param("id")
	holePriceFormulaId, err := strconv.ParseInt(holePriceFormulaIdStr, 10, 64)
	if err != nil || holePriceFormulaId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	holePriceFormula := models.HolePriceFormula{}
	holePriceFormula.Id = holePriceFormulaId
	errF := holePriceFormula.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.HolePriceFormula{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// //Check duplicated
	// if body.IsDuplicated() {
	// 	response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
	// 	return
	// }

	if body.Hole > 0 {
		holePriceFormula.Hole = body.Hole
	}
	if body.StopByRain != "" {
		holePriceFormula.StopByRain = body.StopByRain
	}
	if body.StopBySelf != "" {
		holePriceFormula.StopBySelf = body.StopBySelf
	}

	errUdp := holePriceFormula.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, holePriceFormula)
}

func (_ *CHolePriceFormula) DeleteHolePriceFormula(c *gin.Context, prof models.CmsUser) {
	holePriceFormulaIdStr := c.Param("id")
	holePriceFormulaId, err := strconv.ParseInt(holePriceFormulaIdStr, 10, 64)
	if err != nil || holePriceFormulaId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	holePriceFormula := models.HolePriceFormula{}
	holePriceFormula.Id = holePriceFormulaId
	errF := holePriceFormula.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := holePriceFormula.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
