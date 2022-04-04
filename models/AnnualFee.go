package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Phí thường niên
// TODO: Chú ý logic số tiền phải trả và số tiền trả từng đợt
type AnnualFee struct {
	ModelId
	PartnerUid        string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid         string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	MemberCardUid     string `json:"member_card_uid" gorm:"type:varchar(100);index"` // Member Card Uid
	Year              int    `json:"year"`                                           // Year
	PaymentType       string `json:"payment_type" gorm:"type:varchar(50);index"`     // TM, CK, CC, TM+CK, TM+CC
	BillNumber        string `json:"bill_number" gorm:"type:varchar(100)"`           //
	Note              string `json:"note" gorm:"type:varchar(256)"`                  //
	AnnualQuotaAmount int64  `json:"annual_quota_amount"`                            // Tiền Phí thuờng niên
	PrePaid           int64  `json:"pre_paid"`                                       // A: Số tiền khách nộp trước khi chạy phần mềm
	PaidForfeit       int64  `json:"paid_forfeit"`                                   // B: Số Tiền phạt do thanh toán chậm
	PaidReduce        int64  `json:"paid_reduce"`                                    // C: Số Tiền giảm trừ khi nộp sớm
	LastYearDebit     int64  `json:"last_year_debit"`                                // D: Số tiền nợ từ năm ngoái
	// MustPaid          int64  `json:"must_paid"`                                      // K: Số tiền Phí khách hàng đó pải đóng K = A-B+C-D+E
	TotalPaid int64 `json:"total_paid"` // G: Tổng số tiền các lần khách trả
	// Debit             int64  `json:"debit"`                                          // H: tiền nợ H = K - G
	PlayCountsAdd int    `json:"play_counts_add"`                    //
	DaysPaid      string `json:"days_paid" gorm:"type:varchar(256)"` // Ghi lại các ngày thanh toán của khách
}

func (item *AnnualFee) IsDuplicated() bool {
	modelCheck := AnnualFee{
		PartnerUid:    item.PartnerUid,
		CourseUid:     item.CourseUid,
		MemberCardUid: item.MemberCardUid,
		Year:          item.Year,
	}
	errFind := modelCheck.FindFirst()
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *AnnualFee) IsValidated() bool {
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
	return true
}

func (item *AnnualFee) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *AnnualFee) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AnnualFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *AnnualFee) Count() (int64, error) {
	db := datasources.GetDatabase().Model(AnnualFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AnnualFee) FindList(page Page) ([]AnnualFee, int64, error) {
	db := datasources.GetDatabase().Model(AnnualFee{})
	list := []AnnualFee{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *AnnualFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
