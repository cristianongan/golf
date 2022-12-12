package model_service

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FbPromotionSet
type FbPromotionSet struct {
	models.ModelId
	PartnerUid  string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code        string  `json:"code"`                                       // Mã Set
	Discount    int64   `json:"discount"`
	Note        string  `json:"note"`
	FBList      FBSet   `json:"fb_list,omitempty" gorm:"type:json"` // ds món, liên kết qua fb code
	InputUser   string  `json:"input_user" gorm:"type:varchar(100)"`
	Price       float64 `json:"price"`                                    // Giá Set
	EnglishName string  `json:"english_name" gorm:"type:varchar(256)"`    // Tên Tiếng Anh
	VieName     string  `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Viet
	AccountCode string  `json:"account_code" gorm:"type:varchar(100)"`    // Mã liên kết với Account kế toán
}

type FBItem struct {
	FBCode      string  `json:"fb_code"`
	EnglishName string  `json:"english_name"`
	VieName     string  `json:"vietnamese_name"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit"`
	Type        string  `json:"type"`
	GroupCode   string  `json:"group_code"`
	GroupName   string  `json:"group_name"`
	Quantity    int     `json:"quantity"`
}

type FBSet []FBItem

func (item *FBSet) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item FBSet) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type FbPromotionSetRequest struct {
	FbPromotionSet
	CodeOrName string `form:"code_or_name"`
}

func (item *FbPromotionSet) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}
	return db.Create(item).Error
}

func (item *FbPromotionSet) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *FbPromotionSet) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *FbPromotionSet) Count(database *gorm.DB) (int64, error) {
	db := database.Model(FbPromotionSet{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *FbPromotionSetRequest) FindList(database *gorm.DB, page models.Page) ([]FbPromotionSet, int64, error) {
	db := database.Model(FbPromotionSet{})
	list := []FbPromotionSet{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("fb_promotion_sets.status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("fb_promotion_sets.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("fb_promotion_sets.course_uid = ?", item.CourseUid)
	}
	if item.VieName != "" {
		db = db.Where("fb_promotion_sets.vie_name LIKE ?", "%"+item.VieName+"%")
	}
	if item.CodeOrName != "" {
		query := "fb_promotion_sets.code COLLATE utf8mb4_general_ci LIKE ? OR " +
			"fb_promotion_sets.vie_name COLLATE utf8mb4_general_ci LIKE ? OR " +
			"fb_promotion_sets.english_name COLLATE utf8mb4_general_ci LIKE ?"
		db = db.Where(query, "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *FbPromotionSet) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
