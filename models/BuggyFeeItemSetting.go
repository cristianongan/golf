package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BuggyFeeItemSetting struct {
	ModelId
	PartnerUid     string                `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string                `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	ParentId       int64                 `json:"parent_id"`                                  // id của setting cha
	GuestStyleName string                `json:"guest_style_name" gorm:"type:varchar(256)"`  // Ten Guest style
	GuestStyle     string                `json:"guest_style" gorm:"index;type:varchar(200)"` // Guest style
	Dow            string                `json:"dow" gorm:"type:varchar(100)"`               // Dow
	RentalFee      utils.ListGolfHoleFee `json:"rental_fee" gorm:"type:varchar(256)"`        // Phi Rental
	PrivateCarFee  utils.ListGolfHoleFee `json:"private_car_fee" gorm:"type:varchar(256)"`   // Phi Xe rieng
	OddCarFee      utils.ListGolfHoleFee `json:"odd_car_fee" gorm:"type:varchar(256)"`       // Phi buggy
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
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BuggyFeeItemSetting) Update(db *gorm.DB) error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
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
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BuggyFeeItemSetting) Delete(database *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
