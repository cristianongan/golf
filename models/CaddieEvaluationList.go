package models

import "start/datasources"

type CaddieEvaluationList struct {
	CourseUid  string
	CaddieName string
	CaddieCode string
	Month      string
	BookingUid string
}

func (item *CaddieEvaluationList) FindList(page Page) ([]CaddieEvaluation, int64, error) {
	var list []CaddieEvaluation
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieEvaluation{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" {
		db = db.Where("caddie_name LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("caddie_code = ?", item.CaddieCode)
	}

	if item.Month != "" {
		db = db.Where("DATE_FORMAT(booking_date, '%Y-%m') = ?", item.Month)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieEvaluationList) FindFirst() (CaddieEvaluation, error) {
	var result CaddieEvaluation
	db := datasources.GetDatabase().Model(CaddieEvaluation{})
	err := db.Where(item).First(&result).Error
	return result, err
}
