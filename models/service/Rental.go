package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"
)

// Rental
type Rental struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	EnglishName string `json:"english_name" gorm:"type:varchar(256)"`      // Tên Tiếng Anh
	RenPos      string `json:"ren_pos" gorm:"type:varchar(100)"`
	VieName     string `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Anh
	Type        string `json:"type" gorm:"type:varchar(50)"`             // Loại rental, kiosk, proshop,...
	SystemCode  string `json:"system_code" gorm:"type:varchar(100)"`
	GroupCode   string `json:"group_code" gorm:"type:varchar(100);index"`
	Unit        string `json:"unit" gorm:"type:varchar(100)"`
	Price       int64  `json:"price"`
	ByHoles     bool   `json:"by_holes"`
	ForPos      bool   `json:"for_pos"`
	OnlyForRen  bool   `json:"only_for_ren"`
	InputUser   string `json:"input_user" gorm:"type:varchar(100)"`
	Name        string `json:"name" gorm:"type:varchar(256)"` // Tên
}

func (item *Rental) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Rental) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Rental) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Rental) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Rental{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Rental) FindList(page models.Page) ([]Rental, int64, error) {
	db := datasources.GetDatabase().Model(Rental{})
	list := []Rental{}
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
	if item.EnglishName != "" {
		db = db.Where("english_name LIKE ?", "%"+item.EnglishName+"%")
	}
	if item.VieName != "" {
		db = db.Where("vie_name LIKE ?", "%"+item.VieName+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}
	if item.SystemCode != "" {
		db = db.Where("code = ?", item.SystemCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Rental) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
