package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"
)

// Sân Golf
type Course struct {
	Model
	PartnerUid                string  `json:"partner_uid" gorm:"type:varchar(100);index"`
	Name                      string  `json:"name" gorm:"type:varchar(256)"`
	Hole                      int     `json:"hole"`
	Address                   string  `json:"address" gorm:"type:varchar(500)"`
	Lat                       float64 `json:"lat"`
	Lng                       float64 `json:"lng"`
	Icon                      string  `json:"icon" gorm:"type:varchar(256)"`
	TypeSendInfoBooking       string  `json:"type_send_info_booking" gorm:"type:varchar(100)"`
	TypeSendInfoBookingAgency string  `json:"type_send_info_booking_agency" gorm:"type:varchar(100)"`
	EmailBooking              string  `json:"email_booking" gorm:"type:varchar(100)"`
	RateGolfFee               string  `json:"rate_golf_fee" gorm:"type:varchar(256)"`
	MaxPeopleInFlight         int     `json:"max_people_in_flight"`                          //số người tối đa trong 1 flight. Mặc định để 4 người.
	MemberBooking             *bool   `json:"member_booking" gorm:"default:0"`               // yêu cầu nguồn booking phải có tối thiểu 1 member.
	ApiKey                    string  `json:"api_key" gorm:"->;<-:create;type:varchar(100)"` // Api key, chỉ cho phép read và create
	AutoSendBooking           bool    `json:"auto_send_booking" gorm:"default:0"`
}

type CourseRes struct {
	Model
	PartnerUid          string  `json:"partner_uid" gorm:"type:varchar(100);index"`
	Name                string  `json:"name" gorm:"type:varchar(256)"`
	Hole                int     `json:"hole"`
	Address             string  `json:"address" gorm:"type:varchar(500)"`
	Lat                 float64 `json:"lat"`
	Lng                 float64 `json:"lng"`
	Icon                string  `json:"icon" gorm:"type:varchar(256)"`
	TypeSendInfoBooking string  `json:"type_send_info_booking" gorm:"type:varchar(100)"`
	EmailBooking        string  `json:"email_booking" gorm:"type:varchar(100)"`
	RateGolfFee         string  `json:"rate_golf_fee" gorm:"type:varchar(256)"`
	MaxPeopleInFlight   int     `json:"max_people_in_flight"`            //số người tối đa trong 1 flight. Mặc định để 4 người.
	MemberBooking       *bool   `json:"member_booking" gorm:"default:0"` // yêu cầu nguồn booking phải có tối thiểu 1 member.
}

// ======= CRUD ===========
func (item *Course) Create() error {
	db := datasources.GetDatabaseAuth()
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	item.ApiKey = utils.RandomCharNumberV2(50)

	return db.Create(item).Error
}

func (item *Course) Update() error {
	db := datasources.GetDatabaseAuth()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Course) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	err := db.Where(item).First(item).Error
	item.ApiKey = ""
	return err
}

func (item *Course) FindFirstHaveKey() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *Course) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(Course{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Course) FindList(page Page) ([]CourseRes, int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Table("courses")
	list := []CourseRes{}
	total := int64(0)
	status := item.Status

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%")
	}
	if item.PartnerUid != "" && item.PartnerUid != constants.ROOT_PARTNER_UID {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Course) FindALL() ([]CourseRes, int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Table("courses")
	list := []CourseRes{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	db.Count(&total)

	db = db.Find(&list)

	return list, total, db.Error
}

func (item *Course) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
