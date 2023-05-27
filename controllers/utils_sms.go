package controllers

import (
	base64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"start/config"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/services"
	"start/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	if len(listBooking) < 1 {
		log.Println("genQRCodeListBook can not find any Booking")
		return
	}

	// check config
	courseUid := listBooking[0].CourseUid
	agencyId := listBooking[0].AgencyId
	customerBookingEmail := listBooking[0].CustomerBookingEmail
	customerBookingPhone := listBooking[0].CustomerBookingPhone

	course := models.Course{}

	course.Uid = courseUid

	err := course.FindFirst()

	if err != nil {
		log.Println("genQRCodeListBook - Can not find course with uid: ", courseUid, err)
	}

	listHaveQRURL := []model_booking.Booking{}
	for _, v := range listBooking {
		genQrCodeForBooking(&v)
		listHaveQRURL = append(listHaveQRURL, v)
	}

	// Send socket
	for _, v := range listHaveQRURL {
		// push socket
		cNotification := CNotification{}
		bookingClone := v
		go cNotification.PushMessBoookingForApp(constants.NOTIFICATION_BOOKING_ADD, &bookingClone)
	}

	cNotification := CNotification{}
	go cNotification.PushNotificationCreateBooking(constants.NOTIFICATION_BOOKING_CMS, model_booking.Booking{})

	// check config accept auto send
	if !course.AutoSendBooking {
		log.Println("genQRCodeListBook - config is disabled auto send sms and email ")

		return
	}

	// bắn theo config angency
	if agencyId > 0 {
		// Send email
		if course.TypeSendInfoBookingAgency == constants.SEND_INFOR_GUEST_BOTH || course.TypeSendInfoBookingAgency == constants.SEND_INFOR_GUEST_EMAIL {
			go sendEmailBooking(listHaveQRURL, customerBookingEmail)

		}

		// Send sms
		if course.TypeSendInfoBookingAgency == constants.SEND_INFOR_GUEST_BOTH || course.TypeSendInfoBookingAgency == constants.SEND_INFOR_GUEST_SMS {
			go sendSmsBooking(listHaveQRURL, customerBookingPhone)
		}
	} else {
		// Send email
		if course.TypeSendInfoBooking == constants.SEND_INFOR_GUEST_BOTH || course.TypeSendInfoBooking == constants.SEND_INFOR_GUEST_EMAIL {
			go sendEmailBooking(listHaveQRURL, customerBookingEmail)

		}

		// Send sms
		if course.TypeSendInfoBooking == constants.SEND_INFOR_GUEST_BOTH || course.TypeSendInfoBooking == constants.SEND_INFOR_GUEST_SMS {
			go sendSmsBooking(listHaveQRURL, customerBookingPhone)
		}
	}

}

/*
Giới hạn độ dài tin nhắn
*/
func sendSmsBooking(listBooking []model_booking.Booking, phone string) {
	part := 10 // tối đa 10 booking 1 tin nhắn
	count := 0

	if len(listBooking) <= part {
		makeSendSmsBooking(listBooking, phone)
	} else {
		for {
			end := count + part
			if end > len(listBooking)-1 {
				end = len(listBooking)
			}

			makeSendSmsBooking(listBooking[count:end], phone)
			count = end
			if count >= len(listBooking) {
				break
			}
		}
	}
}

