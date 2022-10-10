package model_gostarter

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Flight
type Flight struct {
	models.ModelId
	PartnerUid  string    `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string    `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Tee         int       `json:"tee"`                                        // Tee
	TeeOff      string    `json:"tee_off" gorm:"type:varchar(30)"`            //
	Turn        string    `json:"turn" gorm:"type:varchar(30)"`               //
	End         string    `json:"end" gorm:"type:varchar(30)"`                //
	DateDisplay string    `json:"date_display" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	Bookings    []Booking `json:"bookings"`
	GroupName   string    `json:"group_name" gorm:"type:varchar(100)"`
}

type Booking BookingForFlight

type BookingForFlight struct {
	models.Model
	CourseType       string                      `json:"course_type"`
	PartnerUid       string                      `json:"partner_uid,omitempty"`
	CourseUid        string                      `json:"course_uid,omitempty"`
	BookingDate      string                      `json:"booking_date,omitempty"`
	Bag              string                      `json:"bag,omitempty"`
	Hole             int                         `json:"hole"`
	HoleBooking      int                         `json:"hole_booking"`     // Số hố khi booking
	HoleTimeOut      int                         `json:"hole_time_out"`    // Số hố khi time out
	HoleMoveFlight   int                         `json:"hole_move_flight"` // Số hố trong đã chơi của flight khi bag move sang
	CustomerName     string                      `json:"customer_name,omitempty"`
	CustomerUid      string                      `json:"customer_uid,omitempty"`
	CustomerInfo     model_booking.CustomerInfo  `json:"customer_info,omitempty"`
	CaddieId         int64                       `json:"caddie_id,omitempty"`
	CaddieInfo       model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	BuggyId          int64                       `json:"buggy_id,omitempty"`
	BuggyInfo        model_booking.BookingBuggy  `json:"buggy_info,omitempty"`
	CaddieStatus     string                      `json:"caddie_status,omitempty"`
	CaddieBuggyInOut []CaddieBuggyInOut          `json:"caddie_buggy_in_out" gorm:"foreignKey:BookingUid;references:Uid"`
	FlightId         int64                       `json:"flight_id"`
	CheckOutTime     int64                       `json:"check_out_time,omitempty"`
	CheckInTime      int64                       `json:"check_in_time,omitempty"`
	CardId           string                      `json:"card_id,omitempty"`
	MemberCardUid    string                      `json:"member_card_uid,omitempty"`
	AgencyId         int64                       `json:"agency_id,omitempty"`
	AgencyInfo       model_booking.BookingAgency `json:"agency_info,omitempty"`
	GuestStyle       string                      `json:"guest_style,omitempty"`
	GuestStyleName   string                      `json:"guest_style_name,omitempty"`
	TimeOutFlight    int64                       `json:"time_out_flight,omitempty"`
	CmsUser          string                      `json:"cms_user,omitempty"`
	CmsUserLog       string                      `json:"cms_user_log,omitempty"`
	NoteOfBag        string                      `json:"note_of_bag"`
	NoteOfBooking    string                      `json:"note_of_booking"`
	NoteOfGo         string                      `json:"note_of_go"`
	BagStatus        string                      `json:"bag_status"`
	IsPrivateBuggy   bool                        `json:"is_private_buggy"`
	MovedFlight      bool                        `json:"moved_flight"`
	AddedRound       bool                        `json:"added_flight"`
}

func (item *Flight) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Flight) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Flight) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Flight) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Flight{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Flight) FindList(page models.Page, from, to int64) ([]Flight, int64, error) {
	db := datasources.GetDatabase().Model(Flight{})
	list := []Flight{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Flight) FindListAll() ([]Flight, error) {
	db := datasources.GetDatabase().Model(Flight{})
	list := []Flight{}
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.DateDisplay != "" {
		db = db.Where("date_display = ?", item.DateDisplay)
	}

	db.Find(&list)
	err := db.Error
	if err != nil {
		log.Println("Flight FindListAll err ", err.Error())
	}
	return list, err
}

func (item *Flight) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
