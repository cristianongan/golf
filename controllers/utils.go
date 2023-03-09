package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	kiosk_inventory "start/models/kiosk-inventory"
	model_payment "start/models/payment"
	model_service "start/models/service"
	"start/utils"
	"strconv"
	"strings"
	"time"

	model_report "start/models/report"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

func checkDuplicateGolfFee(db *gorm.DB, body models.GolfFee, isUdp bool) bool {
	golfFee := models.GolfFee{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		GuestStyle:   body.GuestStyle,
		Dow:          body.Dow,
		TablePriceId: body.TablePriceId,
	}

	if body.ApplyTime != "" {
		// Có set time áp dụng
		golfFee.ApplyTime = body.ApplyTime
		errFind := golfFee.FindFirst(db)
		if errFind == nil || golfFee.Id > 0 {
			log.Print("checkDuplicateGolfFee 0 true")
			return true
		}
		return false
	}

	errFind := golfFee.FindFirst(db)
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
	listTemp := listTempR.GetGuestStyleGolfFeeByGuestStyle(db)

	listDowStr := strings.Split(body.Dow, "")

	isdup := false
	for _, v := range listTemp {
		if isUdp && v.Id == body.Id {
			// Nếu là item đó và udp thì cứ udp
		} else {
			for _, v1 := range listDowStr {
				if strings.Contains(v.Dow, v1) {
					log.Print("checkDuplicateGolfFee1 true")
					isdup = true
					break
				}
			}
		}
	}

	return isdup
}

func getCustomerCategoryFromCustomerType(db *gorm.DB, cusType string) string {
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

// Get log for cms user action booking
func getBookingCmsUserLog(cmsUser string, timeDo int64) string {
	hourStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.HOUR_FORMAT)
	dayStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.DAY_FORMAT)
	yearStr, _ := utils.GetDateFromTimestampWithFormat(timeDo, constants.DATE_FORMAT_1)
	return `(` + cmsUser + `, ` + hourStr + `, ` + dayStr + `)` + " Input book: " + yearStr
}

// Khi add sub bag vào 1 booking thì cần cập nhật lại main bag cho booking sub bag
// Cập nhật lại giá cho SubBag
func updateMainBagForSubBag(db *gorm.DB, mainBooking model_booking.Booking) error {
	var err error
	for _, v := range mainBooking.SubBags {
		booking := model_booking.Booking{}
		booking.Uid = v.BookingUid
		errFind := booking.FindFirst(db)
		if errFind == nil {
			mainBag := utils.BookingSubBag{
				BookingUid: mainBooking.Uid,
				GolfBag:    mainBooking.Bag,
				PlayerName: mainBooking.CustomerName,
			}
			booking.MainBags = utils.ListSubBag{}
			booking.MainBags = append(booking.MainBags, mainBag)

			errUdp := booking.Update(db)

			// Tính lại giá của sub
			booking.UpdatePriceDetailCurrentBag(db)
			booking.UpdateMushPay(db)
			booking.Update(db)
			if errUdp != nil {
				err = errUdp
				log.Println("UpdateMainBagForSubBag errUdp", errUdp.Error())
			} else {
				// Udp lai info payment
				go handlePayment(db, booking)
			}
		} else {
			err = errFind
			log.Println("UpdateMainBagForSubBag errFind", errFind.Error())
		}
	}

	// Tính lại giá của main
	mainBooking.UpdatePriceDetailCurrentBag(db)
	mainBooking.UpdateMushPay(db)
	if errUdpM := mainBooking.Update(db); errUdpM != nil {
		go handlePayment(db, mainBooking)
	}

	return err
}

