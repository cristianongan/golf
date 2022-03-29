package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Loại thẻ
type AnnualFee struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // Member Card Uid
	Year          int    `json:"year"`                                           // Year
	PaymentType   string `json:"payment_type" gorm:"type:varchar(50);index"`     // TM, CK, CC, TM+CK, TM+CC
	BillNumber    string `json:"bill_number" gorm:"type:varchar(100)"`
	Note          string `json:"note" gorm:"type:varchar(256)"`
	Amount        int64  `json:"amount"` // Tiền
}

func (item *AnnualFee) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
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
