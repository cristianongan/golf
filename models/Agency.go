package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Đại lý
type Agency struct {
	ModelId
	Code                 string         `json:"code" gorm:"type:varchar(100);uniqueIndex"`  // Mã code
	PartnerUid           string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid            string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	AgencyId             string         `json:"agency_id" gorm:"type:varchar(100);index"`   // Id Agency
	ShortName            string         `json:"short_name" gorm:"type:varchar(256)"`        // Ten ngắn Dai ly
	Category             string         `json:"category" gorm:"type:varchar(256);index"`    // Category
	GuestStyle           string         `json:"guest_style" gorm:"type:varchar(256);index"` // Guest Style
	Name                 string         `json:"name" gorm:"type:varchar(500)"`              // Ten Dai ly
	Province             string         `json:"province" gorm:"type:varchar(100)"`          //
	PrimaryContactFirst  AgencyContact  `json:"primary_contact_first,omitempty" gorm:"type:json"`
	PrimaryContactSecond AgencyContact  `json:"primary_contact_second,omitempty" gorm:"type:json"`
	ContractDetail       AgencyContract `json:"contract_detail,omitempty" gorm:"type:json"`
}

type AgencyContact struct {
	Name        string `json:"name"`
	JobTile     string `json:"job_title"`
	DirectPhone string `json:"direct_phone"`
	Mail        string `json:"mail"`
}

type AgencyContract struct {
	ContractNo      string `json:"contract_no"`
	ExpDate         int64  `json:"exp_date"`
	ContractDate    int64  `json:"contract_date"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	OfficialAddress string `json:"official_address"`
	TaxCode         string `json:"tax_code"`
	Note            string `json:"note"`
	Rounds          int    `json:"rounds"`
	PrePaid         bool   `json:"pre_paid"`
	ContractAddress string `json:"contract_address"`
}

func (item *Agency) IsDuplicated() bool {
	modelCheck := Agency{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		AgencyId:   item.AgencyId,
		ShortName:  item.ShortName,
	}

	errFind := modelCheck.FindFirst()
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *Agency) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.AgencyId == "" {
		return false
	}
	if item.Code == "" {
		return false
	}
	if item.GuestStyle == "" {
		return false
	}
	return true
}

func (item *Agency) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Agency) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Agency) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Agency) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Agency{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Agency) FindList(page Page) ([]Agency, int64, error) {
	db := datasources.GetDatabase().Model(Agency{})
	list := []Agency{}
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

func (item *Agency) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
