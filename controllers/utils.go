package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getLanguageFromHeader(c *gin.Context) string {
	lang := c.Request.Header.Get(constants.API_HEADER_KEY_LANGUAGE)
	if lang != "" {
		return lang
	}
	return constants.LANGUAGE_DEFAULT
}

func initListStringToStr() string {
	list := models.ListString{}
	strJson, err := json.Marshal(list)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return string(strJson)
}

func ListStringToStr(list models.ListString) string {
	listJson, err := json.Marshal(list)
	if err != nil {
		log.Println("ListStringToStr", err.Error())
		return ""
	}
	return string(listJson)
}

func checkContain(array []int64, val int64) bool {
	for _, item := range array {
		if val == item {
			return true
		}
	}
	return false
}

func checkStringInArray(array []string, val string) bool {
	for _, item := range array {
		if val == item {
			return true
		}
	}
	return false
}

// ===================================================================
type errorResponse struct {
	Message interface{} `json:"message"`
}

type errorResponse1 struct {
	Message interface{} `json:"message"`
	Status  int         `json:"status"`
}

func errorRequest(c *gin.Context, cause interface{}) {
	c.JSON(200, errorResponse1{Message: cause, Status: 400})
	c.Abort()
}

func badRequest(c *gin.Context, cause interface{}) {
	c.JSON(http.StatusBadRequest, errorResponse{cause})
}

func internalError(c *gin.Context, cause interface{}) {
	c.JSON(http.StatusInternalServerError, errorResponse{cause})
}

func notFoundError(c *gin.Context, cause interface{}) {
	c.JSON(http.StatusNotFound, errorResponse{cause})
}

func unauthorizedResponse(c *gin.Context, cause interface{}) {
	c.JSON(http.StatusUnauthorized, errorResponse{cause})
}

func okRes(c *gin.Context) {
	okResponse(c, gin.H{"message": "success"})
}

func okResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func conflictResponse(c *gin.Context, cause interface{}) {
	c.JSON(http.StatusConflict, errorResponse{cause})
}

func noContentResponse(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{})
}

func bindUriId(c *gin.Context) (int64, error) {
	type Uri struct {
		Id int64 `uri:"id" binding:"required"`
	}

	uri := Uri{}
	err := c.ShouldBindUri(&uri)
	return uri.Id, err
}

func udpPartnerUid(input string) string {
	result := strings.ReplaceAll(input, " ", "-")
	result1 := strings.ReplaceAll(result, "_", "-")
	result2 := strings.ToUpper(result1)
	return result2
}

func udpCourseUid(courseUid, partnerUid string) string {
	if strings.Contains(courseUid, partnerUid) {
		return strings.ToUpper(courseUid)
	}

	courseUid1 := strings.ReplaceAll(courseUid, " ", "-")
	courseUid2 := strings.ReplaceAll(courseUid1, "_", "-")
	return strings.ToUpper(partnerUid + "-" + courseUid2)
}

func checkDuplicateGolfFee(body models.GolfFee) bool {
	golfFee := models.GolfFee{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		GuestStyle:   body.GuestStyle,
		Dow:          body.Dow,
		TablePriceId: body.TablePriceId,
	}
	errFind := golfFee.FindFirst()
	if errFind == nil || golfFee.Id > 0 {
		log.Print("checkDuplicateGolfFee true")
		return true
	}

	//Check theo chi tiết ngày
	listTempR := models.GolfFee{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		GuestStyle:   body.GuestStyle,
		TablePriceId: body.TablePriceId,
	}
	listTemp := listTempR.GetGuestStyleGolfFeeByGuestStyle()

	listDowStr := strings.Split(body.Dow, "")

	isdup := false
	for _, v := range listTemp {
		for _, v1 := range listDowStr {
			if strings.Contains(v.Dow, v1) {
				log.Print("checkDuplicateGolfFee1 true")
				isdup = true
				break
			}
		}

	}

	return isdup
}

func getCustomerCategoryFromCustomerType(cusType string) string {
	customerType := models.CustomerType{
		Type: cusType,
	}
	errFind := customerType.FindFirst()
	if errFind != nil {
		log.Println("getCustomerCategoryFromCustomerType err", errFind.Error())
		return constants.CUSTOMER_TYPE_CUSTOMER
	}
	return customerType.Category
}

