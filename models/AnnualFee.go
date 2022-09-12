package models

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Phí thường niên
// TODO: Chú ý logic số tiền phải trả và số tiền trả từng đợt
type AnnualFee struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // Member Card Uid
	Year          int    `json:"year" gorm:"index"`                              // Year
	// PaymentType       string `json:"payment_type" gorm:"type:varchar(50);index"`     // TM, CK, CC, TM+CK, TM+CC
	// BillNumber        string `json:"bill_number" gorm:"type:varchar(100)"`           //
	Note              string `json:"note" gorm:"type:varchar(256)"` //
	AnnualQuotaAmount int64  `json:"annual_quota_amount"`           // Tiền Phí thuờng niên
	PrePaid           int64  `json:"pre_paid"`                      // A: Số tiền khách nộp trước khi chạy phần mềm
	PaidForfeit       int64  `json:"paid_forfeit"`                  // B: Số Tiền phạt do thanh toán chậm
	PaidReduce        int64  `json:"paid_reduce"`                   // C: Số Tiền giảm trừ khi nộp sớm
	LastYearDebit     int64  `json:"last_year_debit"`               // D: Số tiền nợ từ năm ngoái
	// MustPaid          int64  `json:"must_paid"`                                      // K: Số tiền Phí khách hàng đó pải đóng K = A-B+C-D+E
	TotalPaid int64 `json:"total_paid"` // G: Tổng số tiền các lần khách trả
	// Debit             int64  `json:"debit"`                                          // H: tiền nợ H = K - G
	// PlayCountsAdd int    `json:"play_counts_add"`                    // Bỏ, lấy từ adjust_play_count member card
	DaysPaid string `json:"days_paid" gorm:"type:varchar(256)"` // Ghi lại các ngày thanh toán của khách
}

func (item *AnnualFee) IsDuplicated() bool {
	modelCheck := AnnualFee{
		PartnerUid:    item.PartnerUid,
		CourseUid:     item.CourseUid,
		MemberCardUid: item.MemberCardUid,
		Year:          item.Year,
	}
	errFind := modelCheck.FindFirst()
	if errFind == nil || modelCheck.Id > 0 {
		return true
	}
	return false
}

func (item *AnnualFee) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.MemberCardUid == "" {
		return false
	}
	if item.Year <= 0 {
		return false
	}
	return true
}

