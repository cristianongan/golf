package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Thẻ thành viên
type MemberCard struct {
	Model
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	OwnerUid        string `json:"owner_uid" gorm:"type:varchar(100);index"`   // Uid chủ sở hữu
	CardId          string `json:"card_id" gorm:"type:varchar(100);index"`     // Id thẻ
	Type            string `json:"type" gorm:"type:varchar(100);index"`        // Loại thẻ
	McType          string `json:"mc_type" gorm:"type:varchar(100);index"`     // Member Card Type = Member Type
	ValidDate       int64  `json:"valid_date" gorm:"index"`                    // Hieu luc tu ngay
	ExpDate         int64  `json:"exp_date" gorm:"index"`                      // Het hieu luc tu ngay
	ChipCode        string `json:"chip_code" gorm:"type:varchar(200)"`         // Sân tập cho bán chip, là mã thẻ đọc bằng máy đọc thẻ
	Note            string `json:"note" gorm:"type:varchar(500)"`              // Ghi chu them
	Locker          string `json:"locker" gorm:"type:varchar(100)"`            // Mã số tủ gửi đồ
	AdjustPlayCount int    `json:"adjust_play_count"`                          // Trước đó đã chơi bao nhiêu lần

	PriceCode int64 `json:"price_code"` // Giá
	GreenFee  int64 `json:"green_fee"`  // Phí sân cỏ
	CaddieFee int64 `json:"caddie_fee"` // Phí caddie
	BuggyFee  int64 `json:"buggy_fee"`  // Phí Buggy

	StartPrecial int64 `json:"start_precial"` // Khoảng TG được dùng giá riêng
	EndPrecial   int64 `json:"end_precial"`   // Khoảng TG được dùng giá riêng
}

func (item *MemberCard) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *MemberCard) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *MemberCard) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *MemberCard) Count() (int64, error) {
	db := datasources.GetDatabase().Model(MemberCard{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *MemberCard) FindList(page Page) ([]MemberCard, int64, error) {
	db := datasources.GetDatabase().Model(MemberCard{})
	list := []MemberCard{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
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

func (item *MemberCard) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}