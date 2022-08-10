package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Group Fee
type McTypeAnnualFee struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	McTypeId   int64  `json:"mc_type_id" gorm:"index"`                    // Member Card Type id
	Year       int    `json:"year" gorm:"index"`
	Fee        int64  `json:"fee"`
}

func (item *McTypeAnnualFee) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *McTypeAnnualFee) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *McTypeAnnualFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *McTypeAnnualFee) Count() (int64, error) {
	db := datasources.GetDatabase().Model(McTypeAnnualFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *McTypeAnnualFee) FindByMcTypeId() ([]McTypeAnnualFee, error) {
	db := datasources.GetDatabase().Model(McTypeAnnualFee{})
	list := []McTypeAnnualFee{}
	db = db.Where("mc_type_id = ?", item.McTypeId)

	db.Find(&list)

	return list, db.Error
}

func (item *McTypeAnnualFee) FindList(page Page) ([]McTypeAnnualFee, int64, error) {
	db := datasources.GetDatabase().Model(McTypeAnnualFee{})
	list := []McTypeAnnualFee{}
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

func (item *McTypeAnnualFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
