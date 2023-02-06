package mdoel_report

import (
	"start/constants"
	"start/models"
	"start/utils"
	"time"

	"gorm.io/gorm"
)

type ReportRevenueDetailList struct {
	PartnerUid string
	CourseUid  string
	FromDate   string
	ToDate     string
	GuestStyle string
	Month      int
	Year       int
	Zone       string
}

type ResReportCashierAudit struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	TransTime  int64  `json:"trans_time"`
	Bag        string `json:"bag"`
	Cash       int64  `json:"cash"`     // Số tiền mặt
	Card       int64  `json:"card"`     // Số tiền cà thẻ
	Voucher    int64  `json:"voucher"`  // Số tiền Voucher
	Debit      int64  `json:"debit"`    // Số tiền nợ
	Transfer   int64  `json:"transfer"` // Số tiền chuyển khoản
}

type ResReportGolfService struct {
	YearToMonth1 GolfService `json:"year_to_month_1"`
	YearToMonth2 GolfService `json:"year_to_month_2"`
}

type GolfService struct {
	GreenFee  int64 `json:"green_fee"`
	CaddieFee int64 `json:"caddie_fee"`
	BuggyFee  int64 `json:"buggy_fee"`
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

func (item *ReportRevenueDetailList) FindGolfFeeRevenue(database *gorm.DB) ResReportGolfService {
	now := utils.GetTimeNow()
	currentYear, _, _ := now.Date()

	yearToMonth1 := getGolfFeeService(item.Year, item.Month, database)
	yearToMonth2 := getGolfFeeService(currentYear, item.Month, database)

	return ResReportGolfService{
		YearToMonth1: yearToMonth1,
		YearToMonth2: yearToMonth2,
	}
}

func getGolfFeeService(year int, month int, database *gorm.DB) GolfService {
	now := utils.GetTimeNow()
	currentLocation := now.Location()

	firstOfMonth2 := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, currentLocation)
	firstOfMonth2Str := firstOfMonth2.Format(constants.DATE_FORMAT)

	lastOfMonth2 := firstOfMonth2.AddDate(0, 1, -1)
	lastOfMonth2Str := lastOfMonth2.Format(constants.DATE_FORMAT)

	rReportRevenueDetailList1 := ReportRevenueDetailList{
		FromDate: firstOfMonth2Str,
		ToDate:   lastOfMonth2Str,
	}
	db1 := database.Model(ReportRevenueDetail{})
	db1 = addFilter(db1, &rReportRevenueDetailList1)

	var list2 []ReportRevenueDetail
	db1.Find(&list2)

	caddieFee := int64(0)
	greenFee := int64(0)
	buggyFee := int64(0)
	for _, item2 := range list2 {
		caddieFee += item2.CaddieFee
		greenFee += item2.GreenFee
		buggyFee += item2.BuggyFee
	}

	return GolfService{
		CaddieFee: caddieFee,
		GreenFee:  greenFee,
		BuggyFee:  buggyFee,
	}
}