func (item *AnnualFee) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *AnnualFee) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AnnualFee) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *AnnualFee) Count() (int64, error) {
	db := datasources.GetDatabase().Model(AnnualFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AnnualFee) FindListWithGroupMemberCard(page Page) ([]map[string]interface{}, int64, error) {
	db := datasources.GetDatabase().Table("annual_fees")
	list := []map[string]interface{}{}
	total := int64(0)

	queryStr := `select * from (select * from (select * from annual_fees where annual_fees.partner_uid = ` + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and annual_fees.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.MemberCardUid != "" {
		queryStr = queryStr + " and annual_fees.member_card_uid = " + `"` + item.MemberCardUid + `"`
	}

	queryStr = queryStr + " GROUP BY member_card_uid "

	queryStr = queryStr + ") tb0 "
	queryStr = queryStr + `LEFT JOIN (select tb1.*, 
		member_card_types.name as member_card_types_names, 
		member_card_types.type as base_type, 
		customer_users.name as owner_name,
		customer_users.email as owner_email,
		customer_users.address1 as owner_address1,
		customer_users.phone as owner_phone,
		customer_users.sex as owner_sex,
		customer_users.dob as owner_dob
		from (
		select member_cards.uid as mc_uid,  
		member_cards.valid_date as mc_valid_date, 
		member_cards.exp_date as mc_exp_date, 
		member_cards.owner_uid as owner_uid, 
		member_cards.mc_type_id as mc_type_id,
		member_cards.adjust_play_count as adjust_play_count
		from member_cards WHERE member_cards.partner_uid = `

	queryStr = queryStr + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and member_cards.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.MemberCardUid != "" {
		queryStr = queryStr + " and member_cards.uid = " + `"` + item.MemberCardUid + `"`
	}

	queryStr = queryStr + ") tb1 "
	queryStr = queryStr + `LEFT JOIN member_card_types on member_card_types.id = tb1.mc_type_id
	LEFT JOIN customer_users on customer_users.uid = tb1.owner_uid
	) tb2 on tb0.member_card_uid = tb2.mc_uid) tb3`

	// var countReturn CountStruct
	var countReturn utils.CountStruct
	strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	if errCount != nil {
		log.Println("AnnualFee err", errCount.Error())
		return list, total, errCount
	}

	total = countReturn.Count
	//Check if limit large then set to 50
	if page.Limit > 50 {
		page.Limit = 50
	}

	if total > 0 && int64(page.Offset()) < total {
		queryStr = queryStr + " order by tb3." + page.SortBy + " " + page.SortDir + " LIMIT " + strconv.Itoa(page.Limit) + " OFFSET " + strconv.Itoa(page.Offset())
	}
	err := db.Raw(queryStr).Scan(&list).Error
	if err != nil {
		return list, total, err
	}

	return list, total, db.Error
}

func (item *AnnualFee) FindList(page Page) ([]map[string]interface{}, utils.CountAnnualFeeStruct, int64, error) {
	db := datasources.GetDatabase().Table("annual_fees")
	list := []map[string]interface{}{}
	var countTotalAnnualFee utils.CountAnnualFeeStruct
	total := int64(0)

	queryStr := `select * from (select * from (select * from annual_fees where annual_fees.partner_uid = ` + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and annual_fees.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.MemberCardUid != "" {
		queryStr = queryStr + " and annual_fees.member_card_uid = " + `"` + item.MemberCardUid + `"`
	}
	if item.Year > 0 {
		queryStr = queryStr + " and annual_fees.year = " + strconv.Itoa(item.Year)
	}

	queryStr = queryStr + ") tb0 "
	queryStr = queryStr + `LEFT JOIN (select tb1.*, 
		member_card_types.name as member_card_types_names, 
		member_card_types.type as base_type, 
		customer_users.name as owner_name,
		customer_users.email as owner_email,
		customer_users.address1 as owner_address1,
		customer_users.phone as owner_phone,
		customer_users.sex as owner_sex,
		customer_users.dob as owner_dob
		from (
		select member_cards.uid as mc_uid,  
		member_cards.valid_date as mc_valid_date, 
		member_cards.exp_date as mc_exp_date, 
		member_cards.owner_uid as owner_uid, 
		member_cards.mc_type_id as mc_type_id,
		member_cards.card_id as mc_card_id,
		member_cards.adjust_play_count as adjust_play_count
		from member_cards WHERE member_cards.partner_uid = `

	queryStr = queryStr + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and member_cards.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.MemberCardUid != "" {
		queryStr = queryStr + " and member_cards.uid = " + `"` + item.MemberCardUid + `"`
	}

	queryStr = queryStr + ") tb1 "
	queryStr = queryStr + `LEFT JOIN member_card_types on member_card_types.id = tb1.mc_type_id
	LEFT JOIN customer_users on customer_users.uid = tb1.owner_uid
	) tb2 on tb0.member_card_uid = tb2.mc_uid) tb3`

	// var countReturn CountStruct
	var countReturn utils.CountStruct
	strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	if errCount != nil {
		log.Println("AnnualFee err", errCount.Error())
		return list, countTotalAnnualFee, total, errCount
	}

	// Sum Total
	strSQLCountTotalAnnualFee := " select SUM(annual_quota_amount) as total_a, SUM(pre_paid) as total_b, SUM(paid_forfeit) as total_c, SUM(paid_reduce) as total_d, SUM(last_year_debit) as total_e, SUM(total_paid) as total_g from ( " + queryStr + " ) as subTable "
	errCountTotalAnnualFee := db.Raw(strSQLCountTotalAnnualFee).Scan(&countTotalAnnualFee).Error
	if errCountTotalAnnualFee != nil {
		log.Println("AnnualFee errCountTotalAnnualFee", errCountTotalAnnualFee.Error())
		return list, countTotalAnnualFee, total, errCount
	}

	total = countReturn.Count
	//Check if limit large then set to 50
	if page.Limit > 50 {
		page.Limit = 50
	}

	if total > 0 && int64(page.Offset()) < total {
		queryStr = queryStr + " order by tb3." + page.SortBy + " " + page.SortDir + " LIMIT " + strconv.Itoa(page.Limit) + " OFFSET " + strconv.Itoa(page.Offset())
	}
	err := db.Raw(queryStr).Scan(&list).Error
	if err != nil {
		return list, countTotalAnnualFee, total, err
	}

	return list, countTotalAnnualFee, total, db.Error
}

func (item *AnnualFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
