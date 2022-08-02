package models

import (
	"start/constants"
	"start/datasources"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func (item *CaddieFeeSettingGroup) IsDuplicated() bool {
	CaddieFeeSettingGroup := CaddieFeeSettingGroup{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Name:       item.Name,
		FromDate:   item.FromDate,
		ToDate:     item.ToDate,
	}

	errFind := CaddieFeeSettingGroup.FindFirst()
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

func (item *CaddieFeeSettingGroup) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieFeeSettingGroup) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFeeSettingGroup) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieFeeSettingGroup) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CaddieFeeSettingGroup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFeeSettingGroup) FindList(page Page, from, to int64) ([]CaddieFeeSettingGroup, int64, error) {
	db := datasources.GetDatabase().Model(CaddieFeeSettingGroup{})
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

func (item *CaddieFeeSettingGroup) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
