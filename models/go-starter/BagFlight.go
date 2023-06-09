package model_gostarter

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
)

// BagFlight
/*
1 Bag có thể chơi nhiều Flight
Lưu booking_uid, flight_id, hole, sẽ tính lại dc giá của bag đó khi chơi nhiều Flight

Ở Bảng bookings sẽ lưu Flight hiện tại, Buggy và Caddie hiện tại và đa số sẽ truy vấn hầu hết trên này
Bảng bag_flights này sẽ lưu các bag và các flight

*/
type BagFlight struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Bag         string `json:"bag" gorm:"type:varchar(50);index"`          // Bag
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	FlightId    int64  `json:"flight_id" gorm:"index"`                     // Flight Id
	BookingUid  string `json:"booking_uid" gorm:"type:varchar(100);index"` // Booking Uid
	CaddieId    int64  `json:"caddie_id"`                                  // Caddie Id
	BuddyId     int64  `json:"buddy_id"`                                   // Buggy Id
	Hole        int    `json:"hole"`                                       // Số hole
}

func (item *BagFlight) Create() error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BagFlight) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BagFlight) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BagFlight) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BagFlight{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BagFlight) FindList(page models.Page, from, to int64) ([]BagFlight, int64, error) {
	db := datasources.GetDatabase().Model(BagFlight{})
	list := []BagFlight{}
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

func (item *BagFlight) FindListAll() ([]BagFlight, error) {
	db := datasources.GetDatabase().Model(BagFlight{})
	list := []BagFlight{}
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
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Find(&list)
	err := db.Error
	if err != nil {
		log.Println("BagFlight FindListAll err ", err.Error())
	}
	return list, err
}

func (item *BagFlight) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
