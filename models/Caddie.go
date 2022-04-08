package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

type Caddie struct {
	ModelId
	CourseId       string `json:"course_id" gorm:"type:varchar(100);index"`
	Num            string `json:"caddie_num" gorm:"type:varchar(40)"`
	Name           string `json:"name" gorm:"type:varchar(120)"`
	Phone          string `json:"phone" gorm:"type:varchar(20)"`
	Address        string `json:"address" gorm:"type:varchar(200)"`
	Image          string `json:"image" gorm:"type:varchar(200)"`
	Sex            bool   `json:"sex"`
	BirthDay       int64  `json:"birth_day"`
	BirthPlace     string `json:"birth_place" gorm:"type:varchar(200)"`
	IdentityCard   string `json:"identity_card" gorm:"type:varchar(20)"`
	IssuedBy       string `json:"issued_by" gorm:"type:varchar(200)"`
	IssuedDate     int64  `json:"issued_date"`
	ExpiredDate    int64  `json:"expired_date"`
	EducationLevel string `json:"education_level" gorm:"type:varchar(40)"`
	FingerPrint    string `json:"finger_print" gorm:"type:varchar(40)"`
	HrCode         string `json:"hr_code" gorm:"type:varchar(20)"`
	HrPosition     string `json:"hr_position" gorm:"type:varchar(40)"`
	Group          string `json:"group" gorm:"type:varchar(20)"`
	StartedDate    int64  `json:"started_date"`
	WorkingStatus  string `json:"working_status" gorm:"type:varchar(20)"`
	Level          string `json:"level" gorm:"type:varchar(40)"`
	Note           string `json:"note" gorm:"type:varchar(200)"`
}

type CaddieResponse struct {
	ModelId
	CourseId       string `json:"course_id"`
	Num            string `json:"caddie_num"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Image          string `json:"image"`
	Sex            bool   `json:"sex"`
	BirthDay       int64  `json:"birth_day"`
	BirthPlace     string `json:"birth_place"`
	IdentityCard   string `json:"identity_card"`
	IssuedBy       string `json:"issued_by"`
	IssuedDate     int64  `json:"issued_date"`
	ExpiredDate    int64  `json:"expired_date"`
	EducationLevel string `json:"education_level"`
	FingerPrint    string `json:"finger_print"`
	HrCode         string `json:"hr_code"`
	HrPosition     string `json:"hr_position"`
	Group          string `json:"group"`
	StartedDate    int64  `json:"started_date"`
	WorkingStatus  string `json:"working_status"`
	Level          string `json:"level"`
	Note           string `json:"note"`
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
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
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
