package controllers

import (
	"log"
	"net/http"
	"start/config"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strings"

	"github.com/gin-gonic/gin"
)

type CRedis struct{}

func (_ *CRedis) DeleteAllRedisTeeTime(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteRedis{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Xóa tee time lock
	teeTimeLockRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:"
	listKey, _ := datasources.GetAllKeysWith(teeTimeLockRedisKey)
	errTeeTimeLock := datasources.DelCacheByKey(listKey...)
	log.Print(errTeeTimeLock)

	// Xóa row_index
	teeTimeRowIndexRedisKey := config.GetEnvironmentName() + ":" + "tee_time_row_index:"
	listRowIndexKey, _ := datasources.GetAllKeysWith(teeTimeRowIndexRedisKey)
	errTeeTimeRowIndex := datasources.DelCacheByKey(listRowIndexKey...)
	log.Print(errTeeTimeRowIndex)

	// Xóa slot tee time
	teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty:"
	listTeeTimeSlotKey, _ := datasources.GetAllKeysWith(teeTimeSlotEmptyRedisKey)
	errTeeTimeSlot := datasources.DelCacheByKey(listTeeTimeSlotKey...)
	log.Print(errTeeTimeSlot)

	okRes(c)
}

func (_ *CRedis) DeleteTeeTimeRedis(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteLockRequest{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	prefixRedisKey := config.GetEnvironmentName() + ":" + query.TeeTimeType + ":"
	if query.BookingDate != "" {
		prefixRedisKey += query.BookingDate
	}
	if query.CourseUid != "" {
		prefixRedisKey += "_" + query.CourseUid
	}
	if query.TeeType != "" && query.CourseType != "" {
		prefixRedisKey += "_" + query.TeeType + query.CourseType
	}
	if query.TeeTime != "" {
		prefixRedisKey += "_" + query.TeeTime
	}

	log.Println("prefixRedisKey", prefixRedisKey)
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
	if errRedis != nil {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", errRedis.Error(), http.StatusBadRequest)
		return
	}
	errTeeTimeSlot := datasources.DelCacheByKey(listKey...)
	log.Print(errTeeTimeSlot)
	okRes(c)
}

func (_ *CRedis) UpdateTeeTimeRedis(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteLockRequest{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	prefixRedisKey := config.GetEnvironmentName() + ":" + query.TeeTimeType + ":"
	if query.BookingDate != "" {
		prefixRedisKey += query.BookingDate
	}
	if query.CourseUid != "" {
		prefixRedisKey += "_" + query.CourseUid
	}
	if query.TeeType != "" && query.CourseType != "" {
		prefixRedisKey += "_" + query.TeeType + query.CourseType
	}
	if query.TeeTime != "" {
		prefixRedisKey += "_" + query.TeeTime
	}

	log.Println("prefixRedisKey", prefixRedisKey)
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
	if errRedis != nil {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", errRedis.Error(), http.StatusBadRequest)
		return
	}

	newlistKey := []string{}
	for _, key := range listKey {
		newK := strings.Replace(key, "02/02/2023", "03/02/2023", 1)
		newlistKey = append(newlistKey, newK)
	}

	listData := []string{}
	strData, errGet := datasources.GetCaches(listKey...)
	if errGet != nil {
		log.Println("updateSlotTeeTimeWithLock-error", errGet.Error())
	} else {
		for _, data := range strData {
			if data != nil {
				byteData := []byte(data.(string))
				rowIndex := string(byteData)
				listData = append(listData, rowIndex)
			}
		}
	}

	for index, key := range newlistKey {
		if index == 4 {
			break
		}
		rowIndexsRedis := utils.ConvertStringToIntArray(listData[index])
		rowIndexsRaw, _ := rowIndexsRedis.Value()
		if errRedis := datasources.SetCache(key, rowIndexsRaw, 0); errRedis != nil {
			log.Print("fasdfasdfdas")
		}
	}

	okRes(c)
}
