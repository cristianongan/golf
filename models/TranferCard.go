package models

import (
	"log"
	"start/constants"
	"start/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Thẻ thành viên
type TranferCard struct {
	ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"`   // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`    // San Golf
	CardUid     string `json:"card_uid" gorm:"type:varchar(100);index"`      // Uid the
	CardId      string `json:"card_id" gorm:"type:varchar(100);index"`       // Id the
	OwnerUidOld string `json:"owner_uid_old" gorm:"type:varchar(100);index"` // Uid nguoi so huu the cu
	OwnerUid    string `json:"owner_uid" gorm:"type:varchar(100);index"`     // Uid nguoi so huu the moi
	BillNumber  string `json:"bill_number" gorm:"type:varchar(100)"`         // So bill
	BillDate    int64  `json:"bill_date"`                                    // ngay tao bill
	Amount      int64  `json:"amount"`                                       // phi chuyen doi the
	TranferDate int64  `json:"tranfer_date" gorm:"index"`                    // ngay tranfer card
	InputUser   string `json:"input_user" gorm:"type:varchar(200)"`          // nguoi nhap
	Note        string `json:"note" gorm:"type:varchar(500)"`                // Ghi chu them
}

func (item *TranferCard) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *TranferCard) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *TranferCard) Count(database *gorm.DB) (int64, error) {
	db := database.Model(MemberCardType{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TranferCard) FindList(database *gorm.DB,page Page, playerName string) ([]map[string]interface{}, int64, error) {
	db := database.Table("tranfer_cards")
	list := []map[string]interface{}{}
	total := int64(0)

	queryStr := `select * from (select tb0.*, 
	customer_users.name as owner_name,
	customer_users.email as owner_email,
	customer_users.address1 as owner_address,
	customer_users.phone as owner_phone,
	cso.name as owner_name_old,
	cso.email as owner_email_old,
	cso.address1 as owner_address_old,
	cso.phone as owner_phone_old
	from (select * from tranfer_cards WHERE tranfer_cards.partner_uid = ` + `"` + item.PartnerUid + `"`

	if item.CourseUid != "" {
		queryStr = queryStr + " and tranfer_cards.course_uid = " + `"` + item.CourseUid + `"`
	}
	if item.OwnerUid != "" {
		queryStr = queryStr + " and tranfer_cards.owner_uid = " + `"` + item.OwnerUid + `"`
	}
	if item.CardId != "" {
		queryStr = queryStr + " and tranfer_cards.card_id LIKE " + `"%` + item.CardId + `%"`
	}

	queryStr = queryStr + ") tb0 "
	queryStr = queryStr + `LEFT JOIN (select * from customer_users) cso on tb0.owner_uid_old = cso.uid
	LEFT JOIN customer_users on tb0.owner_uid = customer_users.uid `

	queryStr = queryStr + ") tb1 "

	if playerName != "" {
		queryStr = queryStr + " where "
		queryStr = queryStr + " tb1.owner_name LIKE " + `"%` + playerName + `%" `
		queryStr = queryStr + " or "
		queryStr = queryStr + " tb1.owner_name_old LIKE " + `"%` + playerName + `%" `
	}

	// var countReturn CountStruct
	var countReturn utils.CountStruct
	strSQLCount := " select count(*) as count from ( " + queryStr + " ) as subTable "
	errCount := db.Raw(strSQLCount).Scan(&countReturn).Error
	if errCount != nil {
		log.Println("TranferCard err", errCount.Error())
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