// Check Fee Data để lưu vào DB
func formatGolfFee(feeText string) string {
	feeTextFormat0 := strings.TrimSpace(feeText)
	feeTextFormat1 := strings.ReplaceAll(feeTextFormat0, " ", "")
	feeTextFormat2 := strings.ReplaceAll(feeTextFormat1, ",", "")
	feeTextFormatLast := strings.ReplaceAll(feeTextFormat2, ".", "")

	if strings.Contains(feeTextFormatLast, constants.FEE_SEPARATE_CHAR) {
		return feeTextFormatLast
	}
	list1 := strings.Split(feeTextFormatLast, constants.FEE_SEPARATE_CHAR)
	if len(list1) == 0 {
		return feeTextFormat2
	}
	if len(list1) == 1 {
		return list1[0]
	}
	return strings.Join(list1, constants.FEE_SEPARATE_CHAR)
}

// Get log for cms user action booking
func getBookingCmsUserLog(cmsUser string, timeDo int64) string {
	hourStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.HOUR_FORMAT)
	dayStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.DAY_FORMAT)
	yearStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.DATE_FORMAT_1)
	return `(` + cmsUser + `, ` + hourStr + `, ` + dayStr + `)` + " Input book: " + yearStr
}

/*
	  Tính golf fee cho tạo đơn có guest style
		Là phần tử đầu của list golfFee
*/
func getInitListGolfFeeForBooking(uid string, body request.CreateBookingBody, golfFee models.GolfFee) (model_booking.ListBookingGolfFee, model_booking.BookingGolfFee) {
	listBookingGolfFee := model_booking.ListBookingGolfFee{}
	bookingGolfFee := model_booking.BookingGolfFee{}
	bookingGolfFee.BookingUid = uid
	bookingGolfFee.Bag = body.Bag
	bookingGolfFee.PlayerName = body.CustomerName
	bookingGolfFee.RoundIndex = 0

	bookingGolfFee.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, body.Hole)
	bookingGolfFee.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, body.Hole)
	bookingGolfFee.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, body.Hole)

	listBookingGolfFee = append(listBookingGolfFee, bookingGolfFee)
	return listBookingGolfFee, bookingGolfFee
}

func getInitListGolfFeeForAddRound(booking *model_booking.Booking, golfFee models.GolfFee, hole int) {
	bookingGolfFee := booking.ListGolfFee[0]

	bookingGolfFee.BookingUid = booking.Uid
	bookingGolfFee.CaddieFee += utils.GetFeeFromListFee(golfFee.CaddieFee, hole)
	bookingGolfFee.BuggyFee += utils.GetFeeFromListFee(golfFee.BuggyFee, hole)
	bookingGolfFee.GreenFee += utils.GetFeeFromListFee(golfFee.GreenFee, hole)

	booking.ListGolfFee[0] = bookingGolfFee
}

/*
Tính golf fee cho đơn thqay đổi hố
*/
func getInitGolfFeeForChangeHole(body request.ChangeBookingHole, golfFee models.GolfFee) model_booking.BookingGolfFee {
	holePriceFormula := models.HolePriceFormula{}
	holePriceFormula.Hole = body.Hole
	err := holePriceFormula.FindFirst()
	if err != nil {
		log.Println("find hole price err", err.Error())
	}

	bookingGolfFee := model_booking.BookingGolfFee{}

	bookingGolfFee.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, body.Hole)
	bookingGolfFee.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, body.Hole)
	bookingGolfFee.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, body.Hole)

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_SELF && holePriceFormula.StopBySelf != "" {
		bookingGolfFee.CaddieFee = utils.GetFeeWidthHolePrice(golfFee.CaddieFee, body.Hole, holePriceFormula.StopBySelf)
		bookingGolfFee.BuggyFee = utils.GetFeeWidthHolePrice(golfFee.BuggyFee, body.Hole, holePriceFormula.StopBySelf)
		bookingGolfFee.GreenFee = utils.GetFeeWidthHolePrice(golfFee.GreenFee, body.Hole, holePriceFormula.StopBySelf)
	}

	if body.TypeChangeHole == constants.BOOKING_STOP_BY_RAIN && holePriceFormula.StopByRain != "" {
		bookingGolfFee.CaddieFee = utils.GetFeeWidthHolePrice(golfFee.CaddieFee, body.Hole, holePriceFormula.StopByRain)
		bookingGolfFee.BuggyFee = utils.GetFeeWidthHolePrice(golfFee.BuggyFee, body.Hole, holePriceFormula.StopByRain)
		bookingGolfFee.GreenFee = utils.GetFeeWidthHolePrice(golfFee.GreenFee, body.Hole, holePriceFormula.StopByRain)
	}

	return bookingGolfFee
}

