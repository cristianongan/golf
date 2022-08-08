package logger

import (
	"gorm.io/datatypes"
	"start/datasources"
	"start/models"
)

type UpdateActivityLog struct {
	ActivityLog
	Value UpdateLogData `json:"value"`
}

type UpdateLogData struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

type UpdateActivityLogData struct {
	ActivityLog
	Value datatypes.JSON `json:"value"`
}

func (item *UpdateActivityLogData) FindList(page models.Page) ([]UpdateActivityLogData, int64, error) {
	var list []UpdateActivityLogData
	total := int64(0)

	db := datasources.GetDatabase().Model(ActivityLog{})

	if item.Category != "" {
		db = db.Where("category = ?", item.Category)
	}

	if item.Label != "" {
		db = db.Where("label = ?", item.Label)
	}

	if item.Action != "" {
		db = db.Where("action = ?", item.Action)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
