package mdoel_report

import (
	"start/models"

	"gorm.io/gorm"
)

type ReportRevenueDetailList struct {
	PartnerUid string
	CourseUid  string
	FromDate   string
	ToDate     string
	GuestStyle string
}

type ResReportCashierAudit struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	TransTime  int64  `json:"trans_time"`
	Bag        string `json:"bag"`
	Cash       int64  `json:"cash"`    // Số tiền mặt
	Card       int64  `json:"card"`    // Số tiền cà thẻ
	Voucher    int64  `json:"voucher"` // Số tiền Voucher
	Debit      int64  `json:"debit"`   // Số tiền nợ
}

func addFilter(db *gorm.DB, item *ReportRevenueDetailList) *gorm.DB {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.FromDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') >= ?", item.FromDate)
	}

	if item.ToDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') <= ?", item.ToDate)
	}

	if item.GuestStyle != "" {
		db = db.Where("guest_style = ?", item.GuestStyle)
	}

	return db
}

func (item *ReportRevenueDetailList) FindBookingRevenueList(database *gorm.DB, page models.Page) ([]ReportRevenueDetail, int64, error) {
	total := int64(0)
	var list []ReportRevenueDetail

	db := database.Model(ReportRevenueDetail{})

	db = addFilter(db, item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
