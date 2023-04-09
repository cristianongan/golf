package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/models"
	"start/utils"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BookingServiceItem struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hãng golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(150);index"`   // Sân golf
	ItemId         int64  `json:"item_id"  gorm:"index"`                       // Id item
	Unit           string `json:"unit"  gorm:"type:varchar(100)"`              // Unit của item
	ServiceId      string `json:"service_id"  gorm:"type:varchar(100)"`        // uid service
	ServiceType    string `json:"service_type"  gorm:"type:varchar(100)"`      // Loại service gồm FB, Rental, Proshop
	BookingUid     string `json:"booking_uid"  gorm:"type:varchar(100);index"` // Uid booking
	PlayerName     string `json:"player_name" gorm:"type:varchar(256)"`        // Tên người chơi
	Bag            string `json:"bag" gorm:"type:varchar(50)"`                 // Golf Bag
	Type           string `json:"type" gorm:"type:varchar(50)"`                // Loại rental, kiosk, proshop,...
	ItemCode       string `json:"item_code"  gorm:"type:varchar(100)"`         // Mã code của item
	ItemType       string `json:"item_type"  gorm:"type:varchar(100)"`         // Phân loại đồ ăn theo COMBO hoặc NORMAL
	Name           string `json:"name" gorm:"type:varchar(256)"`
	EngName        string `json:"eng_name" gorm:"type:varchar(256)"` // Tên tiếng anh của sản phẩm
	GroupCode      string `json:"group_code" gorm:"type:varchar(100)"`
	Quality        int    `json:"quality"` // Số lượng
	UnitPrice      int64  `json:"unit_price"`
	DiscountType   string `json:"discount_type" gorm:"type:varchar(50)"`
	DiscountValue  int64  `json:"discount_value"`
	DiscountReason string `json:"discount_reason" gorm:"type:varchar(50)"` // Lý do giảm giá
	Amount         int64  `json:"amount"`
	UserAction     string `json:"user_action" gorm:"type:varchar(100)"` // Người tạo
	Input          string `json:"input" gorm:"type:varchar(300)"`       // Note
	BillCode       string `json:"bill_code" gorm:"type:varchar(100);index"`
	ServiceBill    int64  `json:"service_bill" gorm:"index"`               // id service cart
	SaleQuantity   int64  `json:"sale_quantity"`                           // tổng số lượng bán được
	Location       string `json:"location" gorm:"type:varchar(100);index"` // Dc add từ đâu
	PaidBy         string `json:"paid_by" gorm:"type:varchar(50)"`         // Paid by: cho case đại lý thanh toán
	Hole           int    `json:"hole"`                                    // Số hố check in
}

type ListItemInApp struct {
	models.ModelId
	PartnerUid    string `json:"partner_uid"` // Hãng golf
	CourseUid     string `json:"course_uid"`  // Sân golf
	ItemId        int64  `json:"item_id"`     // Id item
	Unit          string `json:"unit"`        // Unit của item
	ItemCode      string `json:"item_code"`   // Mã code của item
	ItemType      string `json:"item_type"`   // Phân loại đồ ăn theo COMBO hoặc NORMAL
	Name          string `json:"name"`
	Quality       int    `json:"quality"` // Số lượng
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	UserAction    string `json:"user_action"`  // Người tạo
	Input         string `json:"input"`        // Note
	ServiceBill   int64  `json:"service_bill"` // id service cart
	Location      string `json:"location"`     // Dc add từ đâu
}

// Response cho FE
type BookingServiceItemResponse struct {
	BookingServiceItem
	CheckInTime int64 `json:"check_in_time"`
}

type BookingServiceItemWithPaidInfo struct {
	BookingServiceItem
	IsPaid bool `json:"is_paid"`
	// IsAgencyPaid bool `json:"is_agency_paid"`
}

// ------- List Booking service ---------
type ListBookingServiceItems []BookingServiceItem

