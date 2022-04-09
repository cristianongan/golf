package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Booking
type Booking struct {
	models.Model
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf

	CreatedDate string `json:"created_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022

	Bag            string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	Hole           int    `json:"hole"`                                       // Số hố
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(200);index"` // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Guest Style Name

	CardId        string `json:"card_id" gorm:"index"`                           // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // MemberCard Uid, Uid object trong Database
	CustomerName  string `json:"customer_name" gorm:"type:varchar(256)"`         // Tên khách hàng
	CustomerUid   string `json:"customer_uid" gorm:"type:varchar(256);index"`    // Uid khách hàng
	// Thêm customer info

	CheckInTime  int64  `json:"check_in_time"`                          // Time Check In
	CheckOutTime int64  `json:"check_out_time"`                         // Time Check Out
	TeeType      string `json:"tee_type" gorm:"type:varchar(50);index"` // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath      string `json:"tee_path" gorm:"type:varchar(50);index"` // MORNING, NOON, NIGHT
	TurnTime     string `json:"turn_time" gorm:"type:varchar(30)"`      // Ex: 16:26
	TeeTime      string `json:"tee_time" gorm:"type:varchar(30)"`       // Ex: 16:26 Tee time là thời gian tee off dự kiến
	TeeOffTime   string `json:"tee_off_time" gorm:"type:varchar(30)"`   // Ex: 16:26 Là thời gian thực tế phát bóng
	RowIndex     int    `json:"row_index"`                              // index trong Flight

	PriceDetail BookingPriceDetail `json:"price_detail" gorm:"type:varchar(500)"` // Thông tin phí++: Tính toán lại phí Service items, Tiền cho Subbag
	GolfFee     BookingGolfFee     `json:"golf_fee" gorm:"type:varchar(200)"`     // Thông tin Golf Fee

	Note   string `json:"note" gorm:"type:varchar(500)"`   // Note
	Locker string `json:"locker" gorm:"type:varchar(100)"` // Locker mã số tủ gửi đồ

	CmsUser    string `json:"cms_user" gorm:"type:varchar(100)"`     // Cms User
	CmsUserLog string `json:"cms_user_log" gorm:"type:varchar(200)"` // Cms User Log

	BookingServiceItems utils.ListBookingServiceItems `json:"booking_service_items" gorm:"type:varchar(1000)"` // List item service: rental, proshop, restaurant, kiosk

	// TODO
	// Caddie Info
	CaddieId int64 `json:"caddie_id" gorm:"index"`

	// Buggy Info
	BuggyId int64 `json:"buggy_id" gorm:"index"`

	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags" gorm:"type:varchar(1000)"` // List Sub Bags

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags" gorm:"type:varchar(200)"` // List Main Bags, thêm main bag sẽ thanh toán những cái gì
	// Main bug for Pay: Mặc định thanh toán all, Nếu có trong list này thì k thanh toán
	MainBagNoPay utils.ListString `json:"main_bag_no_pay" gorm:"type:varchar(100)"` // Main Bag không thanh toán những phần này
}

type BookingGolfFee struct {
	CaddieFee int64 `json:"caddie_fee"`
	BuggyFee  int64 `json:"buggy_fee"`
	GreenFee  int64 `json:"green_fee"`
}

func (item *BookingGolfFee) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingGolfFee) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type BookingPriceDetail struct {
	Kiosk int64 `json:"kiosk"`
}

func (item *BookingPriceDetail) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingPriceDetail) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *Booking) IsDuplicated() bool {
	booking := Booking{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		TeeTime:     item.TeeTime,
		TurnTime:    item.TurnTime,
		CreatedDate: item.CreatedDate,
	}

	errFind := booking.FindFirst()
	if errFind == nil || booking.Uid != "" {
		return true
	}
	return false
}

func (item *Booking) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = item.CourseUid + "-" + utils.HashCodeUuid(uid.String())
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Booking) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Booking) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Booking) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Booking) FindList(page models.Page) ([]Booking, int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	list := []Booking{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Booking) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
