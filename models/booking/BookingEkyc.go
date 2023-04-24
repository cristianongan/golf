package model_booking

import (
	"encoding/json"
	"log"
	"start/models"
)

// Booking
// omitempty: xứ lý khi các field trả về rỗng
type BookingEkycRes struct {
	models.Model
	PartnerUid  string `json:"partner_uid"`  // Hang Golf
	CourseUid   string `json:"course_uid"`   // San Golf
	BookingDate string `json:"booking_date"` // Ex: 06/11/2022

	Bag            string `json:"bag"`              // Golf Bag
	Hole           int    `json:"hole"`             // Số hố check in
	HoleBooking    int    `json:"hole_booking"`     // Số hố khi booking
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name

	// MemberCard
	CardId        string `json:"card_id"`         // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid"` // MemberCard Uid, Uid object trong Database

	// Thêm customer info
	CustomerName string `json:"customer_name"` // Tên khách hàng

	TeeType string `json:"tee_type"` // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath string `json:"tee_path"` // MORNING, NOON, NIGHT
	TeeTime string `json:"tee_time"` // Ex: 16:26 Tee time là thời gian tee off dự kiến

	MushPayInfo BookingMushPay `json:"mush_pay_info,omitempty"` // Mush Pay info

	//Qr code
	CheckInCode string `json:"check_in_code"`

	ListGolfFee      ListBookingGolfFee   `json:"list_golf_fee,omitempty"`      // Thông tin List Golf Fee, Main Bag, Sub Bag
	ListServiceItems []BookingServiceItem `json:"list_service_items,omitempty"` // List service item: rental, proshop, restaurant, kiosk

}

func (item *Booking) CloneBookingEkyc() BookingEkycRes {
	bookingEkyc := BookingEkycRes{}
	bData, errM := json.Marshal(&item)
	if errM != nil {
		log.Println("CloneBookingEkyc errM", errM.Error())
	}
	errUnM := json.Unmarshal(bData, &bookingEkyc)
	if errUnM != nil {
		log.Println("CloneBookingEkyc errUnM", errUnM.Error())
	}

	return bookingEkyc
}
