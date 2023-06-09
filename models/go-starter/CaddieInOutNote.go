package model_gostarter

import (
	"fmt"
	"start/constants"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieBuggyInOut struct {
	models.ModelId
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BookingUid      string `json:"booking_uid" gorm:"type:varchar(50);index"`  // Ex: Booking Uid
	BookingDate     string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	Bag             string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	CaddieId        int64  `json:"caddie_id" gorm:"index"`                     // Caddie Id
	CaddieCode      string `json:"caddie_code" gorm:"type:varchar(256)"`
	BuggyId         int64  `json:"buggy_id"`                                   // Buggy Id
	BuggyCode       string `json:"buggy_code" gorm:"type:varchar(100)"`        // Buggy Code
	Note            string `json:"note" gorm:"type:varchar(500)"`              // note
	CaddieType      string `json:"caddie_type"`                                // Type: IN(undo), OUT, CHANGE
	BuggyType       string `json:"buggy_type"`                                 // Type: IN(undo), OUT, CHANGE
	Hole            int    `json:"hole"`                                       // hole caddie
	HoleBuggy       int    `json:"hole_buggy"`                                 // hole buggy
	BagShareBuggy   string `json:"bag_share_buggy" gorm:"type:varchar(100)"`   // Bag đi chung với buggy
	IsPrivateBuggy  *bool  `json:"is_private_buggy" gorm:"default:0"`          // Bag có dùng buggy riêng không
	BuggyCommonCode string `json:"buggy_common_code" gorm:"type:varchar(100)"` // Đánh dấu record có chung buggy
}

type CaddieBuggyInOutWithBooking struct {
	CaddieBuggyInOut
	Bag            string `json:"bag"`
	TeeOff         string `json:"tee_off"`
	IsPrivateBuggy bool   `json:"is_private_buggy"`
	GuestStyle     string `json:"guest_style"`
	GuestStyleName string `json:"guest_style_name"`
}

type ReportBuggy struct {
	H_9         int    `json:"h_9"`
	H_18        int    `json:"h_18"`
	H_36        int    `json:"h_36"`
	H_45        int    `json:"h_45"`
	H_54        int    `json:"h_54"`
	BookingDate string `json:"booking_date"`
}

type CaddieBuggyInOutRequest struct {
	CaddieBuggyInOut
	Bag            string
	Date           string
	ShareBuggy     *bool
	BagOrBuggyCode string
	BookingDate    string `json:"booking_date"`
}

func (item *CaddieBuggyInOut) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CaddieBuggyInOut) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieBuggyInOut) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieBuggyInOut) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CaddieBuggyInOut{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieBuggyInOut) FindAllCaddieInOutNotes(database *gorm.DB, date string) ([]CaddieBuggyInOut, error) {
	from, _ := time.Parse("02/01/2006 15:04:05", date+" 17:00:00")

	to, _ := time.Parse("02/01/2006 15:04:05", date+" 16:59:59")

	db := database.Model(CaddieBuggyInOut{})
	list := []CaddieBuggyInOut{}

	if item.CaddieId != 0 {
		db = db.Where("caddie_id = ?", item.CaddieId)
	}

	db = db.Where("caddie_type = ?", constants.STATUS_OUT)
	db = db.Where("created_at >= ?", from.AddDate(0, 0, -1).Unix())
	db = db.Where("created_at < ?", to.Unix())

	db.Find(&list)
	return list, db.Error
}

func (item *CaddieBuggyInOut) FindList(database *gorm.DB, page models.Page, from, to int64) ([]CaddieBuggyInOut, int64, error) {
	db := database.Model(CaddieBuggyInOut{})
	list := []CaddieBuggyInOut{}
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

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieBuggyInOut) FindOrderByDateList(database *gorm.DB) ([]CaddieBuggyInOut, int64, error) {
	db := database.Model(CaddieBuggyInOut{})
	list := []CaddieBuggyInOut{}
	total := int64(0)
	item.ModelId.Status = ""

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingUid != "" {
		db = db.Where("booking_uid = ?", item.BookingUid)
	}
	if item.BuggyType != "" {
		db = db.Where("buggy_type = ?", item.BuggyType)
	}
	if item.CaddieType != "" {
		db = db.Where("caddie_type = ?", item.CaddieType)
	}
	if item.CaddieId > 0 {
		db = db.Where("caddie_id = ?", item.CaddieId)
	}
	if item.BuggyId > 0 {
		db = db.Where("buggy_id = ?", item.BuggyId)
	}

	db = db.Order("created_at desc")
	db.Count(&total)
	db.Find(&list)
	return list, total, db.Error
}

