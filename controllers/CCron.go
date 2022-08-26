package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"start/constants"
	"start/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type CCron struct{}

func (_ CCron) CreateCaddieWorkingCalendar(c *gin.Context) {
	var err error

	// get caddie groups
	var caddieWorkingScheduleList []models.CaddieWorkingSchedule
	today := datatypes.Date(time.Now())

	idDayOff := false
	caddieWorkingSchedule := models.CaddieWorkingSchedule{}
	caddieWorkingSchedule.ApplyDate = &today
	caddieWorkingSchedule.IsDayOff = &idDayOff
	caddieWorkingScheduleList, err = caddieWorkingSchedule.FindListWithoutPage()

	if err != nil {
		fmt.Println("[CRON_JOB] [CREATE_CADDIE_WORKING_CALENDAR] [ERROR]", err.Error())
	}

	var groupIds []int

	for _, item := range caddieWorkingScheduleList {
		caddieGroup := models.CaddieGroup{}
		caddieGroup.Code = item.CaddieGroupCode
		if err = caddieGroup.FindFirst(); err != nil {
			continue
		}
		groupIds = append(groupIds, int(caddieGroup.Id))
	}

	countGroup := len(groupIds)

	// get caddies
	caddies := map[int][]models.Caddie{}

	var caddieList models.CaddieList

	maxLengthCaddie := 0

	caddieList = models.CaddieList{}
	caddieList.OrderByGroupIndexDesc = true // Check xem đang ở trí số bao nhiêu, check lại theo tăng dần

	for i := 0; i < countGroup; i++ {
		caddieList.GroupId = int64(groupIds[i])
		caddies[i], err = caddieList.FindListWithoutPage()

		if err != nil {
			fmt.Println("[CRON_JOB] [CREATE_CADDIE_WORKING_CALENDAR] [ERROR]", err.Error())
		}

		if maxLengthCaddie < len(caddies[countGroup-1]) {
			maxLengthCaddie = len(caddies[countGroup-1])
		}
	}

	// Xào group với caddie
	// set result
	result := []models.Caddie{}

	for i := 0; i < maxLengthCaddie; i++ {
		for j := 0; j < countGroup; j++ {
			if itemList, ok := caddies[j]; ok {
				if len(itemList) > i {
					result = append(result, caddies[j][i])
				}
			}
		}
	}

	resultJson, _ := json.Marshal(result)

	fmt.Println("[DEBUG]", string(resultJson))

	// set caddie_working_calendars
	lengthResult := len(result)

	// Add vào db
	// Xếp Nốt
	for i := 0; i < lengthResult; i++ {
		caddieWorkingCalendar := models.CaddieWorkingCalendar{}
		caddieWorkingCalendar.CaddieUid = strconv.FormatInt(result[i].Id, 10)
		caddieWorkingCalendar.CaddieCode = result[i].Code
		caddieWorkingCalendar.PartnerUid = result[i].PartnerUid
		caddieWorkingCalendar.CourseUid = result[i].CourseUid
		caddieWorkingCalendar.CaddieLabel = constants.CADDIE_WORKING_CALENDAR_LABEL_READY
		if ((i + 1) % 10) != 0 {
			caddieWorkingCalendar.CaddieColumn = (i + 1) % 10
		} else {
			caddieWorkingCalendar.CaddieColumn = 10
		}
		caddieWorkingCalendar.CaddieRow = "H" + strconv.FormatInt(int64(math.Round(float64(i+1)/float64(10))), 10)
		caddieWorkingCalendar.ApplyDate = datatypes.Date(time.Now())

		if err := caddieWorkingCalendar.Create(); err != nil {
			fmt.Println("[CRON_JOB] [CREATE_CADDIE_WORKING_CALENDAR] [ERROR]", err.Error())
		}
	}

	okRes(c)
}
