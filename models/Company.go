package models

import (
	"encoding/json"
	"errors"
	"log"
	"start/constants"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

// Company
type Company struct {
	ModelId
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code            string `json:"code" gorm:"type:varchar(256);index"`        // Mã công ty
	Name            string `json:"name" gorm:"type:varchar(256)"`
	Address         string `json:"address" gorm:"type:varchar(500)"`
	Phone           string `json:"phone" gorm:"type:varchar(30);index"`
	Fax             string `json:"fax" gorm:"type:varchar(30);index"`
	FaxCode         string `json:"fax_code" gorm:"type:varchar(30);index"`
	CompanyTypeId   int64  `json:"company_type_id" gorm:"index"`
	CompanyTypeName string `json:"company_type_name" gorm:"type:varchar(300)"`
}

/*
 Clone object
*/
func (item *Company) CloneCompany() Company {
	copyCompany := Company{}
	bData, errM := json.Marshal(&item)
	if errM != nil {
		log.Println("CloneCompany errM", errM.Error())
	}
	errUnM := json.Unmarshal(bData, &copyCompany)
	if errUnM != nil {
		log.Println("CloneCompany errUnM", errUnM.Error())
	}

	return copyCompany
}

func (item *Company) IsDuplicated(db *gorm.DB) bool {
	company := Company{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Code:       item.Code,
	}

	errFind := company.FindFirst(db)
	if errFind == nil || company.Id > 0 {
		return true
	}
	return false
}

// ======= CRUD ===========
func (item *Company) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Company) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Company) FindFirst(db *gorm.DB) error {

	return db.Where(item).First(item).Error
}

func (item *Company) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Company{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Company) FindList(database *gorm.DB, page Page) ([]Company, int64, error) {
	db := database.Model(Company{})
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
	if item.Code != "" {
		db = db.Where("code LIKE ?", "%"+item.Code+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Company) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
