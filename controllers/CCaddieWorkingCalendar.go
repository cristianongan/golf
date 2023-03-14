package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CCaddieWorkingCalendar struct{}

func (_ *CCaddieWorkingCalendar) CreateCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieWorkingCalendar BindJSON error", err)
		response_message.BadRequest(c, "")
		return
	}

	now := utils.GetTimeNow()

	for _, v := range body.CaddieWorkingList {
		caddieWC := models.CaddieWorkingCalendar{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			ApplyDate:  v.ApplyDate,
		}

		list, err := caddieWC.FindAll(db)
		if err != nil {
			response_message.BadRequest(c, "Find all caddie working calendar "+err.Error())
			return
		}

		if body.ActionType == "IMPORT" {
			// Kiểm tra ngày truy vấn có dữ liệu hay không
			if len(list) > 0 {
				// Xóa hết dữ liệu ngày truy vấn
				if errD := caddieWC.DeleteBatch(db); errD != nil {
					response_message.BadRequest(c, "Delete batch caddie working calendar "+errD.Error())
					return
				}

				// Update thông note ngày truy vấn
				caddieWCNote := models.CaddieWorkingCalendarNote{
					PartnerUid: body.PartnerUid,
					CourseUid:  body.CourseUid,
					ApplyDate:  v.ApplyDate,
				}

				if err := caddieWCNote.FindFirst(db); err != nil {
					response_message.BadRequest(c, "Find first caddie working calendar note "+err.Error())
					return
				}
				caddieWCNote.Note = v.Note
				if err := caddieWCNote.Update(db); err != nil {
					response_message.BadRequest(c, "Update caddie working calendar note "+err.Error())
					return
				}
			} else {
				// Tạo lưu ý theo ngày truy vấn
				caddieWCNoteCreate := models.CaddieWorkingCalendarNote{
					PartnerUid: body.PartnerUid,
					CourseUid:  body.CourseUid,
					ApplyDate:  v.ApplyDate,
					Note:       v.Note,
				}

				if err := caddieWCNoteCreate.Create(db); err != nil {
					response_message.BadRequest(c, "Create caddie working calendar note "+err.Error())
					return
				}
			}
		}

		listCreate := []models.CaddieWorkingCalendar{}
		listCaddieCode := []string{}

		for _, data := range v.CaddieList {
			caddieWC := models.CaddieWorkingCalendar{}
			caddieWC.CreatedAt = now.Unix()
			caddieWC.UpdatedAt = now.Unix()
			caddieWC.Status = constants.STATUS_ENABLE
			caddieWC.PartnerUid = body.PartnerUid
			caddieWC.CourseUid = body.CourseUid
			caddieWC.CaddieCode = data.CaddieCode
			caddieWC.ApplyDate = v.ApplyDate
			caddieWC.Row = data.Row
			caddieWC.NumberOrder = data.NumberOrder
			caddieWC.CaddieIncrease = data.CaddieIncrease

			// Check duplicate
			if errCheck := caddieWC.IsDuplicated(db); errCheck {
				continue
			}

			listCreate = append(listCreate, caddieWC)
			listCaddieCode = append(listCaddieCode, data.CaddieCode)
		}

		// create
		if len(listCreate) > 0 {
			caddieWCCreate := models.CaddieWorkingCalendar{}
			if err := caddieWCCreate.BatchInsert(db, listCreate); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			//Update lại ds caddie trong GO
			go updateCaddieWorkingOnDay(listCaddieCode, body.PartnerUid, body.CourseUid, true)
		}
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) ImportCaddieSlotAuto(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.ImportCaddieSlotAutoBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("ImportCaddieSlotAuto BindJSON error", err)
		response_message.BadRequest(c, "")
		return
	}

	//Validate caddie slot
	listDup := utils.CheckDupArray(body.CaddieSlot)

	if len(listDup) > 0 {
		response_message.BadRequest(c, "Duplicate caddie: "+strings.Join(listDup, ", "))
		return
	}

	caddieWCS := models.CaddieWorkingSlot{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		ApplyDate:  body.ApplyDate,
	}

	// Xóa dữ liệu ngày truy vấn
	if errD := caddieWCS.DeleteBatch(db); errD != nil {
		response_message.BadRequest(c, "Delete caddie slot auto "+errD.Error())
		return
	}

	caddieWCS.CaddieSlot = body.CaddieSlot

	err := caddieWCS.Create(db)
	if err != nil {
		log.Println("Create report caddie slot err", err.Error())
	}

	//Update lại ds caddie trong GO
	go updateCaddieWorkingOnDay(body.CaddieSlot, body.PartnerUid, body.CourseUid, true)

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) GetCaddieWorkingCalendarList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// TODO: filter by from and to

	body := request.GetCaddieWorkingCalendarList{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//get caddie working slot today
	caddieWorkingCalendar := models.CaddieWorkingSlot{}
	caddieWorkingCalendar.CourseUid = body.CourseUid
	caddieWorkingCalendar.PartnerUid = body.PartnerUid
	caddieWorkingCalendar.ApplyDate = body.ApplyDate

	list, err := caddieWorkingCalendar.Find(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//get caddie increase
	caddieWCI := models.CaddieWorkingCalendar{}
	caddieWCI.CourseUid = body.CourseUid
	caddieWCI.PartnerUid = body.PartnerUid
	caddieWCI.ApplyDate = body.ApplyDate
	caddieWCI.CaddieIncrease = true

	listIncrease, _, err := caddieWCI.FindAllByDate(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// //get note
	// caddieWCNote := models.CaddieWorkingCalendarNote{
	// 	PartnerUid: body.PartnerUid,
	// 	CourseUid:  body.CourseUid,
	// 	ApplyDate:  body.ApplyDate,
	// }

	// listNote, err := caddieWCNote.Find(db)
	// if err != nil {
	// 	response_message.BadRequest(c, "Find first caddie working calendar note "+err.Error())
	// 	return
	// }
	dataCaddieSlot := models.CaddieWorkingSlot{}

	if len(list) > 0 {
		dataCaddieSlot = list[0]
	}

	listRes := map[string]interface{}{
		"data_caddie":          dataCaddieSlot,
		"data_caddie_increase": listIncrease,
		// "note":                 listNote,
	}

	res := map[string]interface{}{
		"data": listRes,
	}

	okResponse(c, res)
}

func (_ *CCaddieWorkingCalendar) GetCaddieWorkingCalendarListNormal(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// TODO: filter by from and to

	body := request.GetCaddieWorkingCalendarList{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//get caddie working slot today
	caddieWC := models.CaddieWorkingCalendar{}
	caddieWC.CourseUid = body.CourseUid
	caddieWC.PartnerUid = body.PartnerUid
	caddieWC.ApplyDate = body.ApplyDate

	list, _, err := caddieWC.FindAllByDate(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//get caddie increase
	caddieWCI := models.CaddieWorkingCalendar{}
	caddieWCI.CourseUid = body.CourseUid
	caddieWCI.PartnerUid = body.PartnerUid
	caddieWCI.ApplyDate = body.ApplyDate
	caddieWCI.CaddieIncrease = true

	listIncrease, _, err := caddieWCI.FindAllByDate(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// //get note
	// caddieWCNote := models.CaddieWorkingCalendarNote{
	// 	PartnerUid: body.PartnerUid,
	// 	CourseUid:  body.CourseUid,
	// 	ApplyDate:  body.ApplyDate,
	// }

	// listNote, err := caddieWCNote.Find(db)
	// if err != nil {
	// 	response_message.BadRequest(c, "Find first caddie working calendar note "+err.Error())
	// 	return
	// }

	listRes := map[string]interface{}{
		"data_caddie":          list,
		"data_caddie_increase": listIncrease,
		// "note":                 listNote,
	}

	res := map[string]interface{}{
		"data": listRes,
	}

	okResponse(c, res)
}

func (_ *CCaddieWorkingCalendar) GetNoteCaddieSlotByDate(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// Validate body
	body := request.GetNoteCaddieSlotByDateForm{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Get note
	caddieWCNote := models.CaddieWorkingCalendarNote{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		ApplyDate:  body.ApplyDate,
	}

	listNote, err := caddieWCNote.Find(db)
	if err != nil {
		response_message.BadRequest(c, "Find caddie working calendar note "+err.Error())
		return
	}

	res := map[string]interface{}{
		"data": listNote,
	}

	okResponse(c, res)
}

func (_ *CCaddieWorkingCalendar) AddNoteCaddieSlotByDate(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate body
	body := request.AddNoteCaddieSlotByDateForm{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//validate note
	caddieWCNote := models.CaddieWorkingCalendarNote{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		ApplyDate:  body.ApplyDate,
	}

	if caddieWCNote.IsDuplicated(db) {
		caddieWCNote.Note = body.Note

		if err := caddieWCNote.Create(db); err != nil {
			response_message.BadRequest(c, "Create caddie working calendar note "+err.Error())
			return
		}
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) UpdateNoteCaddieSlotByDate(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate
	caddeNoteIdStr := c.Param("id")
	caddeWCId, err := strconv.ParseInt(caddeNoteIdStr, 10, 64)
	if err != nil || caddeWCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	var body request.UpdateNoteCaddieSlotByDateForm
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingCalendar BindJSON error")
		response_message.BadRequest(c, "")
	}

	//Find caddie note
	// Update thông note theo ngày
	caddieWCNote := models.CaddieWorkingCalendarNote{}

	caddieWCNote.Id = caddeWCId

	if err := caddieWCNote.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find first caddie working calendar note "+err.Error())
		return
	}

	caddieWCNote.Note = body.Note
	if err := caddieWCNote.Update(db); err != nil {
		response_message.BadRequest(c, "Update caddie working calendar note "+err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) UpdateCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	caddeWCIdStr := c.Param("id")
	caddeWCId, err := strconv.ParseInt(caddeWCIdStr, 10, 64)
	if err != nil || caddeWCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	var body request.UpdateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingCalendar BindJSON error")
		response_message.BadRequest(c, "")
	}

	caddiWC := models.CaddieWorkingCalendar{}
	caddiWC.Id = caddeWCId
	caddiWC.PartnerUid = prof.PartnerUid
	caddiWC.CourseUid = prof.CourseUid

	if err := caddiWC.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	oldCaddie := caddiWC.CaddieCode
	newCaddie := body.CaddieCode
	caddiWC.CaddieCode = body.CaddieCode

	if err := caddiWC.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//Update lại ds caddie trong GO
	go func() {
		updateCaddieWorkingOnDay([]string{oldCaddie}, prof.PartnerUid, prof.CourseUid, false)
		updateCaddieWorkingOnDay([]string{newCaddie}, prof.PartnerUid, prof.CourseUid, true)
	}()
	okRes(c)
}

func (_ *CCaddieWorkingCalendar) UpdateCaddieSlotAuto(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	// validate body
	var body request.UpdateCaddieWorkingSlotAutoBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingSlotAutoBody BindJSON error")
		response_message.BadRequest(c, "")
	}

	// Find slot caddie with date
	caddieWS := models.CaddieWorkingSlot{}
	caddieWS.PartnerUid = body.PartnerUid
	caddieWS.CourseUid = body.CourseUid
	caddieWS.ApplyDate = body.ApplyDate

	if err := caddieWS.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Swap slot caddie
	caddieWS.CaddieSlot = utils.SwapValue(caddieWS.CaddieSlot, body.CaddieCodeOld, body.CaddieCodeNew)

	// Update slot caddie
	if err := caddieWS.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) DeleteCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	caddeWCIdStr := c.Param("id")
	caddeWCId, err := strconv.ParseInt(caddeWCIdStr, 10, 64)
	if err != nil || caddeWCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	caddiWC := models.CaddieWorkingCalendar{}
	caddiWC.Id = caddeWCId
	caddiWC.PartnerUid = prof.PartnerUid
	caddiWC.CourseUid = prof.CourseUid

	if err := caddiWC.Delete(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//Update lại ds caddie trong GO
	go updateCaddieWorkingOnDay([]string{caddiWC.CaddieCode}, prof.PartnerUid, prof.CourseUid, false)
	okRes(c)
}