/*
Init List Round
*/
func initListRound(db *gorm.DB, booking model_booking.Booking, bookingGolfFee model_booking.BookingGolfFee) {

	// create round and add round
	round := models.Round{}
	round.BillCode = booking.BillCode
	round.Index = 1

	// Check Tồn tại chưa
	errFind := round.FindFirst(db)
	if errFind == nil {
		round.GuestStyle = booking.GuestStyle
		round.BuggyFee = bookingGolfFee.BuggyFee
		round.CaddieFee = bookingGolfFee.CaddieFee
		round.GreenFee = bookingGolfFee.GreenFee
		round.Hole = booking.Hole
		round.MemberCardUid = booking.MemberCardUid
		round.TeeOffTime = booking.CheckInTime
		for _, v := range booking.AgencyPaid {
			if v.Type == constants.BOOKING_AGENCY_GOLF_FEE && v.Fee > 0 {
				round.PaidBy = constants.PAID_BY_AGENCY
			}
		}
		errUdp := round.Update(db)
		if errUdp != nil {
			log.Println("createBagsNote errUdp", errUdp.Error())
		}

		return
	}

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
	for _, v := range booking.AgencyPaid {
		if v.Type == constants.BOOKING_AGENCY_GOLF_FEE && v.Fee > 0 {
			round.PaidBy = constants.PAID_BY_AGENCY
		}
	}

	errCreateRound := round.Create(db)
	if errCreateRound != nil {
		log.Println("create round err", errCreateRound.Error())
	}
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
func createBagsNoteNoteOfBag(db *gorm.DB, booking model_booking.Booking) {
	if booking.NoteOfBag == "" {
		return
	}

	bagsNote := models.BagsNote{
		BookingUid:  booking.Uid,
		GolfBag:     booking.Bag,
		Note:        booking.NoteOfBag,
		PlayerName:  booking.CustomerName,
		Type:        constants.BAGS_NOTE_TYPE_BAG,
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingDate: booking.BookingDate,
	}

	errC := bagsNote.Create(db)
	if errC != nil {
		log.Println("createBagsNote err", errC.Error())
	}
}

/*
Create bags note: Note of Booking
*/
func createBagsNoteNoteOfBooking(db *gorm.DB, booking model_booking.Booking) {
	if booking.NoteOfBooking == "" {
		return
	}

	bagsNote := models.BagsNote{
		BookingUid:  booking.Uid,
		GolfBag:     booking.Bag,
		Note:        booking.NoteOfBooking,
		PlayerName:  booking.CustomerName,
		Type:        constants.BAGS_NOTE_TYPE_BOOKING,
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingDate: booking.BookingDate,
	}

	errC := bagsNote.Create(db)
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

func cloneToBooking(booking model_booking.Booking) model_booking.Booking {
	bookingNew := model_booking.Booking{}
	agencyData, errMAgency := json.Marshal(&booking)
	if errMAgency != nil {
		log.Println("CloneToBooking errMAgency", errMAgency.Error())
	}
	errUnMAgency := json.Unmarshal(agencyData, &bookingNew)
	if errMAgency != nil {
		log.Println("CloneToBooking errUnMAgency", errUnMAgency.Error())
	}

	return bookingNew
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
Bag Status Waiting
*/
func addCaddieBuggyToBooking(db *gorm.DB, partnerUid, courseUid, bookingDate, bag, caddieCode, buggyCode string, isPrivateBuggy bool) (error, response.AddCaddieBuggyToBookingRes) {
	// Get booking
	booking := model_booking.Booking{
		PartnerUid:  partnerUid,
		CourseUid:   courseUid,
		BookingDate: bookingDate,
		Bag:         bag,
	}

	err := booking.FindFirst(db)
	if err != nil {
		return err, response.AddCaddieBuggyToBookingRes{}
	}

	response := response.AddCaddieBuggyToBookingRes{}

	//get old caddie
	if booking.CaddieId > 0 {
		oldCaddie := models.Caddie{}
		oldCaddie.Id = booking.CaddieId
		if errFC := oldCaddie.FindFirst(db); errFC == nil {
			response.OldCaddie = oldCaddie
		}
	}

	//get old buggy
	if booking.BuggyId > 0 {
		oldBuggy := models.Buggy{}
		oldBuggy.Id = booking.BuggyId
		if errFC := oldBuggy.FindFirst(db); errFC == nil {
			response.OldBuggy = oldBuggy
		}
	}

	if !(*booking.ShowCaddieBuggy) {
		booking.ResetCaddieBuggy()
	}

	//Check caddie
	var caddie models.Caddie
	if caddieCode != "" {
		caddie = models.Caddie{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       caddieCode,
		}
		errFC := caddie.FindFirst(db)
		if errFC != nil {
			return errFC, response
		}

		if caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_LOCK {
			if booking.CaddieId != caddie.Id {
				errTitle := fmt.Sprintln("Caddie", caddie.Code, "đang bị LOCK")
				return errors.New(errTitle), response
			}
		} else {
			if errCaddie := checkCaddieReady(booking, caddie); errCaddie != nil {
				return errCaddie, response
			}
		}

		booking.CaddieId = caddie.Id
		booking.CaddieInfo = cloneToCaddieBooking(caddie)
		booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

		if response.OldCaddie.Id == caddie.Id {
			response.OldCaddie = models.Caddie{}
		}
	}

	//Check buggy
	var buggy models.Buggy
	if buggyCode != "" {
		buggy = models.Buggy{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       buggyCode,
		}

		errFB := buggy.FindFirst(db)
		if errFB != nil {
			return errFB, response
		}

		if err := checkBuggyReady(db, buggy, booking, isPrivateBuggy, false); err != nil {
			return err, response
		}

		booking.BuggyId = buggy.Id
		booking.IsPrivateBuggy = setBoolForCursor(isPrivateBuggy)
		booking.BuggyInfo = cloneToBuggyBooking(buggy)

		if response.OldBuggy.Id == buggy.Id {
			response.OldBuggy = models.Buggy{}
		}
	}

	booking.ShowCaddieBuggy = setBoolForCursor(true)
	response.NewCaddie = caddie
	response.NewBuggy = buggy
	response.Booking = booking
	return nil, response
}

/*
Out Buggy
*/
func udpOutBuggy(db *gorm.DB, booking *model_booking.Booking, isOutAll bool) error {
	// Get Caddie
	if booking.BuggyId > 0 {
		bookingR := model_booking.BookingList{
			BookingDate: booking.BookingDate,
			BuggyId:     booking.BuggyId,
			BagStatus:   constants.BAG_STATUS_IN_COURSE,
		}

		_, total, _ := bookingR.FindAllBookingList(db)

		if total > 1 && !isOutAll {
			return errors.New("Buggy còn đang ghép với player khác")
		}

		buggy := models.Buggy{}
		buggy.Id = booking.BuggyId
		err := buggy.FindFirst(db)
		if err == nil {
			buggy.BuggyStatus = constants.BUGGY_CURRENT_STATUS_FINISH
			if errUdp := buggy.Update(db); errUdp != nil {
				log.Println("udpBuggyOut err", err.Error())
				return errUdp
			}
		}
	}

	return nil
}

/*
Update caddie is in course is false
*/
func udpCaddieOut(db *gorm.DB, caddieId int64) {
	// Get Caddie
	if caddieId > 0 {
		caddie := models.Caddie{}
		caddie.Id = caddieId
		err := caddie.FindFirst(db)
		if !(utils.ContainString(constants.LIST_CADDIE_READY_JOIN, caddie.CurrentStatus) > -1) {
			if caddie.CurrentRound == 0 {
				caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
			} else if caddie.CurrentRound == 1 {
				caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH
			} else if caddie.CurrentRound == 2 {
				caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH_R2
			} else if caddie.CurrentRound == 3 {
				caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH_R3
			}
			errUpd := caddie.Update(db)
			if errUpd != nil {
				log.Println("udpCaddieOut err", err.Error())
			}
		}
	}
}

/*
add Caddie In Out Note
*/
func addBuggyCaddieInOutNote(db *gorm.DB, caddieBuggyInOut model_gostarter.CaddieBuggyInOut) {
	newCaddieInOut := model_gostarter.CaddieBuggyInOut{
		PartnerUid: caddieBuggyInOut.PartnerUid,
		CourseUid:  caddieBuggyInOut.CourseUid,
		BookingUid: caddieBuggyInOut.BookingUid,
	}

	list, total, _ := newCaddieInOut.FindOrderByDateList(db)
	if total > 0 {
		lastItem := list[0]
		if caddieBuggyInOut.BuggyId > 0 && caddieBuggyInOut.CaddieId > 0 {
			err := caddieBuggyInOut.Create(db)
			if err != nil {
				log.Println("Create addBuggyCaddieInOutNote", err.Error())
			}
		} else {
			if (lastItem.BuggyId > 0 && lastItem.CaddieId > 0) ||
				lastItem.CaddieType == constants.STATUS_OUT ||
				lastItem.BuggyType == constants.STATUS_OUT ||
				caddieBuggyInOut.BuggyType == constants.STATUS_OUT ||
				caddieBuggyInOut.CaddieType == constants.STATUS_OUT {

				if caddieBuggyInOut.BagShareBuggy == "" {
					caddieBuggyInOut.BagShareBuggy = lastItem.BagShareBuggy
				}

				if caddieBuggyInOut.Hole == 0 {
					caddieBuggyInOut.Hole = lastItem.Hole
				}

				if caddieBuggyInOut.IsPrivateBuggy == nil {
					caddieBuggyInOut.IsPrivateBuggy = lastItem.IsPrivateBuggy
				}

				if caddieBuggyInOut.BuggyId > 0 && caddieBuggyInOut.CaddieId == 0 &&
					lastItem.CaddieType == constants.STATUS_IN {
					caddieBuggyInOut.CaddieId = lastItem.CaddieId
					caddieBuggyInOut.CaddieCode = lastItem.CaddieCode
					caddieBuggyInOut.CaddieType = lastItem.CaddieType
				} else if caddieBuggyInOut.CaddieId > 0 && caddieBuggyInOut.BuggyId == 0 &&
					lastItem.BuggyType == constants.STATUS_IN {
					caddieBuggyInOut.BuggyId = lastItem.BuggyId
					caddieBuggyInOut.BuggyCode = lastItem.BuggyCode
					caddieBuggyInOut.BuggyType = lastItem.BuggyType
				}

				err := caddieBuggyInOut.Create(db)
				if err != nil {
					log.Println("Create addBuggyCaddieInOutNote", err.Error())
				}
			} else {
				if caddieBuggyInOut.CaddieId > 0 {
					lastItem.CaddieId = caddieBuggyInOut.CaddieId
					lastItem.CaddieCode = caddieBuggyInOut.CaddieCode
					lastItem.CaddieType = caddieBuggyInOut.CaddieType
				}
				if caddieBuggyInOut.BuggyId > 0 {
					lastItem.BuggyId = caddieBuggyInOut.BuggyId
					lastItem.BuggyCode = caddieBuggyInOut.BuggyCode
					lastItem.BuggyType = caddieBuggyInOut.BuggyType
				}
				err := lastItem.Update(db)
				if err != nil {
					log.Println("Update addBuggyCaddieInOutNote", err.Error())
				}
			}
		}
	} else {
		err := caddieBuggyInOut.Create(db)
		if err != nil {
			log.Println("err addBuggyCaddieInOutNote", err.Error())
		}
	}

}

/*
add Buggy In Out Note
*/
func addBuggyInOutNote(db *gorm.DB, buggyInOut model_gostarter.BuggyInOut) {
	if buggyInOut.BuggyId != 0 {
		err := buggyInOut.Create(db)
		if err != nil {
			log.Println("err", err.Error())
		}
	}
}

/*
unlock turn time
*/
func unlockTurnTime(db *gorm.DB, booking model_booking.Booking) {
	cLockTeeTim := CLockTeeTime{}
	cLockTeeTim.DeleteLockTurn(db, booking.TeeTime, booking.BookingDate, booking.PartnerUid)
}

/*
Create Locker: Locker for list
*/
func createLocker(db *gorm.DB, booking model_booking.Booking) {
	if booking.LockerNo == "" {
		return
	}

	locker := models.Locker{
		BookingUid: booking.Uid,
	}

	// check tồn tại
	errF := locker.FindFirst(db)
	if errF != nil || locker.Id <= 0 {
		// Tạo mới
		locker.CourseUid = booking.CourseUid
		locker.PartnerUid = booking.PartnerUid
		locker.GolfBag = booking.Bag
		locker.PlayerName = booking.CustomerName
		locker.Locker = booking.LockerNo
		locker.GuestStyle = booking.GuestStyle
		locker.GuestStyleName = booking.GuestStyleName

		errC := locker.Create(db)
		if errC != nil {
			log.Println("createLocker errC", errC.Error())
		}
		return
	}

	if booking.LockerNo != "" {
		locker.PlayerName = booking.CustomerName
		locker.Locker = booking.LockerNo
		locker.GolfBag = booking.Bag
		errU := locker.Update(db)
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
				if utils.CheckDow(v.Dow, "", createdTime) {
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

func updateMemberCard(db *gorm.DB, memberCard models.MemberCard) {
	errUdp := memberCard.Update(db)
	if errUdp != nil {
		log.Println("updateMemberCard errUdp", errUdp.Error())
	}
}

/*
Handle MemberCard for Booking
*/
func handleCheckMemberCardOfGuest(db *gorm.DB, memberUidOfGuest, guestStyle string) (error, models.MemberCard, string) {
	var memberCard models.MemberCard
	memberCard = models.MemberCard{}
	memberCard.Uid = memberUidOfGuest
	errM1, errM2, memberCardType := memberCard.FindFirstWithMemberCardType(db)
	if errM1 != nil {
		return errM1, memberCard, ""
	}
	if errM2 != nil {
		return errM2, memberCard, ""
	}

	// Check còn slot
	isOk, errCheckMember := checkMemberCardGuestOfDay(memberCard, memberCardType, guestStyle, utils.GetTimeNow())
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
	errFC := customer.FindFirst(db)
	if errFC != nil {
		log.Println("handleBookingForMemberCard err", errFC.Error())
	}

	return nil, memberCard, customer.Name
}

func updateAnnualFeeToMcType(db *gorm.DB, yearInt int, mcTypeId, fee int64) {
	if utils.GetTimeNow().Year() == yearInt {
		mcType := models.MemberCardType{}
		mcType.Id = mcTypeId
		errFMCType := mcType.FindFirst(db)
		if errFMCType == nil {
			if mcType.CurrentAnnualFee != fee {
				mcType.CurrentAnnualFee = fee
				errMcTUdp := mcType.Update(db)
				if errMcTUdp != nil {
					log.Println("updateAnnualFeeToMcType errMcTUdp", errMcTUdp.Error())
				}
			}
		} else {
			log.Println("updateAnnualFeeToMcType errFMCType", errFMCType.Error())
		}
	}
}

func validatePartnerAndCourse(db *gorm.DB, partnerUid string, courseUid string) error {
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

/*
find booking with round va service items data
*/
func getBagDetailFromBooking(db *gorm.DB, booking model_booking.Booking) model_booking.BagDetail {
	//Get service items
	booking.FindServiceItems(db)

	bagDetail := model_booking.BagDetail{
		Booking: booking,
	}

	// Get Rounds
	if booking.BillCode != "" {
		round := models.Round{BillCode: booking.BillCode}
		listRound, _ := round.FindAll(db)

		if len(listRound) > 0 {
			bagDetail.Rounds = listRound
		}
	}
	return bagDetail
}

/*
find booking with round va service items data
*/
func getBagWithRoundDetail(db *gorm.DB, booking model_booking.Booking) model_booking.BagRoundNote {
	//Get service items
	booking.FindServiceItems(db)

	bagDetail := model_booking.BagRoundNote{
		Booking: booking,
	}

	// GetMemberCardInfo
	if booking.MemberCardUid != "" {
		memberCard := models.MemberCard{}
		memberCard.Uid = booking.MemberCardUid
		if errFindMB := memberCard.FindFirst(db); errFindMB == nil {
			bagDetail.MemberCardInfo = &memberCard
		}
	}

	// Get Rounds
	if booking.BillCode != "" {
		round := models.Round{BillCode: booking.BillCode}
		listRound, _ := round.FindAll(db)
		listRoundWithNote := []models.RoundWithNote{}
		for _, item := range listRound {
			listRoundWithNote = append(listRoundWithNote, models.RoundWithNote{
				Round: item,
			})
		}

		bookingListR := model_booking.BookingList{
			BillCode: booking.BillCode,
		}

		bookingList := []model_booking.Booking{}
		db1 := datasources.GetDatabaseWithPartner(booking.PartnerUid)
		db2, _, _ := bookingListR.FindAllBookingList(db1)
		db2 = db2.Order("created_at asc")
		db2.Find(&bookingList)

		for index, booking := range bookingList {
			roundIndex := index + 1
			for idx, round := range listRoundWithNote {
				if round.Index == roundIndex {
					listRoundWithNote[idx].Note = booking.NoteOfGo
					break
				}
			}
		}
		if len(listRound) > 0 {
			bagDetail.RoundsWithNote = listRoundWithNote
		}
	}
	return bagDetail
}

/*
find booking with round va service items data
*/
func getBagDetailForPayment(db *gorm.DB, booking model_booking.Booking) model_booking.BagDetail {
	//Get service items
	booking.FindServiceItemsInPayment(db)

	bagDetail := model_booking.BagDetail{
		Booking: booking,
	}

	mainPaidRound1 := false
	mainPaidRound2 := false

	if len(booking.MainBags) > 0 {
		mainBook := model_booking.Booking{
			CourseUid:   booking.CourseUid,
			PartnerUid:  booking.PartnerUid,
			Bag:         booking.MainBags[0].GolfBag,
			BookingDate: booking.BookingDate,
		}
		errFMB := mainBook.FindFirst(db)
		if errFMB != nil {
			log.Println("UpdateMushPay-"+booking.Bag+"-Find Main Bag", errFMB.Error())
		}

		mainPaidRound1 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND) > -1
		mainPaidRound2 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS) > -1
	}

	// Get Rounds
	if booking.BillCode != "" {
		round := models.Round{BillCode: booking.BillCode}
		listRound, _ := round.FindAll(db)

		newListRoundWillAdd := []models.Round{}

		for _, rd := range listRound {
			if rd.Index == 1 && !mainPaidRound1 && !booking.CheckAgencyPaidAll() && !booking.CheckAgencyPaidRound1() {
				newListRoundWillAdd = append(newListRoundWillAdd, rd)
			}
			if rd.Index == 2 && !mainPaidRound2 {
				newListRoundWillAdd = append(newListRoundWillAdd, rd)
			}
		}

		if len(listRound) > 0 {
			bagDetail.Rounds = newListRoundWillAdd
		}
	}
	return bagDetail
}

/*
Mỗi lần thêm đợt thanh toán, update lại totalPaid
*/
func updateTotalPaidAnnualFeeForMemberCard(db *gorm.DB, mcUid string, year int) {
	//Get List paid
	listPaidR := models.AnnualFeePay{
		MemberCardUid: mcUid,
		Year:          year,
	}
	listPaid, errF := listPaidR.FindAll(db)
	if errF != nil {
		log.Println("updateTotalPaidAnnualFeeForMemberCard errF", errF.Error())
	}

	totalPaid := int64(0)
	for _, v := range listPaid {
		totalPaid += v.Amount
	}

	countPaid := len(listPaid)

	// Find memberCard Annual Fee
	mcCardAnnualFee := models.AnnualFee{
		MemberCardUid: mcUid,
		Year:          year,
	}
	errMc := mcCardAnnualFee.FindFirst(db)
	if errMc != nil || mcCardAnnualFee.Id <= 0 {
		// Tạo mới
		mcCardAnnualFee.TotalPaid = totalPaid
		mcCardAnnualFee.CountPaid = countPaid
		errC := mcCardAnnualFee.Create(db)
		if errC != nil {
			log.Println("updateTotalPaidAnnualFeeForMemberCard errC", errC.Error())
		}
	} else {
		mcCardAnnualFee.TotalPaid = totalPaid
		mcCardAnnualFee.CountPaid = countPaid
		errUdp := mcCardAnnualFee.Update(db)
		if errUdp != nil {
			log.Println("updateTotalPaidAnnualFeeForMemberCard errUdp", errUdp.Error())
		}
	}
}

/*
Check Caddie có đang sẵn sàng để ghép không
*/
func checkCaddieReady(booking model_booking.Booking, caddie models.Caddie) error {
	if !(caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_READY ||
		caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_FINISH ||
		caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_FINISH_R2 ||
		caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_FINISH_R3) {
		return errors.New("Caddie " + caddie.Code + " chưa sẵn sàng để ghép ")
	}
	return nil
}

/*
Buggy có thể ghép tối đa 2 player
Check Buggy có đang sẵn sàng để ghép không
*/
func checkBuggyReady(db *gorm.DB, buggy models.Buggy, booking model_booking.Booking, isPrivateBuggy bool, isInCourse bool) error {

	if !(buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_ACTIVE ||
		buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_FINISH ||
		buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_LOCK ||
		buggy.BuggyStatus == constants.BUGGY_CURRENT_STATUS_IN_COURSE) {
		errTitle := fmt.Sprintln("Buggy", buggy.Code, "đang ở trạng thái", buggy.BuggyStatus)
		return errors.New(errTitle)
	}

	bookingList := model_booking.BookingList{
		PartnerUid:            booking.PartnerUid,
		CourseUid:             booking.CourseUid,
		BuggyCode:             buggy.Code,
		BookingDate:           booking.BookingDate,
		IsBuggyPrepareForJoin: "1",
	}

	dbResponse, total, _ := bookingList.FindAllBookingList(db)
	var list []model_booking.Booking
	dbResponse.Find(&list)

	// nếu list booking chứa booking đang ghép thì ko check nữa
	for _, data := range list {
		if booking.Uid == data.Uid {
			return nil
		}
	}

	if total >= 2 {
		errTitle := fmt.Sprintln(buggy.Code, "đã ghép đủ người")
		return errors.New(errTitle)
	}

	if total == 1 {
		bookingBuggy := list[0]
		if isInCourse {
			if bookingBuggy.FlightId != booking.FlightId {
				errTitle := fmt.Sprintln("Buggy", buggy.Code, "đang ở flight khác")
				return errors.New(errTitle)
			}
		}
		if *bookingBuggy.IsPrivateBuggy {
			errTitle := fmt.Sprintln("Buggy", buggy.Code, "đang được dùng private")
			return errors.New(errTitle)
		}
	}

	return nil
}

/*
Tính total Paid của user
*/
func getTotalPaidForCustomerUser(db *gorm.DB, userUid string) int64 {
	totalPaid := int64(0)

	//Get list memberCard của khách hàng
	memberCard := models.MemberCard{
		OwnerUid: userUid,
	}
	errMC, listMC := memberCard.FindAll(db)

	if errMC == nil {
		for _, v := range listMC {
			annualFeePayR := models.AnnualFeePay{
				MemberCardUid: v.Uid,
			}
			listFeePay, errAF := annualFeePayR.FindAll(db)
			if errAF == nil {
				for _, v1 := range listFeePay {
					totalPaid += v1.Amount
				}
			} else {
				log.Println("updateTotalPaidForCustomerUser errAF", errAF.Error())
			}
		}
	} else {
		log.Println("updateTotalPaidForCustomerUser errMC", errMC.Error())
	}

	return totalPaid
}

/*
Update report customer play
*/
func updateReportTotalPaidForCustomerUser(db *gorm.DB, userUid string, partnerUid, courseUid string) {
	totalPaid := getTotalPaidForCustomerUser(db, userUid)

	reportCustomer := model_report.ReportCustomerPlay{
		CustomerUid: userUid,
	}

	errF := reportCustomer.FindFirst()
	if errF != nil || reportCustomer.Id <= 0 {
		reportCustomer.CourseUid = courseUid
		reportCustomer.PartnerUid = partnerUid
		reportCustomer.TotalPaid = totalPaid
		errC := reportCustomer.Create()
		if errC != nil {
			log.Println("updateReportTotalPaidForCustomerUser errC", errC.Error())
		}
	} else {
		reportCustomer.TotalPaid = totalPaid
		errUdp := reportCustomer.Update()
		if errUdp != nil {
			log.Println("updateReportTotalPaidForCustomerUser errUdp", errUdp.Error())
		}
	}
}

/*
Udp report số lần chơi của user
*/
func updateReportTotalPlayCountForCustomerUser(userUid string, cardId string, partnerUid, courseUid string) {
	reportCustomer := model_report.ReportCustomerPlay{
		CustomerUid: userUid,
		CardId:      cardId,
	}

	errF := reportCustomer.FindFirst()
	if errF != nil || reportCustomer.Id <= 0 {
		reportCustomer.CourseUid = courseUid
		reportCustomer.PartnerUid = partnerUid
		reportCustomer.TotalPlayCount = 1
		errC := reportCustomer.Create()
		if errC != nil {
			log.Println("updateReportTotalPlayCountForCustomerUser errC", errC.Error())
		}
	} else {
		totalTemp := reportCustomer.TotalPlayCount
		reportCustomer.TotalPlayCount = totalTemp + 1
		errUdp := reportCustomer.Update()
		if errUdp != nil {
			log.Println("updateReportTotalPlayCountForCustomerUser errUdp", errUdp.Error())
		}
	}
}

/*
Udp report tổng giờ chơi của user
*/
func updateReportTotalHourPlayCountForCustomerUser(booking model_booking.Booking, userUid string, partnerUid, courseUid string) {
	reportCustomer := model_report.ReportCustomerPlay{
		CustomerUid: userUid,
	}

	now := utils.GetTimeNow()

	loc, errLoc := time.LoadLocation(constants.LOCATION_DEFAULT)
	if errLoc != nil {
		log.Println(errLoc)
	}

	// parse tee off
	teeOff := now.Format(constants.DATE_FORMAT) + " " + booking.TeeOffTime
	parseTeeOff, _ := time.Parse(constants.DATE_FORMAT_3, teeOff)

	// parse time out fight
	convertTimeOutFlight := time.Unix(booking.TimeOutFlight, 0).In(loc).Format(constants.HOUR_FORMAT)
	timeOutFlightRaw := now.Format(constants.DATE_FORMAT) + " " + convertTimeOutFlight
	paseTimeOutFlight, _ := time.Parse(constants.DATE_FORMAT_3, timeOutFlightRaw)

	totalHour := paseTimeOutFlight.Sub(parseTeeOff).Hours()

	errF := reportCustomer.FindFirst()
	if errF != nil || reportCustomer.Id <= 0 {
		reportCustomer.CourseUid = courseUid
		reportCustomer.PartnerUid = partnerUid
		reportCustomer.TotalHourPlayCount = math.Round(totalHour*100) / 100
		errC := reportCustomer.Create()
		if errC != nil {
			log.Println("updateReportTotalPlayCountForCustomerUser errC", errC.Error())
		}
	} else {
		totalTemp := reportCustomer.TotalHourPlayCount
		reportCustomer.TotalHourPlayCount = totalTemp + (math.Round(totalHour*100) / 100)
		errUdp := reportCustomer.Update()
		if errUdp != nil {
			log.Println("updateReportTotalPlayCountForCustomerUser errUdp", errUdp.Error())
		}
	}
}

/*
Validate Item Code có tồn tại trong Proshop or FB or Rental hay không
*/
func validateItemCodeInService(db *gorm.DB, serviceType string, itemCode string) error {
	if serviceType == constants.GROUP_PROSHOP {
		proshop := model_service.Proshop{
			ProShopId: itemCode,
		}

		if err := proshop.FindFirst(db); err == nil {
			return errors.New(itemCode + "không tìm thấy")
		}
	} else if serviceType == constants.GROUP_FB {
		fb := model_service.FoodBeverage{
			FBCode: itemCode,
		}

		if err := fb.FindFirst(db); err == nil {
			return errors.New(itemCode + "không tìm thấy")
		}
	} else if serviceType == constants.GROUP_RENTAL {
		rental := model_service.Rental{
			RentalId: itemCode,
		}

		if err := rental.FindFirst(db); err == nil {
			return errors.New(itemCode + " không tìm thấy ")
		}
	}
	return nil
}

/*
Get Item Info
*/
func getItemInfoInService(db *gorm.DB, partnerUid, courseUid, itemCode string) (kiosk_inventory.ItemInfo, error) {

	if itemCode == "" {
		return kiosk_inventory.ItemInfo{}, errors.New("Item Code Empty!")
	}

	code := strings.ReplaceAll(itemCode, " ", "")
	proshop := model_service.Proshop{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		ProShopId:  code,
	}

	if errFindProshop := proshop.FindFirst(db); errFindProshop == nil {
		return kiosk_inventory.ItemInfo{
			GroupCode: proshop.GroupCode,
			ItemName:  proshop.VieName,
			Unit:      proshop.Unit,
			GroupType: proshop.Type,
		}, nil
	}

	fb := model_service.FoodBeverage{
		FBCode: itemCode,
	}

	if err := fb.FindFirst(db); err == nil {
		return kiosk_inventory.ItemInfo{
			GroupCode: fb.GroupCode,
			ItemName:  fb.VieName,
			Unit:      fb.Unit,
			GroupType: fb.Type,
		}, nil
	}

	rental := model_service.Rental{
		RentalId: itemCode,
	}

	if err := rental.FindFirst(db); err == nil {
		return kiosk_inventory.ItemInfo{
			GroupCode: rental.GroupCode,
			ItemName:  rental.VieName,
			Unit:      rental.Unit,
			GroupType: rental.Type,
		}, nil
	}

	return kiosk_inventory.ItemInfo{}, errors.New(fmt.Sprintln(itemCode, "not found"))
}

func setBoolForCursor(b bool) *bool {
	boolVar := b
	return &boolVar
}

func getIntPointer(value int) *int {
	return &value
}

/*
Get Tee Time Lock Redis
*/
func getTeeTimeLockRedis(courseUid string, date string, teeType string) []models.LockTeeTimeWithSlot {
	prefixRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + date + "_" + courseUid
	if teeType != "" {
		prefixRedisKey += "_" + teeType
	}
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
	listTeeTimeLockRedis := []models.LockTeeTimeWithSlot{}
	if errRedis == nil && len(listKey) > 0 {
		strData, errGet := datasources.GetCaches(listKey...)
		if errGet != nil {
			log.Println("updateSlotTeeTimeWithLock-error", errGet.Error())
		} else {
			for _, data := range strData {
				if data != nil {
					byteData := []byte(data.(string))
					teeTime := models.LockTeeTimeWithSlot{}
					err2 := json.Unmarshal(byteData, &teeTime)
					if err2 == nil {
						listTeeTimeLockRedis = append(listTeeTimeLockRedis, teeTime)
					}
				}
			}
		}
	}
	return listTeeTimeLockRedis
}

/*
Đánh dấu lại round đã được trả bởi Main Bag
*/
func bookMarkRoundPaidByMainBag(mainBooking model_booking.Booking, db *gorm.DB) {
	checkIsFirstRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	checkIsNextRound := utils.ContainString(mainBooking.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)

	if len(mainBooking.SubBags) > 0 {
		for _, subBooking := range mainBooking.SubBags {
			round1 := models.Round{BillCode: subBooking.BillCode, Index: 1}
			if errRound1 := round1.FindFirst(db); errRound1 == nil {
				if checkIsFirstRound > -1 {
					round1.MainBagPaid = setBoolForCursor(true)
				} else {
					round1.MainBagPaid = setBoolForCursor(false)
				}
				round1.Update(db)
			}
			round2 := models.Round{BillCode: subBooking.BillCode, Index: 2}
			if errRound2 := round2.FindFirst(db); errRound2 == nil {
				if checkIsNextRound > -1 {
					round2.MainBagPaid = setBoolForCursor(true)
				} else {
					round2.MainBagPaid = setBoolForCursor(false)
				}
				round2.Update(db)
			}
		}
	}
}

/*
Tạo book reservation cho restaurant
*/

func addServiceCart(db *gorm.DB, numberGuest int, partnerUid, courseUid, playerName, phone, bookingDate, staffName string) {
	// create service cart
	kiosk := model_service.Kiosk{
		KioskType: constants.RESTAURANT_SETTING,
	}
	kiosk.FindFirst(db)

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = partnerUid
	serviceCart.CourseUid = courseUid

	date, _ := time.Parse(constants.DATE_FORMAT_1, bookingDate)
	serviceCart.BookingDate = datatypes.Date(date)

	serviceCart.ServiceId = kiosk.Id
	serviceCart.ServiceType = kiosk.KioskType
	serviceCart.BillCode = constants.BILL_NONE
	serviceCart.BillStatus = constants.RES_BILL_STATUS_BOOKING
	serviceCart.Type = constants.RES_TYPE_TABLE
	serviceCart.NumberGuest = numberGuest
	serviceCart.StaffOrder = staffName
	serviceCart.PlayerName = playerName
	serviceCart.Phone = phone
	serviceCart.OrderTime = utils.GetTimeNow().Unix()

	if err := serviceCart.Create(db); err != nil {
		log.Println("add service cart error!")
	}
}

/*
Tạo row index cho booking
*/
func generateRowIndex(rowsCurrent []int) int {
	log.Printf("time %d: %v", utils.GetTimeNow().Unix(), rowsCurrent)
	if !utils.Contains(rowsCurrent, 0) {
		return 0
	} else if !utils.Contains(rowsCurrent, 1) {
		return 1
	} else if !utils.Contains(rowsCurrent, 2) {
		return 2
	}
	return 3
}

/*
Remove row index trong redis
*/
func removeRowIndexRedis(booking model_booking.Booking) {
	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(booking.BookingDate, booking.CourseUid, booking.TeeTime, booking.TeeType+booking.CourseType)
	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
	newRowIndexsRedis := utils.ListInt{}

	for index, item := range rowIndexsRedis {
		if booking.RowIndex != nil {
			if item == *booking.RowIndex {
				newRowIndexsRedis = utils.RemoveIndex(rowIndexsRedis, index)
				break
			}
		} else {
			log.Println("removeRowIndexRedis booking.RowIndex == nil")
		}
	}

	if len(newRowIndexsRedis) > 0 {
		rowIndexsRaw, _ := newRowIndexsRedis.Value()
		errRedis := datasources.SetCache(teeTimeRowIndexRedis, rowIndexsRaw, 0)
		if errRedis != nil {
			log.Println("CreateBookingCommon errRedis", errRedis)
		}
	} else {
		errRedis := datasources.DelCacheByKey(teeTimeRowIndexRedis)
		if errRedis != nil {
			log.Println("CreateBookingCommon errRedis", errRedis)
		}
	}
}

func getBuggyFee(gs string) utils.ListGolfHoleFee {

	partnerUid := "CHI-LINH"
	courseUid := "CHI-LINH-01"

	db := datasources.GetDatabaseWithPartner(partnerUid)
	buggyFeeSettingR := models.BuggyFeeSetting{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
	}

	listBuggySetting, _, _ := buggyFeeSettingR.FindAll(db)
	buggyFeeSetting := models.BuggyFeeSetting{}
	for _, item := range listBuggySetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeSetting = item
			break
		}
	}

	buggyFeeItemSettingR := models.BuggyFeeItemSetting{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		GuestStyle: gs,
		SettingId:  buggyFeeSetting.Id,
	}

	listSetting, _, _ := buggyFeeItemSettingR.FindAll(db)
	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	for _, item := range listSetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeItemSetting = item
			break
		}
	}

	return buggyFeeItemSetting.RentalFee
}

// Update slot caddie

func updateCaddieOutSlot(partnerUid, courseUid string, caddies []string) error {
	var caddieSlotNew []string
	var caddieSlotExist []string
	// Format date
	dateNow, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	caddieWS := models.CaddieWorkingSlot{}
	caddieWS.PartnerUid = partnerUid
	caddieWS.CourseUid = courseUid
	caddieWS.ApplyDate = dateNow

	db := datasources.GetDatabaseWithPartner(partnerUid)

	err := caddieWS.FindFirst(db)
	if err != nil {
		return err
	}

	if len(caddieWS.CaddieSlot) > 0 {
		caddieSlotNew = append(caddieSlotNew, caddieWS.CaddieSlot...)
		for _, item := range caddies {
			index := utils.StringInList(item, caddieSlotNew)
			if index != -1 {
				caddieSlotNew = utils.Remove(caddieSlotNew, index)
				caddieSlotExist = append(caddieSlotExist, item)
			}
		}
	}

	caddieWS.CaddieSlot = append(caddieSlotNew, caddieSlotExist...)
	err = caddieWS.Update(db)
	if err != nil {
		return err
	}

	return nil
}

func undoCaddieOutSlot(partnerUid, courseUid string, caddies []string) error {
	var caddieSlotNew []string
	var caddieSlotExist []string
	// Format date
	dateNow, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	caddieWS := models.CaddieWorkingSlot{}
	caddieWS.PartnerUid = partnerUid
	caddieWS.CourseUid = courseUid
	caddieWS.ApplyDate = dateNow

	db := datasources.GetDatabaseWithPartner(partnerUid)

	err := caddieWS.FindFirst(db)
	if err != nil {
		return err
	}

	if len(caddieWS.CaddieSlot) > 0 {
		caddieSlotNew = append(caddieSlotNew, caddieWS.CaddieSlot...)
		for _, item := range caddies {
			index := utils.StringInList(item, caddieSlotNew)
			if index != -1 {
				caddieSlotNew = utils.Remove(caddieSlotNew, index)
				caddieSlotExist = append(caddieSlotExist, item)
			}
		}
	}

	caddieWS.CaddieSlot = append(caddieSlotExist, caddieSlotNew...)
	err = caddieWS.Update(db)
	if err != nil {
		return err
	}

	return nil
}

func removeCaddieOutSlotOnDate(partnerUid, courseUid, date string, caddieCode string) error {
	var caddieSlotNew []string

	caddieWS := models.CaddieWorkingSlot{}
	caddieWS.PartnerUid = partnerUid
	caddieWS.CourseUid = courseUid
	caddieWS.ApplyDate = date

	db := datasources.GetDatabaseWithPartner(partnerUid)

	err := caddieWS.FindFirst(db)
	if err != nil {
		return err
	}

	if len(caddieWS.CaddieSlot) > 0 {
		caddieSlotNew = append(caddieSlotNew, caddieWS.CaddieSlot...)
		index := utils.StringInList(caddieCode, caddieSlotNew)
		if index != -1 {
			caddieSlotNew = utils.Remove(caddieSlotNew, index)
		}
	}

	caddieWS.CaddieSlot = caddieSlotNew
	err = caddieWS.Update(db)
	if err != nil {
		return err
	}

	return nil
}

func updateCaddieWorkingOnDay(caddieCodeList []string, partnerUid, courseUid string, isWorking bool) {
	db := datasources.GetDatabaseWithPartner("CHI-LINH")
	for _, code := range caddieCodeList {
		caddie := models.Caddie{
			PartnerUid: partnerUid,
			CourseUid:  courseUid,
			Code:       code,
		}

		if err := caddie.FindFirst(db); err == nil {
			if isWorking {
				caddie.IsWorking = 1
			} else {
				caddie.IsWorking = 0
			}
			caddie.Update(db)
		}
	}
}

func lockTeeTimeToRedis(body models.LockTeeTimeWithSlot) {
	teeTimeRedisKey := getKeyTeeTimeLockRedis(body.DateTime, body.CourseUid, body.TeeTime, body.TeeType)

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       body.DateTime,
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		TeeTime:        body.TeeTime,
		CurrentTeeTime: body.CurrentTeeTime,
		TeeType:        body.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Type:           constants.LOCK_CMS,
		CurrentCourse:  body.CurrentCourse,
		Slot:           4,
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {
			log.Println("lockTeeTime", err)
		}
	}
}

