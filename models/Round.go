package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"
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
}

type ListRound []Round

func (item *ListRound) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListRound) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *Round) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Round) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Round) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Round) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Round{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Round) FindList(page Page) ([]Round, int64, error) {
	db := datasources.GetDatabase().Model(Round{})
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

func (item *Round) FindAll() ([]Round, int64, error) {
	db := datasources.GetDatabase().Model(Round{})
	list := []Round{}
	total := int64(0)
	item.Status = ""

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	db.Count(&total)
	db = db.Find(&list)

	return list, total, db.Error
}

func (item *Round) Delete() error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
