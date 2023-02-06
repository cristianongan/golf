package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	Note        string `json:"note" gorm:"type:varchar(256)"`              // note
	IsDayOff    *bool  `json:"is_day_off"`
	TotalAmount int64  `json:"total_amount"` // tông số tiền trong 1 tháng
}

func (item *CaddieFee) IsDuplicated(db *gorm.DB) bool {
	caddieFeeCheck := CaddieFee{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		CaddieId:    item.CaddieId,
		BookingDate: item.BookingDate,
	}
	errFind := caddieFeeCheck.FindFirst(db)
	if errFind == nil || caddieFeeCheck.Id > 0 {
		return true
	}
	return false
}

func (item *CaddieFee) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CaddieFee) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieFee) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieFee) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CaddieFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieFee) FindList(database *gorm.DB, page Page) ([]CaddieFee, int64, error) {
	db := database.Model(CaddieFee{})
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

func (item *CaddieFee) FindAll(database *gorm.DB, month string) ([]CaddieFee, int64, error) {
	db := database.Model(CaddieFee{})
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

	db.Distinct("caddie_id", "booking_date", "hole", "round", "amount", "note", "is_day_off")

	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *CaddieFee) FindAllGroupBy(database *gorm.DB, page Page, month string) ([]map[string]interface{}, int64, error) {
	var list []map[string]interface{}
	total := int64(0)

	// sub query
	subQuery1 := database.Model(CaddieFee{})
	// subQuery1.Select("caddie_fees.*")

	if item.CourseUid != "" {
		subQuery1 = subQuery1.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		subQuery1 = subQuery1.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CaddieCode != "" {
		subQuery1 = subQuery1.Where("caddie_code LIKE ?", "%"+item.CaddieCode+"%")
	}
	if item.CaddieName != "" {
		subQuery1 = subQuery1.Where("caddie_name LIKE ?", "%"+item.CaddieName+"%")
	}
	if month != "" {
		subQuery1 = subQuery1.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", month)
	}

	subQuery1.Distinct("created_at", "caddie_id", "caddie_name", "caddie_code", "booking_date", "hole", "round", "amount", "note", "is_day_off")

	db := database.Table("(?) as tb0", subQuery1).Select("*, sum(tb0.amount) as total_amount")

	db.Group("tb0.caddie_id")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *CaddieFee) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
