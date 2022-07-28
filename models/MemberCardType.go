package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Loại thẻ
type MemberCardType struct {
	ModelId
	PartnerUid         string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid          string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	Name               string `json:"name" gorm:"type:varchar(256)"`                  // Ten Loai Member Card
	GuestStyle         string `json:"guest_style" gorm:"index"`                       // Guest Style ???
	GuestStyleOffGuest string `json:"guest_style_off_guest" gorm:"type:varchar(100)"` // Guest Style Off guest ???
	PromotGuestStyle   string `json:"promot_guest_style" gorm:"type:varchar(100)"`    // Promot guest style ???
	NormalDayTakeGuest string `json:"normal_day_take_guest" gorm:"type:varchar(100)"` // Normal day take guest ???
	WeekendTakeGuest   string `json:"weekend_take_guest" gorm:"type:varchar(100)"`    // Weekend take guest ???
	PlayTimesOnMonth   int    `json:"play_times_on_month"`                            // Số lần chơi trên tháng
	Type               string `json:"type" gorm:"type:varchar(100);index"`            // Type: Friendly, InsideMember, OutsideMember, Promotion...
	PlayTimeOnYear     int    `json:"play_times_on_year"`                             // Số lần chơi trong năm
	AnnualType         string `json:"annual_type" gorm:"type:varchar(100)"`           // loại thường niên: không giới hạn, chơi có giới hạn, thẻ ngủ.
}

func (item *MemberCardType) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.Type == "" {
		return false
	}
	if item.GuestStyle == "" {
		return false
	}
	if item.AnnualType == "" {
		return false
	}
	return true
}

func (item *MemberCardType) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *MemberCardType) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *MemberCardType) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *MemberCardType) Count() (int64, error) {
	db := datasources.GetDatabase().Model(MemberCardType{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *MemberCardType) FindList(page Page) ([]MemberCardType, int64, error) {
	db := datasources.GetDatabase().Model(MemberCardType{})
	list := []MemberCardType{}
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
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%")
	}
	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}
	if item.GuestStyle != "" {
		db = db.Where("guest_style LIKE ?", "%"+item.GuestStyle+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *MemberCardType) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
