package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Bag Note
type BagsNote struct {
	ModelId
	GolfBag    string `json:"golf_bag" gorm:"type:varchar(200)"`
	BookingUid string `json:"booking_uid" gorm:"type:varchar(50);index"`
	Note       string `json:"note" gorm:"type:varchar(2000)"`
	PlayerName string `json:"player_name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *BagsNote) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BagsNote) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BagsNote) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BagsNote) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BagsNote{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BagsNote) FindList(page Page) ([]BagsNote, int64, error) {
	db := datasources.GetDatabase().Model(BagsNote{})
	list := []BagsNote{}
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

func (item *BagsNote) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
