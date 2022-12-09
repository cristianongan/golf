package mdoel_report

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ReportRevenueDetail struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BillNo         string `json:"bill_no" gorm:"type:varchar(50);index"`      // Mã bill
	BookingDate    string `json:"booking_date" gorm:"type:varchar(50);index"` // Ngày chơi
	CustomerId     string `json:"customer_id" gorm:"type:varchar(50);index"`  // Mã KH
	CustomerName   string `json:"customer_name"`                              // Tên KH
	MembershipNo   string `json:"member_ship_no" gorm:"type:varchar(50)"`     // Mã thẻ thành viên
	CustomerType   string `json:"customer_type" gorm:"type:varchar(50)"`      // Loại KH
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(50);index"`  // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(100)"`  // Guest Style Name
	Hole           int    `json:"hole" gorm:"type:varchar(50)"`               // Số Hole
	GreenFee       int64  `json:"green_fee"`                                  // Phí sân cỏ
	CaddieFee      int64  `json:"caddie_fee"`                                 // Phí caddie
	SubBagFee      int64  `json:"sub_bag_fee"`                                // Phí trả cho sub bag
	FBFee          int64  `json:"fb_fee"`                                     // Phí ăn uống
	RentalFee      int64  `json:"rental_fee"`                                 // Phí thuê đồ
	BuggyFee       int64  `json:"buggy_fee"`                                  // Phí thuê xe
	ProshopFee     int64  `json:"proshop_fee"`                                // Phí đồ ở Proshop
	PraticeBallFee int64  `json:"pratice_ball_fee"`                           // Phí bóng tập
	OtherFee       int64  `json:"other_fee"`                                  // Phí khác
	Cash           int64  `json:"cash"`                                       // Số tiền mặt
	Card           int64  `json:"card"`                                       // Số tiền cà thẻ
	Voucher        int64  `json:"voucher"`                                    // Số tiền Voucher
	Debit          int64  `json:"debit"`                                      // Số tiền nợ
	Amout          int64  `json:"amount"`                                     // Tổng tiền
	MushPay        int64  `json:"mush_pay"`                                   // Tổng tiền phải trả
}

// ======= CRUD ===========
func (item *ReportRevenueDetail) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *ReportRevenueDetail) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
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

func (item *ReportRevenueDetail) FindList(page models.Page) ([]ReportRevenueDetail, int64, error) {
	db := datasources.GetDatabase().Model(ReportRevenueDetail{})
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

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *ReportRevenueDetail) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
