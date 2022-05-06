package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Booking
type Booking struct {
	models.Model
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf

	CreatedDate string `json:"created_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022

	Bag            string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	Hole           int    `json:"hole"`                                       // Số hố
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(200);index"` // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Guest Style Name

	CardId        string `json:"card_id" gorm:"index"`                           // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // MemberCard Uid, Uid object trong Database
	CustomerName  string `json:"customer_name" gorm:"type:varchar(256)"`         // Tên khách hàng
	CustomerUid   string `json:"customer_uid" gorm:"type:varchar(256);index"`    // Uid khách hàng
	// Thêm customer info

	CheckInOutStatus string `json:"check_in_out_status" gorm:"type:varchar(50);index"` // Time Check In Out status
	CheckInTime      int64  `json:"check_in_time"`                                     // Time Check In
	CheckOutTime     int64  `json:"check_out_time"`                                    // Time Check Out
	TeeType          string `json:"tee_type" gorm:"type:varchar(50);index"`            // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath          string `json:"tee_path" gorm:"type:varchar(50);index"`            // MORNING, NOON, NIGHT
	TurnTime         string `json:"turn_time" gorm:"type:varchar(30)"`                 // Ex: 16:26
	TeeTime          string `json:"tee_time" gorm:"type:varchar(30)"`                  // Ex: 16:26 Tee time là thời gian tee off dự kiến
	TeeOffTime       string `json:"tee_off_time" gorm:"type:varchar(30)"`              // Ex: 16:26 Là thời gian thực tế phát bóng
	RowIndex         int    `json:"row_index"`                                         // index trong Flight

	CurrentBagPrice  BookingCurrentBagPriceDetail  `json:"current_bag_price" gorm:"type:varchar(500)"`   // Thông tin phí++: Tính toán lại phí Service items, Tiền cho Subbag
	ListGolfFee      ListBookingGolfFee            `json:"list_golf_fee" gorm:"type:varchar(200)"`       // Thông tin List Golf Fee, Main Bag, Sub Bag
	ListServiceItems utils.ListBookingServiceItems `json:"list_service_items" gorm:"type:varchar(1000)"` // List service item: rental, proshop, restaurant, kiosk
	MushPayInfo      BookingMushPay                `json:"mush_pay_info" gorm:"type:varchar(200)"`       // Mush Pay info
	Rounds           ListBookingRound              `json:"rounds" gorm:"type:varchar(500)"`              // List Rounds: Sẽ sinh golf Fee với List GolfFee

	Note   string `json:"note" gorm:"type:varchar(500)"`   // Note
	Locker string `json:"locker" gorm:"type:varchar(100)"` // Locker mã số tủ gửi đồ

	CmsUser    string `json:"cms_user" gorm:"type:varchar(100)"`     // Cms User
	CmsUserLog string `json:"cms_user_log" gorm:"type:varchar(200)"` // Cms User Log

	// TODO
	// Caddie Info
	CaddieId int64 `json:"caddie_id" gorm:"index"`

	// Buggy Info
	BuggyId int64 `json:"buggy_id" gorm:"index"`

	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags" gorm:"type:varchar(1000)"` // List Sub Bags

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags" gorm:"type:varchar(200)"` // List Main Bags, thêm main bag sẽ thanh toán những cái gì
	// Main bug for Pay: Mặc định thanh toán all, Nếu có trong list này thì k thanh toán
	MainBagNoPay utils.ListString `json:"main_bag_no_pay" gorm:"type:varchar(100)"` // Main Bag không thanh toán những phần này
	InitType     string           `json:"init_type" gorm:"type:varchar(50);index"`  // BOOKING: Tạo booking xong checkin, CHECKIN: Check In xong tạo Booking luôn
}

// Booking Mush Pay
type BookingMushPay struct {
	MushPay          int64 `json:"mush_pay"`
	TotalGolfFee     int64 `json:"total_golf_fee"`
	TotalServiceItem int64 `json:"total_service_item"`
}

func (item *BookingMushPay) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingMushPay) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Booking GolfFee
type BookingGolfFee struct {
	BookingUid string `json:"booking_uid"`
	PlayerName string `json:"player_name"`
	Bag        string `json:"bag"`
	CaddieFee  int64  `json:"caddie_fee"`
	BuggyFee   int64  `json:"buggy_fee"`
	GreenFee   int64  `json:"green_fee"`
}

type ListBookingGolfFee []BookingGolfFee

func (item *ListBookingGolfFee) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingGolfFee) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Current Bag Price info
type BookingCurrentBagPriceDetail struct {
	Transfer   int64 `json:"transfer"`
	Debit      int64 `json:"debit"`
	GolfFee    int64 `json:"golf_fee"`
	Restaurant int64 `json:"restaurant"`
	Kiosk      int64 `json:"kiosk"`
	Rental     int64 `json:"rental"`
	Proshop    int64 `json:"proshop"`
	Promotion  int64 `json:"promotion"`
	Amount     int64 `json:"amount"`
}

func (item *BookingCurrentBagPriceDetail) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingCurrentBagPriceDetail) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Booking Round
type BookingRound struct {
	Index         int    `json:"index"`
	CaddieFee     int64  `json:"caddie_fee"`
	BuggyFee      int64  `json:"buggy_fee"`
	GreenFee      int64  `json:"green_fee"`
	Hole          int    `json:"hole"`
	GuestStyle    string `json:"guest_style"` // Nếu là member Card thì lấy guest style của member Card, nếu không thì lấy guest style Của booking đó
	MemberCardId  string `json:"member_card_id"`
	MemberCardUid string `json:"member_card_uid"`
	Pax           int    `json:"pax"`
	TeeOffTime    int64  `json:"tee_off_time"`
}

