package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"
)

// Proshop
type Proshop struct {
	models.ModelId
	ProShopId     string  `json:"proshop_id" gorm:"type:varchar(100)"`
	PartnerUid    string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string  `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupCode     string  `json:"group_code" gorm:"type:varchar(100)"`
	EnglishName   string  `json:"english_name" gorm:"type:varchar(256)"`
	VieName       string  `json:"vietnamese_name" gorm:"type:varchar(256)"`
	Brand         string  `json:"brand" gorm:"type:varchar(100)"`
	Barcode       string  `json:"barcode" gorm:"type:varchar(100)"`
	AccountCode   string  `json:"account_code" gorm:"type:varchar(100)"` // Mã liên kết với Account kế toán
	Price         float64 `json:"price"`
	Unit          string  `json:"unit" gorm:"type:varchar(100)"`
	Note          string  `json:"note" gorm:"type:varchar(256)"`
	NetCost       float64 `json:"net_cost" gorm:"type:varchar(100)"` // Net cost tự tính từ Cost Price ko bao gồm 10% VAT
	CostPrice     float64 `json:"cost_price"`
	ProPrice      float64 `json:"pro_price"`
	PeopleDeposit string  `json:"people_deposit" gorm:"type:varchar(100)"`
	ForKiosk      bool    `json:"for_kiosk"`
	IsDeposit     bool    `json:"is_deposit"`
	IsInventory   bool    `json:"is_inventory"`                         // Có trong kho
	Name          string  `json:"name" gorm:"type:varchar(256)"`        // Tên sp default
	UserUpdate    string  `json:"user_update" gorm:"type:varchar(256)"` // Người update cuối cùng
}

type ProshopRequest struct {
	Proshop
	GroupName string `json:"group_name"`
}

func (item *Proshop) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Proshop) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Proshop) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Proshop) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Proshop{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *ProshopRequest) FindList(page models.Page) ([]ProshopRequest, int64, error) {
	db := datasources.GetDatabase().Model(Proshop{})
	list := []ProshopRequest{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("proshops.status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("proshops.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("proshops.course_uid = ?", item.CourseUid)
	}
	if item.EnglishName != "" {
		db = db.Where("proshops.english_name LIKE ?", "%"+item.EnglishName+"%")
	}
	if item.VieName != "" {
		db = db.Where("proshops.vie_name LIKE ?", "%"+item.VieName+"%")
	}
	if item.ProShopId != "" {
		db = db.Where("proshops.proshop_id = ?", item.ProShopId)
	}
	if item.GroupName != "" {
		db = db.Where("proshops.group_name = ?", item.GroupName)
	}

	db = db.Joins("JOIN group_services ON proshops.group_code = group_services.group_code AND " +
		"proshops.partner_uid = group_services.partner_uid AND " +
		"proshops.course_uid = group_services.course_uid")
	db = db.Select("proshops.*, group_services.group_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Proshop) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
