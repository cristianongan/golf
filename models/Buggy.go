package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

type Buggy struct {
	ModelId
	PartnerUid      string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code            string  `json:"code" gorm:"type:varchar(256);index"`        // Id Buddy vận hành
	Origin          string  `json:"origin" gorm:"type:varchar(200)"`
	Note            string  `json:"note" gorm:"type:varchar(200)"`
	BuggyForVip     bool    `json:"buggy_for_vip"`
	WarrantyPeriod  float64 `json:"warranty_period"`
	MaintenanceFrom int64   `json:"maintenance_from"`
	MaintenanceTo   int64   `json:"maintenance_to"`
	BuggyStatus     string  `json:"buggy_status"`
	// TODO: AvailableStatus
}

type BuggyResponse struct {
	ModelId
	CourseUid string `json:"course_uid"` // San Golf
	Code      string `json:"code"`       // Id Buddy vận hành
	Number    int    `json:"number"`
	Origin    string `json:"origin"`
	Note      string `json:"note"`
}

func (item *Buggy) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Buggy) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Buggy) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Buggy) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Buggy) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Buggy{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Buggy) FindList(page Page) ([]Buggy, int64, error) {
	var list []Buggy
	total := int64(0)

	db := datasources.GetDatabase().Model(Buggy{})
	db = db.Where(item)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}
	if item.BuggyStatus != "" {
		db = db.Where("buggy_status = ?", item.BuggyStatus)
	}
	if item.BuggyForVip == true {
		db = db.Where("buggy_for_vip = ?", item.BuggyForVip)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
