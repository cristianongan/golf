package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Bag Note
type BagsNote struct {
	ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GolfBag     string `json:"golf_bag" gorm:"type:varchar(200)"`
	BookingUid  string `json:"booking_uid" gorm:"type:varchar(50);index"`
	Note        string `json:"note" gorm:"type:varchar(2000)"`
	PlayerName  string `json:"player_name" gorm:"type:varchar(256)"`
	Type        string `json:"type" gorm:"type:varchar(50)"`
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
}

// ======= CRUD ===========
func (item *BagsNote) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BagsNote) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BagsNote) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BagsNote) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BagsNote{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BagsNote) FindList(database *gorm.DB, page Page) ([]BagsNote, int64, error) {
	db := database.Model(BagsNote{})
	list := []BagsNote{}
	total := int64(0)
	status := item.Status
	item.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.GolfBag != "" {
		db = db.Where("golf_bag = ?", item.GolfBag)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BagsNote) Delete(db *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
