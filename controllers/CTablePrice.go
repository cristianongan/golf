package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CTablePrice struct{}

func (_ *CTablePrice) CreateTablePrice(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	errC := tablePrice.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	// Tạo các golf fee từ Old Price id
	if body.OldPriceId > 0 {
		//Use Batch Created
		golfFeeR := models.GolfFee{
			TablePriceId: body.OldPriceId,
		}

		listGolfFee, errList := golfFeeR.FindAllByTablePrice(db)
		if errList == nil {
			listCreate := []models.GolfFee{}
			for _, v := range listGolfFee {
				v.Id = 0
				v.TablePriceId = tablePrice.Id
				listCreate = append(listCreate, v)
			}

			if len(listCreate) > 0 {
				golfFeeC := models.GolfFee{}
				errBatchCreate := golfFeeC.BatchInsert(db, listCreate)
				if errBatchCreate != nil {
					response_message.InternalServerError(c, errBatchCreate.Error())
					return
				}
			}
		}
	}

	okResponse(c, tablePrice)
}

func (_ *CTablePrice) GetListTablePrice(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	list, total, err := tablePriceR.FindList(db, page)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	tablePriceIdStr := c.Param("id")
	tablePriceId, err := strconv.ParseInt(tablePriceIdStr, 10, 64)
	if err != nil || tablePriceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	tablePrice := models.TablePrice{}
	tablePrice.Id = tablePriceId
	errF := tablePrice.FindFirst(db)
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

	errUdp := tablePrice.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, tablePrice)
}

func (_ *CTablePrice) DeleteTablePrice(c *gin.Context, prof models.CmsUser) {
	response_message.BadRequestDynamicKey(c, "TABLE_PRICE_DEL_NOTE", "")
	return

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	tablePriceIdStr := c.Param("id")
	tablePriceId, err := strconv.ParseInt(tablePriceIdStr, 10, 64)
	if err != nil || tablePriceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	tablePrice := models.TablePrice{}
	tablePrice.Id = tablePriceId
	errF := tablePrice.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := tablePrice.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
