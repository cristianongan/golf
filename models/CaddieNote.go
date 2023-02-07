package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieNote struct {
	ModelId
	CourseUid  string `json:"course_uid" gorm:"size:256"`
	PartnerUid string `json:"partner_uid" gorm:"size:256"`
	CaddieId   int64  `json:"caddie_id"`
	AtDate     int64  `json:"at_date"`
	Type       string `json:"type" gorm:"type:varchar(40)"`
	Note       string `json:"note" gorm:"type:varchar(200)"`
}

func (item *CaddieNote) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieNote) Delete(db *gorm.DB) error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *CaddieNote) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieNote) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieNote) Count(database *gorm.DB) (int64, error) {
	total := int64(0)

	db := database.Model(CaddieNote{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieNote) FindList(database *gorm.DB, page Page, from int64, to int64) ([]CaddieNote, int64, error) {
	var list []CaddieNote
	total := int64(0)

	db := database.Model(CaddieNote{})
	db = db.Where(item)

	print(from, to)
	if from > 0 {
		db = db.Where("caddie_notes.at_date >= ?", from)
	}

	if to > 0 {
		db = db.Where("caddie_notes.at_date < ?", to)
	}

	db = db.Joins("JOIN caddies ON caddie_notes.caddie_id = caddies.id")
	db = db.Select("caddie_notes.id, caddie_notes.created_at, caddie_notes.updated_at, " +
		"caddie_notes.status, caddie_notes.caddie_id, caddies.name AS caddie_name, caddies.phone as phone, " +
		"caddie_notes.at_date, caddie_notes.type, caddie_notes.note")

	db.Count(&total)
	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
