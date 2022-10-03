package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/models"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BookingServiceItem struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hãng golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(150);index"`   // Sân golf
	ItemId         int64  `json:"item_id"  gorm:"index"`                       // Id item
	Unit           string `json:"unit"  gorm:"type:varchar(100)"`              // Unit của item
	ServiceId      string `json:"service_id"  gorm:"type:varchar(100)"`        // uid service
	ServiceType    string `json:"service_type"  gorm:"type:varchar(100)"`      // Loại service gồm FB, Rental, Proshop
	BookingUid     string `json:"booking_uid"  gorm:"type:varchar(100);index"` // Uid booking
	PlayerName     string `json:"player_name" gorm:"type:varchar(256)"`        // Tên người chơi
	Bag            string `json:"bag" gorm:"type:varchar(50)"`                 // Golf Bag
	Type           string `json:"type" gorm:"type:varchar(50)"`                // Loại rental, kiosk, proshop,...
	ItemCode       string `json:"item_code"  gorm:"type:varchar(100)"`         // Mã code của item
	Name           string `json:"name" gorm:"type:varchar(256)"`
	EngName        string `json:"eng_name" gorm:"type:varchar(256)"` // Tên tiếng anh của sản phẩm
	GroupCode      string `json:"group_code" gorm:"type:varchar(100)"`
	Quality        int    `json:"quality"` // Số lượng
	UnitPrice      int64  `json:"unit_price"`
	DiscountType   string `json:"discount_type" gorm:"type:varchar(50)"`
	DiscountValue  int64  `json:"discount_value"`
	DiscountReason string `json:"discount_reason" gorm:"type:varchar(50)"` // Lý do giảm giá
	Amount         int64  `json:"amount"`
	UserAction     string `json:"user_action" gorm:"type:varchar(100)"` // Người tạo
	Input          string `json:"input" gorm:"type:varchar(300)"`       // Note
	BillCode       string `json:"bill_code" gorm:"type:varchar(100);index"`
	ServiceBill    int64  `json:"service_bill" gorm:"index"`               // id service cart
	SaleQuantity   int64  `json:"sale_quantity"`                           // tổng số lượng bán được
	Location       string `json:"location" gorm:"type:varchar(100);index"` // Dc add từ đâu
}

// Response cho FE
type BookingServiceItemResponse struct {
	BookingServiceItem
	CheckInTime int64 `json:"check_in_time"` // Time Check In
}

// ------- List Booking service ---------
type ListBookingServiceItems []BookingServiceItem

func (item *ListBookingServiceItems) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingServiceItems) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *BookingServiceItem) IsDuplicated(db *gorm.DB) bool {
	errFind := item.FindFirst(db)
	return errFind == nil
}

func (item *BookingServiceItem) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingServiceItem) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingServiceItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingServiceItem) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingServiceItem{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingServiceItem) FindAll(database *gorm.DB) ([]BookingServiceItem, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	item.Status = ""

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.ServiceBill > 0 {
		db = db.Where("service_bill = ?", item.ServiceBill)
	}

	db = db.Find(&list)

	return list, db.Error
}

func (item *BookingServiceItem) FindList(database *gorm.DB, page models.Page) ([]BookingServiceItem, int64, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	total := int64(0)

	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}
	if item.ServiceId != "" {
		db = db.Where("service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}
	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) FindListWithBooking(database *gorm.DB, page models.Page, fromDate int64, toDate int64) ([]BookingServiceItemResponse, int64, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItemResponse{}
	total := int64(0)

	if item.GroupCode != "" {
		db = db.Where("booking_service_items.group_code = ?", item.GroupCode)
	}
	if item.ServiceId != "" {
		db = db.Where("booking_service_items.service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}
	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	if fromDate > 0 {
		db = db.Where("booking_service_items.created_at >= ?", fromDate)
	}

	if toDate > 0 {
		db = db.Where("booking_service_items.created_at <= ?", toDate)
	}

	db = db.Joins("JOIN bookings ON bookings.uid = booking_service_items.booking_uid")
	db = db.Select("booking_service_items.*, bookings.check_in_time")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

// / ------- BookingServiceItem batch insert to db ------
func (item *BookingServiceItem) BatchInsert(database *gorm.DB, list []BookingServiceItem) error {
	db := database.Table("booking_service_items")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch insert err: ", err.Error())
	}
	return err
}

// ------ Batch Update ------
func (item *BookingServiceItem) BatchUpdate(database *gorm.DB, list []BookingServiceItem) error {
	db := database.Table("booking_service_items")
	var err error
	err = db.Updates(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch update err: ", err.Error())
	}
	return err
}

// ------ Find Best Item ------
func (item *BookingServiceItem) FindBestCartItem(database *gorm.DB, page models.Page) ([]BookingServiceItem, int64, error) {
	now := time.Now().Format("02/01/2006")

	from, _ := time.Parse("02/01/2006 15:04:05", now+" 17:00:00")

	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	total := int64(0)

	db.Select("*, sum(quality) as sale_quantity")

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.ServiceId != "" {
		db = db.Where("service_id = ?", item.ServiceId)
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}

	db = db.Where("created_at >= ?", from.AddDate(0, 0, -8).Unix())

	db.Group("item_code")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
