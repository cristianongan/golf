package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/datatypes"
)

// TODO: add gorm_type
type CaddieConfigFee struct {
	ModelId
	PartnerUid string         `json:"partner_uid" gorm:"size:256"`
	CourseUid  string         `json:"course_uid" gorm:"size:256"`
	Type       string         `json:"type" gorm:"size:256"`
	FeeDetail  string         `json:"fee_detail" gorm:"size:256"`
	ValidDate  datatypes.Date `json:"valid_date"`
	ExpDate    datatypes.Date `json:"exp_date"`
}

func (item *CaddieConfigFee) Create() error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieConfigFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieConfigFee) Update() error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

func (item *CaddieConfigFee) FindList(page Page) ([]CaddieConfigFee, int64, error) {
	db := datasources.GetDatabase().Model(CaddieConfigFee{})
	list := []CaddieConfigFee{}
	total := int64(0)
	// status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieConfigFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
