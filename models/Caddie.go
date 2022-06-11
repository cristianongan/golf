package models

import (
	"start/constants"
	"start/datasources"
	"time"
)

type Caddie struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code          string `json:"code" gorm:"type:varchar(256);index"`        // Id Caddie vận hành
	Name          string `json:"name" gorm:"type:varchar(120)"`
	Sex           bool   `json:"sex"`
	BirthDay      int64  `json:"birth_day"`
	WorkingStatus string `json:"working_status" gorm:"type:varchar(20)"`
	Group         string `json:"group" gorm:"type:varchar(20)"`
	StartedDate   int64  `json:"started_date"`
	IdHr          string `json:"id_hr" gorm:"type:varchar(100)"`
	Phone         string `json:"phone" gorm:"type:varchar(20)"`
	Email         string `json:"email" gorm:"type:varchar(100)"`
	IdentityCard  string `json:"identity_card" gorm:"type:varchar(20)"`
	IssuedBy      string `json:"issued_by" gorm:"type:varchar(200)"`
	ExpiredDate   int64  `json:"expired_date"`
	PlaceOfOrigin string `json:"place_of_origin" gorm:"type:varchar(200)"`
	Address       string `json:"address" gorm:"type:varchar(200)"`
	Level         string `json:"level" gorm:"type:varchar(40)"`
	Note          string `json:"note" gorm:"type:varchar(200)"`
}

type CaddieResponse struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code          string `json:"code" gorm:"type:varchar(256);index"`        // Id Caddie vận hành
	Name          string `json:"name"`
	Sex           bool   `json:"sex"`
	BirthDay      int64  `json:"birth_day"`
	WorkingStatus string `json:"working_status"`
	Group         string `json:"group"`
	StartedDate   int64  `json:"started_date"`
	IdHr          string `json:"id_hr"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	IdentityCard  string `json:"identity_card"`
	IssuedBy      string `json:"issued_by"`
	ExpiredDate   int64  `json:"expired_date"`
	PlaceOfOrigin string `json:"place_of_origin"`
	Address       string `json:"address"`
	Level         string `json:"level"`
	Note          string `json:"note"`
}

func (item *Caddie) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Caddie) CreateBatch(caddies []Caddie) error {
	now := time.Now()
	for i := range caddies {
		c := &caddies[i]
		c.ModelId.CreatedAt = now.Unix()
		c.ModelId.UpdatedAt = now.Unix()
		c.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.CreateInBatches(caddies, 100).Error
}

func (item *Caddie) Delete() error {
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Caddie) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Caddie) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Caddie) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Caddie{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Caddie) FindList(page Page) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := datasources.GetDatabase().Model(Caddie{})
	db = db.Where(item)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
