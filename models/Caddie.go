package models

import (
	"start/constants"
	"time"

	"gorm.io/gorm"
)

type Caddie struct {
	ModelId
	PartnerUid     string           `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string           `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code           string           `json:"code" gorm:"type:varchar(256);index"`        // Id Caddie vận hành
	Name           string           `json:"name" gorm:"type:varchar(120)"`
	Sex            bool             `json:"sex"`
	Avatar         string           `json:"avatar" gorm:"type:varchar(256);index"` // San Golf
	BirthDay       int64            `json:"birth_day"`
	WorkingStatus  string           `json:"working_status" gorm:"type:varchar(128)"` // Active | Inactive
	Group          string           `json:"group" gorm:"-"`                          // Caddie thuộc nhóm nào
	StartedDate    int64            `json:"started_date"`                            // Ngày bắt đầu làm việc của Caddie
	IdHr           string           `json:"id_hr" gorm:"type:varchar(100)"`
	Phone          string           `json:"phone" gorm:"type:varchar(20)"`
	Email          string           `json:"email" gorm:"type:varchar(100)"`
	IdentityCard   string           `json:"identity_card" gorm:"type:varchar(20)"`    // Số CMT/CCCD của caddie
	IssuedBy       string           `json:"issued_by" gorm:"type:varchar(200)"`       // Nơi cấp CMT/CCCD
	ExpiredDate    int64            `json:"expired_date"`                             // Ngày hết hạn của CMT/CCCD
	PlaceOfOrigin  string           `json:"place_of_origin" gorm:"type:varchar(200)"` //Quê quán
	Address        string           `json:"address" gorm:"type:varchar(200)"`         // Địa chỉ của Caddie
	Level          string           `json:"level" gorm:"type:varchar(40)"`            // Hạng của Caddie.(A,B,C,D)
	Note           string           `json:"note" gorm:"type:varchar(200)"`
	CurrentStatus  string           `json:"current_status" gorm:"type:varchar(128)"`
	CurrentRound   int              `json:"current_round" gorm:"size:2"`
	ContractStatus string           `json:"contract_status" gorm:"type:varchar(128);index"`
	RdStatus       string           `json:"rd_status" gorm:"type:varchar(128)"`
	DutyStatus     string           `json:"duty_status" gorm:"type:varchar(128)"`
	CaddieCalendar []CaddieCalendar `json:"caddie_calendar" gorm:"foreignKey:caddie_uid"`
	GroupId        int64            `json:"group_id" gorm:"default:0;index"`
	GroupIndex     uint64           `json:"group_index" gorm:"default:0"`
	GroupInfo      CaddieGroup      `json:"group_info" gorm:"foreignKey:GroupId"`
}

type CaddieResponse struct {
	Caddie
	Booking int64 `json:"booking"`
}

func (item *Caddie) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *Caddie) CreateBatch(db *gorm.DB, caddies []Caddie) error {
	now := time.Now()
	for i := range caddies {
		c := &caddies[i]
		c.ModelId.CreatedAt = now.Unix()
		c.ModelId.UpdatedAt = now.Unix()
		c.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.CreateInBatches(caddies, 100).Error
}

func (item *Caddie) Delete(db *gorm.DB) error {
	return db.Delete(item).Error
}

func (item *Caddie) SolfDelete(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	item.ModelId.Status = constants.STATUS_DELETED
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Caddie) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Caddie) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Caddie) FindCaddieDetail(database *gorm.DB) (CaddieResponse, error) {
	total := int64(0)
	var caddieObj Caddie
	db1 := database.Model(Caddie{})
	db1 = db1.Where("caddies.id = ?", item.Id)
	db1.Preload("GroupInfo")
	db1.Find(&caddieObj)

	// Đếm lượt booking của caddie
	db2 := database.Model(Caddie{})
	db2 = db2.Joins("JOIN bookings ON bookings.caddie_id = caddies.id")
	db2.Count(&total)

	caddieResponse := CaddieResponse{
		Caddie:  caddieObj,
		Booking: total,
	}

	return caddieResponse, db1.Error
}

func (item *Caddie) Count(database *gorm.DB) (int64, error) {
	total := int64(0)

	db := database.Model(Caddie{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Caddie) FindList(database *gorm.DB, page Page) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := database.Model(Caddie{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.Name != "" {
		db = db.Where("name LIKE ?", "%"+item.Name+"%")
	}
	if item.WorkingStatus != "" {
		db = db.Where("working_status = ?", item.WorkingStatus)
	}
	if item.Code != "" {
		db = db.Where("code LIKE ?", "%"+item.Code+"%")
	}
	if item.Level != "" {
		db = db.Where("level = ?", item.Level)
	}
	if item.Phone != "" {
		db = db.Where("phone = ?", item.Phone)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Caddie) FindListCaddieNotReady(database *gorm.DB) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := database.Model(Caddie{})

	db = db.Not("current_status = ?", constants.CADDIE_CURRENT_STATUS_READY)
	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *Caddie) FindAllCaddieContract(database *gorm.DB) ([]Caddie, error) {
	var list []Caddie

	db := database.Model(Caddie{})

	db = db.Where("contract_status IN (?, ?)", constants.CADDIE_CONTRACT_STATUS_PARTTIME, constants.CADDIE_CONTRACT_STATUS_FULLTIME)
	db = db.Find(&list)

	return list, db.Error
}

func (item *Caddie) FindAllCaddieGroup(database *gorm.DB, status string, listGroup []int64) ([]Caddie, error) {
	var list []Caddie

	db := database.Model(Caddie{})

	db = db.Where("current_status = ?", constants.CADDIE_CURRENT_STATUS_READY)
	db = db.Where("contract_status = ?", status)
	db = db.Where("group_id IN ?", listGroup)

	db = db.Not("status = ?", constants.STATUS_DELETED)
	db = db.Order("group_id")

	db = db.Find(&list)

	return list, db.Error
}
