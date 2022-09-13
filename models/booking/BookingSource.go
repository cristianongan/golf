package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"time"

	"github.com/pkg/errors"
)

type BookingSource struct {
	models.ModelId
	PartnerUid        string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid         string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	AgencyId          string `json:"agency_id" gorm:"type:varchar(100)"`         // Id của Agency
	BookingSourceName string `json:"booking_source_name"`                        // Tên source: VNPAY,...
	IsPart1TeeType    bool   `json:"is_part1_tee_type"`                          // Part1 enable hay ko
	IsPart2TeeType    bool   `json:"is_part2_tee_type"`                          // Part2 enable hay ko
	IsPart3TeeType    bool   `json:"is_part3_tee_type"`                          // Part3 enable hay ko
	NormalDay         bool   `json:"normal_day"`                                 // Check ngày thường
	Weekend           bool   `json:"week_end"`                                   // Check cuối tuần
	NumberOfDays      int64  `json:"number_of_days"`                             // Số ngày cho phép booking tính từ ngày đặt booking
}

func (item *BookingSource) ValidateTimeRuleInBookingSource(BookingDate string, TeePart string) error {
	errF := item.FindFirst()
	if errF != nil {
		return errors.New("BookingSource not found")
	}
	currentDInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, utils.GetCurrentDay1())
	lastDInt := currentDInt + item.NumberOfDays*24*60*60

	bookingDateInt := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, BookingDate)

	checkTimeRule := bookingDateInt >= currentDInt && bookingDateInt <= lastDInt

	listTeePart := []string{}
	if item.IsPart1TeeType {
		listTeePart = append(listTeePart, "MORNING")
	}
	if item.IsPart2TeeType {
		listTeePart = append(listTeePart, "NOON")
	}
	if item.IsPart1TeeType {
		listTeePart = append(listTeePart, "NIGHT")
	}

	if utils.Contains(listTeePart, TeePart) {
		return errors.New("Tee Time đang không mở.")
	}

	if item.NormalDay && item.Weekend {
		if checkTimeRule {
			return nil
		}
	} else if item.NormalDay && !item.Weekend {
		if !utils.IsWeekend(bookingDateInt) && checkTimeRule {
			return nil
		}
	} else if !item.NormalDay && item.Weekend {
		if utils.IsWeekend(bookingDateInt) && checkTimeRule {
			return nil
		}
	}

	return errors.New("BookingDate không nằm trong ngày quy định của Booking Source")
}

func (item *BookingSource) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *BookingSource) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingSource) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *BookingSource) Count() (int64, error) {
	db := datasources.GetDatabase().Model(BookingSource{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingSource) FindList(page models.Page) ([]BookingSource, int64, error) {
	db := datasources.GetDatabase().Model(BookingSource{})
	list := []BookingSource{}
	total := int64(0)
	status := item.ModelId.Status
	db = db.Joins("JOIN agencies ON agencies.Id = booking_sources.agency_id")
	db = db.Select("agencies.*,agencies.agency_id")

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if status != "" {
		db = db.Where("status = ?", item.Status)
	}

	if item.BookingSourceName != "" {
		db = db.Where("booking_source_name LIKE ?", "%"+item.BookingSourceName+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingSource) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
