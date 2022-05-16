package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CGolfFee struct{}

func (_ *CGolfFee) CreateGolfFee(c *gin.Context, prof models.CmsUser) {
	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check Exits
	isDupli := checkDuplicateGolfFee(body)
	if isDupli {
		response_message.DuplicateRecord(c, "duplicated golf fee")
		return
	}

	// Check Table Price Exit
	tablePrice := models.TablePrice{}
	tablePrice.Id = body.TablePriceId
	errFind := tablePrice.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, "table price not found")
		return
	}

	// Check group Fee
	groupFee := models.GroupFee{}
	groupFee.Id = body.GroupId
	errFind = groupFee.FindFirst()
	if errFind != nil || groupFee.Id <= 0 {
		response_message.BadRequest(c, "group fee not found")
		return
	}
	errFind = nil

	// Tạo Fee
	golfFee := models.GolfFee{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		GuestStyle:   body.GuestStyle,
		Dow:          body.Dow,
		TablePriceId: body.TablePriceId,
	}

	golfFee.Status = body.Status
	golfFee.GuestStyleName = body.GuestStyleName
	golfFee.Dow = body.Dow
	golfFee.GreenFee = body.GreenFee
	golfFee.CaddieFee = body.CaddieFee
	golfFee.BuggyFee = body.BuggyFee
	golfFee.AccCode = body.AccCode
	golfFee.NodeOdd = body.NodeOdd
	golfFee.Note = body.Note
	golfFee.PaidType = body.PaidType
	golfFee.Idx = body.Idx
	golfFee.AccDebit = body.AccDebit
	golfFee.CustomerType = body.CustomerType
	golfFee.CustomerCategory = getCustomerCategoryFromCustomerType(body.CustomerType)
	golfFee.GroupName = body.GroupName
	golfFee.GroupId = groupFee.Id

	errC := golfFee.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) GetListGolfFee(c *gin.Context, prof models.CmsUser) {
	form := request.GetListGolfFeeForm{}
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

	golfFeeR := models.GolfFee{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	golfFeeR.Status = form.Status
	list, total, err := golfFeeR.FindList(page)
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

func (_ *CGolfFee) UpdateGolfFee(c *gin.Context, prof models.CmsUser) {
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if golfFee.GroupId != body.GroupId {
		groupFee := models.GroupFee{}
		groupFee.Id = body.GroupId
		errFindGroupFee := groupFee.FindFirst()
		if errFindGroupFee != nil || groupFee.Id <= 0 {
			response_message.BadRequest(c, "group fee not found")
			return
		}
		golfFee.GroupId = groupFee.Id
		golfFee.GroupName = groupFee.Name
	}

	if body.GuestStyle != "" && body.GuestStyle != golfFee.GuestStyle {
		golfFee.GuestStyle = body.GuestStyle
	}
	if body.GuestStyleName != "" {
		golfFee.GuestStyleName = body.GuestStyleName
	}
	if body.Status != "" {
		golfFee.Status = body.Status
	}

	golfFee.Dow = body.Dow
	golfFee.GreenFee = body.GreenFee
	golfFee.CaddieFee = body.CaddieFee
	golfFee.BuggyFee = body.BuggyFee
	golfFee.AccCode = body.AccCode
	golfFee.Note = body.Note
	golfFee.NodeOdd = body.NodeOdd
	golfFee.PaidType = body.PaidType
	golfFee.Idx = body.Idx
	golfFee.AccDebit = body.AccDebit

	isDupli := checkDuplicateGolfFee(golfFee)
	if isDupli {
		response_message.DuplicateRecord(c, "duplicated golf fee")
		return
	}

	errUdp := golfFee.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) DeleteGolfFee(c *gin.Context, prof models.CmsUser) {
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := golfFee.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
