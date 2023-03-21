package model_booking

import (
	"start/constants"
	"start/models"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BookingWaiting struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CourseType  string `json:"course_type" gorm:"type:varchar(100)"`       // A,B,C
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022

	Hole           int    `json:"hole"`                                       // Số hố check in
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(200);index"` // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Guest Style Name

	// MemberCard
	CardId        string `json:"card_id" gorm:"index"`                           // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // MemberCard Uid, Uid object trong Database

	// Thêm customer info
	CustomerBookingName  string        `json:"customer_booking_name" gorm:"type:varchar(256)"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string        `json:"customer_booking_phone" gorm:"type:varchar(100)"` // SDT khách hàng đặt booking
	CustomerName         string        `json:"customer_name" gorm:"type:varchar(256)"`          // Tên khách hàng
	CustomerUid          string        `json:"customer_uid" gorm:"type:varchar(256);index"`     // Uid khách hàng
	CustomerType         string        `json:"customer_type" gorm:"type:varchar(256)"`          // Loai khach hang: Member, Guest, Visitor...
	CustomerInfo         *CustomerInfo `json:"customer_info,omitempty" gorm:"type:json"`        // Customer Info

	TeeType  string `json:"tee_type" gorm:"type:varchar(50)"`  // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath  string `json:"tee_path" gorm:"type:varchar(50)"`  // MORNING, NOON, NIGHT
	TurnTime string `json:"turn_time" gorm:"type:varchar(30)"` // Ex: 16:26
	TeeTime  string `json:"tee_time" gorm:"type:varchar(30)"`  // Ex: 16:26 Tee time là thời gian tee off dự kiến

	Note string `json:"note" gorm:"type:varchar(500)"` // Note of Booking

	CmsUser    string `json:"cms_user" gorm:"type:varchar(100)"`     // Cms User
	CmsUserLog string `json:"cms_user_log" gorm:"type:varchar(200)"` // Cms User Log

	// Caddie Id
	CaddieBooking string `json:"caddie_booking" gorm:"type:varchar(50)"`

	// Agency Id
	AgencyId   int64          `json:"agency_id" gorm:"index"` // Agency
	AgencyInfo *BookingAgency `json:"agency_info" gorm:"type:json"`

	BookingCode string `json:"booking_code" gorm:"type:varchar(100);index"` // cho case tạo nhiều booking có cùng booking code

	MemberUidOfGuest  string `json:"member_uid_of_guest" gorm:"type:varchar(50);index"` // Member của Guest đến chơi cùng
	MemberNameOfGuest string `json:"member_name_of_guest" gorm:"type:varchar(200)"`     // Member của Guest đến chơi cùng
}

func (item *BookingWaiting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingWaiting) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingWaiting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingWaiting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingWaiting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingWaiting) FindList(database *gorm.DB, page models.Page) ([]BookingWaiting, int64, error) {
	db := database.Model(BookingWaiting{})
	list := []BookingWaiting{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	if item.CustomerBookingPhone != "" {
		db = db.Where("player_contact COLLATE utf8mb4_general_ci LIKE ?", "%"+item.CustomerBookingPhone+"%")
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code COLLATE utf8mb4_general_ci LIKE ?", "%"+item.BookingCode+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingWaiting) FindAll(database *gorm.DB) ([]BookingWaiting, int64, error) {
	db := database.Model(BookingWaiting{})
	list := []BookingWaiting{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *BookingWaiting) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *BookingWaiting) IsDuplicated(db *gorm.DB, checkTeeTime, checkBag bool) (bool, error) {

	if item.TeeTime == "" {
		return false, nil
	}
	//Check turn time đã tồn tại
	if checkTeeTime {
		booking := BookingWaiting{
			PartnerUid:  item.PartnerUid,
			CourseUid:   item.CourseUid,
			TeeTime:     item.TeeTime,
			TurnTime:    item.TurnTime,
			BookingDate: item.BookingDate,
			TeeType:     item.TeeType,
			CourseType:  item.CourseType,
		}

		errFind := booking.FindFirst(db)
		if errFind == nil {
			return true, errors.New("Duplicated TeeTime")
		}
	}

	return false, nil
}
