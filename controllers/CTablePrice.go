package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CTablePrice struct{}

func (_ *CTablePrice) CreateTablePrice(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTablePriceBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	tablePrice := models.TablePrice{
		Name:       body.Name,
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		FromDate:   body.FromDate,
	}
	tablePrice.Status = body.Status
	year, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.YEAR_FORMAT, body.FromDate)
	yearInt, _ := strconv.Atoi(year)
	tablePrice.Year = yearInt
	errC := tablePrice.Create()
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//TODO: Tao các golf fee từ Old Price id
	if body.OldPriceId > 0 {
		//Use Batch Created

	}

	okResponse(c, tablePrice)
}

func (_ *CTablePrice) GetListTablePrice(c *gin.Context, prof models.CmsUser) {
	form := request.GetListTablePriceForm{}
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

	tablePriceR := models.TablePrice{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Name:       form.TablePriceName,
		Year:       form.Year,
	}
	list, total, err := tablePriceR.FindList(page)
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

func (_ *CTablePrice) UpdateTablePrice(c *gin.Context, prof models.CmsUser) {
	tablePriceIdStr := c.Param("id")
	tablePriceId, err := strconv.ParseInt(tablePriceIdStr, 10, 64)
	if err != nil || tablePriceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	tablePrice := models.TablePrice{}
	tablePrice.Id = tablePriceId
	errF := tablePrice.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.TablePrice{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		tablePrice.Name = body.Name
	}
	if body.FromDate > 0 {
		tablePrice.FromDate = body.FromDate
		year, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.YEAR_FORMAT, body.FromDate)
		yearInt, _ := strconv.Atoi(year)
		tablePrice.Year = yearInt
	}
	if body.Status != "" {
		tablePrice.Status = body.Status
	}

	errUdp := tablePrice.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, tablePrice)
}

func (_ *CTablePrice) DeleteTablePrice(c *gin.Context, prof models.CmsUser) {

	response_message.BadRequest(c, "Không hỗ trợ xoá bảng giá")
	return

	tablePriceIdStr := c.Param("id")
	tablePriceId, err := strconv.ParseInt(tablePriceIdStr, 10, 64)
	if err != nil || tablePriceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	tablePrice := models.TablePrice{}
	tablePrice.Id = tablePriceId
	errF := tablePrice.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := tablePrice.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
