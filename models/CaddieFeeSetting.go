package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CaddieFee setting
type CaddieFeeSetting struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupId    int64  `json:"group_id" gorm:"index"`                      // Id nhóm setting
	Hole       int    `json:"hole"`                                       // số hố
	Fee        int64  `json:"fee"`                                        // phí tương ứng
	Type       string `json:"type" gorm:"type:varchar(256)"`              // Type setting caddie fee
}

func (item *CaddieFeeSetting) IsDuplicated(db *gorm.DB) bool {
	CaddieFeeSetting := CaddieFeeSetting{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		GroupId:    item.GroupId,
		Hole:       item.Hole,
	}

	errFind := CaddieFeeSetting.FindFirst(db)
	if errFind == nil || CaddieFeeSetting.Id > 0 {
		return true
	}
	return false
}

func (item *CaddieFeeSetting) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.Hole == 0 {
		return false
	}
	if item.Fee == 0 {
		return false
	}
	if item.GroupId == 0 {
		return false
	}
	if item.Type == "" {
		return false
	}
	return true
}

func (item *CaddieFeeSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CaddieFeeSetting) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFeeSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieFeeSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CaddieFeeSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFeeSetting) FindAll(database *gorm.DB) ([]CaddieFeeSetting, error) {
	db := database.Model(CaddieFeeSetting{})
	list := []CaddieFeeSetting{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.GroupId != 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}

	db.Order("hole asc")

	db.Find(&list)
	return list, db.Error
}

func (item *CaddieFeeSetting) FindList(database *gorm.DB, page Page) ([]CaddieFeeSetting, int64, error) {
	db := database.Model(CaddieFeeSetting{})
	list := []CaddieFeeSetting{}
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

func (item *CaddieFeeSetting) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
