package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CLocker struct{}

func (_ *CLocker) GetListLocker(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListLockerForm{}
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

	lockerR := models.Locker{
		PartnerUid:   form.PartnerUid,
		CourseUid:    form.CourseUid,
		Locker:       form.Locker,
		GolfBag:      form.GolfBag,
		LockerStatus: form.LockerStatus,
	}

	if form.PageRequest.Limit == 0 {
		// Lấy full theo ngày hôm nay
		list, total, err := lockerR.FindList(db, page, form.From, form.To, true)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)

		return
	}

	list, total, err := lockerR.FindList(db, page, form.From, form.To, false)
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

func (_ *CLocker) ReturnLocker(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.ReturnLockerReq{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	lockerR := models.Locker{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		Locker:       body.LockerNo,
		BookingDate:  body.BookingDate,
		LockerStatus: constants.LOCKER_STATUS_UNRETURNED,
	}

	errF := lockerR.FindFirst(db)
	if errF != nil {
		response_message.BadRequestDynamicKey(c, "LOCKER_RETURNED", "LOCKER_RETURNED")
		return
	}

	lockerR.LockerStatus = constants.LOCKER_STATUS_RETURNED

	errF = lockerR.Update(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	okResponse(c, lockerR)
}

func (_ *CLocker) CheckLocker(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CheckLockerReq{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	lockerR := models.Locker{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		GolfBag:      body.GolfBag,
		BookingDate:  body.BookingDate,
		LockerStatus: constants.LOCKER_STATUS_UNRETURNED,
	}

	_ = lockerR.FindFirst(db)
	if lockerR.Id > 0 {
		response_message.BadRequestDynamicKey(c, "LOCKER_RETURNED", "LOCKER_RETURNED")
		return
	}

	okRes(c)
}
