package model_gostarter

import (
	"start/constants"
	"start/models"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BagAttachCaddie struct {
	models.ModelId
	PartnerUid   string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid    string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	BookingUid   string `json:"booking_uid" gorm:"type:varchar(50);index"`
	BookingDate  string `json:"booking_date" gorm:"type:varchar(100);index"` // ngày booking
	Bag          string `json:"bag" gorm:"type:varchar(100);index"`          // Golf Bag
	BagStatus    string `json:"bag_status" gorm:"type:varchar(50);index"`    //Bag status
	CustomerName string `json:"customer_name" gorm:"type:varchar(256)"`      // Player name
	CaddieCode   string `json:"caddie_code" gorm:"type:varchar(50);index"`   // Mã caddie
}

// ======= CRUD ===========
func (item *BagAttachCaddie) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BagAttachCaddie) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BagAttachCaddie) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BagAttachCaddie) FindList(database *gorm.DB, page models.Page) ([]BagAttachCaddie, int64, error) {
	db := database.Model(BagAttachCaddie{})
	list := []BagAttachCaddie{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Bag != "" {
		db = db.Where("bag COLLATE utf8mb4_general_ci LIKE ? OR customer_name COLLATE utf8mb4_general_ci LIKE ? OR caddie_code COLLATE utf8mb4_general_ci LIKE ?",
			"%"+item.Bag+"%", "%"+item.Bag+"%", "%"+item.Bag+"%")
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BagAttachCaddie) Delete(db *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
