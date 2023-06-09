package controllers

import (
	"encoding/json"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/logger"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBuggy struct{}

func (_ *CBuggy) CreateBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateBuggyBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, partnerErrFind.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.CourseUid = body.CourseUid
	buggyRequest.PartnerUid = body.PartnerUid
	buggyRequest.Code = body.Code // Id Buggy vận hành
	errExist := buggyRequest.FindFirst(db)

	if errExist == nil && buggyRequest.ModelId.Id > 0 {
		response_message.BadRequest(c, "Code existed in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	buggy := models.Buggy{
		ModelId:         base,
		Code:            body.Code,
		CourseUid:       body.CourseUid,
		PartnerUid:      body.PartnerUid,
		Origin:          body.Origin,
		Note:            body.Note,
		BuggyForVip:     body.BuggyForVip,
		MaintenanceFrom: body.MaintenanceFrom,
		MaintenanceTo:   body.MaintenanceTo,
		BuggyStatus:     body.BuggyStatus,
	}

	if buggy.BuggyStatus == "" {
		buggy.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
	}

	err := buggy.Create(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, buggy)
}

func (_ *CBuggy) GetBuggyList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBuggyForm{}
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

	buggyRequest := models.Buggy{}
	buggyRequest.CourseUid = form.CourseUid
	buggyRequest.PartnerUid = form.PartnerUid
	buggyRequest.BuggyStatus = form.BuggyStatus
	buggyRequest.BuggyForVip = form.BuggyForVip
	buggyRequest.Code = form.Code

	list, total, err := buggyRequest.FindList(db, page, form.IsReady)

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

func (_ *CBuggy) GetBuggyDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	buggyIdStr := c.Param("id")
	buggyId, errId := strconv.ParseInt(buggyIdStr, 10, 64)

	if buggyIdStr == "" {
		response_message.BadRequest(c, errId.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.Id = buggyId
	errF := buggyRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if buggyRequest.PartnerUid != prof.PartnerUid || buggyRequest.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
		return
	}

	okResponse(c, buggyRequest)
}

func (_ *CBuggy) GetBuggyReadyList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBuggyForm{}
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

	buggyRequest := models.BuggyRequest{}
	buggyRequest.CourseUid = form.CourseUid
	buggyRequest.PartnerUid = form.PartnerUid
	buggyRequest.FunctionType = form.FunctionType
	buggyRequest.Code = form.Code

	listBuggyReady, total, err := buggyRequest.FindBuggyReadyList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	result := []models.Buggy{}
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	for _, buggy := range listBuggyReady {
		if buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_IN_COURSE ||
			buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_LOCK {
			bookingList := model_booking.BookingList{
				PartnerUid:  form.PartnerUid,
				CourseUid:   form.CourseUid,
				BuggyCode:   buggy.Code,
				BookingDate: dateDisplay,
			}
			if buggyRequest.FunctionType == constants.BAG_STATUS_IN_COURSE {
				bookingList.BagStatus = constants.BAG_STATUS_IN_COURSE
			}

			if buggyRequest.FunctionType == constants.BAG_STATUS_WAITING {
				bookingList.BagStatus = constants.BAG_STATUS_WAITING
			}

			listBooking := []model_booking.Booking{}
			db, total, _ := bookingList.FindAllBookingList(db)
			db.Find(&listBooking)
			if total < 2 {
				if total == 1 {
					if !*listBooking[0].IsPrivateBuggy {
						result = append(result, buggy)
					}
				} else {
					result = append(result, buggy)
				}
			}
		} else {
			result = append(result, buggy)
		}
	}

	res := response.PageResponse{
		Total: total,
		Data:  result,
	}

	c.JSON(200, res)
}

func (_ *CBuggy) DeleteBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	buggyIdStr := c.Param("id")
	buggyId, errId := strconv.ParseInt(buggyIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.Id = buggyId
	buggyRequest.PartnerUid = prof.PartnerUid
	buggyRequest.CourseUid = prof.CourseUid
	errF := buggyRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := buggyRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBuggy) UpdateBuggy(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	buggyIdStr := c.Param("id")
	buggyId, errId := strconv.ParseInt(buggyIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateBuggyBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.Id = buggyId
	buggyRequest.PartnerUid = prof.PartnerUid
	buggyRequest.CourseUid = prof.CourseUid

	errF := buggyRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.CourseUid != nil {
		courseRequest := models.Course{}
		courseRequest.Uid = *body.CourseUid
		errFind := courseRequest.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, "course_uid not found")
			return
		}
	}

	if body.PartnerUid != nil {
		partnerRequest := models.Partner{}
		partnerRequest.Uid = *body.PartnerUid
		partnerErrFind := partnerRequest.FindFirst()
		if partnerErrFind != nil {
			response_message.BadRequest(c, "partner_uid not found")
			return
		}
	}
	// activity log
	updateLogData := logger.UpdateLogData{}
	updateLogData.Old = buggyRequest

	// + update data
	if body.CourseUid != nil {
		buggyRequest.CourseUid = *body.CourseUid
	}
	if body.PartnerUid != nil {
		buggyRequest.PartnerUid = *body.PartnerUid
	}
	if body.Origin != nil {
		buggyRequest.Origin = *body.Origin
	}
	if body.Note != nil {
		buggyRequest.Note = *body.Note
	}
	if body.BuggyStatus != nil {
		buggyRequest.BuggyStatus = *body.BuggyStatus
	}
	if body.BuggyForVip != nil {
		buggyRequest.BuggyForVip = *body.BuggyForVip
	}
	if body.MaintenanceFrom != nil {
		buggyRequest.MaintenanceFrom = *body.MaintenanceFrom
	}
	if body.MaintenanceTo != nil {
		buggyRequest.MaintenanceTo = *body.MaintenanceTo
	}
	if body.WarrantyPeriod != nil {
		buggyRequest.WarrantyPeriod = *body.WarrantyPeriod
	}
	// + END: update data

	err := buggyRequest.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	updateLogData.New = buggyRequest

	updateLogDataJson, _ := json.Marshal(updateLogData)

	logger.Log(db, logger.EVENT_ACTION_UPDATE, logger.EVENT_CATEOGRY_BUGGY, buggyRequest.Code, string(updateLogDataJson), prof)

	okRes(c)
}
