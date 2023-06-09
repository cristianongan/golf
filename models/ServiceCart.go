package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/*
Giỏ Hàng
*/
type ServiceCart struct {
	ModelId
	PartnerUid       string         `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hãng golf
	CourseUid        string         `json:"course_uid" gorm:"type:varchar(150);index"`   // Sân golf
	ServiceId        int64          `json:"service_id" gorm:"index"`                     // Mã của service
	ServiceType      string         `json:"service_type" gorm:"type:varchar(100);index"` // Loại của service
	FromService      int64          `json:"from_service" gorm:"index"`                   // Mã của from service
	FromServiceName  string         `json:"from_service_name" gorm:"type:varchar(150)"`  // Tên của from service
	OrderTime        int64          `json:"order_time" gorm:"index"`                     // Thời gian order
	TimeProcess      int64          `json:"time_process"`                                // Thời gian bắt đầu chế biến
	GolfBag          string         `json:"golf_bag" gorm:"type:varchar(100);index"`     // Số bag order
	BookingDate      datatypes.Date `json:"booking_date" gorm:"index"`                   // Ngày order
	BookingUid       string         `json:"booking_uid" gorm:"type:varchar(100)"`        // Booking uid
	BillCode         string         `json:"bill_code" gorm:"default:NONE;index"`         // Mã hóa đơn
	BillStatus       string         `json:"bill_status" gorm:"type:varchar(50);index"`   // trạng thái đơn
	TypeCode         string         `json:"type_code" gorm:"type:varchar(100);index"`    // Mã dịch vụ của hóa đơn
	Type             string         `json:"type" gorm:"type:varchar(100)"`               // Dịch vụ hóa đơn: BRING, SHIP, TABLE
	StaffOrder       string         `json:"staff_order" gorm:"type:varchar(150)"`        // Người tạo đơn
	PlayerName       string         `json:"player_name" gorm:"type:varchar(150)"`        // Người mua
	Note             string         `json:"note" gorm:"type:varchar(250)"`               // Note của người mua
	Phone            string         `json:"phone" gorm:"type:varchar(100)"`              // Số điện thoại
	NumberGuest      int            `json:"number_guest"`                                // số lượng người đi cùng
	Amount           int64          `json:"amount"`                                      // tổng tiền
	DiscountType     string         `json:"discount_type" gorm:"type:varchar(50)"`       // Loại giảm giá
	DiscountValue    int64          `json:"discount_value"`                              // Giá tiền được giảm
	DiscountReason   string         `json:"discount_reason" gorm:"type:varchar(50)"`     // Lý do giảm giá
	CostPrice        bool           `json:"cost_price"`                                  // Có giá VAT hay ko
	ResFloor         int            `json:"res_floor"`                                   // Số tầng bàn được đặt\
	RentalStatus     string         `json:"rental_status" gorm:"type:varchar(100)"`      // Trạng thái thuê đồ
	CaddieCode       string         `json:"caddie_code" gorm:"type:varchar(100)"`        // Caddie đi cùng bag
	TotalMoveKitchen int            `json:"total_move_kitchen"`                          // Tổng số lần move kitchen của bill
}

type ListBillForApp struct {
	ModelId
	Status        string         `json:"status"`         //ENABLE, DISABLE, TESTING, DELETED
	PartnerUid    string         `json:"partner_uid"`    // Hãng golf
	CourseUid     string         `json:"course_uid"`     // Sân golf
	ServiceId     int64          `json:"service_id"`     // Mã của service
	ServiceType   string         `json:"service_type"`   // Loại của service
	GolfBag       string         `json:"golf_bag"`       // Số bag order
	BookingDate   datatypes.Date `json:"booking_date"`   // Ngày order
	BookingUid    string         `json:"booking_uid"`    // Booking uid
	BillCode      string         `json:"bill_code"`      // Mã hóa đơn
	BillStatus    string         `json:"bill_status"`    // trạng thái đơn
	TypeCode      string         `json:"type_code"`      // Mã dịch vụ của hóa đơn
	Type          string         `json:"type"`           // Dịch vụ hóa đơn: BRING, SHIP, TABLE
	StaffOrder    string         `json:"staff_order"`    // Người tạo đơn
	PlayerName    string         `json:"player_name"`    // Người mua
	Note          string         `json:"note"`           // Note của người mua
	Amount        int64          `json:"amount"`         // tổng tiền
	DiscountType  string         `json:"discount_type"`  // Loại giảm giá
	DiscountValue int64          `json:"discount_value"` // Giá tiền được giảm
	CaddieCode    string         `json:"caddie_code"`    // Caddie đi cùng bag
}

func (item *ServiceCart) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *ServiceCart) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *ServiceCart) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	return db.Save(item).Error
}

func (item *ServiceCart) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
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

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.TypeCode != "" {
		db = db.Where("type_code LIKE ?", "%"+item.TypeCode+"%").Or("bill_code LIKE ?", "%"+item.TypeCode+"%")
	}

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}

	if item.BillStatus == constants.RES_BILL_STATUS_SHOW {
		db = db.Where("bill_status IN ?", []string{constants.RES_STATUS_PROCESS, constants.RES_BILL_STATUS_TRANSFER, constants.RES_STATUS_ORDER, constants.RES_BILL_STATUS_FINISH, constants.RES_BILL_STATUS_OUT})
	} else if item.BillStatus == constants.RES_BILL_STATUS_ACTIVE {
		db = db.Where("bill_status IN ?", []string{constants.RES_STATUS_PROCESS, constants.RES_STATUS_DONE})
	} else if item.BillStatus != "" {
		db = db.Where("bill_status = ?", item.BillStatus)
	}

	if item.ResFloor != 0 {
		db = db.Where("res_floor = ?", item.ResFloor)
	}

	// if item.PlayerName != "" {
	// 	db = db.Where("player_name LIKE ?", "%"+item.PlayerName+"%")
	// }

	if item.GolfBag != "" {
		db = db.Where("golf_bag LIKE ? OR bill_code LIKE ? OR player_name LIKE ?", "%"+item.GolfBag+"%", "%"+item.GolfBag+"%", "%"+item.GolfBag+"%")
	}

	if item.FromService != 0 {
		db = db.Where("from_service = ?", item.FromService)
	}

	if item.RentalStatus != "" {
		db = db.Where("rental_status = ?", item.RentalStatus)
	}

	db = db.Where("booking_date = ?", item.BookingDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *ServiceCart) FindListForApp(database *gorm.DB, page Page) ([]ListBillForApp, int64, error) {
	var list []ListBillForApp
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

	if item.GolfBag != "" {
		db = db.Where("golf_bag LIKE ? OR bill_code LIKE ? OR player_name LIKE ?", "%"+item.GolfBag+"%", "%"+item.GolfBag+"%", "%"+item.GolfBag+"%")
	}

	if item.StaffOrder != "" {
		db = db.Where("staff_order = ?", item.StaffOrder)
	}

	db = db.Where("booking_date = ?", item.BookingDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