func (item *CaddieBuggyInOut) FindCaddieBuggyInOutWithBooking(database *gorm.DB, page models.Page, request CaddieBuggyInOutRequest) ([]CaddieBuggyInOutWithBooking, int64, error) {
	db := database.Model(CaddieBuggyInOut{})
	list := []CaddieBuggyInOutWithBooking{}
	total := int64(0)
	db = db.Joins("JOIN bookings ON bookings.uid = caddie_buggy_in_outs.booking_uid")
	db = db.Select("caddie_buggy_in_outs.*,bookings.bag,bookings.guest_style,bookings.guest_style_name,bookings.is_private_buggy,bookings.tee_off_time as tee_off")

	if item.PartnerUid != "" {
		db = db.Where("caddie_buggy_in_outs.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("caddie_buggy_in_outs.course_uid = ?", item.CourseUid)
	}
	if item.BookingUid != "" {
		db = db.Where("caddie_buggy_in_outs.booking_uid = ?", item.BookingUid)
	}
	if item.BuggyType != "" {
		db = db.Where("caddie_buggy_in_outs.buggy_type = ?", item.BuggyType)
	}
	if item.CaddieType != "" {
		db = db.Where("caddie_buggy_in_outs.caddie_type = ?", item.CaddieType)
	}
	if item.CaddieCode != "" {
		db = db.Where("caddie_buggy_in_outs.caddie_code = ?", item.CaddieCode)
	}
	if request.BagOrBuggyCode != "" {
		db = db.Where("caddie_buggy_in_outs.buggy_code LIKE ? OR bookings.bag LIKE ?", "%"+request.BagOrBuggyCode+"%", "%"+request.BagOrBuggyCode+"%")
	}
	if request.ShareBuggy != nil {
		db = db.Where("bookings.is_private_buggy = ?", *request.ShareBuggy)
	}
	if request.Bag != "" {
		db = db.Where("bookings.bag = ?", request.Bag)
	}
	// localTime, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())
	if request.Date != "" {
		db = db.Where("bookings.booking_date = ?", request.Date)
	}

	db = db.Group("bookings.bag")
	db = db.Group("caddie_buggy_in_outs.buggy_code")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieBuggyInOut) FindReportBuggyUsing(database *gorm.DB, month, year string) ([]ReportBuggy, error) {
	subQuery1 := database.Model(CaddieBuggyInOut{})
	list := []ReportBuggy{}

	if item.PartnerUid != "" {
		subQuery1 = subQuery1.Where("caddie_buggy_in_outs.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		subQuery1 = subQuery1.Where("caddie_buggy_in_outs.course_uid = ?", item.CourseUid)
	}

	subQuery1 = subQuery1.Where("DATE_FORMAT(STR_TO_DATE(bookings.booking_date, '%d/%m/%Y'), '%Y-%m') = ?", fmt.Sprintf("%s-%s", year, month))
	subQuery1 = subQuery1.Where("caddie_buggy_in_outs.buggy_type = ?", constants.STATUS_OUT)
	subQuery1 = subQuery1.Joins("JOIN bookings ON bookings.uid = caddie_buggy_in_outs.booking_uid")
	subQuery1 = subQuery1.Group("caddie_buggy_in_outs.buggy_code")
	subQuery1 = subQuery1.Group("bookings.booking_date")
	subQuery1 = subQuery1.Select("SUM(caddie_buggy_in_outs.hole) as hole_buggy, bookings.booking_date as booking_date")

	subQuery2 := database.Table("(?) as tb1", subQuery1)
	subQuery2 = subQuery2.Select(`
					SUM(if(tb1.hole_buggy = 9, 1, 0)) AS h_9,
					SUM(if(tb1.hole_buggy = 18, 1, 0)) AS h_18,
					SUM(if(tb1.hole_buggy = 36, 1, 0)) AS h_36,
					SUM(if(tb1.hole_buggy = 45, 1, 0)) AS h_45,
					SUM(if(tb1.hole_buggy = 54, 1, 0)) AS h_54,
					tb1.booking_date`)
	subQuery2 = subQuery2.Group("tb1.booking_date")

	subQuery2.Find(&list)
	return list, subQuery1.Error
}

func (item *CaddieBuggyInOut) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
