package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"
)

// FoodBeverage
type FoodBeverage struct {
	models.ModelId
	PartnerUid    string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupId       string  `json:"group_id" gorm:"index"`
	GroupName     string  `json:"group_name" gorm:"type:varchar(256)"`
	FBCode        string  `json:"fb_code" gorm:"type:varchar(100)"`
	EnglishName   string  `json:"english_name" gorm:"type:varchar(256)"`    // Tên Tiếng Anh
	VieName       string  `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Anh
	Unit          string  `json:"unit" gorm:"type:varchar(100)"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost" gorm:"type:varchar(100)"` // Net cost tự tính từ Cost Price ko bao gồm 10% VAT
	CostPrice     float64 `json:"cost_price"`
	Barcode       string  `json:"barcode"`
	AccountCode   string  `json:"account_code" gorm:"type:varchar(100)"` // Mã liên kết với Account kế toán
	BarBeerPrice  float64 `json:"bar_beer_price"`
	Note          string  `json:"note" gorm:"type:varchar(256)"`
	ForKiosk      bool    `json:"for_kiosk"`
	OpenFB        float64 `json:"open_fb"`
	AloneKiosk    string  `json:"alone_kiosk" gorm:"type:varchar(100)"`
	InMenuSet     bool    `json:"in_menu_set"`                   // Món trong combo
	IsInventory   bool    `json:"is_inventory"`                  // Có trong kho
	InternalPrice float64 `json:"internal_price"`                // Giá nội bộ là giá dành cho nhân viên ăn uống và sử dụng
	Name          string  `json:"name" gorm:"type:varchar(256)"` // Tên
}

func (item *FoodBeverage) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *FoodBeverage) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *FoodBeverage) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *FoodBeverage) Count() (int64, error) {
	db := datasources.GetDatabase().Model(FoodBeverage{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *FoodBeverage) FindList(page models.Page) ([]FoodBeverage, int64, error) {
	db := datasources.GetDatabase().Model(FoodBeverage{})
	list := []FoodBeverage{}
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
	if item.FBCode != "" {
		db = db.Where("fb_code = ?", item.FBCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *FoodBeverage) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
