package controllers

import (
	"encoding/json"
	"fmt"
	"start/callservices"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
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

func (_ *CNotification) Admin1ApproveCaddieVacation(c *gin.Context, prof models.CmsUser) {
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

	caddieEx := models.CaddieContentNoti{}
	if err := json.Unmarshal(notification.Content, &caddieEx); err != nil {
		return
	}

	// Update lại trạng thái noti của admin1
	if *form.IsApprove {
		notification.NotificationStatus = constants.NOTIFICATION_APPROVED
		notification.UserApprove = prof.UserName
		notification.DateApproved = utils.GetTimeNow().Unix()
		if errUpdNotification := notification.Update(db); errUpdNotification != nil {
			response_message.InternalServerError(c, errUpdNotification.Error())
			return
		}
		//

		// Tạo noti cho admin2
		notificationAd2 := models.Notification{
			PartnerUid:         notification.PartnerUid,
			CourseUid:          notification.CourseUid,
			Type:               notification.Type,
			Title:              notification.Title,
			NotificationStatus: constants.NOTIFICATION_PENDIND,
			UserCreate:         prof.UserName,
			Content:            notification.Content,
			Role:               constants.NOTIFICATION_CHANNEL_ADMIN_2,
		}

		if errNotification := notificationAd2.Create(db); errNotification != nil {
			response_message.InternalServerError(c, errNotification.Error())
			return
		}

		// parse
		var mess map[string]interface{}
		inrec, _ := json.Marshal(notificationAd2)
		json.Unmarshal(inrec, &mess)

		// push mess socket
		reqSocket := request.MessSocketBody{
			Data: mess,
			Room: constants.NOTIFICATION_CHANNEL_ADMIN_2,
		}

		go callservices.PushMessInSocket(reqSocket)
	} else {
		notification.NotificationStatus = constants.NOTIFICATION_REJECTED
		notification.UserApprove = prof.UserName
		notification.DateApproved = utils.GetTimeNow().Unix()
		if errUpdNotification := notification.Update(db); errUpdNotification != nil {
			response_message.InternalServerError(c, errUpdNotification.Error())
			return
		}

		approvedTitle := ""
		if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_SICK_OFF {
			approvedTitle = "không được duyệt nghỉ phép ốm"
		} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_UNPAID {
			approvedTitle = "không được duyệt nghỉ phép không lương"
		} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_MATERNITY_LEAD {
			approvedTitle = "không được duyệt nghỉ thai sản"
		} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_JOB {
			approvedTitle = "không được duyệt nghỉ đi công tác"
		} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_ANNUAL_LEAVE {
			approvedTitle = "không được duyệt nghỉ thường niên"
		}

		newNotification := models.Notification{}
		newNotification.Title = fmt.Sprintln("Caddie", caddieEx.Code, approvedTitle, caddieEx.NumberDayOff, "ngày", "từ", caddieEx.FromDay, "đến", caddieEx.ToDay)
		newNotification.UserCreate = prof.UserName
		newNotification.Note = form.Note
		newNotification.Role = constants.NOTIFICATION_CHANNEL_CADDIE_MASTER
		newNotification.Type = constants.NOTIFICATION_CADDIE_VACATION_CONFIRM

		if errNotification := newNotification.Create(db); errNotification != nil {
			response_message.InternalServerError(c, errNotification.Error())
			return
		}

		// parse
		var mess map[string]interface{}
		inrec, _ := json.Marshal(newNotification)
		json.Unmarshal(inrec, &mess)

		// push mess socket
		reqSocket := request.MessSocketBody{
			Data: mess,
			Room: constants.NOTIFICATION_CHANNEL_CADDIE_MASTER,
		}

		go callservices.PushMessInSocket(reqSocket)
	}

	okRes(c)
}

func (_ *CNotification) Admin2ApproveCaddieVacation(c *gin.Context, prof models.CmsUser) {
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

	caddieEx := models.CaddieContentNoti{}
	if err := json.Unmarshal(notification.Content, &caddieEx); err != nil {
		return
	}

	approvedTitle := ""
	if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_SICK_OFF {
		if *form.IsApprove {
			approvedTitle = "được duyệt nghỉ phép ốm"
		} else {
			approvedTitle = "không được duyệt nghỉ phép ốm"
		}
	} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_UNPAID {
		if *form.IsApprove {
			approvedTitle = "được duyệt nghỉ phép không lương"
		} else {
			approvedTitle = "không được duyệt nghỉ phép không lương"
		}
	} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_MATERNITY_LEAD {
		if *form.IsApprove {
			approvedTitle = "được duyệt nghỉ thai sản"
		} else {
			approvedTitle = "không được duyệt nghỉ thai sản"
		}
	} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_JOB {
		if *form.IsApprove {
			approvedTitle = "được duyệt nghỉ đi công tác"
		} else {
			approvedTitle = "không được duyệt nghỉ đi công tác"
		}
	} else if notification.Type == constants.NOTIFICATION_CADDIE_VACATION_ANNUAL_LEAVE {
		if *form.IsApprove {
			approvedTitle = "được duyệt nghỉ thường niên"
		} else {
			approvedTitle = "không được duyệt nghỉ thường niên"
		}
	}

	newNotification := models.Notification{}
	newNotification.Title = fmt.Sprintln("Caddie", caddieEx.Code, approvedTitle, caddieEx.NumberDayOff, "ngày", "từ", caddieEx.FromDay, "đến", caddieEx.ToDay)
	newNotification.UserCreate = prof.UserName
	newNotification.Note = form.Note
	newNotification.Role = constants.NOTIFICATION_CHANNEL_ADMIN_2
	newNotification.Type = constants.NOTIFICATION_CADDIE_VACATION_CONFIRM

	if errNotification := newNotification.Create(db); errNotification != nil {
		response_message.InternalServerError(c, errNotification.Error())
		return
	}

	if *form.IsApprove {
		notification.NotificationStatus = constants.NOTIFICATION_APPROVED
	} else {
		notification.NotificationStatus = constants.NOTIFICATION_REJECTED
	}
	notification.UserApprove = prof.UserName
	notification.DateApproved = utils.GetTimeNow().Unix()
	if errUpdNotification := notification.Update(db); errUpdNotification != nil {
		response_message.InternalServerError(c, errUpdNotification.Error())
		return
	}

	cCaddieVacation := CCaddieVacationCalendar{}
	go cCaddieVacation.UpdateCaddieVacationStatus(notification.Content, *form.IsApprove, notification.PartnerUid, prof)

	// parse
	var mess map[string]interface{}
	inrec, _ := json.Marshal(newNotification)
	json.Unmarshal(inrec, &mess)

	// push mess socket
	reqSocket1 := request.MessSocketBody{
		Data: mess,
		Room: constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}

	reqSocket2 := request.MessSocketBody{
		Data: mess,
		Room: constants.NOTIFICATION_CHANNEL_CADDIE_MASTER,
	}

	go callservices.PushMessInSocket(reqSocket1)
	go callservices.PushMessInSocket(reqSocket2)
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
	} else if body.Title == constants.NOTIFICATION_CADDIE_VACATION_MATERNITY_LEAD {
		notiType = constants.NOTIFICATION_CADDIE_VACATION_MATERNITY_LEAD
		extraTitle = "xin nghỉ thai sản"
	} else if body.Title == constants.CADDIE_VACATION_JOB {
		notiType = constants.CADDIE_VACATION_JOB
		extraTitle = "xin nghỉ đi công tác"
	} else if body.Title == constants.NOTIFICATION_CADDIE_VACATION_ANNUAL_LEAVE {
		notiType = constants.NOTIFICATION_CADDIE_VACATION_ANNUAL_LEAVE
		extraTitle = "xin nghỉ nghỉ thường niên"
	}

	fromDay, _ := utils.GetDateFromTimestampWithFormat(body.DateFrom, constants.DATE_FORMAT_1)
	toDay, _ := utils.GetDateFromTimestampWithFormat(body.DateTo, constants.DATE_FORMAT_1)
	title := fmt.Sprintln("Caddie", body.Caddie.Code, extraTitle, body.NumberDayOff, "ngày", "từ", fromDay, "đến", toDay)
	extraInfo := models.CaddieContentNoti{
		Id:           body.Id,
		Code:         body.Caddie.Code,
		Type:         constants.NOTIFICATION_CADDIE_VACATION,
		NumberDayOff: body.NumberDayOff,
		FromDay:      fromDay,
		ToDay:        toDay,
	}

	datas, _ := json.Marshal(extraInfo)
	notiData := models.Notification{
		PartnerUid:         body.PartnerUid,
		CourseUid:          body.CourseUid,
		Type:               notiType,
		Title:              title,
		NotificationStatus: constants.NOTIFICATION_PENDIND,
		UserCreate:         body.UserName,
		Content:            datas,
		Role:               constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}

	notiData.Create(db)

	// parse
	var mess map[string]interface{}
	inrec, _ := json.Marshal(notiData)
	json.Unmarshal(inrec, &mess)

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: mess,
		Room: constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) CreateCaddieWorkingStatusNotification(title string) {
	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_CADDIE_WORKING_STATUS_UPDATE,
		"title": title,
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: constants.NOTIFICATION_CHANNEL_CADDIE_MASTER,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) PushNotificationCreateBooking(bookType string, booking any) {
	notiData := map[string]interface{}{
		"type":  bookType,
		"title": "",
		// "booking": booking,
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: constants.NOTIFICATION_CHANNEL_BOOKING,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) PushNotificationLockTee(lockType string) {
	notiData := map[string]interface{}{
		"type":  lockType,
		"title": "",
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: constants.NOTIFICATION_CHANNEL_BOOKING,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) CreateCaddieVacation(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddie := models.Caddie{}

	caddie.Id = 20
	if err := caddie.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}
	cNotification := CNotification{}
	cNotification.CreateCaddieVacationNotification(db, request.GetCaddieVacationNotification{
		Caddie:       caddie,
		DateFrom:     1677920680,
		DateTo:       1678093480,
		NumberDayOff: 2,
		Title:        constants.NOTIFICATION_CADDIE_VACATION_ANNUAL_LEAVE,
		CreateAt:     1678093480,
		UserName:     prof.UserName,
		Id:           caddie.Id,
	})

	okRes(c)
}

func (_ *CNotification) PushMessPOSForApp(bill models.ServiceCart) {
	// Data push
	reqSocket := request.MessSocketBody{}
	// Data mess
	notiData := map[string]interface{}{
		"bill_id": bill.Id,
	}

	if bill.ServiceType == constants.RESTAURANT_SETTING {
		notiData["type"] = constants.NOTIFICATION_RESTAURANT_BILL_UPDATE
		reqSocket.Room = constants.NOTIFICATION_CHANNEL_RESTAURANT
	}

	if bill.ServiceType == constants.KIOSK_SETTING {
		notiData["type"] = constants.NOTIFICATION_KIOSK_BILL_UPDATE
		reqSocket.Room = constants.NOTIFICATION_CHANNEL_KIOSK
	}

	// push mess socket
	reqSocket.Data = notiData

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) PushMessKIForApp() {
	// Data mess
	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_KIOSK_INVENTORY_UPDATE,
		"title": "",
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: constants.NOTIFICATION_CHANNEL_KIOSK,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) CreateCWCNotification(db *gorm.DB, prof models.CmsUser, applayDate string, caddieCode []string) {
	// noti
	title := fmt.Sprintln("Caddie xin tăng cường ngày ", applayDate, ": ", strings.Join(caddieCode, ", "))

	extraInfo := models.CaddieWCINoti{
		Caddies:   caddieCode,
		ApplyDate: applayDate,
	}

	datas, _ := json.Marshal(extraInfo)

	notiData := models.Notification{
		PartnerUid:         prof.PartnerUid,
		CourseUid:          prof.CourseUid,
		Type:               constants.NOTIFICATION_ADD_CADDIE_WORKING_CALENDAR,
		Title:              title,
		NotificationStatus: constants.NOTIFICATION_PENDIND,
		UserCreate:         prof.UserName,
		Content:            datas,
		Role:               constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}

	notiData.Create(db)

	// parse
	var mess map[string]interface{}
	inrec, _ := json.Marshal(notiData)
	json.Unmarshal(inrec, &mess)

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: mess,
		Room: constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}

	go callservices.PushMessInSocket(reqSocket)
}

