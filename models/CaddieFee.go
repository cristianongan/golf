package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Caddie Fee
type CaddieFee struct {
	ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CaddieId    int64  `json:"caddie_id" gorm:"index"`                     // caddie id
	CaddieCode  string `json:"caddie_code" gorm:"type:varchar(256);index"` // caddie code
	CaddieName  string `json:"caddie_name" gorm:"type:varchar(256)"`       // caddie name
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // ngày booking
	Hole        int    `json:"hole"`                                       // số hố
	Round       int64  `json:"round"`                                      // số round
	Amount      int64  `json:"amount"`                                     // tổng số tiền
	TotalAmount int64  `json:"total_amount"`                               // tông số tiền trong 1 tháng
}

func (item *CaddieFee) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieFee) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieFee) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CaddieFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFee) FindList(page Page) ([]CaddieFee, int64, error) {
	db := datasources.GetDatabase().Model(CaddieFee{})
	list := []CaddieFee{}
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

func (item *CaddieFee) FindAll(month string) ([]CaddieFee, int64, error) {
	db := datasources.GetDatabase().Model(CaddieFee{})
	list := []CaddieFee{}
	total := int64(0)
	db = db.Where(item)

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CaddieCode != "" {
		db = db.Where("caddie_code = ?", item.CaddieCode)
	}
	if month != "" {
		db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", month)
	}

	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *CaddieFee) FindAllGroupBy(page Page, month string) ([]CaddieFee, int64, error) {
	db := datasources.GetDatabase().Model(CaddieFee{})
	list := []CaddieFee{}
	total := int64(0)

	db.Select("*, sum(amount) as total_amount")

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("s", item.PartnerUid)
	}
	if item.CaddieCode != "" {
		db = db.Where("caddie_code LIKE ?", "%"+item.CaddieCode+"%")
	}
	if item.CaddieName != "" {
		db = db.Where("caddie_name LIKE ?", "%"+item.CaddieName+"%")
	}
	if month != "" {
		db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", month)
	}

	db.Group("caddie_id")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *CaddieFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
