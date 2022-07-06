package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type GroupServices struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupCode   string `json:"group_code" gorm:"type:varchar(100)"`        // Mã Group
	GroupName   string `json:"group_name" gorm:"type:varchar(256)"`        // Tên Group
	DetailGroup string `json:"detail_group" gorm:"type:varchar(256)"`      // Tên Group
	Type        string `json:"type" gorm:"type:varchar(100)"`              // Loại service, kiosk, proshop.
}

func (item *GroupServices) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *GroupServices) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GroupServices) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *GroupServices) Count() (int64, error) {
	db := datasources.GetDatabase().Model(GroupServices{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *GroupServices) FindList(page models.Page) ([]GroupServices, int64, error) {
	db := datasources.GetDatabase().Model(GroupServices{})
	list := []GroupServices{}
	total := int64(0)
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.GroupName != "" {
		db = db.Where("group_name LIKE ?", "%"+item.GroupName+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}
	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *GroupServices) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