/*
Send sms
*/
func makeSendSmsBooking(listBooking []model_booking.Booking, phone string) error {
	if len(listBooking) == 0 {
		errEmpty := errors.New("sendSmsBooking Err List Booking Emplty")
		log.Println("sendSmsBooking errEmpty", errEmpty.Error())
		return errEmpty
	}

	// parse standard phone number
	num, errPhone := libphonenumber.Parse(phone, "VN")
	if errPhone != nil {
		log.Println("sendSmsBooking errPhone:", errPhone)

		return errPhone
	}

	message := getSmsGolfName(listBooking[0].CourseUid) + " xac nhan dat cho ngay " + listBooking[0].BookingDate + ": "

	for i, b := range listBooking {
		if b.AgencyId > 0 {
			log.Println("sendSmsBooking Agency disable send sms")
		} else {
			iStr := strconv.Itoa(i + 1)
			message += iStr + ". "
			playerName := ""
			if b.MemberCard != nil {
				playerName = b.MemberCard.CardId
			}

			message += playerName + "-" + "Ma check-in: " + b.CheckInCode + " - QR: "

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

			// bodyModel := services.ShortReq{
			// 	URL:    linkQRCodeFull,
			// 	Domain: "bit.ly",
			// }

			// bodyModelByte, errB := json.Marshal(bodyModel)
			// if errB != nil {
			// 	log.Println("sendSmsBooking errB", errB.Error())
			// }

			// errS, _, resp := services.BitlyShorten(bodyModelByte)
			// if errS != nil {
			// 	log.Println("sendSmsBooking errS", errS.Error())
			// }

			respModel, errResp := services.GenShortLink(linkQRCodeFull)
			if errResp != nil {
				log.Println("sendSmsBooking errS", errResp.Error())
			}

			if respModel.Short != "" {
				shortLink := config.GetShortLinkFe() + respModel.Short
				log.Println("sendSmsBooking short Link", shortLink)
				message += shortLink
			} else {
				message += linkQRCodeFull
			}

			message += " "
		}
	}

	strPhoneNumber := "+" + fmt.Sprint(num.GetCountryCode()) + fmt.Sprint(num.GetNationalNumber())

	provider, errSend := services.SendSmsV2(strPhoneNumber, message)

	log.Println("sendSmsBooking ok", provider)

	return errSend
}

/*
Update image to ekyc server
*/
func ekycUpdateImage(partnerUid, courseUid, sid, memberUid, link string, imgFile *multipart.File) {

	// Current Time
	currentTime := time.Now().Unix()
	currentTimeStr := strconv.FormatInt(currentTime, 10)

	// Request Id
	uid := uuid.New()
	requestId := utils.HashCodeUuid(uid.String())

	dataModel := services.EkycDataModel{
		Sid:        sid,
		IdNumber:   memberUid,
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		Timestamp:  currentTimeStr,
		RequestId:  requestId,
		ImgLink:    link,
	}

	// d = DataModel -> string json
	//
	dataModelByte, errMar := json.Marshal(dataModel)
	if errMar != nil {
		log.Println("ekycUpdateImage errMar", errMar.Error())
		return
	}
	dataModelJsonStr := string(dataModelByte)
	log.Println("ekycUpdateImage D", dataModelJsonStr)

	// S = signature
	errGen, signature := utils.EkycGenSignature(dataModelJsonStr)
	if errGen != nil {
		log.Println("ekycUpdateImage errGen", errGen.Error())
		return
	}
	log.Println("ekycUpdateImage S", signature)

	body := services.EkycUpdateBody{
		D:           dataModelJsonStr,
		S:           signature,
		SelfieImage: link, // Khong dung
	}

	dataByte, errM := json.Marshal(body)

	if errM != nil {
		log.Println("ekycUpdateImage errM", errM.Error())
		return
	}

	errUpload, _ := services.EkycUpdateImage(dataByte, body, imgFile)
	if errUpload != nil {
		log.Println("ekycUpdateImage errUpload", errUpload.Error())
	}
}

