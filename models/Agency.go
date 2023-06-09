package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Đại lý
type Agency struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	AgencyId   string `json:"agency_id" gorm:"type:varchar(100);index"`   // Id Agency
	ShortName  string `json:"short_name" gorm:"type:varchar(256)"`        // Ten ngắn Dai ly
	// Category             string         `json:"category" gorm:"type:varchar(256);index"`    // Category
	Type                 string         `json:"type" gorm:"type:varchar(256);index"`        // AGENCY / OTA / COMPANY
	GuestStyle           string         `json:"guest_style" gorm:"type:varchar(256);index"` // Guest Style
	Name                 string         `json:"name" gorm:"type:varchar(500)"`              // Ten Dai ly
	Province             string         `json:"province" gorm:"type:varchar(100)"`          //
	PrimaryContactFirst  AgencyContact  `json:"primary_contact_first,omitempty" gorm:"type:json"`
	PrimaryContactSecond AgencyContact  `json:"primary_contact_second,omitempty" gorm:"type:json"`
	ContractDetail       AgencyContract `json:"contract_detail,omitempty" gorm:"type:json"`
	Avatar               string         `json:"avatar" gorm:"type:varchar(256)"`
}

type AgencyDetailRes struct {
	Agency
	NumberOfContract int64 `json:"number_of_contract"`
	NumberOfVoucher  int64 `json:"number_of_voucher"`
	NumberOfCustomer int64 `json:"number_of_customer"`
}

type AgencyContact struct {
	Name        string `json:"name"`
	JobTile     string `json:"job_title"`
	DirectPhone string `json:"direct_phone"`
	Mail        string `json:"mail"`
}

func (item *AgencyContact) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item AgencyContact) Value() (driver.Value, error) {
	return json.Marshal(&item)
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

func (item *AgencyContract) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item AgencyContract) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *Agency) IsDuplicated(db *gorm.DB) bool {
	modelCheck := Agency{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		AgencyId:   item.AgencyId,
		// ShortName:  item.ShortName,
	}

	errFind := modelCheck.FindFirst(db)
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *Agency) IsDuplicatedContract(database *gorm.DB, contractNo string) error {
	db := database.Model(Agency{})
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if contractNo != "" {
		db = db.Where("contract_detail->'$.contract_no' = ?", contractNo)
	}

	return db.First(item).Error
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
	if item.ShortName == "" {
		return false
	}
	if item.GuestStyle == "" {
		return false
	}
	return true
}

func (item *Agency) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	// db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Agency) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Agency) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Agency) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Agency{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Agency) FindList(database *gorm.DB, page Page) ([]Agency, int64, error) {
	db := database.Model(Agency{})
	list := []Agency{}
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
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%").Or("short_name = ?", item.Name)
	}
	if item.AgencyId != "" {
		db = db.Where("agency_id = ?", item.AgencyId)
	}
	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Agency) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *Agency) GetNumberCustomer(database *gorm.DB) int64 {
	total := int64(0)
	db := database.Model(CustomerUser{})
	db = db.Where("agency_id = ?", item.Id)
	db.Count(&total)
	return total
}
