package model_service

import (
	"errors"
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FoodBeverage
type FoodBeverage struct {
	models.ModelId
	PartnerUid    string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	FBCode        string  `json:"fb_code" gorm:"type:varchar(100)"`
	EnglishName   string  `json:"english_name" gorm:"type:varchar(256)"`    // Tên Tiếng Anh
	VieName       string  `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Viet
	Barcode       string  `json:"barcode"`
	AccountCode   string  `json:"account_code" gorm:"type:varchar(100)"` // Mã liên kết với Account kế toán
	GroupCode     string  `json:"group_code" gorm:"type:varchar(100);index"`
	GroupName     string  `json:"group_name" gorm:"type:varchar(100)"`
	Unit          string  `json:"unit" gorm:"type:varchar(100)"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost" gorm:"type:varchar(100)"` // Net cost tự tính từ Cost Price ko bao gồm 10% VAT
	CostPrice     float64 `json:"cost_price"`
	BarBeerPrice  float64 `json:"bar_beer_price"`
	InternalPrice float64 `json:"internal_price"` // Giá nội bộ là giá dành cho nhân viên ăn uống và sử dụng
	Note          string  `json:"note" gorm:"type:varchar(256)"`
	AloneKiosk    string  `json:"alone_kiosk" gorm:"type:varchar(100)"`
	ForKiosk      bool    `json:"for_kiosk"`
	OpenFB        bool    `json:"open_fb"`
	InMenuSet     bool    `json:"in_menu_set"`  // Món trong combo
	IsInventory   bool    `json:"is_inventory"` // Có trong kho
	IsKitchen     bool    `json:"is_kitchen"`
	Name          string  `json:"name" gorm:"type:varchar(256)"`         // Tên
	UserUpdate    string  `json:"user_update" gorm:"type:varchar(256)"`  // Người update cuối cùng
	Type          string  `json:"type" gorm:"type:varchar(256)"`         // sub type của F&B
	HotKitchen    *bool   `json:"hot_kitchen" gorm:"type:varchar(256)"`  // Món ăn chế biến trong bếp nóng
	ColdKitchen   *bool   `json:"cold_kitchen" gorm:"type:varchar(256)"` // Món ăn chế biến trong bếp lạnh như salad, gỏi

}
type FoodBeverageResponse struct {
	FoodBeverage
	GroupName string `json:"group_name"`
}
type FoodBeverageRequest struct {
	FoodBeverage
	CodeOrName string   `form:"code_or_name"`
	GroupName  string   `json:"group_name"`
	FBCodeList []string `form:"fb_code_list"`
}

func (item *FoodBeverage) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *FoodBeverage) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *FoodBeverage) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *FoodBeverage) FindFirstInKiosk(db *gorm.DB, kioskId int64) error {

	return db.Where(item).Where("(alone_kiosk = ? OR for_kiosk = ?)", kioskId, 1).First(item).Error
}

func (item *FoodBeverage) Count(database *gorm.DB) (int64, error) {
	db := database.Model(FoodBeverage{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *FoodBeverageRequest) FindList(database *gorm.DB, page models.Page) ([]FoodBeverageResponse, int64, error) {
	db := database.Model(FoodBeverage{})
	list := []FoodBeverageResponse{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("food_beverages.status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("food_beverages.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("food_beverages.course_uid = ?", item.CourseUid)
	}
	if item.EnglishName != "" {
		db = db.Or("food_beverages.english_name LIKE ?", "%"+item.EnglishName+"%")
	}
	if item.FBCode != "" {
		db = db.Or("food_beverages.fb_code = ?", item.FBCode)
	}
	if item.VieName != "" {
		db = db.Where("food_beverages.vie_name LIKE ?", "%"+item.VieName+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("food_beverages.group_code = ?", item.GroupCode)
	}
	if item.Type != "" {
		db = db.Where("food_beverages.type = ?", item.Type)
	}
	if item.CodeOrName != "" {
		query := "food_beverages.fb_code COLLATE utf8mb4_general_ci LIKE ? OR " +
			"food_beverages.vie_name COLLATE utf8mb4_general_ci LIKE ? OR " +
			"food_beverages.english_name COLLATE utf8mb4_general_ci LIKE ?"
		db = db.Where(query, "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%", "%"+item.CodeOrName+"%")
	}
	if len(item.FBCodeList) != 0 {
		db = db.Where("food_beverages.fb_code IN (?)", item.FBCodeList)
	}

	db = db.Joins("JOIN group_services ON food_beverages.group_code = group_services.group_code AND " +
		"food_beverages.partner_uid = group_services.partner_uid AND " +
		"food_beverages.course_uid = group_services.course_uid")
	db = db.Select("food_beverages.*, group_services.group_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *FoodBeverage) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
