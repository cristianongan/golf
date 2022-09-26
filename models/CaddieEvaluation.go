package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TODO: add gorm_type
type CaddieEvaluation struct {
	ModelId
	BookingUid  string         `json:"booking_uid" gorm:"size:256"`
	BookingCode string         `json:"booking_code" gorm:"size:256"`
	BookingDate datatypes.Date `json:"booking_date"`
	CaddieUid   string         `json:"caddie_uid" gorm:"size:256"`
	CaddieCode  string         `json:"caddie_code" gorm:"size:256"`
	CaddieName  string         `json:"caddie_name" gorm:"size:256"`
	CourseUid   string         `json:"course_uid" gorm:"size:256"`
	PartnerUid  string         `json:"partner_uid" gorm:"size:256"`
	GolfBag     string         `json:"golf_bag" gorm:"size:256"`
	Hole        int            `json:"hole" gorm:"size:2"`
	RankType    int            `json:"rank_type" gorm:"size:2"`
}

func (item *CaddieEvaluation) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieEvaluation) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieEvaluation) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	return db.Save(item).Error
}
