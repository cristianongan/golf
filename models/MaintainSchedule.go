package models

import (
	"start/constants"
	"start/utils"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MaintainSchedule struct {
	ModelId
	PartnerUid       string          `json:"partner_uid"`
	CourseUid        string          `json:"course_uid"`
	WeekId           string          `json:"week_id"`
	CourseName       string          `json:"course_name"`
	ApplyDate        *datatypes.Date `json:"apply_date"`
	MorningOff       *bool           `json:"morning_off"`
	AfternoonOff     *bool           `json:"afternoon_off"`
	MorningTimeOff   string          `json:"morning_time_off"`
	AfternoonTimeOff string          `json:"afternoon_time_off"`
}

func (item *MaintainSchedule) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *MaintainSchedule) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *MaintainSchedule) FindList(database *gorm.DB, page Page) ([]MaintainSchedule, int64, error) {
	var list []MaintainSchedule
	total := int64(0)

	db1 := database.Model(MaintainSchedule{})
	db2 := database.Model(MaintainSchedule{})

	if item.WeekId != "" {
		db1 = db1.Where("week_id = ?", item.WeekId)
	}

	if item.CourseUid != "" {
		db1 = db1.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db1 = db1.Where("partner_uid = ?", item.PartnerUid)
	}

	query := db1.Select("MAX(id) as id_latest").Group("course_name, apply_date")

	db2.Joins("JOIN (?) q ON maintain_schedules.id = q.id_latest", query).Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db2 = page.Setup(db2).Find(&list)
	}
	return list, total, db2.Error
}

func (item *MaintainSchedule) FindListWithoutPage(database *gorm.DB) ([]MaintainSchedule, error) {
	var list []MaintainSchedule

	db1 := database.Model(MaintainSchedule{})
	db2 := database.Model(MaintainSchedule{})

	if item.WeekId != "" {
		db1 = db1.Where("week_id = ?", item.WeekId)
	}

	if item.CourseUid != "" {
		db1 = db1.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db1 = db1.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.ApplyDate != nil {
		db1 = db1.Where("apply_date = ?", time.Time(*item.ApplyDate).Format("2006-01-02"))
	}

	if item.MorningOff != nil {
		if *item.MorningOff == true {
			db1 = db1.Where("is_day_off = ?", 1)
		}
		if *item.MorningOff == false {
			db1 = db1.Where("is_day_off = ?", 0)
		}
	}

	query := db1.Select("MAX(id) as id_latest").Group("caddie_group_code, apply_date")

	err := db2.Joins("JOIN (?) q ON maintain_schedules.id = q.id_latest", query).Find(&list).Error

	return list, err
}

func (item *MaintainSchedule) CheckCaddieWorkOnDay(database *gorm.DB) bool {
	var list []MaintainSchedule

	db1 := database.Model(MaintainSchedule{})

	if item.CourseUid != "" {
		db1 = db1.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db1 = db1.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.ApplyDate != nil {
		db1 = db1.Where("apply_date = ?", time.Time(*item.ApplyDate).Format("2006-01-02"))
	}

	db1 = db1.Order("created_at desc")
	db1.Find(&list)

	if len(list) > 0 {
		firstItem := list[0]
		return !*firstItem.MorningOff
	}

	return false
}

func (item *MaintainSchedule) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	return db.Save(item).Error
}