/*
Theo giá đặc biệt, k theo GuestStyle
*/
func getInitListGolfFeeWithOutGuestStyleForBooking(uid, rate string, body request.CreateBookingBody, caddieFee, buggyFee, greenFee int64) (model_booking.ListBookingGolfFee, model_booking.BookingGolfFee) {
	listBookingGolfFee := model_booking.ListBookingGolfFee{}
	bookingGolfFee := model_booking.BookingGolfFee{}
	bookingGolfFee.BookingUid = uid
	bookingGolfFee.Bag = body.Bag
	bookingGolfFee.PlayerName = body.CustomerName
	bookingGolfFee.RoundIndex = 0

	bookingGolfFee.CaddieFee = utils.CalculateFeeByHole(body.Hole, caddieFee, rate)
	bookingGolfFee.BuggyFee = utils.CalculateFeeByHole(body.Hole, buggyFee, rate)
	bookingGolfFee.GreenFee = utils.CalculateFeeByHole(body.Hole, greenFee, rate)

	listBookingGolfFee = append(listBookingGolfFee, bookingGolfFee)
	return listBookingGolfFee, bookingGolfFee
}

func getInitListGolfFeeWithOutGuestStyleForAddRound(booking *model_booking.Booking, rate string, caddieFee, buggyFee, greenFee int64, hole int) {
	bookingGolfFee := booking.ListGolfFee[0]

	bookingGolfFee.BookingUid = booking.Uid
	bookingGolfFee.CaddieFee += utils.CalculateFeeByHole(hole, caddieFee, rate)
	bookingGolfFee.BuggyFee += utils.CalculateFeeByHole(hole, buggyFee, rate)
	bookingGolfFee.GreenFee += utils.CalculateFeeByHole(hole, greenFee, rate)

	booking.ListGolfFee[0] = bookingGolfFee
}

/*
	Booking Init and Update

init price
init Golf Fee
init MushPay
init Rounds
*/
func initPriceForBooking(booking *model_booking.Booking, listBookingGolfFee model_booking.ListBookingGolfFee, bookingGolfFee model_booking.BookingGolfFee, checkInTime int64) {
	if booking == nil {
		log.Println("initPriceForBooking err booking nil")
		return
	}
	var bookingTemp model_booking.Booking
	bookingTempByte, err0 := json.Marshal(booking)
	if err0 != nil {
		log.Println("initPriceForBooking err0", err0.Error())
	}
	err1 := json.Unmarshal(bookingTempByte, &bookingTemp)
	if err1 != nil {
		log.Println("initPriceForBooking err1", err1.Error())
	}

	booking.ListGolfFee = listBookingGolfFee
	bookingTemp.ListGolfFee = listBookingGolfFee

	// Current Bag Price Detail
	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	currentBagPriceDetail.UpdateAmount()

	booking.CurrentBagPrice = currentBagPriceDetail
	bookingTemp.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(bookingTemp)

	booking.MushPayInfo = mushPayInfo
	bookingTemp.MushPayInfo = mushPayInfo

	// Rounds: Init Firsts
	initListRound(bookingTemp, bookingGolfFee, checkInTime)
	// booking.Rounds = listRounds
}

