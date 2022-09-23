package models

import (
	"start/constants"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type HolePriceFormula struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Hole       int    `json:"hole"`                                       // Hố Golf
	StopByRain string `json:"stop_by_rain" gorm:"type:varchar(256)"`      // Dừng bởi trời mưa
	StopBySelf string `json:"stop_by_self" gorm:"type:varchar(256)"`      // Dừng bởi người chơi
}

func (item *HolePriceFormula) IsDuplicated(db *gorm.DB) bool {
	modelCheck := HolePriceFormula{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Hole:       item.Hole,
	}
	errFind := modelCheck.FindFirst(db)
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *HolePriceFormula) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *HolePriceFormula) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *HolePriceFormula) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *HolePriceFormula) Count(database *gorm.DB) (int64, error) {
	db := database.Model(HolePriceFormula{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *HolePriceFormula) FindList(database *gorm.DB, page Page) ([]HolePriceFormula, int64, error) {
	db := database.Model(HolePriceFormula{})
	list := []HolePriceFormula{}
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

func (item *HolePriceFormula) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
