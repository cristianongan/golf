package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
)

type TeeTypeInfo struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"`
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"` // Sân Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`
	Hole       int    `json:"hole"`
	TeeType    string `json:"tee_type" gorm:"type:varchar(20)"`
	Note       string `json:"note"`
	ImageLink  string `json:"image_link" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *TeeTypeInfo) Create() error {
	db := datasources.GetDatabaseAuth()
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *TeeTypeInfo) Update() error {
	db := datasources.GetDatabaseAuth()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TeeTypeInfo) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	err := db.Where(item).First(item).Error
	return err
}

func (item *TeeTypeInfo) FindFirstHaveKey() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *TeeTypeInfo) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(TeeTypeInfo{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TeeTypeInfo) FindList(page Page) ([]TeeTypeInfo, int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Table("tee_type_infos")
	list := []TeeTypeInfo{}
	total := int64(0)
	status := item.Status

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%")
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *TeeTypeInfo) FindALL() ([]TeeTypeInfo, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Table("tee_type_infos")
	list := []TeeTypeInfo{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Find(&list)

	return list, db.Error
}

func (item *TeeTypeInfo) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
