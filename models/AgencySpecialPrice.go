package models

import (
	"start/constants"
	"start/utils"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Phí đặc biệt Agency
type AgencySpecialPrice struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	AgencyId   int64  `json:"agency_id" gorm:"index"`                     // Id Agency
	FromHour   string `json:"from_hour" gorm:"type:varchar(50)"`          // time format : HH:mm
	ToHour     string `json:"to_hour" gorm:"type:varchar(50)"`            // time format: HH:mm
	Dow        string `json:"dow" gorm:"type:varchar(100)"`               // Dow
	GreenFee   int64  `json:"green_fee"`                                  // Phi sân cỏ
	CaddieFee  int64  `json:"caddie_fee"`                                 // Phi Caddie
	BuggyFee   int64  `json:"buggy_fee"`                                  // Phi buggy
	Note       string `json:"note" gorm:"type:varchar(400)"`
	Input      string `json:"input" gorm:"type:varchar(100)"`
}

/*
check giá riêng agency có thoả mãn dk time
*/
func (item *AgencySpecialPrice) FindOtherPriceOnTime(db *gorm.DB) (AgencySpecialPrice, error) {
	listAgency, _, err := item.FindListByAgencyId(db)

	if err != nil {
		return AgencySpecialPrice{}, err
	}
	idx := -1
	// check for today
	for i, v := range listAgency {
		fromH := "00:00"
		if v.FromHour != "" {
			fromH = v.FromHour
		}
		toH := "00:00"
		if v.ToHour != "" {
			toH = v.ToHour
		}

		hourLast := fromH + "," + toH
		if utils.CheckDow(v.Dow, hourLast, time.Now()) {
			idx = i
		}
	}

	if idx >= 0 {
		return listAgency[idx], nil
	}

	return AgencySpecialPrice{}, errors.New("")
}

func (item *AgencySpecialPrice) FindOtherPriceOnDate(db *gorm.DB, date time.Time) (AgencySpecialPrice, error) {
	listAgency, _, err := item.FindListByAgencyId(db)

	if err != nil {
		return AgencySpecialPrice{}, err
	}
	idx := -1
	// check for today
	for i, v := range listAgency {
		fromH := "00:00"
		if v.FromHour != "" {
			fromH = v.FromHour
		}
		toH := "00:00"
		if v.ToHour != "" {
			toH = v.ToHour
		}

		hourLast := fromH + "," + toH
		if utils.CheckDow(v.Dow, hourLast, date) {
			idx = i
		}
	}

	if idx >= 0 {
		return listAgency[idx], nil
	}

	return AgencySpecialPrice{}, errors.New("")
}

func (item *AgencySpecialPrice) IsDuplicated(db *gorm.DB) bool {
	modelCheck := AgencySpecialPrice{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Dow:        item.Dow,
		AgencyId:   item.AgencyId,
	}

	errFind := modelCheck.FindFirst(db)
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *AgencySpecialPrice) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	// if item.CourseUid == "" {
	// 	return false
	// }
	if item.AgencyId <= 0 {
		return false
	}
	return true
}

func (item *AgencySpecialPrice) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *AgencySpecialPrice) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AgencySpecialPrice) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *AgencySpecialPrice) FindListByAgencyId(database *gorm.DB) ([]AgencySpecialPrice, int64, error) {
	db := database.Model(AgencySpecialPrice{})
	list := []AgencySpecialPrice{}
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
	if item.AgencyId > 0 {
		db = db.Where("agency_id = ?", item.AgencyId)
	}

	db.Count(&total)
	db.Find(&list)
	return list, total, db.Error
}

func (item *AgencySpecialPrice) FindList(database *gorm.DB, page Page, agencyIdStr, agencyName string) ([]map[string]interface{}, int64, error) {
	db := database.Table("agency_special_prices")
	list := []map[string]interface{}{}
	total := int64(0)

	db = db.Select(`agency_special_prices.*, 
	agencies.agency_id as agency_id_str,
	agencies.name as agency_name,
	agencies.short_name as short_name,
	agencies.guest_style as guest_style`)

	db = db.Joins("LEFT JOIN agencies on agency_special_prices.agency_id = agencies.id")

	if item.PartnerUid != "" {
		db = db.Where("agency_special_prices.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("agency_special_prices.course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("agency_special_prices.status = ?", item.Status)
	}
	if agencyIdStr != "" {
		db = db.Where("agency_special_prices.agency_id = ?", agencyIdStr)
	}

	// queryStr := `select * from (select tb0.*,
	// agencies.agency_id as agency_id_str,
	// agencies.name as agency_name,
	// agencies.short_name as short_name,
	// agencies.guest_style as guest_style,
	// agencies.category as category
	// from (select * from agency_special_prices WHERE agency_special_prices.partner_uid = ` + `"` + item.PartnerUid + `"`

	// if item.CourseUid != "" {
	// 	queryStr = queryStr + " and agency_special_prices.course_uid = " + `"` + item.CourseUid + `"`
	// }
	// if item.Status != "" {
	// 	queryStr = queryStr + " and agency_special_prices.status = " + `"` + item.Status + `"`
	// }

	// queryStr = queryStr + ") tb0 "
	// queryStr = queryStr + `LEFT JOIN agencies on tb0.agency_id = agencies.id ) tb1 `

	// // TODO: chưa tối ưu truy vấn
	// if agencyIdStr != "" {
	// 	queryStr = queryStr + " WHERE tb1.agency_id_str = " + `"` + agencyIdStr + `"`
	// }
	// // if agencyName != "" {
	// // 	queryStr = queryStr + " WHERE tb1.agency_name = " + `"` + agencyName + `"`
	// // }

	// // queryStr = queryStr + ") af on tb0.uid = af.member_card_uid) tb1 "

	// // var countReturn CountStruct
	// var countReturn utils.CountStruct
	// strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	// errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	// if errCount != nil {
	// 	log.Println("AgencySpecialPrice err", errCount.Error())
	// 	return list, total, errCount
	// }

	// total = countReturn.Count
	// //Check if limit large then set to 50
	// if page.Limit > 50 {
	// 	page.Limit = 50
	// }

	// if total > 0 && int64(page.Offset()) < total {
	// 	queryStr = queryStr + " order by tb1." + page.SortBy + " " + page.SortDir + " LIMIT " + strconv.Itoa(page.Limit) + " OFFSET " + strconv.Itoa(page.Offset())
	// }
	// err := db.Raw(queryStr).Scan(&list).Error
	// if err != nil {
	// 	return list, total, err
	// }

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *AgencySpecialPrice) Count(database *gorm.DB) (int64, error) {
	db := database.Model(AgencySpecialPrice{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AgencySpecialPrice) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