func getKeyTeeTimeLockRedis(bookingDate, courseUid, teeTime, teeType string) string {
	teeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + bookingDate + "_" + courseUid + "_"
	teeTimeRedisKey += teeType + "_" + teeTime

	return teeTimeRedisKey
}

func getKeyTeeTimeSlotRedis(bookingDate, courseUid, teeTime, teeType string) string {
	teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty:" + bookingDate + "_" + courseUid + "_"
	teeTimeSlotEmptyRedisKey += teeType + "_" + teeTime

	return teeTimeSlotEmptyRedisKey
}

func getKeyTeeTimeRowIndex(bookingDate, courseUid, teeTime, teeType string) string {
	teeRowIndexTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_row_index:" + bookingDate + "_" + courseUid + "_"
	teeRowIndexTimeRedisKey += teeType + "_" + teeTime

	return teeRowIndexTimeRedisKey
}

func getTeeTimeList(courseUid, partnerUid, bookingDate string) []string {
	db := datasources.GetDatabaseWithPartner(partnerUid)
	form := request.GetListBookingSettingForm{
		CourseUid:  courseUid,
		PartnerUid: partnerUid,
		OnDate:     bookingDate,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	bookingDateTime, _ := time.Parse(constants.DATE_FORMAT_1, bookingDate)
	weekday := strconv.Itoa(int(bookingDateTime.Weekday() + 1))
	bookSetting := model_booking.BookingSetting{}

	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	timeParts := []response.TeeTimePartOTA{
		{
			IsHideTeePart: bookSetting.IsHideTeePart1,
			StartPart:     bookSetting.StartPart1,
			EndPart:       bookSetting.EndPart1,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart2,
			StartPart:     bookSetting.StartPart2,
			EndPart:       bookSetting.EndPart2,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart3,
			StartPart:     bookSetting.StartPart3,
			EndPart:       bookSetting.EndPart3,
		},
	}

	index := 0
	teeTimeListLL := []string{}

	for _, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1

				hour := teeTimeInit.Hour()
				minute := teeTimeInit.Minute()

				hourStr_ := strconv.Itoa(hour)
				if hour < 10 {
					hourStr_ = "0" + hourStr_
				}
				minuteStr := strconv.Itoa(minute)
				if minute < 10 {
					minuteStr = "0" + minuteStr
				}

				hourStr := hourStr_ + ":" + minuteStr

				teeTimeListLL = append(teeTimeListLL, hourStr)
				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.Unix() > endTime.Unix() {
					break
				}
			}
		}
	}
	return teeTimeListLL
}
func getIdGroup(s []models.CaddieGroup, e string) int64 {
	for _, v := range s {
		if v.Code == e {
			return v.Id
		}
	}
	return 0
}
func getBuggyFeeSetting(PartnerUid, CourseUid, GuestStyle string, Hole int) models.BuggyFeeItemSettingResponse {
	db := datasources.GetDatabaseWithPartner(PartnerUid)
	buggyFeeSettingR := models.BuggyFeeSetting{
		PartnerUid: PartnerUid,
		CourseUid:  CourseUid,
	}

	listBuggySetting, _, _ := buggyFeeSettingR.FindAll(db)
	buggyFeeSetting := models.BuggyFeeSetting{}
	for _, item := range listBuggySetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeSetting = item
			break
		}
	}

	buggyFeeItemSettingR := models.BuggyFeeItemSetting{
		PartnerUid: PartnerUid,
		CourseUid:  CourseUid,
		GuestStyle: GuestStyle,
		SettingId:  buggyFeeSetting.Id,
	}
	listSetting, _, _ := buggyFeeItemSettingR.FindAll(db)
	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	for _, item := range listSetting {
		if item.Status == constants.STATUS_ENABLE {
			buggyFeeItemSetting = item
			break
		}
	}

	rentalFee := utils.GetFeeFromListFee(buggyFeeItemSetting.RentalFee, Hole)
	privateCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.PrivateCarFee, Hole)
	oddCarFee := utils.GetFeeFromListFee(buggyFeeItemSetting.OddCarFee, Hole)

	return models.BuggyFeeItemSettingResponse{
		RentalFee:     rentalFee,
		PrivateCarFee: privateCarFee,
		OddCarFee:     oddCarFee,
	}
}

