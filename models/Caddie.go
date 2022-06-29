package models

import (
	"start/constants"
	"start/datasources"
	"time"
)

type Caddie struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code          string `json:"code" gorm:"type:varchar(256);index"`        // Id Caddie vận hành
	Name          string `json:"name" gorm:"type:varchar(120)"`
	Sex           bool   `json:"sex"`
	Avatar        string `json:"avatar" gorm:"type:varchar(256);index"` // San Golf
	IsInCourse    bool   `json:"is_in_course"`                          // Caddie có đang trên sân không
	BirthDay      int64  `json:"birth_day"`
	WorkingStatus string `json:"working_status" gorm:"type:varchar(20)"` // Trạng thái làm việc của Caddie: FullTime, Partime, Nghỉ...
	Group         string `json:"group" gorm:"type:varchar(20)"`          // Caddie thuộc nhóm nào
	StartedDate   int64  `json:"started_date"`                           // Ngày bắt đầu làm việc của Caddie
	IdHr          string `json:"id_hr" gorm:"type:varchar(100)"`
	Phone         string `json:"phone" gorm:"type:varchar(20)"`
	Email         string `json:"email" gorm:"type:varchar(100)"`
	IdentityCard  string `json:"identity_card" gorm:"type:varchar(20)"`    // Số CMT/CCCD của caddie
	IssuedBy      string `json:"issued_by" gorm:"type:varchar(200)"`       // Nơi cấp CMT/CCCD
	ExpiredDate   int64  `json:"expired_date"`                             // Ngày hết hạn của CMT/CCCD
	PlaceOfOrigin string `json:"place_of_origin" gorm:"type:varchar(200)"` //Quê quán
	Address       string `json:"address" gorm:"type:varchar(200)"`         // Địa chỉ của Caddie
	Level         string `json:"level" gorm:"type:varchar(40)"`            // Hạng của Caddie.(A,B,C,D)
	Note          string `json:"note" gorm:"type:varchar(200)"`
}

type CaddieResponse struct {
	Caddie
	Booking int64 `json:"booking"`
}

func (item *Caddie) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Caddie) CreateBatch(caddies []Caddie) error {
	now := time.Now()
	for i := range caddies {
		c := &caddies[i]
		c.ModelId.CreatedAt = now.Unix()
		c.ModelId.UpdatedAt = now.Unix()
		c.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.CreateInBatches(caddies, 100).Error
}

func (item *Caddie) Delete() error {
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Caddie) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Caddie) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Caddie) FindCaddieDetail() (CaddieResponse, error) {
	total := int64(0)
	var caddieObj Caddie
	db := datasources.GetDatabase().Model(Caddie{})
	db.Where(item).Find(&caddieObj)

	db = db.Where("caddies.id = ?", item.Id)
	db = db.Joins("JOIN bookings ON bookings.caddie_id = caddies.id")
	db = db.Count(&total)

	caddieResponse := CaddieResponse{
		Caddie:  caddieObj,
		Booking: total,
	}

	return caddieResponse, db.Error
}

func (item *Caddie) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Caddie{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Caddie) FindList(page Page) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := datasources.GetDatabase().Model(Caddie{})

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
		db = db.Where("code = ?", item.Code)
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
