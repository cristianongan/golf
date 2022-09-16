package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"gorm.io/datatypes"
)

/*
Giỏ Hàng
*/
type ServiceCart struct {
	ModelId
	PartnerUid     string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng golf
	CourseUid      string         `json:"course_uid" gorm:"type:varchar(150);index"`  // Sân golf
	ServiceId      int64          `json:"service_id" gorm:"index"`                    // Mã của service
	GolfBag        string         `json:"golf_bag" gorm:"type:varchar(100);index"`    // Số bag order
	BookingDate    datatypes.Date `json:"booking_date"`                               // Ngày order
	BookingUid     string         `json:"booking_uid" gorm:"type:varchar(100)"`       // Booking uid
	BillCode       string         `json:"bill_code" gorm:"default:NONE"`              // Mã hóa đơn
	BillStatus     string         `json:"bill_status" gorm:"type:varchar(50)"`        // trạng thái đơn
	TypeCode       string         `json:"type_code" gorm:"type:varchar(100)"`         // Mã dịch vụ của hóa đơn
	Type           string         `json:"type" gorm:"type:varchar(100)"`              // Dịch vụ hóa đơn: BRING, SHIP, TABLE
	StaffOrder     string         `json:"staff_order" gorm:"type:varchar(150)"`
	NumberGuest    int            `json:"number_guest"`                            // số lượng người đi cùng
	Amount         int64          `json:"amount"`                                  // tổng tiền
	DiscountType   string         `json:"discount_type" gorm:"type:varchar(50)"`   // Loại giảm giá
	DiscountValue  int64          `json:"discount_value"`                          // Giá tiền được giảm
	DiscountReason string         `json:"discount_reason" gorm:"type:varchar(50)"` // Lý do giảm giá
	CostPrice      int64          `json:"cost_price"`                              // giá VAT
}

func (item *ServiceCart) Create() error {
	now := time.Now()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *ServiceCart) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *ServiceCart) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

func (item *ServiceCart) FindList(page Page) ([]ServiceCart, int64, error) {
	var list []ServiceCart
	total := int64(0)

	db := datasources.GetDatabase().Model(ServiceCart{})

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

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	if item.BillStatus != "" {
		db = db.Where("bill_status = ? OR bill_status = ?", constants.RES_STATUS_PROCESS, constants.RES_STATUS_DONE)
	}

	db = db.Where("booking_date = ?", item.BookingDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
