package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/*
Giỏ Hàng
*/
type ServiceCart struct {
	ModelId
	PartnerUid      string         `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hãng golf
	CourseUid       string         `json:"course_uid" gorm:"type:varchar(150);index"`   // Sân golf
	ServiceId       int64          `json:"service_id" gorm:"index"`                     // Mã của service
	ServiceType     string         `json:"service_type" gorm:"type:varchar(100);index"` // Loại của service
	FromService     int64          `json:"from_service" gorm:"index"`                   // Mã của from service
	FromServiceName string         `json:"from_service_name" gorm:"type:varchar(150)"`  // Tên của from service
	OrderTime       int64          `json:"order_time" gorm:"index"`                     // Thời gian order
	GolfBag         string         `json:"golf_bag" gorm:"type:varchar(100);index"`     // Số bag order
	BookingDate     datatypes.Date `json:"booking_date"`                                // Ngày order
	BookingUid      string         `json:"booking_uid" gorm:"type:varchar(100)"`        // Booking uid
	BillCode        string         `json:"bill_code" gorm:"default:NONE"`               // Mã hóa đơn
	BillStatus      string         `json:"bill_status" gorm:"type:varchar(50)"`         // trạng thái đơn
	TypeCode        string         `json:"type_code" gorm:"type:varchar(100)"`          // Mã dịch vụ của hóa đơn
	Type            string         `json:"type" gorm:"type:varchar(100)"`               // Dịch vụ hóa đơn: BRING, SHIP, TABLE
	StaffOrder      string         `json:"staff_order" gorm:"type:varchar(150)"`        // Người tạo đơn
	PlayerName      string         `json:"player_name" gorm:"type:varchar(150)"`        // Người mua
	Note            string         `json:"note" gorm:"type:varchar(250)"`               // Note của người mua
	Phone           string         `json:"phone" gorm:"type:varchar(100)"`              // Số điện thoại
	NumberGuest     int            `json:"number_guest"`                                // số lượng người đi cùng
	Amount          int64          `json:"amount"`                                      // tổng tiền
	DiscountType    string         `json:"discount_type" gorm:"type:varchar(50)"`       // Loại giảm giá
	DiscountValue   int64          `json:"discount_value"`                              // Giá tiền được giảm
	DiscountReason  string         `json:"discount_reason" gorm:"type:varchar(50)"`     // Lý do giảm giá
	CostPrice       bool           `json:"cost_price"`                                  // Có giá VAT hay ko
	ResFloor        int            `json:"res_floor"`                                   // Số tầng bàn được đặt
}

func (item *ServiceCart) Create(db *gorm.DB) error {
	now := time.Now()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *ServiceCart) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *ServiceCart) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}

func (item *ServiceCart) FindList(database *gorm.DB, page Page) ([]ServiceCart, int64, error) {
	var list []ServiceCart
	total := int64(0)

	db := database.Model(ServiceCart{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ServiceId != 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}

	if item.BillStatus == "Active" {
		db = db.Where("bill_status = ? OR bill_status = ?", constants.RES_STATUS_PROCESS, constants.RES_STATUS_DONE)
	} else if item.BillStatus != "" {
		db = db.Where("bill_status = ?", item.BillStatus)
	}

	if item.ResFloor != 0 {
		db = db.Where("res_floor = ?", item.ResFloor)
	}

	db = db.Where("booking_date = ?", item.BookingDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
