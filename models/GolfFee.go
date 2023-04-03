package models

import (
	"fmt"
	"log"
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Phí Golf
type GolfFee struct {
	ModelId
	PartnerUid       string                `json:"partner_uid" gorm:"type:varchar(100);index"`      // Hang Golf
	CourseUid        string                `json:"course_uid" gorm:"type:varchar(256);index"`       // San Golf
	TablePriceId     int64                 `json:"table_price_id" gorm:"index"`                     // Id Bang gia
	GuestStyleName   string                `json:"guest_style_name" gorm:"type:varchar(256)"`       // Ten Guest style
	GuestStyle       string                `json:"guest_style" gorm:"index;type:varchar(200)"`      // Guest style
	Dow              string                `json:"dow" gorm:"type:varchar(100)"`                    // Dow
	GreenFee         utils.ListGolfHoleFee `json:"green_fee" gorm:"type:varchar(256)"`              // Phi san cỏ
	CaddieFee        utils.ListGolfHoleFee `json:"caddie_fee" gorm:"type:varchar(256)"`             // Phi Caddie
	BuggyFee         utils.ListGolfHoleFee `json:"buggy_fee" gorm:"type:varchar(256)"`              // Phi buggy
	UpdateUserName   string                `json:"update_user_name"`                                // Nguoi sua
	AccCode          string                `json:"acc_code" gorm:"type:varchar(200)"`               // Kết nối với phần mềm kế toán
	Note             string                `json:"note" gorm:"type:varchar(500)"`                   // Note
	NodeOdd          int                   `json:"node_odd"`                                        // 0 || 1 Chỉ tính hố lẻ thì tick vào đây
	PaidType         string                `json:"paid_type" gorm:"type:varchar(50)"`               // Kiểu thanh toán: NOW / AFTER
	Idx              int                   `json:"idx"`                                             // Xắp xếp thứ tự
	AccDebit         string                `json:"acc_debit"`                                       // Mã kế toán ghi nợ
	CustomerType     string                `json:"customer_type" gorm:"index;type:varchar(100)"`    // Loại khách hàng
	CustomerCategory string                `json:"customer_category" gorm:"index;type:varchar(50)"` // CUSTOMER, AGENCY
	GroupName        string                `json:"group_name" gorm:"index;type:varchar(200)"`       // Tên nhóm Fee
	GroupId          int64                 `json:"group_id" gorm:"index"`                           // Id nhóm Fee
	ApplyTime        string                `json:"apply_time" gorm:"type:varchar(100)"`             // Time áp dụng
	TaxCode          string                `json:"tax_code" gorm:"type:varchar(20)"`                // VAT
}

type GuestStyle struct {
	TablePriceId     int64  `json:"table_price_id"`    // Id Bang gia
	GuestStyleName   string `json:"guest_style_name"`  // Ten Guest style
	GuestStyle       string `json:"guest_style"`       // Guest style
	CustomerType     string `json:"customer_type"`     // Loại khách hàng
	CustomerCategory string `json:"customer_category"` // Loại khách hàng
	Dow              string `json:"dow"`               // Dow
}

func (item *GolfFee) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

// / ------- GolfFee batch insert to db ------
func (item *GolfFee) BatchInsert(database *gorm.DB, list []GolfFee) error {
	db := database.Table("golf_fees")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("GolfFee batch insert err: ", err.Error())
	}
	return err
}

