package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieWorkingTime struct {
	ModelId
	CaddieId     string `json:"caddie_id" gorm:"type:varchar(20)"`
	CheckInTime  int64  `json:"check_in_time"`  // Time Check In
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
}

type CaddieWorkingTimeRequest struct {
	Id           int64  `json:"id"`
	CaddieId     string `json:"caddie_id" gorm:"type:varchar(20)"`
	CaddieName   string `json:"caddie_name"`
	CheckInTime  int64  `json:"check_in_time"`  // Time Check In
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
	WorkingTime  int64  `json:"working_time"`
	OverTime     int64  `json:"over_time"`
}

type CaddieWorkingTimeResponse struct {
	CheckInTime  int64 `json:"check_in_time"`  // Time Check In
	CheckOutTime int64 `json:"check_out_time"` // Time Check Out
	WorkingTime  int64 `json:"working_time"`
	OverTime     int64 `json:"over_time"`
}
type WorkingTimeTotal struct {
	CaddieId              string                      `json:"caddie_id"`
	Total                 int                         `json:"total"`
	OverTime              int64                       `json:"over_time"`
	CaddieName            string                      `json:"caddie_name"`
	CaddieWorkingTimeList []CaddieWorkingTimeResponse `json:"data"`
}

func (item *CaddieWorkingTime) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingTime) CreateBatch(db *gorm.DB, caddies []CaddieWorkingTime) error {
	now := utils.GetTimeNow()
	for i := range caddies {
		c := &caddies[i]
		c.ModelId.CreatedAt = now.Unix()
		c.ModelId.UpdatedAt = now.Unix()
		c.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.CreateInBatches(caddies, 100).Error
}

func (item *CaddieWorkingTime) Delete(db *gorm.DB) error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *CaddieWorkingTime) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieWorkingTime) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingTime) Count(database *gorm.DB) (int64, error) {
	total := int64(0)

	db := database.Model(CaddieWorkingTime{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieWorkingTime) FindCaddieWorkingTimeDetail(database *gorm.DB) *CaddieWorkingTime {
	time := CaddieWorkingTime{}
	db := database.Model(CaddieWorkingTime{})
	db.Where(item).Find(&time)
	return &time
}

func (item *CaddieWorkingTimeRequest) FindList(database *gorm.DB, page Page, from, to int64) ([]WorkingTimeTotal, int64, error) {
	var list []CaddieWorkingTimeRequest
	var results []WorkingTimeTotal

	total := int64(0)

	db := database.Model(CaddieWorkingTime{})

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

	db2 := db
	db = db.Joins("JOIN caddies ON caddie_working_times.caddie_id = caddies.uid")

	db2 = db2.Select("caddie_working_times.id, caddie_working_times.caddie_id, caddie_working_times.check_in_time, " +
		"caddie_working_times.check_out_time, (caddie_working_times.check_out_time - caddie_working_times.check_in_time) as working_time")
	db2 = page.Setup(db2).Find(&list)

	db = db.Group("caddie_working_times.caddie_id")
	db = db.Select("SUM(caddie_working_times.check_out_time - caddie_working_times.check_in_time) as total, caddie_working_times.caddie_id, caddies.name as caddie_name")
	db = page.Setup(db).Find(&results)

	for t := range results {
		d := &results[t]
		d.CaddieWorkingTimeList = []CaddieWorkingTimeResponse{}
		for i := range list {
			c := &list[i]
			if d.CaddieId == c.CaddieId {
				data := CaddieWorkingTimeResponse{
					CheckInTime:  c.CheckInTime,
					CheckOutTime: c.CheckOutTime,
					OverTime:     c.OverTime,
					WorkingTime:  c.WorkingTime,
				}
				d.CaddieWorkingTimeList = append(d.CaddieWorkingTimeList, data)
			}
		}
	}

	total = int64(len(results))

	return results, total, db.Error
}
