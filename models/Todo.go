package models

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log"
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

func (item *Todo) CreateBatch(todos []Todo) error {
	now := time.Now()
	for i, _ := range todos {
		t := &todos[i]
		uid := uuid.New()
		t.Model.Uid = uid.String()
		t.Model.CreatedAt = now.Unix()
		t.Model.UpdatedAt = now.Unix()
		t.Model.Status = constants.STATUS_ENABLE
	}
	log.Println(todos)
	db := datasources.GetDatabase()
	return db.CreateInBatches(todos, 100).Error
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

func (item *Todo) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
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
