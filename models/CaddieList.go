package models

import (
	"start/constants"
	"start/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CaddieList struct {
	PartnerUid            string
	CourseUid             string
	CaddieName            string
	CaddieCode            string
	Month                 string
	WorkingStatus         string
	InCurrentStatus       []string
	CaddieCodeList        []string
	GroupId               int64
	OrderByGroupIndexDesc bool
	Level                 string
	Phone                 string
	IsInGroup             string
	IsReadyForBooking     string
	IsReadyForJoin        string
	ContractStatus        string
	CurrentStatus         string
	IsBooked              string
	GroupList             []int64
}

type CaddieRes struct {
	Number   string `json:"Number"`
	FullName string `json:"FullName"`
	Phone    string `json:"Phone"`
}

func addFilter(db *gorm.DB, item *CaddieList) *gorm.DB {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" || item.CaddieCode != "" {
		db = db.Where("name COLLATE utf8mb4_general_ci LIKE ? OR code LIKE ?", "%"+item.CaddieName+"%", "%"+item.CaddieCode+"%")
	}

	if len(item.InCurrentStatus) > 0 {
		db = db.Where("current_status IN ?", item.InCurrentStatus)
	}

	if item.Level != "" {
		db = db.Where("level = ?", item.Level)
	}

	if item.Phone != "" {
		db = db.Where("phone LIKE ?", "%"+item.Phone+"%")
	}

	if item.WorkingStatus != "" {
		db = db.Where("working_status = ?", item.WorkingStatus)
	}

	if item.GroupId != 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}

	if item.ContractStatus != "" {
		status := strings.Split(item.ContractStatus, ",")
		db = db.Where("contract_status in (?)", status)
	}

	if item.CurrentStatus != "" {
		status := strings.Split(item.CurrentStatus, ",")
		db = db.Where("current_status in (?)", status)
	}

	if item.IsBooked == "1" {
		caddieStatus := []string{
			constants.CADDIE_CURRENT_STATUS_LOCK,
			constants.CADDIE_CURRENT_STATUS_IN_COURSE,
			constants.CADDIE_CURRENT_STATUS_IN_COURSE_R2,
			constants.CADDIE_CURRENT_STATUS_IN_COURSE_R3,
		}

		db = db.Where("current_status IN (?) ", caddieStatus)
	}

	if item.IsInGroup != "" {
		isInGroup, _ := strconv.ParseInt(item.IsInGroup, 10, 8)
		if isInGroup == 1 {
			db = db.Where("group_id <> ?", 0)
		} else if isInGroup == 0 {
			db = db.Where("group_id = ?", 0)
		}
	}

	if item.IsReadyForBooking != "" {
		isReadyForBooking, _ := strconv.ParseInt(item.IsReadyForBooking, 10, 8)
		if isReadyForBooking == 1 {
			db = db.Where("working_status = ?", constants.CADDIE_WORKING_STATUS_ACTIVE).Where("current_status <> ?", constants.CADDIE_CURRENT_STATUS_WORKING_ONLY).Where("current_status <> ?", constants.CADDIE_CURRENT_STATUS_JOB)
		} else if isReadyForBooking == 0 {
			db = db.Where("working_status = ?", constants.CADDIE_WORKING_STATUS_INACTIVE).Or("current_status = ?", constants.CADDIE_CURRENT_STATUS_WORKING_ONLY).Or("current_status = ?", constants.CADDIE_CURRENT_STATUS_JOB)
		}
	}

	if item.IsReadyForJoin != "" {
		caddieStatus := []string{
			constants.CADDIE_CURRENT_STATUS_READY,
			constants.CADDIE_CURRENT_STATUS_FINISH,
			constants.CADDIE_CURRENT_STATUS_FINISH_R2,
			constants.CADDIE_CURRENT_STATUS_FINISH_R3,
		}

		db = db.Where("current_status IN (?) ", caddieStatus)
	}

	if len(item.GroupList) > 0 {
		db = db.Where("group_id IN ?", item.GroupList)
	}

	return db
}

func (item *CaddieList) FindList(database *gorm.DB, page Page) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := database.Model(Caddie{})

	db = addFilter(db, item)

	db.Not("status = ?", constants.STATUS_DELETED)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Preload("GroupInfo").Preload("CaddieVacationCalendar", "number_day_off > 15 OR title = ?", "MATERNITY_LEAD").Find(&list)
	}

	return list, total, db.Error
}

func (item *CaddieList) FindAllCaddieReadyOnDayList(database *gorm.DB) ([]Caddie, int64, error) {
	var list []Caddie

	db := database.Model(Caddie{})

	db = addFilter(db, item)

	db.Not("status = ?", constants.STATUS_DELETED)
	db.Where("is_working = 1")

	db.Preload("GroupInfo").Find(&list)

	return list, int64(len(list)), db.Error
}

func (item *CaddieList) FindAllCaddieReadyOnDayListOTA(database *gorm.DB, date string) ([]CaddieRes, int64, error) {
	var list []Caddie

	db := database.Model(Caddie{})

	db = addFilter(db, item)

	db.Not("status = ?", constants.STATUS_DELETE)

	db.Preload("GroupInfo").Find(&list)

	var timeNow datatypes.Date
	if date != "" {
		timeUnix, _ := time.Parse(constants.DATE_FORMAT, date)
		timeNow = datatypes.Date(timeUnix)
	} else {
		timeNow = datatypes.Date(utils.GetTimeNow())
	}

	listResponse := []CaddieRes{}

	for _, data := range list {
		dbCaddieWorkingSchedule := database.Model(CaddieWorkingSchedule{})
		caddieSchedules := CaddieWorkingSchedule{
			ApplyDate:       &timeNow,
			PartnerUid:      data.PartnerUid,
			CourseUid:       data.CourseUid,
			CaddieGroupCode: data.GroupInfo.Code,
		}

		if caddieSchedules.CheckCaddieWorkOnDay(dbCaddieWorkingSchedule) {
			item := CaddieRes{
				Number:   strconv.FormatInt(data.Id, 10),
				FullName: data.Name,
				Phone:    data.Phone,
			}
			listResponse = append(listResponse, item)
		}
	}

	return listResponse, int64(len(listResponse)), db.Error
}

func (item CaddieList) FindListWithoutPage(database *gorm.DB) ([]Caddie, error) {
	var list []Caddie

	db := database.Model(Caddie{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if len(item.CaddieCodeList) > 0 {
		db = db.Where("code IN ?", item.CaddieCodeList)
	}

	if item.GroupId != 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}

	if item.OrderByGroupIndexDesc {
		db = db.Order("group_index desc")
	}

	err := db.Find(&list).Error

	if err != nil {
		return []Caddie{}, err
	}

	return list, nil
}

func (item CaddieList) FindFirst(database *gorm.DB) (Caddie, error) {
	var result Caddie
	db := database.Model(Caddie{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" {
		db = db.Where("name LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("code = ?", item.CaddieCode)
		item.CaddieCode = ""
	}

	if len(item.InCurrentStatus) > 0 {
		db = db.Where("current_status IN ?", item.InCurrentStatus)
	}

	err := db.First(&result).Error

	return result, err
}
