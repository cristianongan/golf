package model_booking

import (
	"log"
	"start/constants"
	"start/models"
	"start/utils"
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Booking setting
type BookingSettingGroup struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Group Name
	FromDate   int64  `json:"from_date" gorm:"index"`                     // Áp dụng từ ngày
	ToDate     int64  `json:"to_date" gorm:"index"`                       // Áp dụng tới ngày
}

func (item *BookingSettingGroup) IsDuplicated(db *gorm.DB) bool {
	bookingSettingGroup := BookingSettingGroup{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Name:       item.Name,
		FromDate:   item.FromDate,
		ToDate:     item.ToDate,
	}

	errFind := bookingSettingGroup.FindFirst(db)
	if errFind == nil || bookingSettingGroup.Id > 0 {
		return true
	}
	return false
}

func (item *BookingSettingGroup) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	return true
}

func (item *BookingSettingGroup) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingSettingGroup) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingSettingGroup) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingSettingGroup) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingSettingGroup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingSettingGroup) FindList(database *gorm.DB, page models.Page, from, to int64) ([]BookingSettingGroup, int64, error) {
	db := database.Model(BookingSettingGroup{})
	list := []BookingSettingGroup{}
	total := int64(0)
	status := constants.STATUS_ENABLE
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		log.Println("BookingSettingGroup FindList status", status)
		db = db.Where("`status` = ?", status)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	//Search With Time
	if from > 0 && to > 0 {
		db = db.Where("from_date < " + strconv.FormatInt(from+30, 10) + " ")
		db = db.Where("to_date > " + strconv.FormatInt(to-30, 10) + " ")
	}
	if from > 0 && to == 0 {
		db = db.Where("from_date < " + strconv.FormatInt(from+30, 10) + " ")
	}
	if from == 0 && to > 0 {
		db = db.Where("to_date > " + strconv.FormatInt(to-30, 10) + " ")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingSettingGroup) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *BookingSettingGroup) ValidateClose1ST(db *gorm.DB, BookingDate string) error {
	bookingSetting := BookingSettingGroup{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
	}
	from := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, BookingDate)
	to := from + 24*60*60
	page := models.Page{
		Limit:   20,
		Page:    1,
		SortBy:  "created_at",
		SortDir: "desc",
	}
	println(item.PartnerUid)
	listBSG, _, errLBSG := bookingSetting.FindList(db, page, from, to)
	if errLBSG != nil || len(listBSG) == 0 {
		return nil
	}
	bookingSettingGroup := listBSG[0]
	if bookingSettingGroup.Status == constants.STATUS_ENABLE {
		teeTypeClose := models.TeeTypeClose{
			PartnerUid:       bookingSettingGroup.PartnerUid,
			CourseUid:        bookingSettingGroup.CourseUid,
			BookingSettingId: bookingSettingGroup.Id,
			DateTime:         BookingDate,
		}
		if err := teeTypeClose.FindFirst(db); err == nil {
			return errors.New("Tee 1 is closed")
		}
	}
	return nil
}
