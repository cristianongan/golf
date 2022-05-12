package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// User  khách hàng
type CustomerUser struct {
	Model
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Type        string `json:"type" gorm:"type:varchar(100);index"`        // Loai khach hang: Member, Guest, Visitor...
	Name        string `json:"name" gorm:"type:varchar(256);index"`        // Ten KH
	Dob         int64  `json:"dob"`                                        // Ngay sinh
	Sex         int    `json:"sex"`                                        // giới tính
	Avatar      string `json:"avatar" gorm:"type:varchar(150)"`            // ảnh đại diện
	Nationality string `json:"nationality" gorm:"type:varchar(100)"`       // Quốc gia
	Phone       string `json:"phone" gorm:"type:varchar(20);index"`        // So dien thoai
	CellPhone   string `json:"cell_phone" gorm:"type:varchar(20)"`         // So dien thoai
	Fax         string `json:"fax" gorm:"type:varchar(100);index"`         // So Fax
	Email       string `json:"email" gorm:"type:varchar(100)"`             // Email
	Address1    string `json:"address1" gorm:"type:varchar(500)"`          // Dia chi
	Address2    string `json:"address2" gorm:"type:varchar(500)"`          // Dia chi
	Job         string `json:"job" gorm:"type:varchar(200)"`               // Ex: Ngan hang
	Position    string `json:"position" gorm:"type:varchar(200)"`          // Ex: Chu tich
	CompanyName string `json:"company_name" gorm:"type:varchar(200)"`      // Ten cong ty
	Mst         string `json:"mst" gorm:"type:varchar(50)"`                // mã số thuế
	Note        string `json:"note" gorm:"type:varchar(500)"`              // Ghi chu them
}

func (item *CustomerUser) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = item.CourseUid + "-" + utils.HashCodeUuid(uid.String())
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CustomerUser) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CustomerUser) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CustomerUser) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CustomerUser{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CustomerUser) FindList(page Page, partnerUid, courseUid, typeCus, customerUid, name string) ([]CustomerUser, int64, error) {
	db := datasources.GetDatabase().Model(CustomerUser{})
	list := []CustomerUser{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if partnerUid != "" {
		db = db.Where("partner_uid = ?", partnerUid)
	}
	if courseUid != "" {
		db = db.Where("course_uid = ?", courseUid)
	}
	if typeCus != "" {
		db = db.Where("type = ?", typeCus)
	}
	if customerUid != "" {
		db = db.Where("uid = ?", customerUid)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	db = db.Where("type = ?", typeCus)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CustomerUser) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