func initUpdatePriceBookingForChanegHole(booking *model_booking.Booking, bookingGolfFee model_booking.BookingGolfFee) {
	if booking == nil {
		log.Println("initPriceForBooking err booking nil")
		return
	}
	var bookingTemp model_booking.Booking
	bookingTempByte, err0 := json.Marshal(booking)
	if err0 != nil {
		log.Println("initPriceForBooking err0", err0.Error())
	}
	err1 := json.Unmarshal(bookingTempByte, &bookingTemp)
	if err1 != nil {
		log.Println("initPriceForBooking err1", err1.Error())
	}

	// update last golffee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].GreenFee = bookingGolfFee.GreenFee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].CaddieFee = bookingGolfFee.CaddieFee
	booking.ListGolfFee[len(booking.ListGolfFee)-1].BuggyFee = bookingGolfFee.BuggyFee
	bookingTemp.ListGolfFee = booking.ListGolfFee

	// Current Bag Price Detail
	currentBagPriceDetail := model_booking.BookingCurrentBagPriceDetail{}
	currentBagPriceDetail.GolfFee = bookingGolfFee.CaddieFee + bookingGolfFee.BuggyFee + bookingGolfFee.GreenFee
	currentBagPriceDetail.UpdateAmount()

	booking.CurrentBagPrice = currentBagPriceDetail
	bookingTemp.CurrentBagPrice = currentBagPriceDetail

	// MushPayInfo
	mushPayInfo := initBookingMushPayInfo(bookingTemp)

	booking.MushPayInfo = mushPayInfo

	// Rounds
	// booking.Rounds[len(booking.Rounds)-1].GreenFee = bookingGolfFee.GreenFee
	// booking.Rounds[len(booking.Rounds)-1].CaddieFee = bookingGolfFee.CaddieFee
	// booking.Rounds[len(booking.Rounds)-1].BuggyFee = bookingGolfFee.BuggyFee
}

// Khi add sub bag vào 1 booking thì cần cập nhật lại main bag cho booking sub bag
// Cập nhật lại giá cho SubBag
func updateMainBagForSubBag(body request.AddSubBagToBooking, mainBag string, customerPlayer string) error {
	var err error
	for _, v := range body.SubBags {
		booking := model_booking.Booking{}
		booking.Uid = v.BookingUid
		errFind := booking.FindFirst()
		if errFind == nil {
			mainBag := utils.BookingSubBag{
				BookingUid: body.BookingUid,
				GolfBag:    mainBag,
				PlayerName: customerPlayer,
			}
			booking.MainBags = append(booking.MainBags, mainBag)
			booking.MushPayInfo.MushPay = 0
			errUdp := booking.Update()
			if errUdp != nil {
				err = errUdp
				log.Println("UpdateMainBagForSubBag errUdp", errUdp.Error())
			}
		} else {
			err = errFind
			log.Println("UpdateMainBagForSubBag errFind", errFind.Error())
		}
	}

	return err
}

/*
Init List Round
*/
func initListRound(booking model_booking.Booking, bookingGolfFee model_booking.BookingGolfFee, checkInTime int64) {
	// create round and add round
	round := models.Round{}
	round.BillCode = booking.BillCode
	round.Bag = booking.Bag
	round.PartnerUid = booking.PartnerUid
	round.CourseUid = booking.CourseUid
	round.GuestStyle = booking.GuestStyle
	round.BuggyFee = bookingGolfFee.BuggyFee
	round.CaddieFee = bookingGolfFee.CaddieFee
	round.GreenFee = bookingGolfFee.GreenFee
	round.Hole = booking.Hole
	round.MemberCardUid = booking.MemberCardUid
	round.TeeOffTime = booking.CheckInTime
	round.Pax = 1
	round.Index = 1

	errCreateRound := round.Create()
	if errCreateRound != nil {
		log.Println("createBagsNote err", errCreateRound.Error())
	}
}

/*
Init Booking MushPayInfo
*/
func initBookingMushPayInfo(booking model_booking.Booking) model_booking.BookingMushPay {
	mushPayInfo := model_booking.BookingMushPay{}
	mushPayInfo.TotalGolfFee = booking.GetTotalGolfFee()
	mushPayInfo.TotalServiceItem = booking.GetTotalServicesFee()
	mushPayInfo.MushPay = mushPayInfo.TotalGolfFee + mushPayInfo.TotalServiceItem
	return mushPayInfo
}

// Check Duplicated SubBag
func checkCheckSubBagDupli(bookingUid string, booking model_booking.Booking) bool {
	isDupli := false

	if booking.SubBags == nil {
		return isDupli
	}
	if len(booking.SubBags) == 0 {
		return isDupli
	}
	for _, v := range booking.SubBags {
		if v.BookingUid == bookingUid {
			isDupli = true
		}
	}

	return isDupli
}

