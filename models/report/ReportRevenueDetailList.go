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

type ReportBagRevenue struct {
	Data    []ReportRevenueDetail `json:"data"`
	Revenue DayEndRevenue         `json:"revenue"`
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
		db = db.Where("guest_style COLLATE utf8mb4_general_ci LIKE ?", "%"+item.GuestStyle+"%")
	}

	return db
}

func (item *ReportRevenueDetailList) FindReportDayEnd(database *gorm.DB) (DayEndRevenue, error) {
	db := database.Model(ReportRevenueDetail{})

	db = addFilter(db, item)

	db = db.Select(`partner_uid,
					course_uid,
					SUM(paid) AS agency_paid, 
					SUM(green_fee) AS green_fee, 
					SUM(caddie_fee) AS caddie_fee,
					SUM(buggy_fee) AS buggy_fee,
					SUM(pratice_ball_fee) AS pratice_ball_fee,
					SUM(restaurant_fee) AS restaurant_fee,
					SUM(kiosk_fee) AS kiosk_fee,
					SUM(proshop_fee) AS proshop_fee,
					SUM(minibar_fee) AS minibar_fee,
					SUM(booking_caddie_fee) AS booking_caddie_fee,
					SUM(mush_pay) as mush_pay,
					SUM(rental_fee) as rental_fee,
					SUM(other_fee) as other_fee,
					SUM(fb_fee) as fb_fee,
					SUM(total) as all_fee,
					SUM(phi_phat) as phi_phat,
					SUM(cash) as cash,
					SUM(transfer) as transfer,
					SUM(card) as card,
					SUM(debit) as debit,
					SUM(customer_type = 'MEMBER') AS member,
					SUM(customer_type = 'GUEST') AS member_guest,
					SUM(customer_type = 'VISITOR') AS visitor,
					SUM(customer_type = 'FOC') AS foc,
					SUM(green_fee + caddie_fee + buggy_fee + pratice_ball_fee + restaurant_fee + kiosk_fee + proshop_fee + minibar_fee + booking_caddie_fee + rental_fee + other_fee + phi_phat) AS total_fee,
					SUM(customer_type = 'TRADITIONAL' || customer_type = 'OTA') AS tour`)

	dayEnd := DayEndRevenue{}
	db.Find(&dayEnd)

	return dayEnd, db.Error
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

func (item *ReportRevenueDetailList) DeleteByBookingDate(database *gorm.DB) error {
	var list []ReportRevenueDetail

	db := database.Model(ReportRevenueDetail{})
	db = addFilter(db, item)
	db.Find(&list)

	return db.Delete(item).Error
}
