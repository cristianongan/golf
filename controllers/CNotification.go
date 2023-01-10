package controllers

import (
	"encoding/json"
	"fmt"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/socket"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CNotification struct{}

func (_ *CNotification) GetListNotification(c *gin.Context, prof models.CmsUser) {
	form := request.GetListNotificationForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(form.PartnerUid)
	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	notificationR := models.Notification{}
	list, total, err := notificationR.FindList(db, page)
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

func (_ *CNotification) DeleteNotification(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, errId := strconv.ParseInt(IdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	notification := models.Notification{}
	notification.Id = Id
	errF := notification.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := notification.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CNotification) SeenNotification(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idStr := c.Param("id")
	Id, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	notification := models.Notification{}
	notification.Id = Id
	errF := notification.FindFirst(db)
	if errF != nil {
		response_message.BadRequestDynamicKey(c, "NOTI_NOT_FOUND", "")
		return
	}

	notification.IsRead = setBoolForCursor(true)
	if errUpd := notification.Update(db); errUpd != nil {
		response_message.BadRequest(c, errUpd.Error())
		return
	}

	okRes(c)
}

func (_ *CNotification) ApproveCaddieCalendarNotification(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idStr := c.Param("id")
	Id, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	form := request.ApproveCaddieCalendarNotification{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	notification := models.Notification{}
	notification.Id = Id
	errF := notification.FindFirst(db)
	if errF != nil {
		response_message.BadRequestDynamicKey(c, "NOTI_NOT_FOUND", "")
		return
	}

	extraTitle := ""
	approvedTitle := ""
	if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_SICK_OFF {
		extraTitle = "xin nghỉ phép ốm"
		if form.IsApprove {
			approvedTitle = "được duyệt nghỉ phép ốm"
		} else {
			approvedTitle = "được duyệt nghỉ phép ốm"
		}
	} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_UNPAID {
		extraTitle = "xin nghỉ phép không lương"
		if form.IsApprove {
			approvedTitle = "được duyệt nghỉ phép không lương"
		} else {
			approvedTitle = "không được duyệt nghỉ phép không lương"
		}
	}

	newNotification := models.Notification{}
	newNotification.Title = strings.Replace(notification.Title, extraTitle, approvedTitle, 1)
	newNotification.UserCreate = prof.UserName
	newNotification.Note = form.Note
	newNotification.Type = notification.Type

	if errNotification := newNotification.Create(db); errNotification != nil {
		response_message.InternalServerError(c, errNotification.Error())
		return
	}

	if form.IsApprove {
		notification.NotificationStatus = constants.NOTIFICATION_APPROVED
	} else {
		notification.NotificationStatus = constants.NOTIFICATION_REJECTED
	}

	if errUpdNotification := notification.Update(db); errUpdNotification != nil {
		response_message.InternalServerError(c, errUpdNotification.Error())
		return
	}

	cCaddieVacation := CCaddieVacationCalendar{}
	go cCaddieVacation.UpdateCaddieVacationStatus(notification.ExtraInfo.Id, form.IsApprove, notification.PartnerUid, prof)

	newFsConfigBytes, _ := json.Marshal(newNotification)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
	okRes(c)
}

func (_ *CNotification) CreateCaddieVacationNotification(db *gorm.DB, body request.GetCaddieVacationNotification) {
	notiType := ""
	extraTitle := ""
	if body.Title == constants.CADDIE_VACATION_SICK {
		notiType = constants.NOTIFICATION_CADDIE_VACATION_SICK_OFF
		extraTitle = "xin nghỉ phép ốm"
	} else if body.Title == constants.CADDIE_VACATION_UNPAID {
		notiType = constants.NOTIFICATION_CADDIE_VACATION_UNPAID
		extraTitle = "xin nghỉ phép không lương"
	}

	fromDay, _ := utils.GetDateFromTimestampWithFormat(body.DateFrom, constants.DATE_FORMAT_1)
	toDay, _ := utils.GetDateFromTimestampWithFormat(body.DateTo, constants.DATE_FORMAT_1)
	hourStr, _ := utils.GetDateFromTimestampWithFormat(body.CreateAt, constants.HOUR_FORMAT)
	title := fmt.Sprintln("Caddie", body.Caddie.Code, extraTitle, body.NumberDayOff, "ngày", "từ", fromDay, "đến", toDay, ",", hourStr)
	extraInfo := models.ExtraInfo{
		Id: body.Id,
	}

	notiData := models.Notification{
		PartnerUid:         body.PartnerUid,
		CourseUid:          body.CourseUid,
		Type:               notiType,
		Title:              title,
		NotificationStatus: constants.NOTIFICATION_PENDIND,
		UserCreate:         body.UserName,
		ExtraInfo:          extraInfo,
	}

	notiData.Create(db)

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}

func (_ *CNotification) CreateCaddieWorkingStatusNotification(title string) {
	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_CADDIE_WORKING_STATUS_UPDATE,
		"title": title,
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}

func (_ *CNotification) PushNotificationCreateBooking(bookType string, booking any) {
	notiData := map[string]interface{}{
		"type":  bookType,
		"title": "",
		// "booking": booking,
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}

func (_ *CNotification) PushNotificationLockTee(lockType string) {
	notiData := map[string]interface{}{
		"type":  lockType,
		"title": "",
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}
