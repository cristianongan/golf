package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"

	"github.com/pkg/errors"
)

type BuggyDiary struct {
	ModelId
	CourseId      string `json:"course_id" gorm:"type:varchar(100);index"`
	BuggyNumber   int    `json:"buggy_number" gorm:"type:int"`
	AccessoriesId int    `json:"accessories_id" gorm:"type:int"`
	Amount        int    `json:"amount" gorm:"type:int"`
	Note          string `json:"note" gorm:"type:varchar(200)"`
	InputUser     string `json:"input_user" gorm:"type:varchar(20)"`
}

type BuggyDiaryResponse struct {
	ModelId
	CourseId      string `json:"course_id"`
	BuggyNumber   int    `json:"buggy_number"`
	AccessoriesId int    `json:"accessories_id"`
	Amount        int    `json:"amount"`
	Note          string `json:"note"`
	InputUser     string `json:"input_user"`
}

func (item *BuggyDiary) Create() error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BuggyDiary) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *BuggyDiary) Update() error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BuggyDiary) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BuggyDiary) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(BuggyDiary{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BuggyDiary) FindList(page Page, from int64, to int64) ([]BuggyDiary, int64, error) {
	var list []BuggyDiary
	total := int64(0)

	db := datasources.GetDatabase().Model(BuggyDiary{})
	db = db.Where(item)

	if from > 0 {
		db = db.Where("created_at >= ?", from)
	}

	if to > 0 {
		db = db.Where("created_at < ?", to)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
