package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

type CaddieWorkingTime struct {
	ModelId
	CaddieId     string `json:"caddie_id" gorm:"type:varchar(20)"`
	CheckInTime  int64  `json:"check_in_time"`  // Time Check In
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
}

type CaddieWorkingTimeResponse struct {
	Id           int64  `json:"id"`
	CaddieId     string `json:"caddie_id" gorm:"type:varchar(20)"`
	CheckInTime  int64  `json:"check_in_time"`  // Time Check In
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
	WorkingTime  int64  `json:"working_time"`
	OverTime     int64  `json:"over_time"`
	CaddieName   string `json:"caddie_name"`
}

func (item *CaddieWorkingTime) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieWorkingTime) CreateBatch(caddies []CaddieWorkingTime) error {
	now := time.Now()
	for i := range caddies {
		c := &caddies[i]
		c.ModelId.CreatedAt = now.Unix()
		c.ModelId.UpdatedAt = now.Unix()
		c.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.CreateInBatches(caddies, 100).Error
}

func (item *CaddieWorkingTime) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CaddieWorkingTime) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieWorkingTime) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingTime) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieWorkingTime{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieWorkingTimeResponse) FindList(page Page, from, to int64) ([]CaddieWorkingTimeResponse, int64, error) {
	var list []CaddieWorkingTimeResponse

	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieWorkingTime{})

	if item.CaddieId != "" {
		db = db.Where("caddie_working_times.caddie_id = ?", item.CaddieId)
	}

	if item.CaddieName != "" {
		db = db.Where("caddies.name LIKE ?", "%"+item.CaddieName+"%")
	}

	if from > 0 {
		db = db.Where("caddie_working_times.created_at >= ?", from)
	}

	if to > 0 {
		db = db.Where("caddie_working_times.created_at < ?", to)
	}

	db = db.Joins("JOIN caddies ON caddie_working_times.caddie_id = caddies.caddie_id")
	db = db.Select("caddie_working_times.id, caddie_working_times.caddie_id, caddie_working_times.check_in_time, " +
		"caddie_working_times.check_out_time, caddies.name as caddie_name")
	// db = db.Group("caddie_working_times.caddie_id")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	for i := range list {
		c := &list[i]
		c.WorkingTime = c.CheckOutTime - c.CheckInTime - 0
		c.OverTime = int64(0)
	}

	return list, total, db.Error
}