func getBookingCadieFeeSetting(PartnerUid, CourseUid, GuestStyle string, Hole int) models.BookingCaddyFeeSettingRes {
	db := datasources.GetDatabaseWithPartner(PartnerUid)
	// Get Buggy Fee
	bookingCaddieFeeSettingR := models.BookingCaddyFeeSetting{
		PartnerUid: PartnerUid,
		CourseUid:  CourseUid,
	}

	listBookingBuggyCaddySetting, _, _ := bookingCaddieFeeSettingR.FindList(db, models.Page{}, false)
	bookingCaddieFeeSetting := models.BookingCaddyFeeSetting{}
	for _, item := range listBookingBuggyCaddySetting {
		if item.Status == constants.STATUS_ENABLE {
			bookingCaddieFeeSetting = item
		}
	}

	return models.BookingCaddyFeeSettingRes{
		Fee:  bookingCaddieFeeSetting.Fee,
		Name: bookingCaddieFeeSetting.Name,
	}
}

func checkForCheckOut(bag model_booking.Booking) (bool, string) {
	db := datasources.GetDatabaseWithPartner(bag.PartnerUid)
	isCanCheckOut := false
	errMessage := "ok"

	if bag.BagStatus == constants.BAG_STATUS_TIMEOUT || bag.BagStatus == constants.BAG_STATUS_WAITING {
		isCanCheckOut = true

		// Check service items
		// Find bag detail
		if isCanCheckOut {
			// Check tiep service items
			bagDetail := getBagDetailFromBooking(db, bag)
			if bagDetail.ListServiceItems != nil && len(bagDetail.ListServiceItems) > 0 {
				for _, v1 := range bagDetail.ListServiceItems {
					serviceCart := models.ServiceCart{}
					serviceCart.Id = v1.ServiceBill

					errSC := serviceCart.FindFirst(db)
					if errSC != nil {
						log.Println("FindFristServiceCart errSC", errSC.Error())
						return false, "FindFristServiceCart errSC"
					}

					// Check trong MainBag có trả mới add
					if v1.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
						// ok
					} else {
						if serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH ||
							serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE ||
							serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
							serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT {
							// ok
						} else {
							if v1.BillCode != bag.BillCode {
								errMessage = "Dich vụ của sub-bag chưa đủ điều kiện được checkout"
							} else {
								errMessage = "Dich vụ của bag chưa đủ điều kiện được checkout"
							}

							isCanCheckOut = false
							break
						}
					}
				}
			}
		}
	} else {
		isCanCheckOut = false
		errMessage = "Trạng thái bag không được checkout"
	}

	//Check sub bag
	// if bag.SubBags != nil && len(bag.SubBags) > 0 && isCanCheckOut {
	// 	for _, v := range bag.SubBags {
	// 		subBag := model_booking.Booking{}
	// 		subBag.Uid = v.BookingUid
	// 		errF := subBag.FindFirst(db)

	// 		if errF == nil {
	// 			if subBag.BagStatus == constants.BAG_STATUS_CHECK_OUT || subBag.BagStatus == constants.BAG_STATUS_CANCEL {
	// 			} else {
	// 				errMessage = "Sub-bag chưa check checkout"
	// 				isCanCheckOut = false
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	return isCanCheckOut, errMessage
}

