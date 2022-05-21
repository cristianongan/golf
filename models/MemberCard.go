package models

import (
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Thẻ thành viên
type MemberCard struct {
	Model
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	OwnerUid   string `json:"owner_uid" gorm:"type:varchar(100);index"`   // Uid chủ sở hữu
	CardId     string `json:"card_id" gorm:"type:varchar(100);index"`     // Id thẻ
	// Type            string `json:"type" gorm:"type:varchar(100);index"`        // Loại thẻ - > Lấy từ MemberCardType.Type = Base Type
	// McType          string `json:"mc_type" gorm:"type:varchar(100);index"` // Member Card Type = Member Type
	McTypeId        int64  `json:"mc_type_id" gorm:"index"`            // Member Card Type id
	ValidDate       int64  `json:"valid_date" gorm:"index"`            // Hieu luc tu ngay
	ExpDate         int64  `json:"exp_date" gorm:"index"`              // Het hieu luc tu ngay
	ChipCode        string `json:"chip_code" gorm:"type:varchar(200)"` // Sân tập cho bán chip, là mã thẻ đọc bằng máy đọc thẻ
	Note            string `json:"note" gorm:"type:varchar(500)"`      // Ghi chu them
	Locker          string `json:"locker" gorm:"type:varchar(100)"`    // Mã số tủ gửi đồ
	AdjustPlayCount int    `json:"adjust_play_count"`                  // Trước đó đã chơi bao nhiêu lần

	PriceCode int64 `json:"price_code"` // 0|1 Check cái này có thì tính theo giá riêng -> theo cuộc họp suggest nên bỏ - Ko bỏ dc
	GreenFee  int64 `json:"green_fee"`  // Phí sân cỏ
	CaddieFee int64 `json:"caddie_fee"` // Phí caddie
	BuggyFee  int64 `json:"buggy_fee"`  // Phí Buggy

	StartPrecial int64 `json:"start_precial"` // Khoảng TG được dùng giá riêng
	EndPrecial   int64 `json:"end_precial"`   // Khoảng TG được dùng giá riêng
}

type MemberCardDetailRes struct {
	Model
	PartnerUid      string `json:"partner_uid"`       // Hang Golf
	CourseUid       string `json:"course_uid"`        // San Golf
	OwnerUid        string `json:"owner_uid"`         // Uid chủ sở hữu
	CardId          string `json:"card_id"`           // Id thẻ
	Type            string `json:"type"`              // Loại thẻ - > Lấy từ MemberCardType.Type = Base Type
	McType          string `json:"mc_type"`           // Member Card Type = Member Type
	McTypeId        int64  `json:"mc_type_id"`        // Member Card Type id
	ValidDate       int64  `json:"valid_date"`        // Hieu luc tu ngay
	ExpDate         int64  `json:"exp_date"`          // Het hieu luc tu ngay
	ChipCode        string `json:"chip_code"`         // Sân tập cho bán chip, là mã thẻ đọc bằng máy đọc thẻ
	Note            string `json:"note"`              // Ghi chu them
	Locker          string `json:"locker"`            // Mã số tủ gửi đồ
	AdjustPlayCount int    `json:"adjust_play_count"` // Trước đó đã chơi bao nhiêu lần

	PriceCode int64 `json:"price_code"` // Check cái này có thì tính theo giá riêng -> theo cuộc họp suggest nên bỏ - Ko bỏ dc
	GreenFee  int64 `json:"green_fee"`  // Phí sân cỏ
	CaddieFee int64 `json:"caddie_fee"` // Phí caddie
	BuggyFee  int64 `json:"buggy_fee"`  // Phí Buggy

	StartPrecial int64 `json:"start_precial"` // Khoảng TG được dùng giá riêng
	EndPrecial   int64 `json:"end_precial"`   // Khoảng TG được dùng giá riêng

	//MemberCardType Info
	MemberCardTypeInfo MemberCardType `json:"member_card_type_info"`

	//Owner Info
	OwnerInfo CustomerUser `json:"owner_info"`
}

// Find member card detail with info card type and owner
func (item *MemberCard) FindDetail() (MemberCardDetailRes, error) {
	memberCardDetailRes := MemberCardDetailRes{}
	memberCardByte, err := json.Marshal(item)
	if err != nil {
		return memberCardDetailRes, err
	}

	errUnM := json.Unmarshal(memberCardByte, &memberCardDetailRes)
	if errUnM != nil {
		return memberCardDetailRes, errUnM
	}

	//Find MemberCardType
	memberCardType := MemberCardType{}
	memberCardType.Id = item.McTypeId
	errFind := memberCardType.FindFirst()
	if errFind != nil {
		log.Println("FindDetail errFind ", errFind.Error())
	}

	//Find Owner
	owner := CustomerUser{}
	owner.Uid = item.OwnerUid
	errFind1 := owner.FindFirst()
	if errFind1 != nil {
		log.Println("FindDetail errFind1", errFind1.Error())
	}

	memberCardDetailRes.MemberCardTypeInfo = memberCardType
	memberCardDetailRes.OwnerInfo = owner

	return memberCardDetailRes, nil
}

func (item *MemberCard) IsValidated() bool {
	if item.CardId == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.OwnerUid == "" {
		return false
	}
	if item.McTypeId <= 0 {
		return false
	}
	return true
}

func (item *MemberCard) IsDuplicated() bool {
	memberCard := MemberCard{
		CardId:   item.CardId,
		McTypeId: item.McTypeId,
	}
	//Check Duplicated
	errFind := memberCard.FindFirst()
	if errFind == nil || memberCard.Uid != "" {
		return true
	}
	return false
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

func (item *MemberCard) FindList(page Page, playerName string) ([]map[string]interface{}, int64, error) {
	db := datasources.GetDatabase().Table("member_cards")
	list := []map[string]interface{}{}
	total := int64(0)

	queryStr := `select * from (select tb0.*, 
	member_card_types.name as mc_types_name,
	member_card_types.type as base_type,
	member_card_types.guest_style as guest_style,
	customer_users.name as owner_name,
	customer_users.email as owner_email,
	customer_users.address1 as owner_address1,
	customer_users.address2 as owner_address2,
	customer_users.phone as owner_phone,
	customer_users.dob as owner_dob,
	customer_users.sex as owner_sex,
	customer_users.job as owner_job,
	customer_users.position as owner_position,
	customer_users.identify as owner_identify,
	customer_users.company_id as owner_company_id,
	customer_users.company_name as owner_company_name
	from (select * from member_cards WHERE member_cards.partner_uid = ` + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and member_cards.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.OwnerUid != "" {
		queryStr = queryStr + " and member_cards.owner_uid = " + `"` + item.OwnerUid + `"`
	}
	if item.Status != "" {
		queryStr = queryStr + " and member_cards.status = " + `"` + item.Status + `"`
	}
	if item.CardId != "" {
		queryStr = queryStr + " and member_cards.card_id = " + `"` + item.CardId + `"`
	}
	if item.McTypeId > 0 {
		queryStr = queryStr + " and member_cards.mc_type_id = " + strconv.Itoa(int(item.McTypeId))
	}

	queryStr = queryStr + ") tb0 "
	queryStr = queryStr + `LEFT JOIN member_card_types on tb0.mc_type_id = member_card_types.id
	LEFT JOIN customer_users on tb0.owner_uid = customer_users.uid) tb1 `

	if playerName != "" {
		queryStr = queryStr + " where "
		queryStr = queryStr + " tb1.owner_name LIKE " + `%` + playerName + `%`
	}

	// var countReturn CountStruct
	var countReturn utils.CountStruct
	strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	if errCount != nil {
		log.Println("Membercard err", errCount.Error())
		return list, total, errCount
	}

	total = countReturn.Count
	//Check if limit large then set to 50
	if page.Limit > 50 {
		page.Limit = 50
	}

	if total > 0 && int64(page.Offset()) < total {
		queryStr = queryStr + " order by tb1." + page.SortBy + " " + page.SortDir + " LIMIT " + strconv.Itoa(page.Limit) + " OFFSET " + strconv.Itoa(page.Offset())
	}
	err := db.Raw(queryStr).Scan(&list).Error
	if err != nil {
		return list, total, err
	}

	return list, total, db.Error
}

func (item *MemberCard) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *MemberCard) GetOwner() (CustomerUser, error) {
	cusUser := CustomerUser{}
	if item.OwnerUid == "" {
		return cusUser, errors.New("Customer uid invalid")
	}
	cusUser.Uid = item.OwnerUid
	errFind := cusUser.FindFirst()
	if errFind != nil {
		return cusUser, errFind
	}
	return cusUser, nil
}
