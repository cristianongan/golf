package models

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/utils"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	Uid       string `gorm:"primary_key" sql:"not null;" json:"uid"`
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status" gorm:"index;type:varchar(50)"` //ENABLE, DISABLE, TESTING
}

type ModelId struct {
	Id        int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status"  gorm:"type:varchar(50)"` //ENABLE, DISABLE, TESTING, DELETED
}

type ModelLog struct {
	Id        int64 `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt int64 `json:"created_at" gorm:"index"`
}

type ListInt64 []int64

func (item *ListInt64) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListInt64) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type ListInt []int

func (item *ListInt) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item *ListInt) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ==================================================
type ListString []string

func (item *ListString) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListString) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ================= YtElmInfo ======================
type YtElmInfo struct {
	YtElmId      string `json:"yt_eml_id"`
	YtElmKeyword string `json:"yt_eml_keyword"`
	Duration     string `json:"duration"`
}

func (item *YtElmInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item YtElmInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ================= ListYtElmInfo ======================
type ListYtElmInfo []YtElmInfo

func (item *ListYtElmInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListYtElmInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ====================================================

/*
Check ngày và h
*/
func CheckDow(dow, hour string, timeCheck time.Time, partnerUid, courseUid string) bool {
	if dow == "" {
		return false
	}

	list := strings.Split(dow, "")
	// log.Println("Check Dow ", list, len(list))

	if len(list) == 0 {
		return false
	}

	// Check nếu dow chứa 0 thì check ngày truyền vào có phải là holiday hay không

	isOk := false
	for _, v := range list {
		dayInt, err := strconv.Atoi(v)
		if err != nil {
			log.Println("CheckDow err", err.Error())
		}

		if dayInt == 0 {
			// check ngay truyen vao co phai la holiday hay khong
			dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())
			if CheckHoliday(partnerUid, courseUid, dateDisplay) {
				isOk = true
				break
			}
		}

		if dayInt == int(timeCheck.Weekday()+1) {
			if hour != "" {
				if CheckHour(hour, timeCheck) {
					isOk = true
				}
			} else {
				isOk = true
			}
		}
	}

	return isOk
}

/*
Check giờ: format 13:00,23:00
*/
func CheckHour(hour string, timeCheck time.Time) bool {

	currentHour := timeCheck.Hour()
	currentMinute := timeCheck.Minute()

	// Parse Hour
	fromHour := -1
	fromMinute := -1
	toHour := -1
	toMinute := -1
	if strings.Contains(hour, ",") {
		listH := strings.Split(hour, ",")
		for i, v := range listH {
			if i == 0 {
				timeHour, err := utils.ConvertHourToTime(v)
				if err == nil {
					fromHour = timeHour.Hour()
					fromMinute = timeHour.Minute()
				} else {
					log.Println("CheckHour err0", err.Error())
				}
			} else if i == 1 {
				timeHour, err := utils.ConvertHourToTime(v)
				if err == nil {
					toHour = timeHour.Hour()
					toMinute = timeHour.Minute()
				} else {
					log.Println("CheckHour err1", err.Error())
				}
			}
		}
	}

	if fromHour >= 0 && toHour == -1 {
		if currentHour > fromHour {
			return true
		}
		if currentHour == fromHour && currentMinute >= fromMinute {
			return true
		}
	}

	if fromHour == -1 && toHour >= 0 {
		if currentHour < toHour {
			return true
		}
		if currentHour == toHour && currentMinute <= toMinute {
			return true
		}
	}
	if fromHour >= 0 && toHour >= 0 {
		if fromHour <= currentHour && currentHour <= toHour {
			return true
		}

	}
	return false
}
