package models

import (
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"time"
)

type CaddieWorkingSchedule struct {
	ModelId
	PartnerUid      string         `json:"partner_uid"`
	CourseUid       string         `json:"course_uid"`
	WeekId          string         `json:"week_id"`
	CaddieGroupName string         `json:"caddie_group_name"`
	CaddieGroupCode string         `json:"caddie_group_code"`
	ApplyDate       datatypes.Date `json:"apply_date"`
	IsDayOff        bool           `json:"is_day_off"`
}

func (item *CaddieWorkingSchedule) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieWorkingSchedule) FindList(page Page) ([]CaddieWorkingSchedule, int64, error) {
	var list []CaddieWorkingSchedule
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieWorkingSchedule{})

	if item.WeekId != "" {
		db.Where("week_id = ?", item.WeekId)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
