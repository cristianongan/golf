package models

import (
	"start/constants"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	IsInCourse      bool    `json:"is_in_course"`
	// TODO: AvailableStatus
}

type BuggyRequest struct {
	Buggy
	FunctionType string `form:"function_type"`
}

type BuggyResponse struct {
	ModelId
	CourseUid string `json:"course_uid"` // San Golf
	Code      string `json:"code"`       // Id Buddy vận hành
	Number    int    `json:"number"`
	Origin    string `json:"origin"`
	Note      string `json:"note"`
}

func (item *Buggy) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *Buggy) Delete(db *gorm.DB) error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *Buggy) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Buggy) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Buggy) Count(database *gorm.DB) (int64, error) {
	total := int64(0)

	db := database.Model(Buggy{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Buggy) FindList(database *gorm.DB, page Page, isReady string) ([]Buggy, int64, error) {
	var list []Buggy
	total := int64(0)

	db := database.Model(Buggy{})

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}
	if item.BuggyStatus != "" {
		db = db.Where("buggy_status = ?", item.BuggyStatus)
	}
	if item.BuggyForVip == true {
		db = db.Where("buggy_for_vip = ?", item.BuggyForVip)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if isReady != "" {
		buggyReadyStatus := []string{
			constants.BUGGY_CURRENT_STATUS_ACTIVE,
			constants.BUGGY_CURRENT_STATUS_FINISH,
			// constants.BUGGY_CURRENT_STATUS_IN_COURSE,
		}

		db = db.Where("buggy_status IN (?) ", buggyReadyStatus)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BuggyRequest) FindBuggyReadyList(database *gorm.DB) ([]Buggy, int64, error) {
	var list []Buggy
	total := int64(0)

	db := database.Model(Buggy{})
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.FunctionType == constants.GO_IN_WAITING {
		buggyReadyStatus := []string{
			constants.BUGGY_CURRENT_STATUS_ACTIVE,
			constants.BUGGY_CURRENT_STATUS_FINISH,
			constants.BUGGY_CURRENT_STATUS_LOCK,
		}

		db = db.Where("buggy_status IN (?) ", buggyReadyStatus)
	}

	if item.FunctionType == constants.GO_IN_COURSE {
		buggyReadyStatus := []string{
			constants.BUGGY_CURRENT_STATUS_ACTIVE,
			constants.BUGGY_CURRENT_STATUS_FINISH,
			constants.BUGGY_CURRENT_STATUS_IN_COURSE,
		}

		db = db.Where("buggy_status IN (?) ", buggyReadyStatus)
	}

	db.Count(&total)
	db.Find(&list)
	return list, total, db.Error
}

func (item *Buggy) FindListBuggyNotReady(database *gorm.DB) ([]Buggy, int64, error) {
	var list []Buggy
	total := int64(0)

	db := database.Model(Buggy{})

	buggyReadyStatus := []string{
		constants.BUGGY_CURRENT_STATUS_LOCK,
		constants.BUGGY_CURRENT_STATUS_FINISH,
		constants.BUGGY_CURRENT_STATUS_IN_COURSE,
	}
	db = db.Where("buggy_status IN (?)", buggyReadyStatus)
	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}
