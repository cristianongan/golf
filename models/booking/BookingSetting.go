package model_booking

import (
	"start/constants"
	"start/models"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Booking setting
type BookingSetting struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Dow            string `json:"dow" gorm:"type:varchar(100)"`               // Ngày áp dụng
	GroupId        int64  `json:"group_id" gorm:"index"`                      // Id nhóm setting
	TeeMinutes     int    `json:"tee_minutes"`                                // Tee minutes
	TurnLength     int    `json:"turn_length"`                                // Config theo minute 2H Xphút
	IsHideTeePart1 bool   `json:"is_hide_tee_part_1"`                         // show hide tee fart 1 - Morning sáng
	IsHideTeePart2 bool   `json:"is_hide_tee_part_2"`                         // show hide tee fart 2 - Noon trưa
	IsHideTeePart3 bool   `json:"is_hide_tee_part_3"`                         // show hide tee fart 3 - Night tối
	StartPart1     string `json:"start_part1" gorm:"type:varchar(50)"`        // Ex: 18:26"
	StartPart2     string `json:"start_part2" gorm:"type:varchar(50)"`
	StartPart3     string `json:"start_part3" gorm:"type:varchar(50)"`
	EndPart1       string `json:"end_part1" gorm:"type:varchar(50)"`
	EndPart2       string `json:"end_part2" gorm:"type:varchar(50)"`
	EndPart3       string `json:"end_part3" gorm:"type:varchar(50)"`

	Part1TeeType string `json:"part1_tee_type" gorm:"type:varchar(50)"` // ALL, 1, 10
	Part2TeeType string `json:"part2_tee_type" gorm:"type:varchar(50)"`
	Part3TeeType string `json:"part3_tee_type" gorm:"type:varchar(50)"`

	IncludeDays int `json:"include_days"`
}

func (item *BookingSetting) IsDuplicated(db *gorm.DB) bool {
	bookingSetting := BookingSetting{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		GroupId:    item.GroupId,
		Dow:        item.Dow,
	}

	errFind := bookingSetting.FindFirst(db)
	if errFind == nil || bookingSetting.Id > 0 {
		return true
	}
	return false
}

func (item *BookingSetting) IsValidated() bool {
	if item.Dow == "" {
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

func (item *BookingSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingSetting) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingSetting) FindList(database *gorm.DB, page models.Page) ([]BookingSetting, int64, error) {
	db := database.Model(BookingSetting{})
	list := []BookingSetting{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingSetting) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
