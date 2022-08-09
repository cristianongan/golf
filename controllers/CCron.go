package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"start/models"
	"time"
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
	caddieList.OrderByGroupIndexDesc = true

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

	okRes(c)
}
