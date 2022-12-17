package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"start/constants"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Hãng Golf
type Round struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CaddieFee     int64  `json:"caddie_fee"`
	BuggyFee      int64  `json:"buggy_fee"`
	GreenFee      int64  `json:"green_fee"`
	GuestStyle    string `json:"guest_style" gorm:"type:varchar(200);"` // Nếu là member Card thì lấy guest style của member Card, nếu không thì lấy guest style Của booking đó
	MemberCardId  string `json:"member_card_id"`
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100)"`
	Pax           int    `json:"pax"`
	TeeOffTime    int64  `json:"tee_off_time"`
	Hole          int    `json:"hole"`
	Index         int    `json:"index"`
	Bag           string `json:"bag" gorm:"type:varchar(100);index"` // Golf Bag
	BillCode      string `json:"bill_code" gorm:"type:varchar(100);index"`
	MainBagPaid   *bool  `json:"main_bag_paid" gorm:"default:0"`
	PaidBy        string `json:"paid_by" gorm:"type:varchar(50)"` // Paid by: cho cây đại lý thanh toán
	IsPaid        bool   `json:"is_paid" gorm:"-:migration"`      // Đánh dấu đã được trả bởi main bag or agency (Không migrate db)
}

type FeeOfRound struct {
	CaddieFee int64 `json:"caddie_fee"`
	BuggyFee  int64 `json:"buggy_fee"`
	GreenFee  int64 `json:"green_fee"`
}

type ListRound []Round

func (item *ListRound) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListRound) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *Round) Create(db *gorm.DB) error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Round) Update(db *gorm.DB) error {
	item.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Round) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Round) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Round{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Round) CountWithBillCode(database *gorm.DB) (int64, error) {
	db := database.Model(Round{})
	total := int64(0)
	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}
	db = db.Count(&total)
	return total, db.Error
}

func (item *Round) FindList(database *gorm.DB, page Page) ([]Round, int64, error) {
	db := database.Model(Round{})
	list := []Round{}
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

func (item *Round) FindAll(database *gorm.DB) ([]Round, error) {
	db := database.Model(Round{})
	list := []Round{}

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}
	db = db.Find(&list)

	return list, db.Error
}

func (item *Round) LastRound(database *gorm.DB) error {
	db := database.Order("created_at desc")
	return db.Where(item).First(item).Error
}

func (item *Round) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *Round) GetAmountGolfFee() int64 {
	return item.CaddieFee + item.GreenFee + item.BuggyFee
}
