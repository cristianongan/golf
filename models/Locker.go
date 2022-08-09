package models

import (
	"start/constants"
	"start/datasources"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Locker
type Locker struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GolfBag    string `json:"golf_bag" gorm:"type:varchar(200)"`
	BookingUid string `json:"booking_uid" gorm:"type:varchar(50);index"`
	Locker     string `json:"locker" gorm:"type:varchar(500)"`
	PlayerName string `json:"player_name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *Locker) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Locker) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Locker) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Locker) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Locker{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Locker) FindList(page Page, from, to int64, isFullDay bool) ([]Locker, int64, error) {
	db := datasources.GetDatabase().Model(Locker{})
	list := []Locker{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.GolfBag != "" {
		db = db.Where("golf_bag = ?", item.GolfBag)
	}
	if item.Locker != "" {
		db = db.Where("locker = ?", item.Locker)
	}
	//Search With Time
	if from > 0 && to > 0 {
		db = db.Where("created_at between " + strconv.FormatInt(from, 10) + " and " + strconv.FormatInt(to, 10) + " ")
	}

	if from > 0 && to == 0 {
		db = db.Where("created_at > " + strconv.FormatInt(from, 10) + " ")
	}

	if from == 0 && to > 0 {
		db = db.Where("created_at < " + strconv.FormatInt(to, 10) + " ")
	}

	db.Count(&total)

	if isFullDay {
		db.Order("created_at desc")
		db.Find(&list)
		return list, total, db.Error
	}

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Locker) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
