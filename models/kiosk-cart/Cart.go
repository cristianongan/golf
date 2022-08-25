package kiosk_cart

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type Cart struct {
	models.ModelId
	PartnerUid  string         `json:"partner_uid"`
	CourseUid   string         `json:"course_uid"`
	KioskCode   string         `json:"kiosk_code"`
	Code        string         `json:"code"`
	GolfBag     string         `json:"golf_bag"`
	BookingDate datatypes.Date `json:"booking_date"`
	BookingUid  string         `json:"booking_uid"`
	BillingCode string         `json:"billing_code" gorm:"default:NONE"`
}

type CartItem struct {
	models.ModelId
	KioskCartId    int64   `json:"kiosk_cart_id"`
	KioskCartCode  string  `json:"kiosk_cart_code"`
	PartnerUid     string  `json:"partner_uid"`
	CourseUid      string  `json:"course_uid"`
	KioskCode      string  `json:"kiosk_code"`
	ItemCode       string  `json:"item_code"`
	Quantity       int64   `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	TotalPrice     float64 `json:"total_price"`
	ActionBy       string  `json:"action_by"`
	DiscountPrice  float64 `json:"discount_price"`
	DiscountType   string  `json:"discount_type"`
	DiscountReason string  `json:"discount_reason"`
	Note           string  `json:"note"`
}

func (_ Cart) TableName() string {
	return "kiosk_cart"
}

func (_ CartItem) TableName() string {
	return "kiosk_cart_item"
}

func (item *Cart) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Code = uid.String()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Cart) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Cart) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

func (item *CartItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CartItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CartItem) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

func (item *CartItem) FindList(page models.Page) ([]CartItem, int64, error) {
	var list []CartItem
	total := int64(0)

	db := datasources.GetDatabase().Model(CartItem{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.KioskCartId != 0 {
		db = db.Where("kiosk_cart_id = ?", item.KioskCartId)
	}

	if item.KioskCartCode != "" {
		db = db.Where("kiosk_cart_code = ?", item.KioskCartCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *CartItem) Delete() error {
	return datasources.GetDatabase().Delete(item).Error
}
