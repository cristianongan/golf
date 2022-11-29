package model_service

import (
	"errors"
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Rental
type Rental struct {
	models.ModelId
	RentalId    string  `json:"rental_id" gorm:"type:varchar(100);index"`
	PartnerUid  string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	EnglishName string  `json:"english_name" gorm:"type:varchar(256)"`      // Tên Tiếng Anh
	RenPos      string  `json:"ren_pos" gorm:"type:varchar(100)"`
	VieName     string  `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Anh
	SystemCode  string  `json:"system_code" gorm:"type:varchar(100)"`
	Unit        string  `json:"unit" gorm:"type:varchar(100)"`
	Price       float64 `json:"price"`
	ByHoles     bool    `json:"by_holes"`
	ForPos      bool    `json:"for_pos"`
	OnlyForRen  bool    `json:"only_for_ren"`
	InputUser   string  `json:"input_user" gorm:"type:varchar(100)"`
	Name        string  `json:"name" gorm:"type:varchar(256)"` // Tên
	Type        string  `json:"type" gorm:"type:varchar(50)"`  // sub type của Rental
	GroupCode   string  `json:"group_code" gorm:"type:varchar(100);index"`
	GroupName   string  `json:"group_name" gorm:"type:varchar(100)"`
	IsDriving   *bool   `json:"is_driving" gorm:"default:0"`
	Rate        string  `json:"rate" gorm:"type:varchar(100)"`
}

type RentalRequest struct {
	Rental
	CodeOrName string `form:"code_or_name"`
	GroupName  string `json:"group_name"`
}

func (item *Rental) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Rental) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Rental) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Rental) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Rental{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *RentalRequest) FindList(database *gorm.DB, page models.Page) ([]RentalRequest, int64, error) {
	db := database.Model(Rental{})
	list := []RentalRequest{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("rentals.status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("rentals.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("rentals.course_uid = ?", item.CourseUid)
	}
	if item.EnglishName != "" {
		db = db.Where("rentals.english_name LIKE ?", "%"+item.EnglishName+"%")
	}
	if item.VieName != "" {
		db = db.Where("rentals.vie_name LIKE ?", "%"+item.VieName+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("rentals.group_code = ?", item.GroupCode)
	}
	if item.SystemCode != "" {
		db = db.Where("rentals.system_code = ?", item.SystemCode)
	}
	if item.GroupCode != "" {
		db = db.Where("rentals.group_code = ?", item.GroupCode)
	}
	if item.Type != "" {
		db = db.Where("rentals.type = ?", item.Type)
	}
	if item.IsDriving != nil {
		db = db.Where("rentals.is_driving = ?", item.IsDriving)
	}
	if item.CodeOrName != "" {
		query := "rentals.rental_id COLLATE utf8mb4_general_ci LIKE ? OR " +
			"rentals.vie_name COLLATE utf8mb4_general_ci LIKE ? OR " +
			"rentals.english_name COLLATE utf8mb4_general_ci LIKE ?"
		db = db.Where(query, "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%")
	}

	db = db.Joins("JOIN group_services ON rentals.group_code = group_services.group_code AND " +
		"rentals.partner_uid = group_services.partner_uid AND " +
		"rentals.course_uid = group_services.course_uid")
	db = db.Select("rentals.*, group_services.group_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Rental) FindALL(database *gorm.DB) ([]Rental, int64, error) {
	db := database.Model(Rental{})
	list := []Rental{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("rentals.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("rentals.course_uid = ?", item.CourseUid)
	}
	if item.IsDriving != nil {
		db = db.Where("rentals.is_driving = ?", item.IsDriving)
	}

	db.Count(&total)

	db = db.Find(&list)

	return list, total, db.Error
}

func (item *Rental) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
