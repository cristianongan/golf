package mdoel_report

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
)

// Report thông tin chơi của khách hàng
type ReportCustomerPlay struct {
	models.ModelId
	PartnerUid         string  `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid          string  `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	CustomerUid        string  `json:"customer_uid" gorm:"type:varchar(100);index"` // Uid customer
	CardId             string  `json:"card_id" gorm:"type:varchar(50);index"`       // Uid customer
	TotalPaid          int64   `json:"total_paid"`                                  // Tổng thanh toán
	TotalPlayCount     int     `json:"total_play_count"`                            // Tổng lượt chơi
	TotalHourPlayCount float64 `json:"total_hour_play_count"`                       // Tổng giờ chơi
}

// ======= CRUD ===========
func (item *ReportCustomerPlay) Create() error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *ReportCustomerPlay) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *ReportCustomerPlay) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *ReportCustomerPlay) Count() (int64, error) {
	db := datasources.GetDatabase().Model(ReportCustomerPlay{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *ReportCustomerPlay) FindList(page models.Page) ([]ReportCustomerPlay, int64, error) {
	db := datasources.GetDatabase().Model(ReportCustomerPlay{})
	list := []ReportCustomerPlay{}
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

func (item *ReportCustomerPlay) FindAllList() ([]ReportCustomerPlay, int64, error) {
	db := datasources.GetDatabase().Model(ReportCustomerPlay{})
	list := []ReportCustomerPlay{}
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
	if item.CardId != "" {
		db = db.Where("card_id = ?", item.CardId)
	}
	if item.CustomerUid != "" {
		db = db.Where("customer_uid = ?", item.CustomerUid)
	}

	db.Count(&total)
	db.Find(&list)
	return list, total, db.Error
}

func (item *ReportCustomerPlay) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
