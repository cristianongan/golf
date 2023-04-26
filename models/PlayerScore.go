package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Bảng điểm của người chơi
type PlayerScore struct {
	ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	BookingDate string `json:"booking_date" gorm:"type:varchar(50);index"`
	Bag         string `json:"bag" gorm:"type:varchar(100);index"`
	Course      string `json:"course"`            //  Sân
	Hole        int    `json:"hole" gorm:"index"` // Số hố
	Par         int    `json:"par"`               // Số lần chạm gậy
	Shots       int    `json:"shots"`             // Số gậy đánh
	Index       int    `json:"index"`             // Độ khó
	TimeStart   int64  `json:"time_start"`        // Thời gian bắt đầu
	TimeEnd     int64  `json:"time_end"`          // Thời gian end
}

type ListPlayerScore struct {
	ModelId
	PartnerUid   string `json:"partner_uid"` // Hãng Golf
	CourseUid    string `json:"course_uid"`  // Sân Golf
	BookingDate  string `json:"booking_date"`
	Bag          string `json:"bag"`
	CustomerName string `json:"customer_name"` // Tên khách hàng
	Course       string `json:"course"`        //  Sân
	Hole         int    `json:"hole"`          // Số hố
	Par          int    `json:"par"`           // Số lần chạm gậy
	Shots        int    `json:"shots"`         // Số gậy đánh
	Index        int    `json:"index"`         // Độ khó
	TimeStart    int64  `json:"time_start"`    // Thời gian bắt đầu
	TimeEnd      int64  `json:"time_end"`      // Thời gian end
}

func (item *PlayerScore) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

// / ------- CaddieWorkingCalendar batch insert to db ------
func (item *PlayerScore) BatchInsert(database *gorm.DB, list []PlayerScore) error {
	db := database.Model(PlayerScore{})

	return db.Create(&list).Error
}

func (item *PlayerScore) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *PlayerScore) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *PlayerScore) FindList(database *gorm.DB, page Page) ([]ListPlayerScore, int64, error) {
	db := database.Model(PlayerScore{})
	list := []ListPlayerScore{}
	total := int64(0)

	db = db.Select("player_scores.*, bookings.customer_name")

	if item.PartnerUid != "" {
		db = db.Where("player_scores.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("player_scores.course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("player_scores.booking_date = ?", item.BookingDate)
	}
	if item.Bag != "" {
		db = db.Where("player_scores.bag = ?", item.Bag)
	}

	db.Joins("INNER JOIN bookings on bookings.bag = player_scores.bag and bookings.booking_date = player_scores.booking_date")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *PlayerScore) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
