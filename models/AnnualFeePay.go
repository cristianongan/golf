package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Số tiền trả từng đợt
type AnnualFeePay struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // Member Card Uid
	Year          int    `json:"year" gorm:"index"`                              // Year
	PaymentType   string `json:"payment_type" gorm:"type:varchar(50);index"`     // TM, CK, CC, TM+CK, TM+CC
	BillNumber    string `json:"bill_number" gorm:"type:varchar(100)"`           // Hoá đơn
	Note          string `json:"note" gorm:"type:varchar(256)"`                  // ghi chú
	PayDate       string `json:"pay_date" gorm:"type:varchar(256)"`              // Ngày thanh toán
	Amount        int64  `json:"amount"`                                         // Số tiền thanh toán đợt này
	CmsUser       string `json:"cms_user" gorm:"type:varchar(100)"`              // Cms User
}

func (item *AnnualFeePay) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.MemberCardUid == "" {
		return false
	}
	if item.Year <= 0 {
		return false
	}
	if item.Amount <= 0 {
		return false
	}
	return true
}

func (item *AnnualFeePay) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *AnnualFeePay) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AnnualFeePay) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *AnnualFeePay) Count(database *gorm.DB) (int64, error) {
	db := database.Model(AnnualFeePay{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AnnualFeePay) FindList(database *gorm.DB, page Page) ([]AnnualFeePay, int64, error) {
	db := database.Model(AnnualFeePay{})
	list := []AnnualFeePay{}
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
	if item.MemberCardUid != "" {
		db = db.Where("member_card_uid = ?", item.MemberCardUid)
	}
	if item.Year > 0 {
		db = db.Where("year = ?", item.Year)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *AnnualFeePay) FindTotalPaid(database *gorm.DB) int64 {
	db := database.Model(AnnualFeePay{})

	sumStr := utils.TotalStruct{}

	db = db.Select("sum(amount) as total_amount")

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.MemberCardUid != "" {
		db = db.Where("member_card_uid = ?", item.MemberCardUid)
	}
	if item.Year > 0 {
		db = db.Where("year = ?", item.Year)
	}

	db = db.Group("member_card_uid").First(&sumStr)

	return sumStr.TotalAmount
}

func (item *AnnualFeePay) FindAll(database *gorm.DB) ([]AnnualFeePay, error) {
	db := database.Model(AnnualFeePay{})
	list := []AnnualFeePay{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.MemberCardUid != "" {
		db = db.Where("member_card_uid = ?", item.MemberCardUid)
	}
	if item.Year > 0 {
		db = db.Where("year = ?", item.Year)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *AnnualFeePay) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