func (item *ListBookingServiceItems) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingServiceItems) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *BookingServiceItem) IsDuplicated(db *gorm.DB) bool {
	errFind := item.FindFirst(db)
	return errFind == nil
}

func (item *BookingServiceItem) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingServiceItem) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingServiceItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingServiceItem) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingServiceItem{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingServiceItem) FindAll(database *gorm.DB) (ListBookingServiceItems, error) {
	db := database.Model(BookingServiceItem{})
	list := ListBookingServiceItems{}
	item.Status = ""

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.ServiceBill > 0 {
		db = db.Where("service_bill = ?", item.ServiceBill)
	}

	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}

	db = db.Find(&list)
	return list, db.Error
}

func (item *BookingServiceItem) FindAllWithPaidInfo(database *gorm.DB) ([]BookingServiceItemWithPaidInfo, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	item.Status = ""

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.ServiceBill > 0 {
		db = db.Where("service_bill = ?", item.ServiceBill)
	}

	db = db.Find(&list)

	res := []BookingServiceItemWithPaidInfo{}
	for _, item := range list {
		res = append(res, BookingServiceItemWithPaidInfo{
			BookingServiceItem: item,
		})
	}
	return res, db.Error
}

func (item *BookingServiceItem) FindList(database *gorm.DB, page models.Page) ([]map[string]interface{}, int64, error) {
	db := database.Model(BookingServiceItem{})
	var list []map[string]interface{}
	total := int64(0)

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}
	if item.ServiceId != "" {
		db = db.Where("service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}
	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}
	if item.ServiceBill > 0 {
		db = db.Where("service_bill = ?", item.ServiceBill)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) FindReportSalePOS(database *gorm.DB, date string) ([]map[string]interface{}, error) {
	db := database.Table("booking_service_items")
	var list []map[string]interface{}

	db = db.Select(`
		booking_service_items.item_id, booking_service_items.name, booking_service_items.eng_name,
		booking_service_items.unit, booking_service_items.group_code, tb2.group_name,
		SUM(if(tb3.bill_status <> 'CANCEL', booking_service_items.quality, 0)) AS export_sale,
		SUM(if(tb3.bill_status = 'CANCEL', booking_service_items.quality, 0)) AS export_cancel
	`)

	if item.CourseUid != "" {
		db = db.Where("booking_service_items.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("booking_service_items.partner_uid = ?", item.PartnerUid)
	}
	if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}

	subQuery := database.Table("bookings")

	if item.PartnerUid != "" {
		subQuery = subQuery.Where("bookings.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		subQuery = subQuery.Where("bookings.course_uid = ?", item.CourseUid)
	}
	if date != "" {
		subQuery = subQuery.Where("bookings.booking_date = ?", date)
	}

	subQuery = subQuery.Where("bookings.check_in_time > 0")
	// subQuery = subQuery.Where("bookings.check_out_time > 0")

	db = db.Joins("LEFT JOIN (?) as tb1 on tb1.uid = booking_service_items.booking_uid", subQuery)

	db = db.Joins("INNER JOIN group_services as tb2 on tb2.group_code = booking_service_items.group_code")

	db = db.Joins("INNER JOIN service_carts as tb3 on tb3.id = booking_service_items.service_bill")

	db.Group("booking_service_items.item_id")
	db.Order("booking_service_items.group_code asc")

	db = db.Find(&list)

	return list, db.Error
}

func (item *BookingServiceItem) FindListWithStatus(database *gorm.DB, page models.Page) ([]map[string]interface{}, int64, error) {
	db := database.Model(BookingServiceItem{})
	var list []map[string]interface{}
	total := int64(0)

	db = db.Select("booking_service_items.*", "COUNT(tb1.item_id) as order_counts", "COUNT(tb2.item_id) as process_counts")

	if item.CourseUid != "" {
		db = db.Where("booking_service_items.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("booking_service_items.partner_uid = ?", item.PartnerUid)
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
	if item.ItemCode != "" {
		db = db.Where("booking_service_items.item_code = ?", item.ItemCode)
	}
	if item.ServiceBill > 0 {
		db = db.Where("booking_service_items.service_bill = ?", item.ServiceBill)
	}

	// sub query
	subQuery1 := database.Table("restaurant_items")

	subQuery1 = subQuery1.Where("bill_id = ?", item.ServiceBill)

	subQuery1 = subQuery1.Where("item_status = ?", constants.RES_STATUS_ORDER)

	subQuery2 := database.Table("restaurant_items")

	subQuery2 = subQuery2.Where("bill_id = ?", item.ServiceBill)

	subQuery2 = subQuery2.Where("item_status = ?", constants.RES_STATUS_PROCESS)

	db = db.Joins("LEFT JOIN (?) as tb1 on tb1.item_id = booking_service_items.id", subQuery1)

	db = db.Joins("LEFT JOIN (?) as tb2 on tb2.item_id = booking_service_items.id", subQuery2)

	db.Group("booking_service_items.id")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) FindListWithBooking(database *gorm.DB, page models.Page, fromDate int64, toDate int64) ([]BookingServiceItemResponse, int64, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItemResponse{}
	total := int64(0)

	if item.GroupCode != "" {
		db = db.Where("booking_service_items.group_code = ?", item.GroupCode)
	}
	if item.ServiceId != "" {
		db = db.Where("booking_service_items.service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}
	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	if fromDate > 0 {
		db = db.Where("booking_service_items.created_at >= ?", fromDate)
	}

	if toDate > 0 {
		db = db.Where("booking_service_items.created_at <= ?", toDate)
	}

	db = db.Joins("JOIN bookings ON bookings.uid = booking_service_items.booking_uid")
	db = db.Select("booking_service_items.*, bookings.check_in_time")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingServiceItem) FindListBookingOrder(database *gorm.DB, date string) ([]BookingServiceItemResponse, int64, error) {
	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItemResponse{}
	total := int64(0)

	if item.CourseUid != "" {
		db = db.Where("booking_service_items.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("booking_service_items.partner_uid = ?", item.PartnerUid)
	}
	if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}

	db = db.Joins("RIGHT JOIN (select * from bookings where partner_uid = ? and course_uid = ? and booking_date = ?) b ON b.uid = booking_service_items.booking_uid", item.PartnerUid, item.CourseUid, date)

	db.Group("b.uid")
	db.Count(&total)

	db = db.Find(&list)

	return list, total, db.Error
}

func (item *BookingServiceItem) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *BookingServiceItem) DeleteWhere(database *gorm.DB) error {
	db := database.Model(BookingServiceItem{})
	return db.Delete(item).Error
}

// / ------- BookingServiceItem batch insert to db ------
func (item *BookingServiceItem) BatchInsert(database *gorm.DB, list []BookingServiceItem) error {
	db := database.Table("booking_service_items")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch insert err: ", err.Error())
	}
	return err
}

// ------ Batch Update ------
func (item *BookingServiceItem) BatchUpdate(database *gorm.DB, list []BookingServiceItem) error {
	db := database.Table("booking_service_items")
	var err error
	err = db.Updates(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch update err: ", err.Error())
	}
	return err
}

// ------ Find Best Item ------
func (item *BookingServiceItem) FindBestCartItem(database *gorm.DB, page models.Page) ([]BookingServiceItem, int64, error) {
	now := utils.GetTimeNow().Format("02/01/2006")

	from, _ := time.Parse("02/01/2006 15:04:05", now+" 17:00:00")

	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	total := int64(0)

	db.Select("*, sum(quality) as sale_quantity")

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.ServiceId != "" {
		db = db.Where("service_id = ?", item.ServiceId)
	}
	if item.GroupCode != "" {
		db = db.Where("group_code = ?", item.GroupCode)
	}

	db = db.Where("created_at >= ?", from.AddDate(0, 0, -8).Unix())

	db.Group("item_code")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

// ------ Find Best Group ------
func (item *BookingServiceItem) FindBestGroup(database *gorm.DB, page models.Page) ([]BookingServiceItem, int64, error) {
	now := utils.GetTimeNow().Format("02/01/2006")

	from, _ := time.Parse("02/01/2006 15:04:05", now+" 17:00:00")

	db := database.Model(BookingServiceItem{})
	list := []BookingServiceItem{}
	total := int64(0)

	db.Select("*, sum(quality) as sale_quantity")

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.ServiceId != "" {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	db = db.Where("created_at >= ?", from.AddDate(0, 0, -8).Unix())

	db.Group("group_code")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *BookingServiceItem) FindReportRevenuePOS(database *gorm.DB, formDate, toDate string) ([]map[string]interface{}, error) {
	db := database.Table("booking_service_items")
	var list []map[string]interface{}

	if item.Type == "KIOSK" || item.Type == "PROSHOP" {
		db.Select(`booking_service_items.name, booking_service_items.unit, tb3.group_name, booking_service_items.location, 
			sum(booking_service_items.quality) as quantity, 
			booking_service_items.unit_price, sum(booking_service_items.amount) as amount,
			booking_service_items.discount_type,
			booking_service_items.discount_value
		`)
	} else {
		db.Select(`booking_service_items.name, 
			booking_service_items.unit, tb3.group_name, 
			sum(booking_service_items.quality) as quantity, 
			booking_service_items.unit_price, sum(booking_service_items.amount) as amount,
			booking_service_items.discount_type,
			booking_service_items.discount_value
		`)
	}

	// if item.CourseUid != "" {
	// 	db = db.Where("booking_service_items.course_uid = ?", item.CourseUid)
	// }
	// if item.PartnerUid != "" {
	// 	db = db.Where("booking_service_items.partner_uid = ?", item.PartnerUid)
	// }
	if item.ServiceId != "" {
		db = db.Where("booking_service_items.service_id = ?", item.ServiceId)
	}
	if item.Name != "" {
		db = db.Where("booking_service_items.name LIKE ?", "%"+item.Name+"%")
	}
	if item.Type == constants.RESTAURANT_SETTING {
		db = db.Where("booking_service_items.type IN ?", []string{constants.RESTAURANT_SETTING, constants.MINI_R_SETTING})
	} else if item.Type != "" {
		db = db.Where("booking_service_items.type = ?", item.Type)
	}

	// sub query
	subQuery := database.Table("bookings")

	if item.CourseUid != "" {
		subQuery = subQuery.Where("bookings.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		subQuery = subQuery.Where("bookings.partner_uid = ?", item.PartnerUid)
	}
	if formDate != "" {
		subQuery = subQuery.Where("STR_TO_DATE(bookings.booking_date, '%d/%m/%Y') >= STR_TO_DATE(?, '%d/%m/%Y')", formDate)
	}
	if toDate != "" {
		subQuery = subQuery.Where("STR_TO_DATE(bookings.booking_date, '%d/%m/%Y') <= STR_TO_DATE(?, '%d/%m/%Y')", toDate)
	}

	subQuery1 := database.Table("group_services")

	if item.CourseUid != "" {
		subQuery1 = subQuery1.Where("group_services.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		subQuery1 = subQuery1.Where("group_services.partner_uid = ?", item.PartnerUid)
	}

	db = db.Joins(`INNER JOIN (?) as tb1 on booking_service_items.booking_uid = tb1.uid`, subQuery)
	db = db.Joins(`LEFT JOIN service_carts as tb2 on booking_service_items.service_bill = tb2.id`)
	db = db.Joins(`LEFT JOIN (?) as tb3 on booking_service_items.group_code = tb3.group_code`, subQuery1)

	db = db.Where("tb1.check_in_time > 0")
	db = db.Where("tb1.bag_status <> 'CANCEL'")
	db = db.Where("(tb2.bill_status NOT IN ? OR tb2.bill_status IS NULL)", []string{constants.RES_BILL_STATUS_CANCEL, constants.RES_BILL_STATUS_ORDER, constants.RES_BILL_STATUS_BOOKING, constants.POS_BILL_STATUS_PENDING})

	if item.Type == "KIOSK" || item.Type == "PROSHOP" {
		db.Group("booking_service_items.service_id, booking_service_items.item_code, booking_service_items.name, booking_service_items.unit_price, booking_service_items.discount_type, booking_service_items.discount_value")
		db.Order("booking_service_items.location, booking_service_items.name")
	} else {
		db.Group("booking_service_items.item_code, booking_service_items.name, booking_service_items.unit_price, booking_service_items.discount_type, booking_service_items.discount_value")
		db.Order("booking_service_items.name")
	}

	db = db.Find(&list)

	return list, db.Error
}

func (item *BookingServiceItem) FindReportDetailFB(database *gorm.DB, date string) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	db := database.Table("booking_service_items as tb")

	db = db.Select(`tb.bag, tb.player_name, tb.name, tb.location, tb.unit,
		SUM(tb.quality) as quality, tb.unit_price, tb.discount_type, tb.discount_value,
		SUM(tb.amount) as amount
	`)

	// if item.CourseUid != "" {
	// 	db = db.Where("tb.course_uid = ?", item.CourseUid)
	// }
	// if item.PartnerUid != "" {
	// 	db = db.Where("tb.partner_uid = ?", item.PartnerUid)
	// }

	if item.Type == constants.RESTAURANT_SETTING {
		db = db.Where("tb.type IN ?", []string{constants.RESTAURANT_SETTING, constants.MINI_R_SETTING})
	} else if item.Type != "" {
		db = db.Where("tb.type = ?", item.Type)
	} else {
		db = db.Where("tb.type IN ?", []string{constants.KIOSK_SETTING, constants.MINI_B_SETTING, constants.RESTAURANT_SETTING, constants.MINI_R_SETTING})
	}

	if item.GroupCode != "" {
		db = db.Where("tb.group_code = ?", item.GroupCode)
	}

	if item.ItemCode != "" {
		db = db.Where("tb.item_code = ?", item.ItemCode)
	}

	if item.Bag != "" {
		db = db.Where("tb.bag = ?", item.Bag)
	}

	// sub query
	subQuery := database.Table("bookings")

	if item.CourseUid != "" {
		subQuery = subQuery.Where("bookings.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		subQuery = subQuery.Where("bookings.partner_uid = ?", item.PartnerUid)
	}
	if date != "" {
		subQuery = subQuery.Where("bookings.booking_date = ?", date)
	}

	// subQuery = subQuery.Where("bookings.added_round = 0")

	db = db.Joins(`INNER JOIN (?) as tb1 on tb.booking_uid = tb1.uid`, subQuery)
	db = db.Joins(`LEFT JOIN service_carts as tb2 on tb.service_bill = tb2.id`)

	db = db.Where("tb1.check_in_time > 0")
	db = db.Where("tb1.bag_status <> 'CANCEL'")
	db = db.Where("(tb2.bill_status NOT IN ? OR tb2.bill_status IS NULL)", []string{constants.RES_BILL_STATUS_CANCEL, constants.RES_BILL_STATUS_ORDER, constants.RES_BILL_STATUS_BOOKING, constants.POS_BILL_STATUS_PENDING})

	db = db.Group("tb.bag, tb.item_code, tb.name, tb.location, tb.unit_price,  tb.discount_type, tb.discount_value")

	db = db.Find(&list)

	return list, db.Error
}

func (item *BookingServiceItem) FindListInApp(database *gorm.DB, page models.Page) ([]ListItemInApp, int64, error) {
	db := database.Model(BookingServiceItem{})
	list := []ListItemInApp{}
	total := int64(0)

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.ServiceBill > 0 {
		db = db.Where("service_bill = ?", item.ServiceBill)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
