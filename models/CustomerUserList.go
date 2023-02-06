package models

import (
	"start/utils"
	"strconv"

	"gorm.io/gorm"
)

type CustomerUserList struct {
	FromBirthDate string
	ToBirthDate   string
	CourseUid     string
}

func (item CustomerUserList) FindCustomerList(database *gorm.DB, page Page) ([]CustomerUser, int64, error) {
	var list []CustomerUser
	total := int64(0)

	db := database.Model(CustomerUser{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.FromBirthDate != "" {
		db = db.Where("DATE(DATE_FORMAT(FROM_UNIXTIME(dob), ?)) >= ?", strconv.FormatInt(int64(utils.GetTimeNow().Year()), 10)+"-%m-%d", item.FromBirthDate)
	}

	if item.ToBirthDate != "" {
		db = db.Where("DATE(DATE_FORMAT(FROM_UNIXTIME(dob), ?)) <= ?", strconv.FormatInt(int64(utils.GetTimeNow().Year()), 10)+"-%m-%d", item.ToBirthDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
