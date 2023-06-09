package mdoel_report

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ReportRevenueDetail struct {
	models.ModelId
	PartnerUid       string        `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid        string        `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BillNo           string        `json:"bill_no" gorm:"type:varchar(50);index"`      // Mã bill
	BookingDate      string        `json:"booking_date" gorm:"type:varchar(50);index"` // Ngày chơi
	Bag              string        `json:"bag" gorm:"type:varchar(50);index"`          // Mã KH
	AgencyInfo       BookingAgency `json:"agency_info" gorm:"type:json"`
	CustomerId       string        `json:"customer_id" gorm:"type:varchar(50);index"` // Mã KH
	CustomerName     string        `json:"customer_name"`                             // Tên KH
	MembershipNo     string        `json:"member_ship_no" gorm:"type:varchar(50)"`    // Mã thẻ thành viên
	CustomerType     string        `json:"customer_type" gorm:"type:varchar(50)"`     // Loại KH
	GuestStyle       string        `json:"guest_style" gorm:"type:varchar(50);index"` // Guest Style
	GuestStyleName   string        `json:"guest_style_name" gorm:"type:varchar(100)"` // Guest Style Name
	Hole             int           `json:"hole" gorm:"type:varchar(50)"`              // Số Hole
	GreenFee         int64         `json:"green_fee"`                                 // Phí sân cỏ
	CaddieFee        int64         `json:"caddie_fee"`                                // Phí caddie
	SubBagFee        int64         `json:"sub_bag_fee"`                               // Phí trả cho sub bag
	FBFee            int64         `json:"fb_fee"`                                    // Phí ăn uống
	RentalFee        int64         `json:"rental_fee"`                                // Phí thuê đồ
	BuggyFee         int64         `json:"buggy_fee"`                                 // Phí thuê xe
	BookingCaddieFee int64         `json:"booking_caddie_fee"`                        // Phí booking caddie
	ProshopFee       int64         `json:"proshop_fee"`                               // Phí đồ ở Proshop
	RestaurantFee    int64         `json:"restaurant_fee"`                            // Phí nhà hàng
	KioskFee         int64         `json:"kiosk_fee"`                                 // Phí kiosk
	MinibarFee       int64         `json:"minibar_fee"`                               // Phí minibar
	PraticeBallFee   int64         `json:"pratice_ball_fee"`                          // Phí bóng tập
	OtherFee         int64         `json:"other_fee"`                                 // Phí khác
	Cash             int64         `json:"cash"`                                      // Số tiền mặt
	Card             int64         `json:"card"`                                      // Số tiền cà thẻ
	Voucher          int64         `json:"voucher"`                                   // Số tiền Voucher
	Debit            int64         `json:"debit"`                                     // Số tiền nợ
	MushPay          int64         `json:"mush_pay"`                                  // Tổng tiền phải trả
	Paid             int64         `json:"paid"`                                      // Tổng tiền phải trả
	Transfer         int64         `json:"transfer"`                                  // Số tiền chuyển khoản
	PhiPhat          int64         `json:"phi_phat"`
	Total            int64         `json:"total"`
	CheckOutTime     int64         `json:"check_out_time"` // Time Check Out
	HoleBooking      int           `json:"hole_booking"`   // Hole booking
}

type DayEndRevenue struct {
	PartnerUid       string `json:"partner_uid"`        // Hang Golf
	CourseUid        string `json:"course_uid"`         // San GolfGreenFee         int64  `json:"green_fee"`
	AgencyPaid       int64  `json:"agency_paid"`        // Phí sân cỏ
	GreenFee         int64  `json:"green_fee"`          // Phí sân cỏ
	CaddieFee        int64  `json:"caddie_fee"`         // Phí caddie
	RentalFee        int64  `json:"rental_fee"`         // Phí thuê đồ
	BuggyFee         int64  `json:"buggy_fee"`          // Phí thuê xe
	BookingCaddieFee int64  `json:"booking_caddie_fee"` // Phí booking caddie
	ProshopFee       int64  `json:"proshop_fee"`        // Phí đồ ở Proshop
	RestaurantFee    int64  `json:"restaurant_fee"`     // Phí nhà hàng
	KioskFee         int64  `json:"kiosk_fee"`          // Phí kiosk
	MinibarFee       int64  `json:"minibar_fee"`        // Phí minibar
	PraticeBallFee   int64  `json:"pratice_ball_fee"`   // Phí bóng tập
	OtherFee         int64  `json:"other_fee"`
	MushPay          int64  `json:"mush_pay"` // Tổng tiền phải trả
	Member           int64  `json:"member"`
	Visitor          int64  `json:"visitor"`
	Foc              int64  `json:"foc"`
	Tour             int64  `json:"tour"`
	HK               int64  `json:"hk"`
	Cash             int64  `json:"cash"`
	MemberGuest      int64  `json:"member_guest"`
	TotalFee         int64  `json:"total_fee"`
	AllFee           int64  `json:"all_fee"`
	PhiPhat          int64  `json:"phi_phat"`
	Transfer         int64  `json:"transfer"` // Số tiền chuyển khoản
	Debit            int64  `json:"debit"`    // Số tiền nợ
	Card             int64  `json:"card"`     // Số tiền cà thẻ
	TotalPlayer      int64  `json:"total_player"`
}