/*
Create bags note: Note of Bag
*/
func createBagsNoteNoteOfBag(booking model_booking.Booking) {
	if booking.NoteOfBag == "" {
		return
	}

	bagsNote := models.BagsNote{
		BookingUid: booking.Uid,
		GolfBag:    booking.Bag,
		Note:       booking.NoteOfBag,
		PlayerName: booking.CustomerName,
		Type:       constants.BAGS_NOTE_TYPE_BAG,
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
	}

	errC := bagsNote.Create()
	if errC != nil {
		log.Println("createBagsNote err", errC.Error())
	}
}

/*
Create bags note: Note of Booking
*/
func createBagsNoteNoteOfBooking(booking model_booking.Booking) {
	if booking.NoteOfBooking == "" {
		return
	}

	bagsNote := models.BagsNote{
		BookingUid: booking.Uid,
		GolfBag:    booking.Bag,
		Note:       booking.NoteOfBooking,
		PlayerName: booking.CustomerName,
		Type:       constants.BAGS_NOTE_TYPE_BOOKING,
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
	}

	errC := bagsNote.Create()
	if errC != nil {
		log.Println("createBagsNote err", errC.Error())
	}
}

func convertToCustomerSqlIntoBooking(customerSql models.CustomerUser) model_booking.CustomerInfo {
	cusBook := model_booking.CustomerInfo{}
	byteData, err := json.Marshal(customerSql)
	if err != nil {
		log.Println("convertToCustomerSqlIntoBooking err", err.Error())
		return cusBook
	}
	errUnM := json.Unmarshal(byteData, &cusBook)
	if errUnM != nil {
		log.Println("convertToCustomerSqlIntoBooking errUnM", errUnM.Error())
	}

	return cusBook
}

func cloneToAgencyBooking(agency models.Agency) model_booking.BookingAgency {
	agencyBooking := model_booking.BookingAgency{}
	agencyData, errMAgency := json.Marshal(&agency)
	if errMAgency != nil {
		log.Println("CloneToAgencyBooking errMAgency", errMAgency.Error())
	}
	errUnMAgency := json.Unmarshal(agencyData, &agencyBooking)
	if errMAgency != nil {
		log.Println("CloneToAgencyBooking errUnMAgency", errUnMAgency.Error())
	}

	return agencyBooking
}

func cloneToCustomerBooking(cus models.CustomerUser) model_booking.CustomerInfo {
	cusBooking := model_booking.CustomerInfo{}
	cusData, errMCus := json.Marshal(&cus)
	if errMCus != nil {
		log.Println("CloneToCustomerBooking errMCus", errMCus.Error())
	}
	errUnMCus := json.Unmarshal(cusData, &cusBooking)
	if errUnMCus != nil {
		log.Println("CloneToCustomerBooking errUnMCus", errUnMCus.Error())
	}

	return cusBooking
}

func cloneToCaddieBooking(caddie models.Caddie) model_booking.BookingCaddie {
	caddieBooking := model_booking.BookingCaddie{}
	caddieData, errM := json.Marshal(&caddie)
	if errM != nil {
		log.Println("cloneToCaddieBooking errM", errM.Error())
	}
	errUnM := json.Unmarshal(caddieData, &caddieBooking)
	if errUnM != nil {
		log.Println("cloneToCaddieBooking errUnM", errUnM.Error())
	}

	return caddieBooking
}

func cloneToBuggyBooking(buggy models.Buggy) model_booking.BookingBuggy {
	buggyBooking := model_booking.BookingBuggy{}
	buggyData, errM := json.Marshal(&buggy)
	if errM != nil {
		log.Println("cloneToBuggyBooking errM", errM.Error())
	}
	errUnM := json.Unmarshal(buggyData, &buggyBooking)
	if errUnM != nil {
		log.Println("cloneToBuggyBooking errUnM", errUnM.Error())
	}

	return buggyBooking
}

