package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Booking setting
type BookingSettingGroup struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Group Name
	From       int64  `json:"from" gorm:"index"`                          // Áp dụng từ ngày
	To         int64  `json:"to" gorm:"index"`                            // Áp dụng tới ngày
}

func (item *BookingSettingGroup) IsDuplicated() bool {
	bookingSettingGroup := BookingSettingGroup{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Name:       item.Name,
		From:       item.From,
		To:         item.To,
	}

	errFind := bookingSettingGroup.FindFirst()
	if errFind == nil || bookingSettingGroup.Id > 0 {
		return true
	}
	return false
}

func (item *BookingSettingGroup) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	return true
}

func (item *BookingSettingGroup) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BookingSettingGroup) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingSettingGroup) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BookingSettingGroup) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BookingSettingGroup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingSettingGroup) FindList(page models.Page, from, to int64) ([]BookingSettingGroup, int64, error) {
	db := datasources.GetDatabase().Model(BookingSettingGroup{})
	list := []BookingSettingGroup{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	//Search With Time
	if from > 0 && to > 0 {
		db = db.Where("from < " + strconv.FormatInt(from+30, 10) + " ")
		// db = db.Where("to > " + strconv.FormatInt(to-30, 10) + " ")
	}
	if from > 0 && to == 0 {
		db = db.Where("from < " + strconv.FormatInt(from+30, 10) + " ")
	}
	if from == 0 && to > 0 {
		db = db.Where("to > " + strconv.FormatInt(to-30, 10) + " ")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingSettingGroup) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
