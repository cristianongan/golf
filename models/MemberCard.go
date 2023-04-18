package models

import (
	"encoding/json"
	"log"
	"start/constants"
	"start/utils"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Thẻ thành viên
type MemberCard struct {
	Model
	PartnerUid      string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid       string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	OwnerUid        string `json:"owner_uid" gorm:"type:varchar(100);index"`   // Uid chủ sở hữu
	CardId          string `json:"card_id" gorm:"type:varchar(100);index"`     // Id thẻ
	McTypeId        int64  `json:"mc_type_id" gorm:"index"`                    // Member Card Type id
	ValidDate       int64  `json:"valid_date" gorm:"index"`                    // Hieu luc tu ngay
	ExpDate         int64  `json:"exp_date" gorm:"index"`                      // Het hieu luc tu ngay
	ChipCode        string `json:"chip_code" gorm:"type:varchar(200)"`         // Sân tập cho bán chip, là mã thẻ đọc bằng máy đọc thẻ
	Note            string `json:"note" gorm:"type:varchar(500)"`              // Ghi chu them
	ReasonUnactive  string `json:"reason_unactive" gorm:"type:varchar(500)"`   // Ghi chu khi Unactive
	Locker          string `json:"locker" gorm:"type:varchar(100)"`            // Mã số tủ gửi đồ
	AdjustPlayCount int    `json:"adjust_play_count" gorm:"type:varchar(100)"` // Trước đó đã chơi bao nhiêu lần
	Float           int64  `json:"float"`                                      // Thẻ không định danh
	PromotionCode   string `json:"promotion_code" gorm:"type:varchar(100)"`    // mã giảm giá
	UserEdit        string `json:"user_edit" gorm:"type:varchar(150)"`         // user cập nhật

	// Company
	CompanyName string `json:"company_name" gorm:"type:varchar(200)"` // Ten cong ty
	CompanyId   int64  `json:"company_id" gorm:"index"`               // Id cong ty

	//
	AnnualType    string `json:"annual_type" gorm:"type:varchar(100)"`          // loại thường niên: UN_LIMITED (không giới hạn), LIMITED (chơi có giới hạn), SLEEP (thẻ ngủ).
	MemberConnect string `json:"member_connect" gorm:"type:varchar(100);index"` // member connect uid
	Relationship  string `json:"relationship" gorm:"type:varchar(100)"`         // Mối quan hệ của member: WIFE, HUSBAND, CHILD

	PriceCode int64 `json:"price_code"` // 0|1 Check cái này có thì tính theo giá riêng -> theo cuộc họp suggest nên bỏ - Ko bỏ dc
	GreenFee  int64 `json:"green_fee"`  // Phí sân cỏ
	CaddieFee int64 `json:"caddie_fee"` // Phí caddie
	BuggyFee  int64 `json:"buggy_fee"`  // Phí Buggy

	StartPrecial int64 `json:"start_precial"` // Khoảng TG được dùng giá riêng
	EndPrecial   int64 `json:"end_precial"`   // Khoảng TG được dùng giá riêng

	TotalGuestOfDay int `json:"total_guest_of_day"`            // Số khách đi cùng trong ngày
	TotalPlayOfYear int `json:"total_play_of_year"`            // Số lần đã chơi trong năm
	IsContacted     int `json:"is_contacted" gorm:"default:0"` // Đánh dấu đã liên hệ KH
}

type MemberCardDetailRes struct {
	Model
	PartnerUid      string `json:"partner_uid"`                             // Hang Golf
	CourseUid       string `json:"course_uid"`                              // San Golf
	OwnerUid        string `json:"owner_uid"`                               // Uid chủ sở hữu
	CardId          string `json:"card_id"`                                 // Id thẻ
	Type            string `json:"type"`                                    // Loại thẻ - > Lấy từ MemberCardType.Type = Base Type
	McType          string `json:"mc_type"`                                 // Member Card Type = Member Type
	McTypeId        int64  `json:"mc_type_id"`                              // Member Card Type id
	ValidDate       int64  `json:"valid_date"`                              // Hieu luc tu ngay
	ExpDate         int64  `json:"exp_date"`                                // Het hieu luc tu ngay
	ChipCode        string `json:"chip_code"`                               // Sân tập cho bán chip, là mã thẻ đọc bằng máy đọc thẻ
	Note            string `json:"note"`                                    // Ghi chu them
	Locker          string `json:"locker"`                                  // Mã số tủ gửi đồ
	AdjustPlayCount int    `json:"adjust_play_count"`                       // Trước đó đã chơi bao nhiêu lần
	Float           int64  `json:"float"`                                   // Thẻ không định danh
	PromotionCode   string `json:"promotion_code" gorm:"type:varchar(100)"` // mã giảm giá
	UserEdit        string `json:"user_edit" gorm:"type:varchar(150)"`      // user cập nhật

	// Company
	CompanyName string `json:"company_name" gorm:"type:varchar(200)"` // Ten cong ty
	CompanyId   int64  `json:"company_id" gorm:"index"`               // Id cong ty

	//
	AnnualType    string `json:"annual_type" gorm:"type:varchar(100)"`    // loại thường niên: UN_LIMITED (không giới hạn), LIMITED (chơi có giới hạn), SLEEP (thẻ ngủ).
	MemberConnect string `json:"member_connect" gorm:"type:varchar(250)"` // member uid
	Relationship  string `json:"relationship" gorm:"type:varchar(100)"`   // Mối quan hệ của member: WIFE, HUSBAND, CHILD

	PriceCode int64 `json:"price_code"` // Check cái này có thì tính theo giá riêng -> theo cuộc họp suggest nên bỏ - Ko bỏ dc
	GreenFee  int64 `json:"green_fee"`  // Phí sân cỏ
	CaddieFee int64 `json:"caddie_fee"` // Phí caddie
	BuggyFee  int64 `json:"buggy_fee"`  // Phí Buggy

	StartPrecial int64 `json:"start_precial"` // Khoảng TG được dùng giá riêng
	EndPrecial   int64 `json:"end_precial"`   // Khoảng TG được dùng giá riêng

	TotalGuestOfDay int `json:"total_guest_of_day"` // Số khách đi cùng trong ngày
	TotalPlayOfYear int `json:"total_play_of_year"` // Số lần đã chơi trong năm

	//MemberCardType Info
	MemberCardTypeInfo MemberCardType `json:"member_card_type_info"`

	//Owner Info
	OwnerInfo CustomerUser `json:"owner_info"`
}

/*
 Clone object
*/
func (item *MemberCard) CloneMemberCard() MemberCard {
	copyMemberCard := MemberCard{}
	bData, errM := json.Marshal(&item)
	if errM != nil {
		log.Println("CloneMemberCard errM", errM.Error())
	}
	errUnM := json.Unmarshal(bData, &copyMemberCard)
	if errUnM != nil {
		log.Println("CloneMemberCard errUnM", errUnM.Error())
	}

	return copyMemberCard
}

/*
Check time có thể sử dụng giá riêng
*/
func (item *MemberCard) IsValidTimePrecial() bool {

	currentTime := utils.GetTimeNow().Unix()

	if item.StartPrecial == 0 && item.EndPrecial == 0 {
		return true
	}

	if item.StartPrecial > 0 && item.EndPrecial == 0 {
		if item.StartPrecial > currentTime {
			return false
		}
		return true
	}

	if item.EndPrecial > 0 && item.StartPrecial == 0 {
		if item.EndPrecial < currentTime {
			return false
		}
		return true
	}

	if item.StartPrecial <= currentTime && currentTime <= item.EndPrecial {
		return true
	}

	return false
}

// Find member card detail with info card type and owner
func (item *MemberCard) FindDetail(db *gorm.DB) (MemberCardDetailRes, error) {
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
	errFind := memberCardType.FindFirst(db)
	if errFind != nil {
		log.Println("FindDetail errFind ", errFind.Error())
	}

	//Find Owner
	owner := CustomerUser{}
	owner.Uid = item.OwnerUid
	errFind1 := owner.FindFirst(db)
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

func (item *MemberCard) IsDuplicated(db *gorm.DB) bool {
	memberCard := MemberCard{
		CardId:   item.CardId,
		McTypeId: item.McTypeId,
	}
	//Check Duplicated
	errFind := memberCard.FindFirst(db)
	if errFind == nil || memberCard.Uid != "" {
		return true
	}
	return false
}

func (item *MemberCard) Create(db *gorm.DB) error {
	uid := uuid.New()
	now := utils.GetTimeNow()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *MemberCard) Update(db *gorm.DB) error {
	item.Model.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *MemberCard) FindAll(database *gorm.DB) (error, []MemberCard) {
	db := database.Model(MemberCard{})
	list := []MemberCard{}
	if item.OwnerUid != "" {
		db = db.Where("owner_uid = ?", item.OwnerUid)
	}
	db.Find(&list)
	return db.Error, list
}

func (item *MemberCard) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *MemberCard) FindFirstWithMemberCardType(db *gorm.DB) (error, error, MemberCardType) {
	err1 := db.Where(item).First(item).Error
	memberCardType := MemberCardType{}
	memberCardType.Id = item.McTypeId
	err2 := memberCardType.FindFirst(db)
	return err1, err2, memberCardType
}

func (item *MemberCard) Count(database *gorm.DB) (int64, error) {
	db := database.Model(MemberCard{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *MemberCard) FindList(database *gorm.DB, page Page, playerName string) ([]map[string]interface{}, int64, error) {
	db := database.Table("member_cards")
	list := []map[string]interface{}{}
	total := int64(0)
	currentYear := utils.GetCurrentYear()

	db = db.Select("member_cards.*, member_card_types.name as mc_types_name,member_card_types.type as base_type,member_card_types.guest_style as guest_style,member_card_types.guest_style_of_guest as guest_style_of_guest,member_card_types.play_time_on_year as play_time_on_year,customer_users.name as owner_name,customer_users.email as owner_email,customer_users.address2 as owner_address2,customer_users.phone as owner_phone,customer_users.dob as owner_dob,customer_users.sex as owner_sex,customer_users.job as owner_job,customer_users.position as owner_position,customer_users.identify as owner_identify,customer_users.company_id as owner_company_id,customer_users.company_name as owner_company_name,member_connect.name as member_connect_name,member_connect.email as member_connect_email,member_connect.address1 as member_connect_address1,member_connect.address2 as member_connect_address2,member_connect.phone as member_connect_phone,report_customer_plays.total_paid as report_total_paid,report_customer_plays.total_play_count as report_total_play_count,report_customer_plays.total_hour_play_count as report_total_hour_play_count,af.annual_quota_amount as annual_quota_amount,af.total_paid as total_paid,af.play_counts_add as play_counts_add")

	db = db.Joins("LEFT JOIN member_card_types on member_cards.mc_type_id = member_card_types.id")

	db = db.Joins("LEFT JOIN customer_users on member_cards.owner_uid = customer_users.uid")

	db = db.Joins("LEFT JOIN customer_users as member_connect on member_cards.member_connect = member_connect.uid")

	db = db.Joins("LEFT JOIN report_customer_plays on member_cards.card_id = report_customer_plays.card_id")

	db = db.Joins("LEFT JOIN (select * from annual_fees where partner_uid = ? and course_uid = ? and annual_fees.year = ?) af on member_cards.uid = af.member_card_uid", item.PartnerUid, item.CourseUid, currentYear)

	if item.CourseUid != "" {
		db = db.Where("member_cards.course_uid = ?", item.CourseUid)
	}
	if item.OwnerUid != "" {
		db = db.Where("member_cards.owner_uid = ?", item.OwnerUid)
	}
	if item.Status != "" {
		db = db.Where("member_cards.status = ?", item.Status)
	}
	if item.CardId != "" {
		db = db.Where("member_cards.card_id LIKE ?", "%"+item.CardId+"%")
	}
	if item.McTypeId > 0 {
		db = db.Where("member_cards.mc_type_id = ?", strconv.Itoa(int(item.McTypeId)))
	}
	if item.MemberConnect == constants.MEMBER_CONNECT_NONE {
		db = db.Where("member_cards.member_connect NOT LIKE ''")
	}
	if playerName != "" {
		db = db.Where("customer_users.name LIKE ? OR member_cards.card_id LIKE ?", "%"+playerName+"%", "%"+playerName+"%")
	}

	// queryStr := `select * from (select tb0.*,
	// member_card_types.name as mc_types_name,
	// member_card_types.type as base_type,
	// member_card_types.guest_style as guest_style,
	// member_card_types.guest_style_of_guest as guest_style_of_guest,
	// member_card_types.play_time_on_year as play_time_on_year,
	// customer_users.name as owner_name,
	// customer_users.email as owner_email,
	// customer_users.address1 as owner_address1,
	// customer_users.address2 as owner_address2,
	// customer_users.phone as owner_phone,
	// customer_users.dob as owner_dob,
	// customer_users.sex as owner_sex,
	// customer_users.job as owner_job,
	// customer_users.position as owner_position,
	// customer_users.identify as owner_identify,
	// customer_users.company_id as owner_company_id,
	// customer_users.company_name as owner_company_name,
	// member_connect.name as member_connect_name,
	// member_connect.email as member_connect_email,
	// member_connect.address1 as member_connect_address1,
	// member_connect.address2 as member_connect_address2,
	// member_connect.phone as member_connect_phone,
	// report_customer_plays.total_paid as report_total_paid,
	// report_customer_plays.total_play_count as report_total_play_count,
	// report_customer_plays.total_hour_play_count as report_total_hour_play_count,
	// af.annual_quota_amount as annual_quota_amount,
	// af.total_paid as total_paid,
	// af.play_counts_add as play_counts_add
	// from (select * from member_cards WHERE member_cards.partner_uid = ` + `"` + item.PartnerUid + `"`

	// if item.CourseUid != "" {
	// 	queryStr = queryStr + " and member_cards.course_uid = " + `"` + item.CourseUid + `"`
	// }
	// if item.OwnerUid != "" {
	// 	queryStr = queryStr + " and member_cards.owner_uid = " + `"` + item.OwnerUid + `"`
	// }
	// if item.Status != "" {
	// 	queryStr = queryStr + " and member_cards.status = " + `"` + item.Status + `"`
	// }
	// if item.CardId != "" {
	// 	queryStr = queryStr + " and member_cards.card_id LIKE " + `"%` + item.CardId + `%"`
	// }
	// if item.McTypeId > 0 {
	// 	queryStr = queryStr + " and member_cards.mc_type_id = " + strconv.Itoa(int(item.McTypeId))
	// }
	// if item.MemberConnect == constants.MEMBER_CONNECT_NONE {
	// 	queryStr = queryStr + " and member_cards.member_connect NOT LIKE ''"
	// }

	// queryStr = queryStr + ") tb0 "
	// queryStr = queryStr + `LEFT JOIN member_card_types on tb0.mc_type_id = member_card_types.id
	// LEFT JOIN customer_users on tb0.owner_uid = customer_users.uid `

	// queryStr = queryStr + `LEFT JOIN customer_users as member_connect on tb0.member_connect = member_connect.uid `

	// queryStr = queryStr + `LEFT JOIN report_customer_plays on tb0.owner_uid = report_customer_plays.customer_uid `

	// queryStr = queryStr + " LEFT JOIN (select * from annual_fees where annual_fees.partner_uid = " + `"` + item.PartnerUid + `"`
	// if item.CourseUid != "" {
	// 	queryStr = queryStr + " and annual_fees.course_uid = " + `"` + item.CourseUid + `"`
	// }
	// if currentYear != "" {
	// 	queryStr = queryStr + " and annual_fees.year = " + currentYear
	// }

	// queryStr = queryStr + ") af on tb0.uid = af.member_card_uid) tb1 "

	// if playerName != "" {
	// 	queryStr = queryStr + " where "
	// 	queryStr = queryStr + " tb1.owner_name LIKE " + `"%` + playerName + `%"`
	// 	queryStr = queryStr + " or tb1.card_id LIKE " + `"%` + playerName + `%"`
	// }

	// // var countReturn CountStruct
	// var countReturn utils.CountStruct
	// strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	// errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	// if errCount != nil {
	// 	log.Println("Membercard err", errCount.Error())
	// 	return list, total, errCount
	// }

	// total = countReturn.Count
	// //Check if limit large then set to 50
	// if page.Limit > 50 {
	// 	page.Limit = 50
	// }

	// if total > 0 && int64(page.Offset()) < total {
	// 	queryStr = queryStr + " order by tb1." + page.SortBy + " " + page.SortDir + " LIMIT " + strconv.Itoa(page.Limit) + " OFFSET " + strconv.Itoa(page.Offset())
	// }
	// err := db.Raw(queryStr).Scan(&list).Error
	// if err != nil {
	// 	return list, total, err
	// }

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *MemberCard) FindAllMemberCardContacted(database *gorm.DB) ([]MemberCard, int64, error) {
	db := database.Table("member_cards")
	list := []MemberCard{}
	total := int64(0)

	db = db.Where("is_contacted = ?", item.IsContacted)
	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}

// ------ Batch Update ------
func (item *MemberCard) BatchUpdate(database *gorm.DB, list []MemberCard) error {
	db := database.Table("member_cards")
	var err error
	err = db.Save(&list).Error

	if err != nil {
		log.Println("member_cards batch update err: ", err.Error())
	}
	return err
}

func (item *MemberCard) Delete(db *gorm.DB) error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *MemberCard) GetOwner(db *gorm.DB) (CustomerUser, error) {
	cusUser := CustomerUser{}
	if item.OwnerUid == "" {
		return cusUser, errors.New("Customer uid invalid")
	}
	cusUser.Uid = item.OwnerUid
	errFind := cusUser.FindFirst(db)
	if errFind != nil {
		return cusUser, errFind
	}
	return cusUser, nil
}

func (item *MemberCard) GetGuestStyle(db *gorm.DB) string {
	// Get Member card Type
	memberCardType := MemberCardType{}
	memberCardType.Id = item.McTypeId
	err := memberCardType.FindFirst(db)
	if err != nil {
		log.Println("Member card get guestStyle err", err.Error())
		return ""
	}
	return memberCardType.GuestStyle
}

func (item *MemberCard) FindListForEkyc(database *gorm.DB) ([]map[string]interface{}, error) {
	db := database.Table("member_cards")
	list := []map[string]interface{}{}
	// total := int64(0)

	db = db.Select("member_cards.uid as uid,member_cards.partner_uid as partner_uid,member_cards.course_uid as course_uid,member_cards.card_id as card_id,member_card_types.name as mc_types_name,member_card_types.guest_style as guest_style,customer_users.name as owner_name,customer_users.email as owner_email,customer_users.phone as owner_phone,customer_users.dob as owner_dob,customer_users.sex as owner_sex,customer_users.avatar as owner_avatar")

	db = db.Joins("LEFT JOIN member_card_types on member_cards.mc_type_id = member_card_types.id")

	db = db.Joins("LEFT JOIN customer_users on member_cards.owner_uid = customer_users.uid")

	if item.CourseUid != "" {
		db = db.Where("member_cards.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("member_cards.partner_uid = ?", item.PartnerUid)
	}
	if item.OwnerUid != "" {
		db = db.Where("member_cards.owner_uid = ?", item.OwnerUid)
	}
	if item.Status != "" {
		db = db.Where("member_cards.status = ?", item.Status)
	}

	// db.Count(&total)
	db = db.Find(&list)

	return list, db.Error
}
