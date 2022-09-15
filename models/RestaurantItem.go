package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

/*
Giỏ Hàng
*/
type RestaurantItem struct {
	ModelId
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(150);index"`  // Sân golf
	ServiceId       int64  `json:"service_id" gorm:"index"`                    // Mã của service
	OrderDate       string `json:"order_date" gorm:"type:varchar(30);index"`   // Ex: 06/11/2022
	Type            string `json:"type" gorm:"type:varchar(100)"`              // Loại sản phẩm: FOOD, DRINK
	BillId          int64  `json:"bill_id" gorm:"index"`                       // id hóa đơn
	ItemId          int64  `json:"item_id" gorm:"index"`                       // id sản phẩm
	ItemCode        string `json:"iten_code" gorm:"type:varchar(100)"`         // Tên sản phẩm
	ItemName        string `json:"iten_name" gorm:"type:varchar(100)"`         // Tên sản phẩm
	ItemStatus      string `json:"iten_staus" gorm:"type:varchar(100)"`        // Trạng thái sản phẩm
	ItemNote        string `json:"iten_note" gorm:"type:varchar(200)"`         // Yêu cầu của khách hàng
	Quatity         int    `json:"quatity"`                                    // Số lượng order
	QuatityProgress int    `json:"quatity_progress"`                           // Số lương đang tiến hành
}

func (item *RestaurantItem) Create() error {
	now := time.Now()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *RestaurantItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *RestaurantItem) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

func (item *RestaurantItem) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *RestaurantItem) FindList(page Page) ([]RestaurantItem, int64, error) {
	var list []RestaurantItem
	total := int64(0)

	db := datasources.GetDatabase().Model(RestaurantItem{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ServiceId != 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	db = db.Where("order_date = ?", item.OrderDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
