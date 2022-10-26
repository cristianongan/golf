package model_role

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Role
type Role struct {
	Id          int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key" json:"id"`
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Status      string `json:"status" gorm:"index;type:varchar(50)"`       // ENABLE, DISABLE, TESTING
	Name        string `json:"name" gorm:"type:varchar(200)"`              // Name Role
	Description string `json:"description" gorm:"type:varchar(200)"`       // description
}

// ======= CRUD ===========
func (item *Role) Create(db *gorm.DB) error {

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Role) Update() error {
	db := datasources.GetDatabaseRole()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Role) FindFirst() error {
	db := datasources.GetDatabaseRole()
	return db.Where(item).First(item).Error
}

func (item *Role) Count() (int64, error) {
	database := datasources.GetDatabaseRole()
	db := database.Model(Role{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Role) FindList(page models.Page) ([]Role, int64, error) {
	database := datasources.GetDatabaseRole()
	db := database.Model(Role{})
	list := []Role{}
	total := int64(0)
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
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

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Role) Delete() error {
	db := datasources.GetDatabaseRole()
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