type ReportDebit struct {
	BookingDate  string `json:"booking_date"`  // Ngày chơi
	Bag          string `json:"bag"`           // Mã KH
	CustomerName string `json:"customer_name"` // Tên KH
	Debit        int64  `json:"debit"`         // Số tiền nợ
}

type BookingAgency struct {
	AgencyId    string `json:"agency_id"`    // Id Agency
	ShortName   string `json:"short_name"`   // Ten ngắn Dai ly
	Category    string `json:"category"`     // Category
	Name        string `json:"name"`         // Ten Dai ly
	BookingCode string `json:"booking_code"` // Ten Dai ly
}

func (item *BookingAgency) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingAgency) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *ReportRevenueDetail) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *ReportRevenueDetail) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *ReportRevenueDetail) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *ReportRevenueDetail) Count(database *gorm.DB) (int64, error) {
	db := database.Model(ReportRevenueDetail{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *ReportRevenueDetail) FindList(dbase *gorm.DB, page models.Page) ([]ReportRevenueDetail, int64, error) {
	db := dbase.Model(ReportRevenueDetail{})
	list := []ReportRevenueDetail{}
	total := int64(0)
	status := item.Status
	item.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.CourseUid != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *ReportRevenueDetail) FindAll(dbase *gorm.DB) ([]ReportRevenueDetail, int64, error) {
	db := dbase.Model(ReportRevenueDetail{})
	list := []ReportRevenueDetail{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.CourseUid != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}

func (item *ReportRevenueDetail) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *ReportRevenueDetail) DeleteByBookingDate() error {
	db := datasources.GetDatabase().Model(ReportRevenueDetail{})
	db = db.Where("booking_date = ?", item.BookingDate)
	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("course_uid = ?", item.CourseUid)
	return db.Delete(item).Error
}

func (item *ReportRevenueDetail) FindReportDayEnd(database *gorm.DB) (DayEndRevenue, error) {
	db := datasources.GetDatabase().Model(ReportRevenueDetail{})
	db = db.Select(`partner_uid,
					course_uid,
					SUM(paid) AS agency_paid, 
					SUM(green_fee) AS green_fee, 
					SUM(caddie_fee) AS caddie_fee,
					SUM(buggy_fee) AS buggy_fee,
					SUM(pratice_ball_fee) AS pratice_ball_fee,
					SUM(restaurant_fee) AS restaurant_fee,
					SUM(kiosk_fee) AS kiosk_fee,
					SUM(proshop_fee) AS proshop_fee,
					SUM(minibar_fee) AS minibar_fee,
					SUM(booking_caddie_fee) AS booking_caddie_fee,
					SUM(mush_pay) as mush_pay,
					SUM(rental_fee) as rental_fee,
					SUM(other_fee) as other_fee,
					SUM(fb_fee) as fb_fee,
					SUM(total) as all_fee,
					SUM(phi_phat) as phi_phat,
					SUM(cash) as cash,
					SUM(debit) as debit,
					SUM(transfer) as transfer,
					SUM(customer_type = 'GUEST') AS member_guest,
					SUM(customer_type = 'VISITOR') AS visitor,
					SUM(customer_type = 'FOC') AS foc,
					SUM(guest_style = '2D'|| guest_style = '2E') AS hk,
					SUM(customer_type = 'MEMBER') - SUM(guest_style = '2D'|| guest_style = '2E') AS member,
					SUM(green_fee + caddie_fee + buggy_fee + pratice_ball_fee + restaurant_fee + kiosk_fee + proshop_fee + minibar_fee + booking_caddie_fee + rental_fee + other_fee + phi_phat) AS total_fee,
					SUM(customer_type = 'TRADITIONAL' || customer_type = 'OTA') AS tour`)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	dayEnd := DayEndRevenue{}
	db.Find(&dayEnd)

	return dayEnd, db.Error
}

func (item *ReportRevenueDetail) FindAllDebit(database *gorm.DB) ([]ReportDebit, error) {
	var list []ReportDebit

	db := database.Model(ReportRevenueDetail{})

	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("course_uid = ?", item.CourseUid)
	db = db.Where("booking_date = ?", item.BookingDate)
	db = db.Where("debit != 0")

	db.Find(&list)

	return list, db.Error
}
