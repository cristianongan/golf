package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"

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

	now := time.Now()

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
			caddieWCNote := models.CaddieWorkingCalendarNote{
				PartnerUid: body.PartnerUid,
				CourseUid:  body.CourseUid,
				ApplyDate:  v.ApplyDate,
				Note:       v.Note,
			}

			if err := caddieWCNote.Create(db); err != nil {
				response_message.BadRequest(c, "Create caddie working calendar note "+err.Error())
				return
			}
		}

		listCreate := []models.CaddieWorkingCalendar{}
		for _, data := range v.CaddieList {
			caddieWC := models.CaddieWorkingCalendar{}
			caddieWC.CreatedAt = now.Unix()
			caddieWC.UpdatedAt = now.Unix()
			caddieWC.Status = constants.STATUS_ENABLE
			caddieWC.PartnerUid = body.PartnerUid
			caddieWC.CourseUid = body.CourseUid
			caddieWC.CaddieCode = data.CaddieCode
			caddieWC.ApplyDate = v.ApplyDate
			caddieWC.NumberOrder = data.NumberOrder
			caddieWC.CaddieIncrease = data.CaddieIncrease
			listCreate = append(listCreate, caddieWC)
		}

		// create
		caddieWCCreate := models.CaddieWorkingCalendar{}
		if err := caddieWCCreate.BatchInsert(db, listCreate); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

	}

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

	caddieWorkingCalendar := models.CaddieWorkingCalendar{}
	caddieWorkingCalendar.CourseUid = body.CourseUid
	caddieWorkingCalendar.PartnerUid = body.PartnerUid
	caddieWorkingCalendar.ApplyDate = body.ApplyDate

	list, total, err := caddieWorkingCalendar.FindAllByDate(db)

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

	if err := caddiWC.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddiWC.CaddieCode = body.CaddieCode

	if err := caddiWC.Update(db); err != nil {
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

	if err := caddiWC.Delete(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}