func (item *GolfFee) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GolfFee) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *GolfFee) Count(database *gorm.DB) (int64, error) {
	db := database.Model(GolfFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

/*
Get Golf Fee valid in to day
*/
func (item *GolfFee) GetGuestStyleOnDay(database *gorm.DB) (GolfFee, error) {
	golfFee := GolfFee{
		GuestStyle: item.GuestStyle,
	}

	if item.GuestStyle == "" {
		return golfFee, errors.New("Guest style invalid")
	}

	// Get table price avaible trước
	tablePriceR := TablePrice{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
	}
	tablePrice, errFTB := tablePriceR.FindCurrentUse(database)
	if errFTB != nil {
		return golfFee, errFTB
	}

	list := []GolfFee{}
	db := database.Model(GolfFee{})
	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("course_uid = ?", item.CourseUid)
	db = db.Where("guest_style = ?", item.GuestStyle)
	db = db.Where("table_price_id = ?", tablePrice.Id)
	err := db.Find(&list).Error

	if err != nil {
		return golfFee, err
	}

	//Check có setup time theo h không
	isHaveHour := false
	for _, v := range list {
		if v.ApplyTime != "" {
			isHaveHour = true
		}
	}

	if isHaveHour {
		// Xử lý check theo giờ
		// check nhung row có hour trước
		idxTemp := -1

		for i, gf := range list {
			if gf.ApplyTime != "" {
				if idxTemp < 0 {
					if CheckDow(gf.Dow, gf.ApplyTime, utils.GetLocalUnixTime(), item.PartnerUid, item.CourseUid) {
						idxTemp = i
					}
				}
			}
		}

		if idxTemp >= 0 {
			return list[idxTemp], nil
		}
	}

	// Không có hour check theo ngày như bt
	idxTemp := -1

	for i, golfFee_ := range list {
		if idxTemp < 0 {
			if CheckDow(golfFee_.Dow, "", utils.GetLocalUnixTime(), item.PartnerUid, item.CourseUid) {
				idxTemp = i
			}
		}
	}

	if idxTemp >= 0 {
		return list[idxTemp], nil
	}

	return golfFee, errors.New("No guest style on day")
}

func (item *GolfFee) GetGuestStyleOnTime(database *gorm.DB, time time.Time) (GolfFee, error) {
	golfFee := GolfFee{
		GuestStyle: item.GuestStyle,
	}

	if item.GuestStyle == "" {
		return golfFee, errors.New("Guest style invalid")
	}

	// Get table price avaible trước
	tablePriceR := TablePrice{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
	}
	tablePrice, errFTB := tablePriceR.FindCurrentUse(database)
	if errFTB != nil {
		return golfFee, errFTB
	}

	list := []GolfFee{}
	db := database.Model(GolfFee{})
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	db = db.Where("course_uid = ?", item.CourseUid)
	db = db.Where("guest_style = ?", item.GuestStyle)
	db = db.Where("table_price_id = ?", tablePrice.Id)
	err := db.Find(&list).Error

	if err != nil {
		return golfFee, err
	}

	//Check có setup time theo h không
	isHaveHour := false
	for _, v := range list {
		if v.ApplyTime != "" {
			isHaveHour = true
		}
	}

	if isHaveHour {
		// Xử lý check theo giờ
		// check nhung row có hour trước
		idxTemp := -1

		for i, gf := range list {
			if gf.ApplyTime != "" {
				if idxTemp < 0 {
					if CheckDow(gf.Dow, gf.ApplyTime, time, item.PartnerUid, item.CourseUid) {
						idxTemp = i
					}
				}
			}
		}

		if idxTemp >= 0 {
			return list[idxTemp], nil
		}
	}

	// Không có hour check theo ngày như bt
	idxTemp := -1

	for i, golfFee_ := range list {
		if idxTemp < 0 {
			if CheckDow(golfFee_.Dow, "", time, item.PartnerUid, item.CourseUid) {
				idxTemp = i
			}
		}
	}

	if idxTemp >= 0 {
		return list[idxTemp], nil
	}

	return golfFee, errors.New("No guest style on day")
}

func (item *GolfFee) FindList(database *gorm.DB, page Page, isToday, dow string) ([]GolfFee, int64, error) {
	db := database.Model(GolfFee{})
	list := []GolfFee{}
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
	if item.GroupId > 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}
	if item.TablePriceId > 0 {
		db = db.Where("table_price_id = ?", item.TablePriceId)
	}
	if item.GuestStyle != "" {
		db = db.Where("guest_style = ?", item.GuestStyle)
	}
	if item.GuestStyleName != "" {
		db = db.Where("guest_style_name COLLATE utf8mb4_general_ci LIKE ?", "%"+item.GuestStyleName+"%")
	}
	if isToday != "" {
		// Lấy table Price hiện tại
		db1 := datasources.GetDatabaseWithPartner(item.PartnerUid)
		tablePriceR := TablePrice{
			PartnerUid: item.PartnerUid,
			CourseUid:  item.CourseUid,
		}
		currentTablePrice, errTB := tablePriceR.FindCurrentUse(db1)
		if errTB == nil {
			if currentTablePrice.Id > 0 {
				db = db.Where("table_price_id = ?", currentTablePrice.Id)
			}
		}
		db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
	}

	if dow != "" {
		// Lấy table Price hiện tại
		db1 := datasources.GetDatabaseWithPartner(item.PartnerUid)
		tablePriceR := TablePrice{
			PartnerUid: item.PartnerUid,
			CourseUid:  item.CourseUid,
		}
		currentTablePrice, errTB := tablePriceR.FindCurrentUse(db1)
		if errTB == nil {
			if currentTablePrice.Id > 0 {
				db = db.Where("table_price_id = ?", currentTablePrice.Id)
			}
		}
		db = db.Where("dow LIKE ?", "%"+dow+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *GolfFee) FindAllByTablePrice(database *gorm.DB) ([]GolfFee, error) {
	db := database.Model(GolfFee{})
	list := []GolfFee{}
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
	if item.GroupId > 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}
	if item.TablePriceId > 0 {
		db = db.Where("table_price_id = ?", item.TablePriceId)
	}
	db = db.Find(&list)
	return list, db.Error
}

func (item *GolfFee) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *GolfFee) GetGuestStyleList(database *gorm.DB, time string) []GuestStyle {
	db := database.Table("golf_fees")
	list := []GuestStyle{}
	status := item.ModelId.Status
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.CustomerType != "" {
		db = db.Where("customer_type = ?", item.CustomerType)
	}
	if item.CustomerCategory != "" {
		db = db.Where("customer_category = ?", item.CustomerCategory)
	}

	if item.TablePriceId > 0 {
		db = db.Where("table_price_id = ?", item.TablePriceId)
	}

	// Filter guest style theo ngày trong tuần
	if len(time) == 0 {
		toDayDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		if CheckHoliday(item.PartnerUid, item.CourseUid, toDayDate) {
			db = db.Where("dow LIKE ? OR dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%", "%0%")
		} else {
			db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
		}
	} else {
		dayOfWeek := utils.GetDayOfWeek(time)
		if dayOfWeek != "" {
			if CheckHoliday(item.PartnerUid, item.CourseUid, time) {
				db = db.Where("dow LIKE ? OR dow LIKE ?", "%"+dayOfWeek+"%", "%0%")
			} else {
				db = db.Where("dow LIKE ?", "%"+dayOfWeek+"%")
			}
		} else {
			if CheckHoliday(item.PartnerUid, item.CourseUid, time) {
				db = db.Where("dow LIKE ? OR dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%", "%0%")
			} else {
				db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
			}
		}
	}

	db = db.Group("guest_style")

	db.Find(&list)

	return list
}

func (item *GolfFee) GetGuestStyleGolfFeeByGuestStyle(database *gorm.DB) []GolfFee {
	db := database.Table("golf_fees")
	list := []GolfFee{}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.GuestStyle != "" {
		db = db.Where("guest_style = ?", item.GuestStyle)
	}
	if item.TablePriceId > 0 {
		db = db.Where("table_price_id = ?", item.TablePriceId)
	}

	db.Find(&list)

	return list
}

func (item *GolfFee) FindFirstWithCusType(database *gorm.DB) error {
	db := database.Model(GolfFee{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.CustomerType != "" {
		db = db.Where("customer_type = ?", item.CustomerType)
	}

	// Lấy table Price hiện tại
	db1 := datasources.GetDatabaseWithPartner(item.PartnerUid)
	tablePriceR := TablePrice{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
	}
	currentTablePrice, errTB := tablePriceR.FindCurrentUse(db1)
	if errTB == nil {
		if currentTablePrice.Id > 0 {
			db = db.Where("table_price_id = ?", currentTablePrice.Id)
		}
	}
	db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")

	return db.First(item).Error
}

func (item *GolfFee) FindAll(database *gorm.DB) ([]GolfFee, error) {
	db := database.Model(GolfFee{})
	list := []GolfFee{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	err := db.Find(&list).Error
	return list, err
}

func CheckHoliday(partnerUid, courseUid, date string) bool {
	db := datasources.GetDatabaseWithPartner(partnerUid)
	bookingDate := ""
	if date != "" {
		bookingDate = date
	} else {
		toDayDate, errD := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		if errD != nil {
			log.Print("CHoliday CheckCurrentDay ", errD.Error())
		} else {
			bookingDate = toDayDate
		}
	}

	if bookingDate != "" {
		applyDate, _ := time.Parse(constants.DATE_FORMAT_1, date)
		year := fmt.Sprint(applyDate.Year())

		holiday := Holiday{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Year:       year,
		}

		_, total, _ := holiday.FindListInRange(db, bookingDate)
		return total > 0
	}
	return false
}
