package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BuggyFeeItemSetting struct {
	ModelId
	PartnerUid     string                `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string                `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	SettingId      int64                 `json:"setting_id"`
	GuestStyleName string                `json:"guest_style_name" gorm:"type:varchar(256)"`  // Ten Guest style
	GuestStyle     string                `json:"guest_style" gorm:"index;type:varchar(200)"` // Guest style
	Dow            string                `json:"dow" gorm:"type:varchar(100)"`               // Dow
	RentalFee      utils.ListGolfHoleFee `json:"rental_fee" gorm:"type:json"`                // Phi Rental
	PrivateCarFee  utils.ListGolfHoleFee `json:"private_car_fee" gorm:"type:json"`           // Phi Xe rieng
	OddCarFee      utils.ListGolfHoleFee `json:"odd_car_fee" gorm:"type:json"`               // Phi buggy
	RateGolfFee    string                `json:"rate_golf_fee" gorm:"type:json"`
}

type BuggyFeeItemSettingResponse struct {
	RentalFee     int64  `json:"rental_fee"`
	PrivateCarFee int64  `json:"private_car_fee"`
	OddCarFee     int64  `json:"odd_car_fee"`
	GuestStyle    string `json:"guest_style"`
}

type BuggyFeeItemSettingResForRental struct {
	RentalFee     utils.ListGolfHoleFee `json:"rental_fee"`
	PrivateCarFee utils.ListGolfHoleFee `json:"private_car_fee"`
	OddCarFee     utils.ListGolfHoleFee `json:"odd_car_fee"`
	GuestStyle    string                `json:"guest_style"`
}

type ListBuggyFeeItemSetting []BuggyFeeItemSetting

func (item *ListBuggyFeeItemSetting) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBuggyFeeItemSetting) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *BuggyFeeItemSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BuggyFeeItemSetting) ValidateCreate(db *gorm.DB) error {
	data1 := BuggyFeeItemSetting{
		GuestStyle: item.GuestStyle,
		Dow:        item.Dow,
	}
	_, total, _ := data1.FindAll(db)

	if total > 0 {
		return errors.New("All Guest for Dow existed!")
	}

	return nil
}

func (item *BuggyFeeItemSetting) Update(db *gorm.DB) error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BuggyFeeItemSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BuggyFeeItemSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BuggyFeeItemSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BuggyFeeItemSetting) FindList(database *gorm.DB, page Page) ([]BuggyFeeItemSetting, int64, error) {
	db := database.Model(BuggyFeeItemSetting{})
	list := []BuggyFeeItemSetting{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.GuestStyle != "" {
		db = db.Where("guest_style = ?", item.GuestStyle)
	}
	if item.SettingId > 0 {
		db = db.Where("setting_id = ?", item.SettingId)
	}

	db.Count(&total)
	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BuggyFeeItemSetting) FindAll(database *gorm.DB) ([]BuggyFeeItemSetting, int64, error) {
	db := database.Model(BuggyFeeItemSetting{})
	list := []BuggyFeeItemSetting{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.SettingId > 0 {
		db = db.Where("setting_id = ?", item.SettingId)
	}

	db = db.Where("guest_style = ? OR guest_style = ?", item.GuestStyle, "")
	db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
	db.Count(&total)
	db = db.Find(&list)
	return list, total, db.Error
}

func (item *BuggyFeeItemSetting) FindBuggyFeeOnDate(database *gorm.DB, time string) ([]BuggyFeeItemSetting, error) {
	db := database.Model(BuggyFeeItemSetting{})
	list := []BuggyFeeItemSetting{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.SettingId > 0 {
		db = db.Where("setting_id = ?", item.SettingId)
	}

	if item.GuestStyle != "" {
		db = db.Where("guest_style = ? OR guest_style = ?", item.GuestStyle, "")
	}

	if time == "" {
		db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
	} else {
		dayOfWeek := utils.GetDayOfWeek(time)
		if dayOfWeek != "" {
			db = db.Where("dow LIKE ?", "%"+dayOfWeek+"%")
		} else {
			db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
		}
	}

	if CheckHoliday(item.PartnerUid, item.CourseUid, time) {
		db = db.Or("dow LIKE ?", "%0%")
	}

	db = db.Where("status = ?", constants.STATUS_ENABLE)
	db = db.Order("created_at desc")

	db.Limit(1).Debug().Find(&list)

	return list, db.Error
}

func (item *BuggyFeeItemSetting) FindAllToday(database *gorm.DB) ([]BuggyFeeItemSettingResForRental, int64, error) {
	db := database.Model(BuggyFeeItemSetting{})
	list := []BuggyFeeItemSettingResForRental{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.SettingId > 0 {
		db = db.Where("setting_id = ?", item.SettingId)
	}

	db = db.Where("dow LIKE ?", "%"+utils.GetCurrentDayStrWithMap()+"%")
	db.Count(&total)
	db = db.Find(&list)
	return list, total, db.Error
}

func (item *BuggyFeeItemSetting) Delete(database *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
