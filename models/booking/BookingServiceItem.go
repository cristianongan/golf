package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type BookingServiceItem struct {
	models.ModelId
	ItemId        int64  `json:"item_id"`     // Id item
	BookingUid    string `json:"booking_uid"` // Uid booking
	PlayerName    string `json:"player_name"` // Tên người chơi
	Bag           string `json:"bag"`         // Golf Bag
	Type          string `json:"type"`        // Loại rental, kiosk, proshop,...
	Order         string `json:"order"`       // Có thể là mã
	Name          string `json:"name"`
	GroupCode     string `json:"group_code"`
	Quality       int    `json:"quality"` // Số lượng
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	Input         string `json:"input"` // Note
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
	if errFind == nil {
		return true
	}
	return false
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

func (item *BookingServiceItem) FindList(page models.Page) ([]BookingServiceItem, int64, error) {
	db := datasources.GetDatabase().Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}

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
