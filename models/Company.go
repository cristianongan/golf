package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"
)

// Company
type Company struct {
	ModelId
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code            string `json:"code" gorm:"type:varchar(256);index"`        // Mã công ty
	Name            string `json:"name" gorm:"type:varchar(256)"`
	Address         string `json:"addresss" gorm:"type:varchar(500)"`
	Phone           string `json:"phone" gorm:"type:varchar(30);index"`
	Fax             string `json:"fax" gorm:"type:varchar(30);index"`
	FaxCode         string `json:"fax_code" gorm:"type:varchar(30);index"`
	CompanyTypeId   int64  `json:"company_type_id" gorm:"index"`
	CompanyTypeName string `json:"company_type_name" gorm:"type:varchar(300)"`
}

func (item *Company) IsDuplicated() bool {
	company := Company{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Code:       item.Code,
	}

	errFind := company.FindFirst()
	if errFind == nil || company.Id > 0 {
		return true
	}
	return false
}

// ======= CRUD ===========
func (item *Company) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Company) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Company) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Company) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Company{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Company) FindList(page Page) ([]Company, int64, error) {
	db := datasources.GetDatabase().Model(Company{})
	list := []Company{}
	total := int64(0)
	status := item.Status
	item.Status = ""

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
	if item.CompanyTypeId > 0 {
		db = db.Where("type = ?", item.CompanyTypeId)
	}
	if item.Phone != "" {
		db = db.Where("phone LIKE ?", "%"+item.Phone+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Company) Delete() error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
