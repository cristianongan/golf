package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type BookingServiceItem struct {
	models.ModelId
	ItemId        int64  `json:"item_id"  gorm:"index"`                       // Id item
	ServiceId     string `json:"service_id"`                                  // uid service
	BookingUid    string `json:"booking_uid"  gorm:"type:varchar(100);index"` // Uid booking
	PlayerName    string `json:"player_name"`                                 // Tên người chơi
	Bag           string `json:"bag"`                                         // Golf Bag
	Type          string `json:"type"`                                        // Loại rental, kiosk, proshop,...
	Order         string `json:"order"`                                       // Có thể là mã
	Name          string `json:"name"`
	GroupCode     string `json:"group_code"`
	Quality       int    `json:"quality"` // Số lượng
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	Input         string `json:"input"` // Note
	BillCode      string `json:"bill_code" gorm:"type:varchar(100);index"`
}

// Response cho FE
type BookingServiceItemResponse struct {
	BookingServiceItem
	CheckInTime  int64  `json:"check_in_time"` // Time Check In
	Bag          string `json:"bag"`           // Golf Bag
	CustomerName string `json:"customer_name"` // Tên khách hàng
}

// ------- List Booking service ---------
type ListBookingServiceItems []BookingServiceItem

func (item *ListBookingServiceItems) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingServiceItems) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *BookingServiceItem) IsDuplicated() bool {
	errFind := item.FindFirst()
	return errFind == nil
}

func (item *BookingServiceItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BookingServiceItem) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingServiceItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BookingServiceItem) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BookingServiceItem{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingServiceItem) FindAll() ([]BookingServiceItem, error) {
	db := datasources.GetDatabase().Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	item.Status = ""

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	db = db.Find(&list)

	return list, db.Error
}

func (item *BookingServiceItem) FindList(page models.Page) ([]BookingServiceItemResponse, int64, error) {
	db := datasources.GetDatabase().Model(BookingServiceItem{})
	list := []BookingServiceItemResponse{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.GroupCode != "" {
		db = db.Where("booking_service_items.group_code = ?", item.GroupCode)
	}
	if item.ServiceId != "" {
		db = db.Where("booking_service_items.service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}

	db = db.Joins("JOIN bookings ON bookings.uid = booking_service_items.booking_uid")
	db = db.Select("booking_service_items.*, bookings.bag, bookings.check_in_time, bookings.customer_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

/// ------- BookingServiceItem batch insert to db ------
func (item *BookingServiceItem) BatchInsert(list []BookingServiceItem) error {
	db := datasources.GetDatabase().Table("booking_service_items")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch insert err: ", err.Error())
	}
	return err
}

// ------ Batch Update ------
func (item *BookingServiceItem) BatchUpdate(list []BookingServiceItem) error {
	db := datasources.GetDatabase().Table("booking_service_items")
	var err error
	err = db.Updates(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch update err: ", err.Error())
	}
	return err
}
