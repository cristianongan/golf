package controllers

import (
	base64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"start/config"
	model_booking "start/models/booking"
	"start/services"
	"strconv"

	"github.com/ttacon/libphonenumber"
)

type QrCodeUrlModel struct {
	QrImg       string `json:"qr_img"`
	Date        string `json:"date"`
	CheckInCode string `json:"check_in_code"`
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
}

/*
 Gen QR URL -> send sms
*/
func genQRCodeListBook(listBooking []model_booking.Booking) {
	listHaveQRURL := []model_booking.Booking{}
	for _, v := range listBooking {
		genQrCodeForBooking(&v)
		listHaveQRURL = append(listHaveQRURL, v)
	}
	sendSmsBooking(listHaveQRURL)
}

/*
 Send sms
*/
func sendSmsBooking(listBooking []model_booking.Booking) error {

	if len(listBooking) == 0 {
		errEmpty := errors.New("sendSmsBooking Err List Booking Emplty")
		log.Println("sendSmsBooking errEmpty", errEmpty.Error())
		return errEmpty
	}

	// parse standard phone number
	num, errPhone := libphonenumber.Parse(listBooking[0].CustomerBookingPhone, "VN")
	if errPhone != nil {
		log.Println("sendSmsBooking errPhone:", errPhone)

		return errPhone
	}

	message := "Sân " + listBooking[0].CourseUid + " xác nhận đặt chỗ ngày " + listBooking[0].BookingDate + ": "

	for i, b := range listBooking {
		iStr := strconv.Itoa(i)
		message += iStr + ". Player " + iStr + ": "
		playerName := ""
		if b.MemberCard != nil {
			playerName = b.MemberCard.CardId
		}
		if playerName == "" {
			playerName = b.CustomerInfo.Name
		}

		message += playerName + " - " + "Mã check-in: " + b.CheckInCode + " - QR: "

		qrCodeUrlModel := QrCodeUrlModel{
			QrImg:       b.QrcodeUrl,
			CheckInCode: b.CheckInCode,
			Date:        b.BookingDate,
			PartnerUid:  b.PartnerUid,
			CourseUid:   b.CourseUid,
		}

		byteQrCodeUrlModel, errMas := json.Marshal(&qrCodeUrlModel)

		if errMas != nil {
			log.Println("sendSmsBooking errMas", errMas.Error())
		}

		encodedurlQrCodeChecking := base64.StdEncoding.EncodeToString(byteQrCodeUrlModel)

		linkQRCodeFull := config.GetPortalCmsUrl() + "qr-ci/" + encodedurlQrCodeChecking

		bodyModel := services.ShortReq{
			URL:    linkQRCodeFull,
			Domain: "bit.ly",
		}

		bodyModelByte, errB := json.Marshal(bodyModel)
		if errB != nil {
			log.Println("sendSmsBooking errB", errB.Error())
		}

		errS, _, resp := services.BitlyShorten(bodyModelByte)
		if errS != nil {
			log.Println("sendSmsBooking errS", errS.Error())
		}

		if resp.URL != "" {
			log.Println("sendSmsBooking short Link", resp.URL)
			message += resp.URL
		} else {
			message += linkQRCodeFull
		}

	}

	strPhoneNumber := "+" + fmt.Sprint(num.GetCountryCode()) + fmt.Sprint(num.GetNationalNumber())

	provider, errSend := services.VNPaySendSmsV2(strPhoneNumber, message)

	log.Println("sendSmsBooking ok", provider)

	return errSend
}
