package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"
)

type ForceUpdate struct {
	ModelId
	Version     string `json:"version" gorm:"type:varchar(50)"`      // X.X.X
	OsType      string `json:"os_type" gorm:"type:varchar(50)"`      // IOS, ANDROID
	DeviceType  string `json:"device_type" gorm:"type:varchar(50)"`  // PHONE, TABLET
	Description string `json:"description" gorm:"type:varchar(200)"` // description
	IsForce     int    `json:"is_force"`                             // 0, 1 la force update
}

// ======= CRUD ===========
func (item *ForceUpdate) Create() error {
	db := datasources.GetDatabaseAuth()
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()

	return db.Create(item).Error
}

func (item *ForceUpdate) Update() error {
	db := datasources.GetDatabaseAuth()
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *ForceUpdate) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *ForceUpdate) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(ForceUpdate{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *ForceUpdate) FindList() ([]ForceUpdate, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(ForceUpdate{})
	list := []ForceUpdate{}

	if item.OsType != "" {
		db = db.Where("os_type = ?", item.OsType)
	}

	if item.DeviceType != "" {
		db = db.Where("device_type = ?", item.DeviceType)
	}

	db.Find(&list)

	return list, db.Error
}

// func (item *ForceUpdate) Delete() error {
// 	db := datasources.GetDatabaseAuth()
// 	if item.Id <= 0 {
// 		return errors.New("Primary key is undefined!")
// 	}
// 	return db.Delete(item).Error
// }
