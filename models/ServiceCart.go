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

	if item.BillStatus == constants.RES_BILL_STATUS_ACTIVE {
		db = db.Where("bill_status = ? OR bill_status = ?", constants.RES_STATUS_PROCESS, constants.RES_STATUS_DONE)
	} else if item.BillStatus != "" {
		db = db.Where("bill_status = ?", item.BillStatus)
	}

	if item.ResFloor != 0 {
		db = db.Where("res_floor = ?", item.ResFloor)
	}

	if item.PlayerName != "" {
		db = db.Where("player_name LIKE ?", "%"+item.PlayerName+"%")
	}

	if item.GolfBag != "" {
		db = db.Where("golf_bag LIKE ? OR bill_code LIKE ?", "%"+item.GolfBag+"%", "%"+item.GolfBag+"%")
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

func (item *ServiceCart) FindReportDetailFBBag(database *gorm.DB, page Page, fromDate, toDate, typeService string) ([]map[string]interface{}, int64, error) {
	var list []map[string]interface{}
	total := int64(0)

	db := database.Table("service_carts")

	db = db.Select("service_carts.booking_date", "service_carts.golf_bag", "service_carts.player_name", "tb1.total_res", "tb2.total_kiosk", "tb3.total_mini_bar")

	if fromDate != "" {
		db = db.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
	}

	if toDate != "" {
		db = db.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
	}

	db = db.Group("service_carts.booking_uid")

	// sum revenue kiosk
	if typeService == "ALL" || typeService == "RES" {
		if typeService == "RES" {
			db = db.Select("service_carts.booking_date", "service_carts.golf_bag", "service_carts.player_name", "tb1.total_res")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, service_carts.booking_uid, SUM(service_carts.amount) as total_res")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("bill_status = ?", constants.RES_BILL_STATUS_FINISH)

		subQuery = subQuery.Where("service_type = ?", constants.RESTAURANT_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("booking_uid")

		db = db.Joins(`LEFT JOIN (?) as tb1 on service_carts.booking_uid = tb1.booking_uid`, subQuery)
	}

	// sum revenue kiosk
	if typeService == "ALL" || typeService == "KIOSK" {
		if typeService == "KIOSK" {
			db = db.Select("service_carts.booking_date", "service_carts.golf_bag", "service_carts.player_name", "tb2.total_kiosk")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, service_carts.booking_uid, SUM(service_carts.amount) as total_kiosk")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("service_carts.partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("service_carts.course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("service_carts.bill_status = ?", constants.POS_BILL_STATUS_ACTIVE)

		subQuery = subQuery.Where("service_carts.service_type = ?", constants.KIOSK_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("service_carts.booking_uid")

		db = db.Joins(`LEFT JOIN (?) as tb2 on service_carts.booking_uid = tb2.booking_uid`, subQuery)
	}

	// sum revenue mini bar
	if typeService == "ALL" || typeService == "MINI_B" {
		if typeService == "MINI_B" {
			db = db.Select("service_carts.booking_date", "service_carts.golf_bag", "service_carts.player_name", "tb3.total_mini_bar")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, service_carts.booking_uid, SUM(service_carts.amount) as total_mini_bar")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("service_carts.partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("service_carts.course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("service_carts.bill_status = ?", constants.POS_BILL_STATUS_ACTIVE)

		subQuery = subQuery.Where("service_carts.service_type = ?", constants.MINI_B_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("service_carts.booking_uid")

		db = db.Joins(`LEFT JOIN (?) as tb3 on service_carts.booking_uid = tb3.booking_uid`, subQuery)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *ServiceCart) FindReport(database *gorm.DB, page Page, fromDate, toDate, typeService string) ([]map[string]interface{}, int64, error) {
	var list []map[string]interface{}
	total := int64(0)

	db := database.Table("service_carts")

	db = db.Select("service_carts.booking_date", "tb1.total_res", "tb2.total_kiosk", "tb3.total_mini_bar")

	if item.PartnerUid != "" {
		db = db.Where("service_carts.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("service_carts.course_uid = ?", item.CourseUid)
	}

	if fromDate != "" {
		db = db.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
	}

	if toDate != "" {
		db = db.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
	}

	db = db.Group("service_carts.booking_date")

	// sum revenue kiosk
	if typeService == "ALL" || typeService == "RES" {
		if typeService == "RES" {
			db = db.Select("service_carts.booking_date", "tb1.total_res")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, SUM(service_carts.amount) as total_res")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("bill_status = ?", constants.RES_BILL_STATUS_FINISH)

		subQuery = subQuery.Where("service_type = ?", constants.RESTAURANT_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("booking_date")

		db = db.Joins(`LEFT JOIN (?) as tb1 on service_carts.booking_date = tb1.booking_date`, subQuery)
	}

	// sum revenue kiosk
	if typeService == "ALL" || typeService == "KIOSK" {
		if typeService == "KIOSK" {
			db = db.Select("service_carts.booking_date", "tb2.total_kiosk")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, SUM(service_carts.amount) as total_kiosk")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("service_carts.partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("service_carts.course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("service_carts.bill_status = ?", constants.POS_BILL_STATUS_ACTIVE)

		subQuery = subQuery.Where("service_carts.service_type = ?", constants.KIOSK_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("service_carts.booking_date")

		db = db.Joins(`LEFT JOIN (?) as tb2 on service_carts.booking_date = tb2.booking_date`, subQuery)
	}

	// sum revenue mini bar
	if typeService == "ALL" || typeService == "MINI_B" {
		if typeService == "MINI_B" {
			db = db.Select("service_carts.booking_date", "tb3.total_mini_bar")
		}

		subQuery := database.Table("service_carts").Select("service_carts.booking_date, SUM(service_carts.amount) as total_mini_bar")

		if item.PartnerUid != "" {
			subQuery = subQuery.Where("service_carts.partner_uid = ?", item.PartnerUid)
		}

		if item.CourseUid != "" {
			subQuery = subQuery.Where("service_carts.course_uid = ?", item.CourseUid)
		}

		subQuery = subQuery.Where("service_carts.bill_status = ?", constants.POS_BILL_STATUS_ACTIVE)

		subQuery = subQuery.Where("service_carts.service_type = ?", constants.MINI_B_SETTING)

		if fromDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
		}

		if toDate != "" {
			subQuery = subQuery.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
		}

		subQuery = subQuery.Group("service_carts.booking_date")

		db = db.Joins(`LEFT JOIN (?) as tb3 on service_carts.booking_date = tb3.booking_date`, subQuery)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *ServiceCart) FindReportDetailFB(database *gorm.DB, page Page, fromDate, toDate, name string) ([]map[string]interface{}, int64, error) {
	var list []map[string]interface{}
	total := int64(0)
	db := database.Table("service_carts")

	db = db.Select("service_carts.booking_date, tb1.*")

	if item.CourseUid != "" {
		db = db.Where("service_carts.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("service_carts.partner_uid = ?", item.PartnerUid)
	}

	if fromDate != "" {
		db = db.Where("service_carts.booking_date >= STR_TO_DATE(?, '%Y-%m-%d')", fromDate)
	}

	if toDate != "" {
		db = db.Where("service_carts.booking_date <= STR_TO_DATE(?, '%Y-%m-%d')", toDate)
	}

	if item.ServiceType != "" {
		db = db.Where("service_carts.service_type = ?", item.ServiceType)
	} else {
		db = db.Where("service_carts.service_type IN ?", []string{constants.KIOSK_SETTING, constants.MINI_B_SETTING, constants.RESTAURANT_SETTING})
	}

	db = db.Where("service_carts.bill_status IN ?", []string{constants.RES_BILL_STATUS_FINISH, constants.POS_BILL_STATUS_ACTIVE})

	// sub query
	subQuery := database.Table("booking_service_items")

	if name != "" {
		subQuery = subQuery.Where("name LIKE ?", "%"+name+"%")
	}

	db = db.Joins(`INNER JOIN (?) as tb1 on service_carts.id = tb1.service_bill`, subQuery)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
