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
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils"
	"strings"
	"time"

	model_report "start/models/report"

	"github.com/gin-gonic/gin"
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

func checkDuplicateGolfFee(db *gorm.DB, body models.GolfFee) bool {
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
			booking.UpdatePriceForBagHaveMainBags(db)
			errUdp := booking.Update(db)
			if errUdp != nil {
				err = errUdp
				log.Println("UpdateMainBagForSubBag errUdp", errUdp.Error())
			} else {
				// Udp lai info payment
				handlePayment(db, booking)
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
		booking.IsPrivateBuggy = newTrue(isPrivateBuggy)
		booking.BuggyInfo = cloneToBuggyBooking(buggy)
	}

	booking.ShowCaddieBuggy = newTrue(true)
	response.NewCaddie = caddie
	response.NewBuggy = buggy
	response.Booking = booking
	return nil, response
}

/*
Out caddie
*/
func udpOutCaddieBooking(db *gorm.DB, booking *model_booking.Booking) error {

	errCd := udpCaddieOut(db, booking.CaddieId)
	if errCd != nil {
		return errCd
	}
	// Udp booking
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_OUT

	return nil
}

/*
Out Buggy
*/
func udpOutBuggy(db *gorm.DB, booking *model_booking.Booking, isOutAll bool) error {
	// Get Caddie

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

	return nil
}

/*
Update caddie is in course is false
*/
func udpCaddieOut(db *gorm.DB, caddieId int64) error {
	// Get Caddie
	caddie := models.Caddie{}
	caddie.Id = caddieId
	err := caddie.FindFirst(db)
	if caddie.CurrentRound == 0 {
		caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
	} else {
		if caddie.CurrentRound > 1 {
			caddie.CurrentStatus = fmt.Sprintln(constants.CADDIE_CURRENT_STATUS_FINISH, "R", caddie.CurrentRound)
		} else {
			caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_FINISH
		}
	}
	err = caddie.Update(db)
	if err != nil {
		log.Println("udpCaddieOut err", err.Error())
	}
	return err
}

/*
add Caddie In Out Note
*/
func addBuggyCaddieInOutNote(db *gorm.DB, caddieInOut model_gostarter.CaddieBuggyInOut) {
	newCaddieInOut := model_gostarter.CaddieBuggyInOut{
		PartnerUid: caddieInOut.PartnerUid,
		CourseUid:  caddieInOut.CourseUid,
		BookingUid: caddieInOut.BookingUid,
	}

	list, total, _ := newCaddieInOut.FindOrderByDateList(db)
	if total > 0 {
		lastItem := list[0]
		if caddieInOut.BuggyId > 0 && caddieInOut.CaddieId > 0 {
			err := caddieInOut.Create(db)
			if err != nil {
				log.Println("Create addBuggyCaddieInOutNote", err.Error())
			}
		} else {
			if (lastItem.BuggyId > 0 && lastItem.CaddieId > 0) ||
				lastItem.CaddieType == constants.STATUS_OUT ||
				lastItem.BuggyType == constants.STATUS_OUT ||
				caddieInOut.BuggyType == constants.STATUS_OUT ||
				caddieInOut.CaddieType == constants.STATUS_OUT {

				if caddieInOut.BuggyId > 0 && caddieInOut.CaddieId == 0 &&
					lastItem.CaddieType == constants.STATUS_IN {
					caddieInOut.CaddieId = lastItem.CaddieId
					caddieInOut.CaddieCode = lastItem.CaddieCode
					caddieInOut.CaddieType = lastItem.CaddieType
				} else if caddieInOut.CaddieId > 0 && caddieInOut.BuggyId == 0 &&
					lastItem.BuggyType == constants.STATUS_IN {
					caddieInOut.BuggyId = lastItem.BuggyId
					caddieInOut.BuggyCode = lastItem.BuggyCode
					caddieInOut.BuggyType = lastItem.BuggyType
				}

				err := caddieInOut.Create(db)
				if err != nil {
					log.Println("Create addBuggyCaddieInOutNote", err.Error())
				}
			} else {
				if caddieInOut.CaddieId > 0 {
					lastItem.CaddieId = caddieInOut.CaddieId
					lastItem.CaddieCode = caddieInOut.CaddieCode
					lastItem.CaddieType = caddieInOut.CaddieType
				}
				if caddieInOut.BuggyId > 0 {
					lastItem.BuggyId = caddieInOut.BuggyId
					lastItem.BuggyCode = caddieInOut.BuggyCode
					lastItem.BuggyType = caddieInOut.BuggyType
				}
				err := lastItem.Update(db)
				if err != nil {
					log.Println("Update addBuggyCaddieInOutNote", err.Error())
				}
			}
		}
	} else {
		err := caddieInOut.Create(db)
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
	cLockTeeTim.DeleteLockTurn(db, booking.TeeTime, booking.BookingDate)
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

	if booking.LockerNo != "" && locker.Locker != booking.LockerNo {
		locker.PlayerName = booking.CustomerName
		locker.Locker = booking.LockerNo
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
	errFC := customer.FindFirst(db)
	if errFC != nil {
		log.Println("handleBookingForMemberCard err", errFC.Error())
	}

	return nil, memberCard, customer.Name
}

func updateAnnualFeeToMcType(db *gorm.DB, yearInt int, mcTypeId, fee int64) {
	if time.Now().Year() == yearInt {
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
		caddie.CurrentStatus == constants.CADDIE_CURRENT_STATUS_FINISH) {
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
func updateReportTotalPlayCountForCustomerUser(userUid string, partnerUid, courseUid string) {
	reportCustomer := model_report.ReportCustomerPlay{
		CustomerUid: userUid,
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

	now := time.Now()

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
	proshop := model_service.Proshop{
		PartnerUid: partnerUid,
		CourseUid:  courseUid,
		ProShopId:  itemCode,
	}

	if errFindProshop := proshop.FindFirst(db); errFindProshop == nil {
		return kiosk_inventory.ItemInfo{
			GroupCode: proshop.GroupCode,
			ItemName:  proshop.Name,
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
			ItemName:  fb.Name,
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
			ItemName:  rental.Name,
			Unit:      rental.Unit,
			GroupType: rental.Type,
		}, nil
	}

	return kiosk_inventory.ItemInfo{}, errors.New(fmt.Sprintln(itemCode, "not found"))
}

func newTrue(b bool) *bool {
	boolVar := b
	return &boolVar
}

/*
Get Tee Time Lock Redis
*/
func getTeeTimeLockRedis(courseUid string, date string) []models.LockTeeTime {
	prefixRedisKey := config.GetEnvironmentName() + ":" + courseUid + "_" + date
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
	listTeeTimeLockRedis := []models.LockTeeTime{}
	if errRedis == nil && len(listKey) > 0 {
		strData, _ := datasources.GetCaches(listKey...)
		for _, data := range strData {

			byteData := []byte(data.(string))
			teeTime := models.LockTeeTime{}
			err2 := json.Unmarshal(byteData, &teeTime)
			if err2 == nil {
				listTeeTimeLockRedis = append(listTeeTimeLockRedis, teeTime)
			}
		}
	}
	return listTeeTimeLockRedis
}
