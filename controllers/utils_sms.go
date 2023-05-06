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
	//disable for prod
	// sendSmsBooking(listHaveQRURL)
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

	message := "San " + listBooking[0].CourseUid + " xac nhan dat cho ngay " + listBooking[0].BookingDate + ": "

	for i, b := range listBooking {
		iStr := strconv.Itoa(i + 1)
		message += iStr + ". Player " + iStr + ": "
		playerName := ""
		if b.MemberCard != nil {
			playerName = b.MemberCard.CardId
		}
		// if playerName == "" {
		// 	playerName = b.CustomerInfo.Name
		// }

		message += playerName + " - " + "Ma check-in: " + b.CheckInCode + " - QR: "

		// base64 qr image
		encodeQrUrl := base64.StdEncoding.EncodeToString([]byte(b.QrcodeUrl))

		qrCodeUrlModel := QrCodeUrlModel{
			QrImg:       encodeQrUrl,
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

/*
Update image to ekyc server
*/
func ekycUpdateImage(memberUid, link string) {
	body := services.EkycUpdateBody{
		D:           memberUid,
		S:           memberUid,
		SelfieImage: link,
	}

	dataByte, errM := json.Marshal(body)

	if errM != nil {
		log.Println("ekycUpdateImage errM", errM.Error())
		return
	}

	_, _ = services.EkycUpdateImage(dataByte)
}
