package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type CancelBookingSetting struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	PeopleFrom int    `json:"people_from"`                                // Số người từ bao nhiêu
	PeopleTo   int    `json:"people_to"`                                  // Đến Số người bao nhiêu
	TimeMin    string `json:"time_min" gorm:"type:varchar(100)"`          // Thời gian min cho phép cancel vd: 120:15,...
	TimeMax    string `json:"time_max" gorm:"type:varchar(100)"`          // Thời gian max cho phép cancel vd: 120:15,...
}

func (item *CancelBookingSetting) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CancelBookingSetting) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CancelBookingSetting) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CancelBookingSetting) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CancelBookingSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CancelBookingSetting) FindList() ([]CancelBookingSetting, int64, error) {
	db := datasources.GetDatabase().Model(CancelBookingSetting{})
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

	if item.PeopleFrom > 0 {
		db = db.Where("people_from <= ? AND (people_to = 0 OR people_to >= ?)", item.PeopleFrom, item.PeopleFrom)
	}

	db.Count(&total)
	db = db.Debug().Find(&list)

	return list, total, db.Error
}

func (item *CancelBookingSetting) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CancelBookingSetting) ValidateBookingCancel(booking Booking) error {
	// Tính ra số giờ từ lúc cancel so với booking date
	bookingDate := booking.BookingDate
	teeTime := booking.TeeTime
	fullTimeBooking := bookingDate + " " + teeTime
	bookingDateUnixT := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_2, fullTimeBooking)
	rangeTime := time.Now().Unix() - bookingDateUnixT

	// Nếu là Agency
	if booking.AgencyId > 0 {
		bookingList := BookingList{
			AgencyId:    booking.AgencyId,
			BookingDate: booking.BookingDate,
		}

		_, total, err := bookingList.FindAllBookingList()
		if err != nil {
			return err
		}

		cancelBookingSetting := CancelBookingSetting{
			PeopleFrom: int(total),
		}

		list, _, cancelErrBooking := cancelBookingSetting.FindList()

		if cancelErrBooking != nil {
			return cancelErrBooking
		}

		if len(list) == 0 {
			return nil
		}

		cancelSetting := list[0]

		timeMax := strings.Split(cancelSetting.TimeMax, ":")
		timeMaxH, _ := strconv.ParseInt(timeMax[0], 10, 64)
		timeMaxM := int64(0)
		if len(timeMax) > 1 {
			timeMaxM, _ = strconv.ParseInt(timeMax[1], 10, 64)
		}
		timeMaxUnix := timeMaxH*3600 + timeMaxM*60

		timeMin := strings.Split(cancelSetting.TimeMin, ":")
		timeMinH, _ := strconv.ParseInt(timeMin[0], 10, 64)
		timeMinM := int64(0)
		if len(timeMax) > 1 {
			timeMinM, _ = strconv.ParseInt(timeMax[1], 10, 64)
		}
		timeMixUnix := timeMinH*3600 + timeMinM*60

		if rangeTime >= timeMixUnix && rangeTime <= timeMaxUnix {
			return nil
		}
		return errors.New("Booking chưa đủ thời gian hủy.")
	}

	// Hội viên muốn hủy đặt chỗ chơi golf đều phải thông báo trước 24h
	oneDayTimeUnix := int64(24 * 3600)
	if rangeTime < oneDayTimeUnix {
		return errors.New("Booking chưa đủ thời gian hủy.")
	}

	return nil
}