type ListBookingRound []BookingRound

func (item *ListBookingRound) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingRound) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// -------- Booking Logic --------
func (item *Booking) GetCurrentBagGolfFee() BookingGolfFee {
	golfFee := BookingGolfFee{}
	if item.ListGolfFee == nil {
		return golfFee
	}
	if len(item.ListGolfFee) <= 0 {
		return golfFee
	}

	return item.ListGolfFee[0]
}

func (item *Booking) GetTotalServicesFee() int64 {
	total := int64(0)
	if item.ListServiceItems != nil {
		for _, v := range item.ListServiceItems {
			temp := v.Amount
			total += temp
		}
	}

	return total
}

func (item *Booking) GetTotalGolfFee() int64 {
	total := int64(0)
	if item.ListGolfFee != nil {
		for _, v := range item.ListGolfFee {
			golfFeeTemp := v.BuggyFee + v.CaddieFee + v.GreenFee
			total += golfFeeTemp
		}
	}

	return total
}

func (item *Booking) AddRound(memberCardUid string, golfFee models.GolfFee) error {
	lengthRound := len(item.Rounds)

	if memberCardUid == "" {
		// Guest

	}

	// Member
	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUid
	errFind := memberCard.FindFirst()
	if errFind != nil {
		return errFind
	}

	bookingRound := BookingRound{
		Index:         lengthRound + 1,
		Hole:          item.Hole,
		Pax:           1,
		MemberCardId:  memberCard.CardId,
		MemberCardUid: memberCardUid,
		TeeOffTime:    time.Now().Unix(),
	}
	bookingRound.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, bookingRound.Hole)
	bookingRound.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, bookingRound.Hole)
	bookingRound.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, bookingRound.Hole)

	item.Rounds = append(item.Rounds, bookingRound)

	return nil
}

func (item *Booking) UpdateBagGolfFee() {
	if len(item.ListGolfFee) > 0 {
		item.ListGolfFee[0].Bag = item.Bag
	}
}

// Udp MushPay
func (item *Booking) UpdateMushPay() {
	mushPay := BookingMushPay{}

	totalGolfFee := int64(0)
	for _, v := range item.ListGolfFee {
		totalGolfFee += (v.BuggyFee + v.CaddieFee + v.GreenFee)
	}
	mushPay.TotalGolfFee = totalGolfFee

	// SubBag

	// Sub Service Item của current Bag
	for _, v := range item.ListServiceItems {
		isNeedPay := true
		if len(item.MainBagNoPay) > 0 {
			for _, v1 := range item.MainBagNoPay {
				// TODO: Tính Fee cho sub bag fee
				if v1 == constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS {
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND {
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
				} else {
					if v1 == v.Type {
						isNeedPay = false
					}
				}
			}
		}
		if isNeedPay {
			mushPay.TotalServiceItem += v.Amount
		}
	}

	item.MushPayInfo = mushPay
}

// Udp lại giá cho Booking
// Udp sub bag price
func (item *Booking) UpdatePriceDetailCurrentBag() {
	priceDetail := BookingCurrentBagPriceDetail{}

	if len(item.ListGolfFee) > 0 {
		priceDetail.GolfFee = item.ListGolfFee[0].BuggyFee + item.ListGolfFee[0].CaddieFee + item.ListGolfFee[0].GreenFee
	}

	for _, serviceItem := range item.ListServiceItems {
		if serviceItem.BookingUid == item.Uid {
			// Udp service detail cho booking uid
			if serviceItem.Type == constants.GOLF_SERVICE_RENTAL {
				priceDetail.Rental += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_PROSHOP {
				priceDetail.Proshop += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_RESTAURANT {
				priceDetail.Restaurant += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_KIOSK {
				priceDetail.Kiosk += serviceItem.Amount
			}
		}
	}

	priceDetail.Amount = priceDetail.Transfer +
		priceDetail.Debit +
		priceDetail.GolfFee +
		priceDetail.Restaurant +
		priceDetail.Kiosk +
		priceDetail.Rental +
		priceDetail.Proshop +
		priceDetail.Promotion

	item.CurrentBagPrice = priceDetail
}

// Check Duplicated
func (item *Booking) IsDuplicated(checkTeeTime, checkBag bool) (bool, error) {
	//Check turn time đã tồn tại
	if checkTeeTime {
		booking := Booking{
			PartnerUid:  item.PartnerUid,
			CourseUid:   item.CourseUid,
			TeeTime:     item.TeeTime,
			TurnTime:    item.TurnTime,
			CreatedDate: item.CreatedDate,
		}

		errFind := booking.FindFirst()
		if errFind == nil || booking.Uid != "" {
			return true, errors.New("Duplicated TeeTime")
		}
	}

	//Check Bag đã tồn tại
	if checkBag {
		if item.Bag != "" {
			booking := Booking{
				PartnerUid:  item.PartnerUid,
				CourseUid:   item.CourseUid,
				CreatedDate: item.CreatedDate,
				Bag:         item.Bag,
			}
			errBagFind := booking.FindFirst()
			if errBagFind == nil || booking.Uid != "" {
				return true, errors.New("Duplicated Bag")
			}
		}
	}

	return false, nil
}

// ----------- CRUD ------------
func (item *Booking) Create(uid string) error {
	item.Model.Uid = uid
	now := time.Now()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Booking) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Booking) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Booking) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Booking) FindList(page models.Page) ([]Booking, int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	list := []Booking{}
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

func (item *Booking) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
