package models

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Phí thường niên
// TODO: Chú ý logic số tiền phải trả và số tiền trả từng đợt
type AnnualFee struct {
	ModelId
	PartnerUid        string `json:"partner_uid" gorm:"type:varchar(100);index"`     // Hang Golf
	CourseUid         string `json:"course_uid" gorm:"type:varchar(256);index"`      // San Golf
	MemberCardUid     string `json:"member_card_uid" gorm:"type:varchar(100);index"` // Member Card Uid
	Year              int    `json:"year" gorm:"index"`                              // Year
	PaymentType       string `json:"payment_type" gorm:"type:varchar(50);index"`     // TM, CK, CC, TM+CK, TM+CC
	BillNumber        string `json:"bill_number" gorm:"type:varchar(100)"`           //
	Note              string `json:"note" gorm:"type:varchar(256)"`                  //
	AnnualQuotaAmount int64  `json:"annual_quota_amount"`                            // Tiền Phí thuờng niên
	PrePaid           int64  `json:"pre_paid"`                                       // A: Số tiền khách nộp trước khi chạy phần mềm
	PaidForfeit       int64  `json:"paid_forfeit"`                                   // B: Số Tiền phạt do thanh toán chậm
	PaidReduce        int64  `json:"paid_reduce"`                                    // C: Số Tiền giảm trừ khi nộp sớm
	LastYearDebit     int64  `json:"last_year_debit"`                                // D: Số tiền nợ từ năm ngoái
	// MustPaid          int64  `json:"must_paid"`                                      // K: Số tiền Phí khách hàng đó pải đóng K = A-B+C-D+E
	TotalPaid int64 `json:"total_paid"` // G: Tổng số tiền các lần khách trả
	// Debit             int64  `json:"debit"`                                          // H: tiền nợ H = K - G
	PlayCountsAdd int    `json:"play_counts_add"`                    //
	DaysPaid      string `json:"days_paid" gorm:"type:varchar(256)"` // Ghi lại các ngày thanh toán của khách
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

func (item *AnnualFee) FindListWithGroupMemberCard(page Page) ([]AnnualFee, int64, error) {
	db := datasources.GetDatabase().Model(AnnualFee{})
	list := []AnnualFee{}
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
	if item.MemberCardUid != "" {
		db = db.Where("member_card_uid = ?", item.MemberCardUid)
	}
	if item.Year > 0 {
		db = db.Where("year = ?", item.Year)
	}
	db = db.Group("member_card_uid")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *AnnualFee) FindList(page Page) ([]map[string]interface{}, int64, error) {
	db := datasources.GetDatabase().Table("annual_fees")
	list := []map[string]interface{}{}
	total := int64(0)
	// status := item.ModelId.Status
	// item.ModelId.Status = ""
	// db = db.Where(item)
	// if status != "" {
	// 	db = db.Where("status in (?)", strings.Split(status, ","))
	// }

	// if item.PartnerUid != "" {
	// 	db = db.Where("partner_uid = ?", item.PartnerUid)
	// }
	// if item.CourseUid != "" {
	// 	db = db.Where("course_uid = ?", item.CourseUid)
	// }
	// if item.MemberCardUid != "" {
	// 	db = db.Where("member_card_uid = ?", item.MemberCardUid)
	// }
	// if item.Year > 0 {
	// 	db = db.Where("year = ?", item.Year)
	// }

	queryStr := `select * from (select * from (select * from annual_fees where annual_fees.partner_uid = "FLC" and annual_fees.course_uid = "FLC-HA-LONG") tb0
	LEFT JOIN (select tb1.*, 
	member_card_types.name as member_card_types_names, 
	member_card_types.type as base_type, 
	customer_users.name as owner_name,
	customer_users.email as owner_email,
	customer_users.address1 as owner_address1,
	customer_users.phone as owner_phone
	from (
	select member_cards.uid as mc_uid,  
	member_cards.valid_date as mc_valid_date, 
	member_cards.exp_date as mc_exp_date, 
	member_cards.owner_uid as owner_uid, 
	member_cards.mc_type_id as mc_type_id
	from member_cards WHERE member_cards.partner_uid = "FLC" and member_cards.course_uid = "FLC-HA-LONG") tb1 
	LEFT JOIN member_card_types on member_card_types.id = tb1.mc_type_id
	LEFT JOIN customer_users on customer_users.uid = tb1.owner_uid
	) tb2 on tb0.member_card_uid = tb2.mc_uid) tb3`

	// var countReturn CountStruct
	var countReturn utils.CountStruct
	strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	if errCount != nil {
		log.Println(errCount)
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

func (item *AnnualFee) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
