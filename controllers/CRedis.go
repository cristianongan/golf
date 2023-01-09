package controllers

import (
	"log"
	"net/http"
	"start/config"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"

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
	teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_"
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
