package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Bảng điểm của người chơi
type RestaurantSetting struct {
	ModelId
	PartnerUid    string           `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid     string           `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	ServiceId     int64            `json:"service_id" gorm:"index"`                    // Id nhà hàng
	Name          string           `json:"name" gorm:"type:varchar(256)"`              // Tên setting
	NumberTables  int              `json:"number_tables"`                              // Số bàn
	PeopleInTable int              `json:"people_in_table"`                            //  Tổng số người trong 1 bàn
	Type          string           `json:"type" gorm:"type:varchar(100)"`              // Loại setting
	Time          int              `json:"time"`                                       // Số phút setting
	Symbol        string           `json:"symbol" gorm:"type:varchar(100)"`            // Ký hiệu
	TableFrom     int              `json:"table_from"`                                 //
	DataTables    utils.ListString `json:"data_tables,omitempty" gorm:"type:json"`     //
}

func (item *RestaurantSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

// / ------- CaddieWorkingCalendar batch insert to db ------
func (item *RestaurantSetting) BatchInsert(database *gorm.DB, list []RestaurantSetting) error {
	db := database.Model(RestaurantSetting{})

	return db.Create(&list).Error
}

func (item *RestaurantSetting) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *RestaurantSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *RestaurantSetting) FindList(database *gorm.DB, page Page) ([]RestaurantSetting, int64, error) {
	db := database.Model(RestaurantSetting{})
	list := []RestaurantSetting{}
	total := int64(0)

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

func (item *RestaurantSetting) Delete(db *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
