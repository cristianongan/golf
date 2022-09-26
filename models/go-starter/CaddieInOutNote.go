package model_gostarter

import (
	"start/constants"
	"start/models"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieInOutNote struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BookingUid string `json:"booking_uid" gorm:"type:varchar(50);index"`  // Ex: Booking Uid
	CaddieId   int64  `json:"caddie_id" gorm:"index"`                     // Caddie Id
	CaddieCode string `json:"caddie_code" gorm:"type:varchar(256)"`
	Note       string `json:"note" gorm:"type:varchar(500)"`      // note
	Type       string `json:"type" gorm:"type:varchar(50);index"` // Type: IN(undo), OUT, CHANGE
	Hole       int    `json:"hole" gorm:"type:bigint"`
}

func (item *CaddieInOutNote) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CaddieInOutNote) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieInOutNote) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieInOutNote) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CaddieInOutNote{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieInOutNote) FindAllCaddieInOutNotes(database *gorm.DB) ([]CaddieInOutNote, error) {
	now := time.Now().Format("02/01/2006")

	from, _ := time.Parse("02/01/2006 15:04:05", now+" 17:00:00")

	to, _ := time.Parse("02/01/2006 15:04:05", now+" 16:59:59")

	db := database.Model(CaddieInOutNote{})
	list := []CaddieInOutNote{}

	db = db.Where("type = ?", constants.STATUS_OUT)
	db = db.Where("created_at >= ?", from.AddDate(0, 0, -1).Unix())
	db = db.Where("created_at < ?", to.Unix())

	db.Find(&list)
	return list, db.Error
}

func (item *CaddieInOutNote) FindList(database *gorm.DB, page models.Page, from, to int64) ([]CaddieInOutNote, int64, error) {
	db := database.Model(CaddieInOutNote{})
	list := []CaddieInOutNote{}
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

func (item *CaddieInOutNote) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
