package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/models"
	"start/utils"

	// "start/utils"

	// "strconv"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CancelBookingSetting struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	PeopleFrom int    `json:"people_from"`                                // Số người từ bao nhiêu
	PeopleTo   int    `json:"people_to"`                                  // Đến Số người bao nhiêu
	Time       int    `json:"time"`                                       // Thời gian min cho phép cancel vd: 120:15,...
	Type       int64  `json:"type"`                                       // Xác định Setting nào cùng loại
}

type ListCancelBookingSetting []CancelBookingSetting

func (item *ListCancelBookingSetting) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListCancelBookingSetting) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *CancelBookingSetting) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	return true
}

func (item *CancelBookingSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *CancelBookingSetting) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CancelBookingSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CancelBookingSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(CancelBookingSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CancelBookingSetting) FindList(database *gorm.DB) ([]CancelBookingSetting, int64, error) {
	db := database.Model(CancelBookingSetting{})
	list := []CancelBookingSetting{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.Type > 0 {
		db = db.Where("type = ?", item.Type)
	}

	if item.PeopleFrom > 0 {
		db = db.Where("people_from <= ? AND (people_to = 0 OR people_to >= ?)", item.PeopleFrom, item.PeopleFrom)
	}

	db.Count(&total)
	db = db.Find(&list)

	return list, total, db.Error
}

func (item *CancelBookingSetting) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *CancelBookingSetting) ValidateBookingCancel(db *gorm.DB, booking Booking) error {
	// Tính ra số giờ từ lúc cancel so với lúc tạo booking CreateAt
	rangeTime := utils.GetTimeNow().Unix() - booking.CreatedAt
	oneDayTimeUnix := int64(24 * 3600)
	twoHourTimeUnix := int64(2 * 3600)

	// Nếu là Agency
	if booking.AgencyId > 0 {
		bookingList := BookingList{
			AgencyId:    booking.AgencyId,
			BookingDate: booking.BookingDate,
		}

		_, total, err := bookingList.FindAllBookingList(db)
		if err != nil {
			return err
		}

		cancelBookingSetting := CancelBookingSetting{
			PeopleFrom: int(total),
		}

		list, _, cancelErrBooking := cancelBookingSetting.FindList(db)

		if cancelErrBooking != nil {
			return cancelErrBooking
		}

		if len(list) == 0 {
			return nil
		}

		cancelSetting := list[0]

		if cancelSetting.Status == constants.STATUS_DISABLE {
			return nil
		}

		if rangeTime >= twoHourTimeUnix && rangeTime <= oneDayTimeUnix {
			return errors.New("Book trước 2h và dưới 24h thì sẽ không được hủy booking.")
		}

		if rangeTime > int64(cancelBookingSetting.Time) {
			return errors.New("Booking đã quá thời gian hủy.")
		}
	}

	// Hội viên muốn hủy đặt chỗ chơi golf đều phải thông báo trước 24h
	if rangeTime > oneDayTimeUnix {
		return errors.New("Booking đã quá thời gian hủy.")
	}

	return nil
}