/*
Add Caddie, Buggy To Booking
*/
func addCaddieBuggyToBooking(partnerUid, courseUid, bookingDate, bag, caddieCode, buggyCode string) (error, model_booking.Booking, models.Caddie, models.Buggy) {
	//if partnerUid == "" || courseUid == "" || bookingDate == "" || bag == "" {
	//	return errors.New(constants.API_ERR_INVALID_BODY_DATA), model_booking.Booking{}, models.Caddie{}, models.Buggy{}
	//}
	// -> COMMENT REASON: duplicate validate

	// Get booking
	booking := model_booking.Booking{
		PartnerUid:  partnerUid,
		CourseUid:   courseUid,
		BookingDate: bookingDate,
		Bag:         bag,
	}

	err := booking.FindFirst()
	if err != nil {
		return err, booking, models.Caddie{}, models.Buggy{}
	}

	//Check caddie
	var caddie models.Caddie
	if caddieCode != "" {
		caddie = models.Caddie{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       caddieCode,
		}
		errFC := caddie.FindFirst()
		if errFC != nil {
			return errFC, booking, caddie, models.Buggy{}
		}

		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

		// if caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_READY {
		// booking.CaddieId = caddie.Id
		// booking.CaddieInfo = cloneToCaddieBooking(caddie)
		// booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN
		// caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_IN_COURSE
		// } else {
		// 	return errors.New("Caddie đang trong IN_COURSE"), booking, caddie, models.Buggy{}
		// }
	}

	// TODO: validate current_status

	//Check buggy
	var buggy models.Buggy
	if buggyCode != "" {
		buggy = models.Buggy{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       buggyCode,
		}
		errFB := buggy.FindFirst()
		if errFB != nil {
			return errFB, booking, caddie, buggy
		}
		booking.BuggyId = buggy.Id
		booking.BuggyInfo = cloneToBuggyBooking(buggy)
	}

	return nil, booking, caddie, buggy
}

/*
Out caddie
*/
func udpOutCaddieBooking(booking *model_booking.Booking) error {
	// Get Caddie
	errCd := udpCaddieOut(booking.CaddieId)
	if errCd != nil {
		return errCd
	}
	// Udp booking
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_OUT

	booking.BagStatus = constants.BAG_STATUS_TIMEOUT

	return nil
}

/*
Update caddie is in course is false
*/
func udpCaddieOut(caddieId int64) error {
	// Get Caddie
	caddie := models.Caddie{}
	caddie.Id = caddieId
	err := caddie.FindFirst()
	//caddie.IsInCourse = false
	if caddie.CurrentRound == 0 {
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
	} else {
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH
	}
	err = caddie.Update()
	if err != nil {
		log.Println("udpCaddieOut err", err.Error())
	}
	return err
}

/*
add Caddie In Out Note
*/
func addCaddieInOutNote(caddieInOut model_gostarter.CaddieInOutNote) {
	err := caddieInOut.Create()
	if err != nil {
		log.Println("err addCaddieInOutNote", err.Error())
	}
}

func udpBuggyOut(buggyId int64) error {
	buggy := models.Buggy{}
	buggy.Id = buggyId
	err := buggy.FindFirst()
	//buggy.IsInCourse = false
	if err := buggy.Update(); err != nil {
		log.Println("udpBuggyOut err", err.Error())
	}
	return err
}

/*
Create Locker: Locker for list
*/
func createLocker(booking model_booking.Booking) {
	if booking.LockerNo == "" {
		return
	}

	locker := models.Locker{
		BookingUid: booking.Uid,
	}

	// check tồn tại
	errF := locker.FindFirst()
	if errF != nil || locker.Id <= 0 {
		// Tạo mới
		locker.CourseUid = booking.CourseUid
		locker.PartnerUid = booking.PartnerUid
		locker.GolfBag = booking.Bag
		locker.PlayerName = booking.CustomerName
		locker.Locker = booking.LockerNo
		locker.GuestStyle = booking.GuestStyle
		locker.GuestStyleName = booking.GuestStyleName

		errC := locker.Create()
		if errC != nil {
			log.Println("createLocker errC", errC.Error())
		}
		return
	}

	if booking.LockerNo != "" && locker.Locker != booking.LockerNo {
		locker.PlayerName = booking.CustomerName
		locker.Locker = booking.LockerNo
		errU := locker.Update()
		if errU != nil {
			log.Println("createLocker errU", errU.Error())
		}
	}
}

