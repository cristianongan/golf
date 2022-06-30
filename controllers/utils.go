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
	return false
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

	bookingGolfFee.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, body.Hole)
	bookingGolfFee.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, body.Hole)
	bookingGolfFee.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, body.Hole)

	listBookingGolfFee = append(listBookingGolfFee, bookingGolfFee)
	return listBookingGolfFee, bookingGolfFee
}

// Khi add sub bag vào 1 booking thì cần cập nhật lại main bag cho booking sub bag
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
func initListRound(booking model_booking.Booking, bookingGolfFee model_booking.BookingGolfFee, checkInTime int64) model_booking.ListBookingRound {
	round := model_booking.BookingRound{}
	round.GuestStyle = booking.GuestStyle
	round.BuggyFee = bookingGolfFee.BuggyFee
	round.CaddieFee = bookingGolfFee.CaddieFee
	round.GreenFee = bookingGolfFee.GreenFee
	round.Hole = booking.Hole
	round.MemberCardUid = booking.MemberCardUid
	round.TeeOffTime = checkInTime
	round.Pax = 1

	listRounds := model_booking.ListBookingRound{}
	listRounds = append(listRounds, round)
	return listRounds
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
	if partnerUid == "" || courseUid == "" || bookingDate == "" || bag == "" {
		return errors.New(constants.API_ERR_INVALID_BODY_DATA), model_booking.Booking{}, models.Caddie{}, models.Buggy{}
	}

	// Get booking
	booking := model_booking.Booking{
		PartnerUid:  partnerUid,
		CourseUid:   courseUid,
		BookingDate: bookingDate,
		Bag:         bag,
	}

	err := booking.FindFirst()
	if err != nil {
		return errors.New("Không "), booking, models.Caddie{}, models.Buggy{}
	}

	//Check caddie
	//TODO: check caddie avaible
	caddie := models.Caddie{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		Code:       caddieCode,
	}
	errFC := caddie.FindFirst()
	if errFC != nil {
		return errFC, booking, caddie, models.Buggy{}
	}

	// Caddie đang trên sân rồi
	if caddie.IsInCourse {
		return errors.New("Caddie in course"), booking, caddie, models.Buggy{}
	}

	//Check buggy
	buggy := models.Buggy{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		Code:       buggyCode,
	}
	errFB := buggy.FindFirst()
	if errFC != nil {
		return errFB, booking, caddie, buggy
	}

	//Caddie
	booking.CaddieId = caddie.Id
	booking.CaddieInfo = cloneToCaddieBooking(caddie)
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

	//Buggy
	booking.BuggyId = buggy.Id
	booking.BuggyInfo = cloneToBuggyBooking(buggy)

	return nil, booking, caddie, buggy
}

/*
	Out caddie
*/
func udpOutCaddieBooking(booking model_booking.Booking) error {
	// Get Caddie
	errCd := udpCaddieOut(booking.CaddieId)
	if errCd != nil {
		return errCd
	}
	// Udp booking
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_OUT

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
	caddie.IsInCourse = false
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
