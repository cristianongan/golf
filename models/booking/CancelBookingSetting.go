package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type CancelBookingSetting struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	PeopleFrom int64  `json:"people_from"`
	PeopleTo   int64  `json:"people_to"`
	TimeMin    int64  `json:"time_min"`
	TimeMax    int64  `json:"time_max"`
}

func (item *CancelBookingSetting) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CancelBookingSetting) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CancelBookingSetting) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CancelBookingSetting) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CancelBookingSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CancelBookingSetting) FindList() ([]CancelBookingSetting, int64, error) {
	db := datasources.GetDatabase().Model(CancelBookingSetting{})
	list := []CancelBookingSetting{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)
	db = db.Find(&list)

	return list, total, db.Error
}

func (item *CancelBookingSetting) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
