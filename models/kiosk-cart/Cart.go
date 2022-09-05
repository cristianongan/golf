package kiosk_cart

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

/*
Giỏ Hàng
*/
type Cart struct {
	models.ModelId
	PartnerUid  string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng golf
	CourseUid   string         `json:"course_uid" gorm:"type:varchar(150);index"`  // Sân golf
	KioskCode   int64          `json:"kiosk_code" gorm:"index"`                    // Mã của kiosk
	Code        string         `json:"code" gorm:"type:varchar(100);index"`        // Mã giỏ hàng
	GolfBag     string         `json:"golf_bag" gorm:"type:varchar(100);index"`    // Số bag order
	BookingDate datatypes.Date `json:"booking_date"`                               // Ngày order
	BookingUid  string         `json:"booking_uid" gorm:"type:varchar(100)"`       // Booking uid
	BillingCode string         `json:"billing_code" gorm:"default:NONE"`           // Mã hóa đơn
}

/*
Các sản phẩm trong giỏ hàng
*/
type CartItem struct {
	models.ModelId
	KioskCartId    int64   `json:"kiosk_cart_id" gorm:"type:varchar(100);index"`   // Mã id giỏ hàng
	KioskCartCode  string  `json:"kiosk_cart_code" gorm:"type:varchar(100);index"` // Mã code giỏ hàng
	PartnerUid     string  `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hãng golf
	CourseUid      string  `json:"course_uid" gorm:"type:varchar(150);index"`      // Sân golf
	KioskCode      int64   `json:"kiosk_code" gorm:"index"`                        // Mã của kiosk
	ItemCode       string  `json:"item_code" gorm:"type:varchar(100);index"`       // Mã sản phẩm
	ItemGroupId    string  `json:"item_group_id" gorm:"type:varchar(100);index"`   // Nhóm sản phẩm
	ItemName       string  `json:"item_name" gorm:"type:varchar(250)"`             // Tên sản phẩm
	Quantity       int64   `json:"quantity"`                                       // Số lượng
	UnitPrice      float64 `json:"unit_price"`                                     // Giá sản phẩm
	TotalPrice     float64 `json:"total_price"`                                    // Tổng giá tiền
	ActionBy       string  `json:"action_by" gorm:"type:varchar(250);index"`       // Người tạo
	DiscountPrice  float64 `json:"discount_price"`                                 // Giảm giá
	DiscountType   string  `json:"discount_type" gorm:"type:varchar(100);index"`   // ???
	DiscountReason string  `json:"discount_reason" gorm:"type:varchar(250);index"` // Lý do giảm giá
	Note           string  `json:"note" gorm:"type:varchar(250);index"`            // lưu ý của khách hàng
}

func (_ Cart) TableName() string {
	return "kiosk_cart"
}

func (_ CartItem) TableName() string {
	return "kiosk_cart_item"
}

// Cart
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

func (item *Cart) FindList(page models.Page) ([]Cart, int64, error) {
	var list []Cart
	total := int64(0)

	db := datasources.GetDatabase().Model(Cart{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.KioskCode != 0 {
		db = db.Where("kiosk_cart_code = ?", item.KioskCode)
	}

	db = db.Where("booking_date = ?", item.BookingDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

// CartItem
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

func (item *CartItem) FindBestCartItem(page models.Page) ([]CartItem, int64, error) {
	now := time.Now().Format("02/01/2006")

	from, _ := time.Parse("02/01/2006 15:04:05", now+" 17:00:00")

	db := datasources.GetDatabase().Model(CartItem{})
	list := []CartItem{}
	total := int64(0)

	db.Select("*, sum(quantity) as sale_quantity")

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.KioskCode != 0 {
		db = db.Where("kiosk_code = ?", item.KioskCode)
	}
	if item.ItemGroupId != "" {
		db = db.Where("item_group_id = ?", item.ItemGroupId)
	}

	db = db.Where("created_at >= ?", from.AddDate(0, 0, -8).Unix())

	db.Group("item_code")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
