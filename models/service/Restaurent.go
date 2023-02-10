package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strconv"
	"strings"
)

// Restaurent
type Restaurent struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Tên
	Type       string `json:"type" gorm:"type:varchar(50)"`               // Loại rental, kiosk, proshop,...
	Code       string `json:"code" gorm:"type:varchar(100)"`
	GroupId    int64  `json:"group_id" gorm:"index"`
	GroupCode  string `json:"group_code" gorm:"type:varchar(100);index"`
	GroupName  string `json:"group_name" gorm:"type:varchar(256)"`
	Unit       string `json:"unit" gorm:"type:varchar(100);index"`
	Price      int64  `json:"price"`
}

func (item *Restaurent) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.Type == "" {
		return false
	}
	if item.Code == "" {
		return false
	}
	if item.GroupId <= 0 {
		return false
	}
	return true
}

func (item *Restaurent) Create() error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Restaurent) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Restaurent) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Restaurent) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Restaurent{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Restaurent) FindList(page models.Page) ([]Restaurent, int64, error) {
	db := datasources.GetDatabase().Model(Restaurent{})
	list := []Restaurent{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}
	if item.GroupId > 0 {
		db = db.Where("group_id = ?", strconv.FormatInt(item.GroupId, 10))
	}
	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Restaurent) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
