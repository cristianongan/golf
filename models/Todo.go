package models

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"start/constants"
	"start/datasources"
	"time"
)

type Todo struct {
	Model
	Content string `json:"content" gorm:"type:varchar(200)"`
	Done    bool   `json:"done" gorm:"type:boolean"`
}

type TodoResponse struct {
	Model
	Content string `json:"content" gorm:"type:varchar(200)"`
	Done    bool   `json:"done" gorm:"type:boolean"`
}

func (item *Todo) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	item.Model.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Todo) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Todo) Update() error {
	item.Model.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Todo) FindFirst() ([]Todo, error) {
	var single []Todo

	db := datasources.GetDatabase()
	db.Where(item).First(&single)
	return single, db.Error
}

func (item *Todo) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Todo{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Todo) FindList(page Page, done *bool) ([]Todo, int64, error) {
	var list []Todo
	total := int64(0)

	db := datasources.GetDatabase().Model(Todo{})
	db = db.Where(item)
	if done != nil {
		db = db.Where("done = ?", *done)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
