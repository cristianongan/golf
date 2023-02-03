package controllers

import (
	"fmt"
	"log"
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CLockTeeTime struct{}

func (_ *CLockTeeTime) CreateTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTeeTimeSettings{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	teeTimeRedisKey := getKeyTeeTimeLockRedis(body.DateTime, body.CourseUid, body.TeeTime, body.TeeType)

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       body.DateTime,
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		TeeTime:        body.TeeTime,
		CurrentTeeTime: body.TeeTime,
		TeeType:        body.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Type:           constants.LOCK_CMS,
		Slot:           4,
		Note:           body.Note,
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {

		}
	}

	okResponse(c, teeTimeRedis)
}
func (_ *CLockTeeTime) GetTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	query := request.GetListTeeTimeSettings{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	teeTimeSetting := models.LockTeeTime{}
	teeTimeSetting.TeeTime = query.TeeTime
	teeTimeSetting.TeeTimeStatus = query.TeeTimeStatus
	teeTimeSetting.DateTime = query.DateTime
	list := []models.LockTeeTimeWithSlot{}

	// get các teetime đang bị khóa ở redis
	listTeeTimeLockRedis := getTeeTimeLockRedis(query.CourseUid, query.DateTime, "")

	if query.RequestType == "TURN_TIME" {
		for _, teeTime := range listTeeTimeLockRedis {
			if teeTime.CurrentTeeTime != teeTime.TeeTime {
				list = append(list, teeTime)
			}
		}
	}

	if query.RequestType == "TEE_TIME" {
		for _, teeTime := range listTeeTimeLockRedis {
			if teeTime.CurrentTeeTime == teeTime.TeeTime || teeTime.TeeTime == "" {
				list = append(list, teeTime)
			}
		}
	}

	res := response.PageResponse{
		Total: int64(len(list)),
		Data:  list,
	}

	c.JSON(200, res)
}
func (_ *CLockTeeTime) LockTurn(body request.CreateLockTurn, hole int, c *gin.Context, prof models.CmsUser) error {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	form := request.GetListBookingSettingForm{
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
		OnDate:     body.BookingDate,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	bookingDateTime, _ := time.Parse(constants.DATE_FORMAT_1, body.BookingDate)
	weekday := strconv.Itoa(int(bookingDateTime.Weekday() + 1))

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = body.BookingDate
	bookings.TeeTime = body.TeeTime
	bookings.TeeType = body.TeeType
	bookings.CourseType = body.CourseType

	dbB := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	dbB2, _, _ := bookings.FindAllBookingList(dbB)
	dbB2 = dbB2.Where("bag_status <> ?", constants.BAG_STATUS_CANCEL)
	listBooking := []model_booking.Booking{}
	dbB2.Find(&listBooking)

	countHole18 := 0
	countHole27 := 0

	for _, booking := range listBooking {
		if booking.Hole == 18 {
			countHole18 += 1
		}
		if booking.Hole == 27 {
			countHole27 += 1
		}
	}

	log.Println("LockTurn-weekday:", weekday)
	turnTimeH := 2
	bookSetting := model_booking.BookingSetting{}

	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	currentTeeTimeDate, _ := utils.ConvertHourToTime(body.TeeTime)
	teeList := []string{}
	teeType := fmt.Sprint(body.TeeType, body.CourseType)

	if countHole18 >= countHole27 {

		if teeType == "1A" {
			teeList = []string{"1B"}
		} else if teeType == "1B" {
			teeList = []string{"1C"}
		} else if teeType == "1C" {
			teeList = []string{"1A"}
		}
	} else {
		if teeType == "1A" {
			teeList = []string{"1B", "1C"}
		} else if teeType == "1B" {
			teeList = []string{"1C", "1A"}
		} else if teeType == "1C" {
			teeList = []string{"1A", "1B"}
		}

	}

	timeParts := []response.TeeTimePartOTA{
		{
			IsHideTeePart: bookSetting.IsHideTeePart1,
			StartPart:     bookSetting.StartPart1,
			EndPart:       bookSetting.EndPart1,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart2,
			StartPart:     bookSetting.StartPart2,
			EndPart:       bookSetting.EndPart2,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart3,
			StartPart:     bookSetting.StartPart3,
			EndPart:       bookSetting.EndPart3,
		},
	}

	index := 0
	teeTimeListLL := []string{}

	for _, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1

				hour := teeTimeInit.Hour()
				minute := teeTimeInit.Minute()

				hourStr_ := strconv.Itoa(hour)
				if hour < 10 {
					hourStr_ = "0" + hourStr_
				}
				minuteStr := strconv.Itoa(minute)
				if minute < 10 {
					minuteStr = "0" + minuteStr
				}

				hourStr := hourStr_ + ":" + minuteStr

				teeTimeListLL = append(teeTimeListLL, hourStr)
				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.Unix() > endTime.Unix() {
					break
				}
			}
		}
	}

	for index, teeTypeLock := range teeList {

		t := currentTeeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(bookSetting.TurnLength)) * time.Duration(index+1))

		hour := t.Hour()
		minute := t.Minute()

		hourStr_ := strconv.Itoa(hour)
		if hour < 10 {
			hourStr_ = "0" + hourStr_
		}
		minuteStr := strconv.Itoa(minute)
		if minute < 10 {
			minuteStr = "0" + minuteStr
		}

		teeTime1B := hourStr_ + ":" + minuteStr

		if utils.Contains(teeTimeListLL, teeTime1B) {
			lockTeeTime := models.LockTeeTimeWithSlot{
				PartnerUid:     body.PartnerUid,
				CourseUid:      body.CourseUid,
				TeeTime:        teeTime1B,
				TeeTimeStatus:  "LOCKED",
				DateTime:       body.BookingDate,
				CurrentTeeTime: body.TeeTime,
				TeeType:        teeTypeLock,
				Type:           constants.LOCK_CMS,
				CurrentCourse:  teeType,
			}

			lockTeeTimeToRedis(lockTeeTime)
		}
	}

	return nil
}
func (_ *CLockTeeTime) DeleteLockTurn(db *gorm.DB, teeTime string, bookingDate string, courseUid string) error {
	listTeeTimeLockRedis := getTeeTimeLockRedis(courseUid, bookingDate, "")

	for _, teeTimeR := range listTeeTimeLockRedis {
		if teeTimeR.CurrentTeeTime == teeTime {
			teeTimeRedisKey := getKeyTeeTimeLockRedis(teeTimeR.DateTime, teeTimeR.CourseUid, teeTimeR.TeeTime, teeTimeR.TeeType)
			err := datasources.DelCacheByKey(teeTimeRedisKey)

			log.Print(err)
			if err := datasources.DelCacheByKey(teeTimeRedisKey); err != nil {
				log.Println("DeleteLockTurn", err)
			}
		}
	}
	return nil
}

func (_ *CLockTeeTime) DeleteLockTeeTime(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteLockRequest{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if query.Type == constants.LOCK_OTA {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", "Unlock Fail", constants.ERROR_DELETE_LOCK_OTA)
		return
	}

	teeTimeRedisKey := getKeyTeeTimeLockRedis(query.BookingDate, query.CourseUid, query.TeeTime, query.TeeType+query.CourseType)
	err := datasources.DelCacheByKey(teeTimeRedisKey)
	log.Print(err)
	okRes(c)
}
