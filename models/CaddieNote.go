package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

type CaddieNote struct {
	ModelId
	CourseId string `json:"course_id" gorm:"type:varchar(100);index"`
	CaddieId int64  `json:"caddie_id" gorm:"type:varchar(100);index"`
	AtDate   int64  `json:"at_date"`
	Type     string `json:"type" gorm:"type:varchar(40)"`
	Note     string `json:"note" gorm:"type:varchar(200)"`
}

type CaddieNoteResponse struct {
	ModelId
	CourseId   string `json:"course_id"`
	CaddieId   int64  `json:"caddie_id"`
	CaddieNum  string `json:"caddie_num"`
	CaddieName string `json:"caddie_name"`
	Phone      string `json:"phone"`
	AtDate     int64  `json:"at_date"`
	Type       string `json:"type"`
	Note       string `json:"note"`
}

func (item *CaddieNote) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieNote) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CaddieNote) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieNote) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieNote) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieNote{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieNote) FindList(page Page, from int64, to int64) ([]CaddieNoteResponse, int64, error) {
	var list []CaddieNoteResponse
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieNote{})
	db = db.Where(item)

	if from > 0 {
		db = db.Where("at_date >= ?", from)
	}

	if to > 0 {
		db = db.Where("at_date < ?", to)
	}

	db.Count(&total)

	db = db.Joins("JOIN caddies ON caddie_notes.caddie_id = caddies.id")
	db = db.Select("caddie_notes.id, caddie_notes.created_at, caddie_notes.updated_at, " +
		"caddie_notes.status, caddie_notes.course_id, caddies.id AS caddie_id, " +
		"caddies.num AS caddie_num, caddies.name AS caddie_name, caddies.phone as phone, " +
		"caddie_notes.at_date, caddie_notes.type, caddie_notes.note")

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
