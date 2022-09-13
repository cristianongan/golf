package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type BookingWaiting struct {
	models.ModelId
	PartnerUid    string           `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid     string           `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode   string           `json:"booking_code" gorm:"type:varchar(100);index"` //
	BookingTime   string           `json:"booking_time" gorm:"type:varchar(100)"`       // Ngày tạo booking waiting
	PlayerName    string           `json:"player_name" gorm:"type:varchar(256)"`        // Tên người đặt booking waiting
	PlayerContact string           `json:"player_contact" gorm:"type:varchar(256)"`     // SĐT người đặt booking waiting
	PeopleList    utils.ListString `json:"people_list,omitempty" gorm:"type:json"`      // Danh sách người chơi
	Note          string           `json:"note" gorm:"type:varchar(256)"`               // Ghi chú
}

func (item *BookingWaiting) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BookingWaiting) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingWaiting) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BookingWaiting) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BookingWaiting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingWaiting) FindList(page models.Page) ([]BookingWaiting, int64, error) {
	db := datasources.GetDatabase().Model(BookingWaiting{})
	list := []BookingWaiting{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.BookingTime != "" {
		db = db.Where("booking_time = ?", item.BookingTime)
	}

	if item.PlayerName != "" {
		db = db.Where("player_name LIKE ?", "%"+item.PlayerName+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingWaiting) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
