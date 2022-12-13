package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CGolfFee struct{}

func (_ *CGolfFee) CreateGolfFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check Exits
	isDupli := checkDuplicateGolfFee(db, body, false)
	if isDupli {
		response_message.DuplicateRecord(c, "duplicated golf fee")
		return
	}

	// Check Table Price Exit
	tablePrice := models.TablePrice{}
	tablePrice.Id = body.TablePriceId
	errFind := tablePrice.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, "table price not found")
		return
	}

	// Check group Fee
	groupFee := models.GroupFee{}
	groupFee.Id = body.GroupId
	errFind = groupFee.FindFirst(db)
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
	golfFee.CustomerCategory = getCustomerCategoryFromCustomerType(db, body.CustomerType)
	golfFee.GroupName = body.GroupName
	golfFee.GroupId = groupFee.Id
	golfFee.UpdateUserName = prof.UserName
	golfFee.ApplyTime = strings.TrimSpace(body.ApplyTime)
	golfFee.TaxCode = body.TaxCode

	errC := golfFee.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) GetListGolfFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		PartnerUid:     form.PartnerUid,
		CourseUid:      form.CourseUid,
		TablePriceId:   form.TablePriceId,
		GroupId:        form.GroupId,
		GuestStyle:     form.GuestStyle,
		GuestStyleName: form.GuestStyleName,
	}
	golfFeeR.Status = form.Status
	list, total, err := golfFeeR.FindList(db, page, form.IsToday)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if golfFee.Dow != body.Dow || golfFee.GuestStyle != body.GuestStyle {
		//Get list GS
		//Check theo chi tiết ngày
		listTempR := models.GolfFee{
			PartnerUid:   body.PartnerUid,
			CourseUid:    body.CourseUid,
			GuestStyle:   body.GuestStyle,
			TablePriceId: body.TablePriceId,
		}
		listTemp := listTempR.GetGuestStyleGolfFeeByGuestStyle(db)
		// Cho trường hợp sửa golf fee có 1 row
		if len(listTemp) > 1 {
			isDupli := checkDuplicateGolfFee(db, body, true)
			if isDupli {
				response_message.DuplicateRecord(c, "duplicated golf fee")
				return
			}
		}
	}

	if golfFee.GroupId != body.GroupId {
		groupFee := models.GroupFee{}
		groupFee.Id = body.GroupId
		errFindGroupFee := groupFee.FindFirst(db)
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
	if body.CustomerType != "" {
		golfFee.CustomerType = body.CustomerType
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
	golfFee.ApplyTime = strings.TrimSpace(body.ApplyTime)
	golfFee.UpdateUserName = prof.UserName
	golfFee.TaxCode = body.TaxCode

	errUdp := golfFee.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) DeleteGolfFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := golfFee.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CGolfFee) GetListGuestStyle(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListGolfFeeForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" {
		response_message.BadRequest(c, "partner uid invalid")
		return
	}

	// Lấy table Price hợp lệ
	tablePriceR := models.TablePrice{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	tablePrice, err := tablePriceR.FindCurrentUse(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	golfFeeR := models.GolfFee{
		PartnerUid:       form.PartnerUid,
		CourseUid:        form.CourseUid,
		TablePriceId:     tablePrice.Id,
		CustomerType:     form.CustomerType,
		CustomerCategory: form.CustomerCategory,
	}
	golfFeeR.Status = constants.STATUS_ENABLE
	guestStyles := golfFeeR.GetGuestStyleList(db)
	okResponse(c, guestStyles)
}

func (_ *CGolfFee) GetGolfFeeByGuestStyle(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListGolfFeeByGuestStyleForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" || form.GuestStyle == "" {
		response_message.BadRequest(c, "data invalid")
		return
	}

	// Lấy table Price hợp lệ
	tablePriceR := models.TablePrice{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	tablePrice, err := tablePriceR.FindCurrentUse(db)
	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	golfFeeR := models.GolfFee{
		PartnerUid:   form.PartnerUid,
		CourseUid:    form.CourseUid,
		TablePriceId: tablePrice.Id,
		GuestStyle:   form.GuestStyle,
	}
	guestStyles := golfFeeR.GetGuestStyleGolfFeeByGuestStyle(db)

	for i, v := range guestStyles {
		groupGS := models.GroupFee{}
		groupGS.Id = v.GroupId
		errFGGS := groupGS.FindFirst(db)
		if errFGGS != nil {
			log.Println("GetGolfFeeByGuestStyle", errFGGS.Error())
		} else {
			guestStyles[i].GroupName = groupGS.Name
		}
	}

	okResponse(c, guestStyles)
}
