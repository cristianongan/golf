package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"
)

type JsonDataLog struct {
	Data interface{} `json:"data"`
}

type OperationLog struct {
	ModelId
	PartnerUid  string      `json:"partner_uid" gorm:"type:varchar(100);index"`
	CourseUid   string      `json:"course_uid" gorm:"type:varchar(100);index"`
	UserUid     string      `json:"user_uid" gorm:"type:varchar(100);index"`
	UserName    string      `json:"user_name" gorm:"type:varchar(200);index"`   // cms user name
	BookingDate string      `json:"booking_date" gorm:"type:varchar(30);index"` // 06/11/2022
	Bag         string      `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	BookingUid  string      `json:"booking_uid" gorm:"type:varchar(50);index"`  // Booking uid
	BillCode    string      `json:"bill_code" gorm:"type:varchar(50);index"`    // Bill Code
	Action      string      `json:"action" gorm:"type:varchar(200);index"`      // hoạt động
	Function    string      `json:"function" gorm:"type:varchar(100);index"`    // Booking, Checkin
	Module      string      `json:"module" gorm:"type:varchar(100);index"`      // GO, RECEPTION
	Method      string      `json:"method" gorm:"type:varchar(30);index"`       // create, update, delete
	Path        string      `json:"path" gorm:"type:varchar(100)"`              // Path Api
	Body        JsonDataLog `json:"body" gorm:"type:json"`                      // Body Api
	ValueOld    JsonDataLog `json:"value_old" gorm:"type:json"`                 // Value Old Object
	ValueNew    JsonDataLog `json:"value_new" gorm:"type:json"`                 // Value New Object
}

func (item *OperationLog) Create() error {
	now := utils.GetTimeNow()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *OperationLog) Update() error {
	mydb := datasources.GetDatabaseAuth()
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *OperationLog) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *OperationLog) Count() (int64, error) {
	db := datasources.GetDatabaseAuth().Model(OperationLog{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *OperationLog) FindList(page Page, search string) ([]OperationLog, int64, error) {
	db := datasources.GetDatabaseAuth().Model(OperationLog{})
	list := []OperationLog{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if search != "" {
		db = db.Where("(user_name LIKE ? OR full_name LIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *OperationLog) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabaseAuth().Delete(item).Error
}