/*
Check ngày của guest
guest_style_of_guest
- ngày dc đi:
2,2B:2345 :
GS = GuestStyle
Mô tả - GS 2 đc đi tất cả các ngày trong tuần, GS 2B dc đi các thứ 2345

- Số lượng Guest dc đi trong ngày của member card đó
+ Check ngày thường(normal_day_take_guest): Ex: 7,3: ý nghĩa ngày thường mã 2 dc đưa 7 khách, mã 2B dc đưa 3 khách
+ Check ngày cuối tuần(weekend_take_guest): Ex 3: Ý nghĩa cuối tuần mã 3 được đưa 2 khách, Mã 2B không được đưa khách nào
*/
func checkMemberCardGuestOfDay(memberCard models.MemberCard, memberCardType models.MemberCardType, guestStyle string, createdTime time.Time) (bool, error) {
	// Parse guest_style_of_guest
	// 2, 2B:2345 -> to List
	listGsOfGuest := memberCardType.ParseGsOfGuest()

	if len(listGsOfGuest) == 0 {
		return true, nil
	}

	isOk := true
	var err error

	for i, v := range listGsOfGuest {
		// Check GuestStyle có không
		if v.GuestStyle == guestStyle {
			if v.Dow != "" {
				if utils.CheckDow(v.Dow, createdTime) {
					// Ngày hợp lệ
					listTotal := []int{}
					if utils.IsWeekend(createdTime.Unix()) {
						// Check nếu cuối tuần
						listTotal = memberCardType.ParseWeekendTakeGuest()
					} else {
						listTotal = memberCardType.ParseNormalDayTakeGuest()
					}
					if i < len(listTotal) {
						if memberCard.TotalGuestOfDay >= listTotal[i] {
							isOk = false
							err = errors.New("Qua so lan choi trong ngay")
						}
					}
				} else {
					isOk = false
					err = errors.New("Ngay khong cho phep")
				}
			} else {
				// ok tất cả các ngày
			}
		}

	}

	return isOk, err
}

func updateMemberCard(memberCard models.MemberCard) {
	errUdp := memberCard.Update()
	if errUdp != nil {
		log.Println("updateMemberCard errUdp", errUdp.Error())
	}
}

/*
Handle MemberCard for Booking
*/
func handleCheckMemberCardOfGuest(memberUidOfGuest, guestStyle string) (error, models.MemberCard, string) {
	var memberCard models.MemberCard
	memberCard = models.MemberCard{}
	memberCard.Uid = memberUidOfGuest
	errM1, errM2, memberCardType := memberCard.FindFirstWithMemberCardType()
	if errM1 != nil {
		return errM1, memberCard, ""
	}
	if errM2 != nil {
		return errM2, memberCard, ""
	}

	// Check còn slot
	isOk, errCheckMember := checkMemberCardGuestOfDay(memberCard, memberCardType, guestStyle, time.Now())
	if !isOk {
		if errCheckMember == nil {
			errCheckMember = errors.New("not ok")
		}
		return errCheckMember, memberCard, ""
	}

	totalTemp := memberCard.TotalGuestOfDay
	memberCard.TotalGuestOfDay = totalTemp + 1

	customer := models.CustomerUser{}
	customer.Uid = memberCard.OwnerUid
	errFC := customer.FindFirst()
	if errFC != nil {
		log.Println("handleBookingForMemberCard err", errFC.Error())
	}

	return nil, memberCard, customer.Name
}

func updateAnnualFeeToMcType(yearInt int, mcTypeId, fee int64) {
	if time.Now().Year() == yearInt {
		mcType := models.MemberCardType{}
		mcType.Id = mcTypeId
		errFMCType := mcType.FindFirst()
		if errFMCType == nil {
			if mcType.CurrentAnnualFee != fee {
				mcType.CurrentAnnualFee = fee
				errMcTUdp := mcType.Update()
				if errMcTUdp != nil {
					log.Println("updateAnnualFeeToMcType errMcTUdp", errMcTUdp.Error())
				}
			}
		} else {
			log.Println("updateAnnualFeeToMcType errFMCType", errFMCType.Error())
		}
	}
}

func validatePartnerAndCourse(partnerUid string, courseUid string) error {
	partnerRequest := models.Partner{}
	partnerRequest.Uid = partnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		return partnerErrFind
	}

	courseRequest := models.Course{}
	courseRequest.Uid = courseUid
	errCourseFind := courseRequest.FindFirst()
	if errCourseFind != nil {
		return errCourseFind
	}
	return nil
}

/*
Init data main Bag For pay for booking
*/
func initMainBagForPay() utils.ListString {
	listPays := utils.ListString{}
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_RENTAL)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_KIOSK)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP)
	listPays = append(listPays, constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE)
	return listPays
}