func (_ *CNotification) PushMessBoookingForApp(typeMess string, bag *model_booking.Booking) {
	// Data mess
	notiData := map[string]interface{}{
		"type": typeMess,
		"data": bag,
	}

	// Check date
	dateNow, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	if dateNow == bag.BookingDate {
		// push mess socket
		reqSocket := request.MessSocketBody{
			Data: notiData,
			Room: constants.NOTIFICATION_CHANNEL_BOOKING_APP,
		}

		go callservices.PushMessInSocket(reqSocket)
	}

}

func (_ *CNotification) Admin1ApproveCaddieWC(c *gin.Context, prof models.CmsUser) {
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

	// Title noti
	title := ""

	//
	data := models.CaddieWCINoti{}
	if err := json.Unmarshal(notification.Content, &data); err != nil {
		return
	}

	// Update lại trạng thái noti của admin1
	if *form.IsApprove {
		notification.NotificationStatus = constants.NOTIFICATION_APPROVED
		notification.UserApprove = prof.UserName
		notification.DateApproved = utils.GetTimeNow().Unix()
		if errUpdNotification := notification.Update(db); errUpdNotification != nil {
			response_message.InternalServerError(c, errUpdNotification.Error())
			return
		}

		caddieWCI := models.CaddieWorkingCalendar{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			ApplyDate:  data.ApplyDate,
		}

		if errUpd := caddieWCI.UpdateBatchCaddieCode(db, data.Caddies, prof.UserName); errUpd != nil {
			response_message.InternalServerError(c, errUpd.Error())
			return
		}

		title = fmt.Sprintln(notification.Title, " đã được duyệt.")
	} else {
		notification.NotificationStatus = constants.NOTIFICATION_REJECTED
		notification.UserApprove = prof.UserName
		notification.DateApproved = utils.GetTimeNow().Unix()
		if errUpdNotification := notification.Update(db); errUpdNotification != nil {
			response_message.InternalServerError(c, errUpdNotification.Error())
			return
		}

		caddieWCI := models.CaddieWorkingCalendar{
			PartnerUid: prof.PartnerUid,
			CourseUid:  prof.CourseUid,
			ApplyDate:  data.ApplyDate,
		}

		if errDel := caddieWCI.DeleteBatchCaddies(db, data.Caddies); errDel != nil {
			response_message.InternalServerError(c, errDel.Error())
			return
		}

		title = fmt.Sprintln(notification.Title, " không được duyệt.")
	}

	newNotification := models.Notification{
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
		Type:       constants.NOTIFICATION_ADD_CADDIE_WORKING_CALENDAR_CONFIRM,
		Title:      title,
		UserCreate: prof.UserName,
		Role:       constants.NOTIFICATION_CHANNEL_CADDIE_MASTER,
		Note:       form.Note,
	}

	// parse
	var mess map[string]interface{}
	inrec, _ := json.Marshal(newNotification)
	json.Unmarshal(inrec, &mess)

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: mess,
		Room: constants.NOTIFICATION_CHANNEL_CADDIE_MASTER,
	}

	go callservices.PushMessInSocket(reqSocket)

	okRes(c)
}
