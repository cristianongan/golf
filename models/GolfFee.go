package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"start/utils"

	"github.com/pkg/errors"
)

// Phí Golf
type GolfFee struct {
	ModelId
	PartnerUid     string                `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid      string                `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	TablePriceId   int64                 `json:"table_price_id" gorm:"index"`                // Id Bang gia
	GuestStyleName string                `json:"guest_style_name" gorm:"type:varchar(256)"`  // Ten Guest style
	GuestStyle     string                `json:"guest_style" gorm:"index;type:varchar(200)"` // Guest style
	Dow            string                `json:"dow" gorm:"type:varchar(100)"`               // Dow
	GreenFee       utils.ListGolfHoleFee `json:"green_fee" gorm:"type:varchar(256)"`         // Phi san cỏ
	CaddieFee      utils.ListGolfHoleFee `json:"caddie_fee" gorm:"type:varchar(256)"`        // Phi Caddie
	BuggyFee       utils.ListGolfHoleFee `json:"buggy_fee" gorm:"type:varchar(256)"`         // Phi buggy
	//UpdateUserUid    string                `json:"update_user_uid" gorm:"index"`                    // Nguoi sua
	UpdateUserName   string `json:"update_user_name"`                                // Nguoi sua
	AccCode          string `json:"acc_code" gorm:"type:varchar(200)"`               // Kết nối với phần mềm kế toán
	Note             string `json:"note" gorm:"type:varchar(500)"`                   // Note
	NodeOdd          int    `json:"node_odd"`                                        // 0 || 1 Chỉ tính hố lẻ thì tick vào đây
	PaidType         string `json:"paid_type" gorm:"type:varchar(50)"`               // Kiểu thanh toán: NOW / AFTER
	Idx              int    `json:"idx"`                                             // Xắp xếp thứ tự
	AccDebit         string `json:"acc_debit"`                                       // Mã kế toán ghi nợ
	CustomerType     string `json:"customer_type" gorm:"index;type:varchar(100)"`    // Loại khách hàng
	CustomerCategory string `json:"customer_category" gorm:"index;type:varchar(50)"` // CUSTOMER, AGENCY
	GroupName        string `json:"group_name" gorm:"index;type:varchar(200)"`       // Tên nhóm Fee
	GroupId          int64  `json:"group_id" gorm:"index"`                           // Id nhóm Fee
}

type GuestStyle struct {
	TablePriceId   int64  `json:"table_price_id"`   // Id Bang gia
	GuestStyleName string `json:"guest_style_name"` // Ten Guest style
	GuestStyle     string `json:"guest_style"`      // Guest style
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

/*
 Get Golf Fee valid in to day
*/
func (item *GolfFee) GetGuestStyleOnDay() (GolfFee, error) {
	golfFee := GolfFee{
		GuestStyle: item.GuestStyle,
	}

	if item.GuestStyle == "" {
		return golfFee, errors.New("Guest style invalid")
	}

	list := []GolfFee{}
	db := datasources.GetDatabase().Model(GolfFee{})
	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("course_uid = ?", item.CourseUid)
	db = db.Where("guest_style = ?", item.GuestStyle)
	err := db.Find(&list).Error

	if err != nil {
		return golfFee, err
	}

	idxTemp := -1

	for i, golfFee_ := range list {
		if idxTemp < 0 {
			if utils.CheckDow(golfFee_.Dow, time.Now()) {
				idxTemp = i
			}
		}
	}

	if idxTemp >= 0 {
		return list[idxTemp], nil
	}

	return golfFee, nil
}

func (item *GolfFee) FindList(page Page) ([]GolfFee, int64, error) {
	db := datasources.GetDatabase().Model(GolfFee{})
	list := []GolfFee{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.GroupId > 0 {
		db = db.Where("group_id = ?", item.GroupId)
	}
	if item.TablePriceId > 0 {
		db = db.Where("table_price_id = ?", item.TablePriceId)
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

func (item *GolfFee) GetGuestStyleList() []GuestStyle {
	db := datasources.GetDatabase().Table("golf_fees")
	list := []GuestStyle{}
	status := item.ModelId.Status
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Group("guest_style")
	db.Find(&list)

	return list
}
