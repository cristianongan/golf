package models

import (
	"start/constants"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieWorkingCalendar struct {
	ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CaddieCode     string `json:"caddie_code" gorm:"type:varchar(100);index"` // caddie code
	ApplyDate      string `json:"apply_date"  gorm:"type:varchar(100)"`       // ngày áp dụng
	Row            int    `json:"row"`                                        // thứ tự hàng
	NumberOrder    int64  `json:"number_order"`                               // số thứ tự caddie\
	CaddieIncrease bool   `json:"caddie_increase"`                            // caddie tăng cường
}

func (item *CaddieWorkingCalendar) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

// / ------- CaddieWorkingCalendar batch insert to db ------
func (item *CaddieWorkingCalendar) BatchInsert(database *gorm.DB, list []CaddieWorkingCalendar) error {
	db := database.Table("caddie_working_calendars")

	return db.Create(&list).Error
}

func (item *CaddieWorkingCalendar) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingCalendar) FindAll(database *gorm.DB) ([]CaddieWorkingCalendar, error) {
	list := []CaddieWorkingCalendar{}

	db := database.Model(CaddieWorkingCalendar{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("apply_date = ?", item.ApplyDate)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *CaddieWorkingCalendar) FindAllByDate(database *gorm.DB) ([]map[string]interface{}, int64, error) {
	list := []map[string]interface{}{}
	total := int64(0)

	db := database.Table("caddie_working_calendars")

	db.Select("caddie_working_calendars.*, caddies.current_status")

	if item.PartnerUid != "" {
		db = db.Where("caddie_working_calendars.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("caddie_working_calendars.course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("caddie_working_calendars.apply_date = ?", item.ApplyDate)
	}

	if item.CaddieIncrease {
		db = db.Where("caddie_working_calendars.caddie_increase = 1")
	} else {
		db = db.Where("caddie_working_calendars.caddie_increase = 0")
	}

	db.Joins("left join caddies on caddies.code = caddie_working_calendars.caddie_code")

	db.Order("caddie_working_calendars.row asc")

	db.Count(&total)

	db.Find(&list)

	return list, total, db.Error
}

func (item *CaddieWorkingCalendar) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}

func (item *CaddieWorkingCalendar) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *CaddieWorkingCalendar) DeleteBatch(db *gorm.DB) error {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("apply_date = ?", item.ApplyDate)
	}

	return db.Delete(item).Error
}
