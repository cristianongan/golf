package models

import (
	"start/constants"
	"start/utils"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CaddieFee setting
type CaddieFeeSettingGroup struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Group name setting caddie fee
	FromDate   int64  `json:"from_date" gorm:"index"`                     // Áp dụng từ ngày
	ToDate     int64  `json:"to_date" gorm:"index"`                       // Áp dụng tới ngày
}

func (item *CaddieFeeSettingGroup) IsDuplicated(db *gorm.DB) bool {
	CaddieFeeSettingGroup := CaddieFeeSettingGroup{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Name:       item.Name,
		FromDate:   item.FromDate,
		ToDate:     item.ToDate,
	}

	errFind := CaddieFeeSettingGroup.FindFirst(db)
	if errFind == nil || CaddieFeeSettingGroup.Id > 0 {
		return true
	}
	return false
}

func (item *CaddieFeeSettingGroup) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	return true
}

func (item *CaddieFeeSettingGroup) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CaddieFeeSettingGroup) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFeeSettingGroup) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieFeeSettingGroup) FindFirstByDate(database *gorm.DB, date int64) error {
	db := database

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Where("status = ?", constants.STATUS_ENABLE)
	db = db.Where("from_date < " + strconv.FormatInt(date, 10) + " ")
	db = db.Where("to_date > " + strconv.FormatInt(date, 10) + " ")

	return db.First(item).Error
}

func (item *CaddieFeeSettingGroup) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CaddieFeeSettingGroup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFeeSettingGroup) FindList(database *gorm.DB, page Page, from, to int64) ([]CaddieFeeSettingGroup, int64, error) {
	db := database.Model(CaddieFeeSettingGroup{})
	list := []CaddieFeeSettingGroup{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	//Search With Time
	if from > 0 && to > 0 {
		db = db.Where("from_date < " + strconv.FormatInt(from+30, 10) + " ")
		db = db.Where("to_date > " + strconv.FormatInt(to-30, 10) + " ")
	}
	if from > 0 && to == 0 {
		db = db.Where("from_date < " + strconv.FormatInt(from+30, 10) + " ")
	}
	if from == 0 && to > 0 {
		db = db.Where("to_date > " + strconv.FormatInt(to-30, 10) + " ")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieFeeSettingGroup) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
