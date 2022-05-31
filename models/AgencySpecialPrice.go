package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func (item *AgencySpecialPrice) IsDuplicated() bool {
	modelCheck := AgencySpecialPrice{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		Dow:        item.Dow,
		AgencyId:   item.AgencyId,
	}

	errFind := modelCheck.FindFirst()
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

func (item *AgencySpecialPrice) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *AgencySpecialPrice) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AgencySpecialPrice) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *AgencySpecialPrice) FindList(page Page) ([]AgencySpecialPrice, int64, error) {
	db := datasources.GetDatabase().Model(AgencySpecialPrice{})
	list := []AgencySpecialPrice{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *AgencySpecialPrice) Count() (int64, error) {
	db := datasources.GetDatabase().Model(AgencySpecialPrice{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AgencySpecialPrice) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
