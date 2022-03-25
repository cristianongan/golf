package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Phí Golf
type GolfFee struct {
	ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	TablePriceId   int64  `json:"table_price_id" gorm:"index"`                // Id Bang gia
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Ten Guest style
	GuestStyle     string `json:"guest_style" gorm:"index;type:varchar(200)"` // Guest style
	Dow            string `json:"dow" gorm:"type:varchar(100)"`               // Dow
	GreenFee       string `json:"green_fee" gorm:"type:varchar(256)"`         // Phi san cỏ
	CaddieFee      string `json:"caddie_fee" gorm:"type:varchar(256)"`        // Phi Caddie
	BuggyFee       string `json:"buggy_fee" gorm:"type:varchar(256)"`         // Phi buggy
	UpdateUserUid  string `json:"update_user_uid" gorm:"index"`               // Nguoi sua
	UpdateUserName string `json:"update_user_name"`                           // Nguoi sua
	AccCode        string `json:"acc_code" gorm:"type:varchar(200)"`          // Kết nối với phần mềm kế toán
	Note           string `json:"note" gorm:"type:varchar(500)"`              // Note
	NodeOdd        int    `json:"node_odd"`                                   // 0 || 1 Chỉ tính hố lẻ thì tick vào đây
	PaidType       string `json:"paid_type" gorm:"type:varchar(50)"`          // Kiểu thanh toán: NOW / AFTER
	Idx            int    `json:"idx"`                                        // Xắp xếp thứ tự
	AccDebit       string `json:"acc_debit"`                                  // Mã kế toán ghi nợ
}

func (item *GolfFee) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *GolfFee) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GolfFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *GolfFee) Count() (int64, error) {
	db := datasources.GetDatabase().Model(GolfFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *GolfFee) FindList(page Page) ([]GolfFee, int64, error) {
	db := datasources.GetDatabase().Model(GolfFee{})
	list := []GolfFee{}
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

func (item *GolfFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