func deleteBuggyFee(booking model_booking.Booking) {
	db := datasources.GetDatabaseWithPartner(booking.PartnerUid)
	bookingServiceItemsR := model_booking.BookingServiceItem{
		PartnerUid: booking.PartnerUid,
		CourseUid:  booking.CourseUid,
		BillCode:   booking.BillCode,
	}

	list, _ := bookingServiceItemsR.FindAll(db)
	name1 := "Thuê xe (1/2 xe)"
	name2 := "Thuê lẻ xe"
	name3 := "Thuê riêng xe"

	for _, item := range list {
		if item.ServiceType == constants.BUGGY_SETTING {
			if item.Name == name1 {
				item.Delete(db)
				name1 = ""
			}

			if item.Name == name2 {
				item.Delete(db)
				name2 = ""
			}

			if item.Name == name3 {
				item.Delete(db)
				name3 = ""
			}
		}
	}
}
func validateBooking(db *gorm.DB, bookindUid string) (model_booking.Booking, error) {
	bookingR := model_booking.Booking{}
	bookingR.Uid = bookindUid
	booking, err := bookingR.FindFirstByUId(db)
	if err != nil {
		return booking, err
	}

	if *booking.LockBill {
		return booking, errors.New("Bag " + booking.Bag + " đã lock")
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		return booking, errors.New("Bag " + booking.Bag + " đã check out!")
	}

	return booking, nil
}

func updateAgencyInfoInPayment(db *gorm.DB, booking model_booking.Booking) {
	agency := model_payment.AgencyPayment{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingCode: booking.BookingCode,
	}

	if errFindAgency := agency.FindFirst(db); errFindAgency == nil {
		agency.AgencyId = booking.AgencyId
		agency.PlayerBook = booking.CustomerBookingName
		agency.AgencyInfo = model_payment.PaymentAgencyInfo{
			Name:           booking.AgencyInfo.Name,
			GuestStyle:     booking.GuestStyle,
			GuestStyleName: booking.GuestStyleName,
		}
		agency.Update(db)
	}
}
