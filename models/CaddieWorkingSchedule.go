package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CaddieWorkingSchedule struct {
	ModelId
	PartnerUid      string          `json:"partner_uid"`
	CourseUid       string          `json:"course_uid"`
	WeekId          string          `json:"week_id"`
	CaddieGroupName string          `json:"caddie_group_name"`
	CaddieGroupCode string          `json:"caddie_group_code"`
	ApplyDate       *datatypes.Date `json:"apply_date"`
	IsDayOff        *bool           `json:"is_day_off"`
}

func (item *CaddieWorkingSchedule) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingSchedule) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingSchedule) FindList(database *gorm.DB, page Page) ([]CaddieWorkingSchedule, int64, error) {
	var list []CaddieWorkingSchedule
	total := int64(0)

	db1 := database.Model(CaddieWorkingSchedule{})
	db2 := database.Model(CaddieWorkingSchedule{})

	if item.WeekId != "" {
		db1 = db1.Where("week_id = ?", item.WeekId)
	}

	if item.CourseUid != "" {
		db1 = db1.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db1 = db1.Where("partner_uid = ?", item.PartnerUid)
	}

	query := db1.Select("MAX(id) as id_latest").Group("caddie_group_code, apply_date")

	db2.Joins("JOIN (?) q ON caddie_working_schedules.id = q.id_latest", query).Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db2 = page.Setup(db2).Find(&list)
	}
	return list, total, db2.Error
}

func (item *CaddieWorkingSchedule) FindListWithoutPage(database *gorm.DB) ([]CaddieWorkingSchedule, error) {
	var list []CaddieWorkingSchedule

	db1 := database.Model(CaddieWorkingSchedule{})
	db2 := database.Model(CaddieWorkingSchedule{})

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

	if item.IsDayOff != nil {
		if *item.IsDayOff == true {
			db1 = db1.Where("is_day_off = ?", 1)
		}
		if *item.IsDayOff == false {
			db1 = db1.Where("is_day_off = ?", 0)
		}
	}

	query := db1.Select("MAX(id) as id_latest").Group("caddie_group_code, apply_date")

	err := db2.Joins("JOIN (?) q ON caddie_working_schedules.id = q.id_latest", query).Find(&list).Error

	return list, err
}

func (item *CaddieWorkingSchedule) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}
