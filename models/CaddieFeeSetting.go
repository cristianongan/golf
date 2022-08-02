package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// CaddieFee setting
type CaddieFeeSetting struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupId    int64  `json:"group_id" gorm:"index"`                      // Id nhóm setting
	Hole       int64  `json:"hole"`                                       // số hố
	Fee        int64  `json:"fee"`                                        // phí tương ứng
	Type       string `json:"type" gorm:"type:varchar(256)"`              // Type setting caddie fee
}

func (item *CaddieFeeSetting) IsDuplicated() bool {
	CaddieFeeSetting := CaddieFeeSetting{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		GroupId:    item.GroupId,
		Hole:       item.Hole,
	}

	errFind := CaddieFeeSetting.FindFirst()
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
	if item.Type == "" {
		return false
	}
	return true
}

func (item *CaddieFeeSetting) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieFeeSetting) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFeeSetting) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieFeeSetting) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CaddieFeeSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFeeSetting) FindList(page Page) ([]CaddieFeeSetting, int64, error) {
	db := datasources.GetDatabase().Model(CaddieFeeSetting{})
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

func (item *CaddieFeeSetting) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