/*
Send sms
*/
func sendEmailBooking(listBooking []model_booking.Booking, email string) error {

	if len(listBooking) == 0 {
		errEmpty := errors.New("sendEmailBooking Err List Booking Emplty")
		log.Println("sendEmailBooking errEmpty", errEmpty.Error())
		return errEmpty
	}

	course := models.Course{}
	course.Uid = listBooking[0].CourseUid
	course.PartnerUid = listBooking[0].PartnerUid
	errF := course.FindFirst()
	if errF != nil {
		log.Println("sendEmailBooking errEmpty", errF.Error())
		return errF
	}

	// Sender
	// sender := "hotro@ caro .vn"
	sender := course.EmailBooking
	if sender == "" {
		log.Println("sendEmailBooking Sender is empty")
		return errors.New("Sender is empty")
	}

	// subject
	subject := getSmsGolfName(course.Uid) + " xác nhận đặt chỗ ngày " + listBooking[0].BookingDate

	// message
	message := ""

	if listBooking[0].AgencyId > 0 {
		message = fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		</head>
		<body>
		<h4 style="margin-bottom:20px;">Kính gửi Quý đối tác %s,</h4>
		<p><span style="font-weight: bold;">%s</span> xác nhận đặt chỗ ngày <span style="font-weight: bold;">%s</span> :</p>
		<p>- Mã đặt chỗ: <span style="font-weight: bold;">%s</span></p>
		<p>- Người đặt: <span style="font-weight: bold;">%s(%s)</span></p>
		<p style="margin-bottom:20px;">- Số lượng: <span style="font-weight: bold;">%d</span></p>
	`, listBooking[0].AgencyInfo.ShortName, course.Name, listBooking[0].BookingDate, listBooking[0].BookingCode,
			listBooking[0].AgencyInfo.ShortName, listBooking[0].AgencyInfo.AgencyId, len(listBooking))
	} else {
		message = fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			</head>
			<body>
			<h4 style="margin-bottom:20px;">Kính gửi anh/chị %s,</h4>
			<p><span style="font-weight: bold;">%s</span> xác nhận đặt chỗ ngày <span style="font-weight: bold;">%s</span> :</p>
		`, listBooking[0].CustomerBookingName, course.Name, listBooking[0].BookingDate)

		if listBooking[0].MemberUidOfGuest != "" {
			db := datasources.GetDatabaseWithPartner(listBooking[0].PartnerUid)

			memberCard := models.MemberCard{}
			memberCard.Uid = listBooking[0].MemberUidOfGuest
			if errFindMB := memberCard.FindFirst(db); errFindMB != nil {
				message += fmt.Sprintf(`<p>- Người đặt: <span style="font-weight: bold;">%s</span></p>`, listBooking[0].CustomerBookingName)
			} else {
				message += fmt.Sprintf(`<p>- Người đặt: <span style="font-weight: bold;">%s(%s)</span></p>`, listBooking[0].CustomerBookingName, memberCard.CardId)
			}
		} else {
			message += fmt.Sprintf(`<p>- Người đặt: <span style="font-weight: bold;">%s</span></p>`, listBooking[0].CustomerBookingName)
		}

		message += fmt.Sprintf(`<p style="margin-bottom:20px;">- Số lượng: <span style="font-weight: bold;">%d</span></p>`, len(listBooking))
	}

	for i, b := range listBooking {
		iStr := strconv.Itoa(i + 1)
		message += `<p>` + iStr + ". Player " + b.CustomerName + ""
		// playerName := ""
		// if b.MemberCard != nil {
		// 	message += fmt.Sprintf(`(<span style="font-weight: bold;">%s</span>)`, b.MemberCard.CardId)
		// }

		message += fmt.Sprintf(` - Mã check-in: <span style="font-weight: bold;">%s</span> - (QR Check-in: "`, b.CheckInCode)

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
			log.Println("sendEmailBooking errMas", errMas.Error())
		}

		encodedurlQrCodeChecking := base64.StdEncoding.EncodeToString(byteQrCodeUrlModel)

		linkQRCodeFull := config.GetPortalCmsUrl() + "qr-ci/" + encodedurlQrCodeChecking

		// bodyModel := services.ShortReq{
		// 	URL:    linkQRCodeFull,
		// 	Domain: "bit.ly",
		// }

		// bodyModelByte, errB := json.Marshal(bodyModel)
		// if errB != nil {
		// 	log.Println("sendEmailBooking errB", errB.Error())
		// }

		// errS, _, resp := services.BitlyShorten(bodyModelByte)
		// if errS != nil {
		// 	log.Println("sendEmailBooking errS", errS.Error())
		// }

		respModel, errResp := services.GenShortLink(linkQRCodeFull)
		if errResp != nil {
			log.Println("sendSmsBooking errS", errResp.Error())
		}

		if respModel.Short != "" {
			shortLink := config.GetShortLinkFe() + respModel.Short
			log.Println("sendEmailBooking short Link", shortLink)
			message += shortLink + ")</p>"
		} else {
			message += linkQRCodeFull + ")</p>"
		}

	}

	message = message + `
		<h4 style="color: red; margin-top:20px; font-style: italic;">Lưu ý: Quý khách vui lòng cung cấp mã QR Check-in hoặc đọc Mã check-in để được phục vụ khi đến sân. Xin cảm ơn !</h4>
		</body>
		</html>`

	// Send mail
	errSend := datasources.SendEmail(email, subject, message, sender)

	if errSend != nil {
		log.Println("sendEmailBooking errSend", errSend.Error())
	}

	return errSend
}

func getSmsGolfName(coureUid string) string {
	if coureUid == "CHI-LINH-01" {
		return "CHILINH GOLF"
	}
	return coureUid
}
