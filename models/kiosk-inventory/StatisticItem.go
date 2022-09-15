package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

/*
 Để lưu thông tin xuất kho
*/
type StatisticItem struct {
	models.ModelId
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	ServiceId       int64  `json:"service_id" gorm:"index"`                    // mã service
	ServiceName     string `json:"service_name" gorm:"type:varchar(256)"`      // tên service
	ItemCode        string `json:"item_code" gorm:"type:varchar(100);index"`   // Mã Item
	EndingInventory int64  `json:"ending_inventory"`                           // Số lượng item cuối ngày
	Import          int64  `json:"import"`                                     // Số lượng đã Import cuối ngày
	Export          int64  `json:"export"`                                     // Số lượng đã Export cuối ngày
	Total           int64  `json:"total"`                                      // Tổng số lượng cuối ngày
	Time            string `json:"time" gorm:"type:varchar(256)"`              // Ngày thống kê
}

func (item *StatisticItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *StatisticItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *StatisticItem) FindList(page models.Page) ([]StatisticItem, int64, error) {
	db := datasources.GetDatabase().Model(StatisticItem{})
	list := []StatisticItem{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	if item.ServiceId > 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
