package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"time"

	"gorm.io/datatypes"
)

func runCreateCaddieWorkingSlotJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerCreateCaddieWorkingSlot(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runCreateCaddieWorkingSlot()
}

// Tạo nốt caddie theo ngày
func runCreateCaddieWorkingSlot() {
	db := datasources.GetDatabase()

	var dataGroupWorking []int64
	var slotPrioritize []int64

	// statusFull := []string{constants.CADDIE_CONTRACT_STATUS_FULLTIME}
	// statusAll := []string{constants.CADDIE_CONTRACT_STATUS_FULLTIME, constants.CADDIE_CONTRACT_STATUS_PARTTIME}

	// Format date
	dateNow, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
	dateConvert, _ := time.Parse(constants.DATE_FORMAT_1, dateNow)
	dayNow := int(dateConvert.Weekday())

	// Get group caddie work today
	applyDate1 := datatypes.Date(dateConvert)
	idDayOff1 := false

	// get caddie work sechedule
	caddieWCN := models.CaddieWorkingSchedule{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
		ApplyDate:  &(applyDate1),
		IsDayOff:   &idDayOff1,
	}

	listCWSNow, err := caddieWCN.FindListWithoutPage(db)
	if err != nil {
		log.Println("Find list caddie working schedule today", err.Error())
	}

	var listCWSYes []models.CaddieWorkingSchedule

	if dayNow != 6 && dayNow != 0 {
		// get group caddie day off yesterday
		var dateYesterday string

		if dayNow == 1 {
			dateYesterday, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -3).Unix())
		} else {
			dateYesterday, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -1).Unix())
		}

		dateConvert2, _ := time.Parse(constants.DATE_FORMAT_1, dateYesterday)
		applyDate2 := datatypes.Date(dateConvert2)
		idDayOff2 := true

		// get caddie work sechedule
		caddieWSY := models.CaddieWorkingSchedule{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			ApplyDate:  &(applyDate2),
			IsDayOff:   &idDayOff2,
		}

		listCWSYes, err = caddieWSY.FindListWithoutPage(db)
		if err != nil {
			log.Println("Find frist caddie working schedule", err.Error())
		}
	}

	//get all group
	caddieGroup := models.CaddieGroup{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
	}

	listCaddieGroup, err := caddieGroup.FindListWithoutPage(db)
	if err != nil {
		log.Println("Find frist caddie working schedule", err.Error())
	}

	//add group caddie
	for _, item := range listCWSNow {
		id := getIdGroup(listCaddieGroup, item.CaddieGroupCode)

		if id > 0 {
			// Check group prioritize
			check := ContainsCaddie(listCWSYes, item.CaddieGroupCode)

			if check {
				slotPrioritize = append(slotPrioritize, id)
			} else {
				dataGroupWorking = append(dataGroupWorking, id)
			}
		}
	}

	//Check caddie vacation today
	caddieVC := models.CaddieVacationCalendar{
		PartnerUid:    "CHI-LINH",
		CourseUid:     "CHI-LINH-01",
		ApproveStatus: constants.CADDIE_VACATION_APPROVED,
	}

	// Caddie nghỉ hôm nay
	listCVCLeave, err := caddieVC.FindAllWithDate(db, "LEAVE", dateConvert)

	if err != nil {
		log.Println("Find caddie vacation calendar err", err.Error())
	}

	// Caddie nghỉ hôm qua và đi làm hôm nay
	listCVCWork, err := caddieVC.FindAllWithDate(db, "WORK", dateConvert)

	if err != nil {
		log.Println("Find caddie vacation calendar err", err.Error())
	}

	// Get caddie code
	var caddiePrioritize []string
	var caddieWorking []string
	caddieWork := GetCaddieCodeFromVacation(listCVCWork)
	caddieLeave := GetCaddieCodeFromVacation(listCVCLeave)

	caddies := models.Caddie{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
	}

	if len(slotPrioritize) > 0 {
		listCaddies, err := caddies.FindAllCaddieGroup(db, constants.CADDIE_CONTRACT_STATUS_FULLTIME, slotPrioritize)

		if err != nil {
			log.Println("Find all caddie group err", err.Error())
		}

		caddieCodes := GetCaddieCode(listCaddies)

		// Lấy data xếp nốt
		var applyDate string

		if dayNow == 1 || dayNow == 2 {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -4).Unix())
		} else {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -2).Unix())
		}

		caddieSlot := models.CaddieWorkingSlot{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			ApplyDate:  applyDate,
		}

		err = caddieSlot.FindFirst(db)

		if err != nil {
			caddiePrioritize = append(caddiePrioritize, caddieCodes...)
		} else {
			caddieMerge := MergeCaddieCode(caddieSlot.CaddieSlot, caddieCodes, caddieLeave)

			caddiePrioritize = append(caddiePrioritize, caddieMerge...)
		}
	}

	if len(dataGroupWorking) > 0 && dayNow != 6 && dayNow != 0 {
		listCaddies, err := caddies.FindAllCaddieGroup(db, constants.CADDIE_CONTRACT_STATUS_FULLTIME, dataGroupWorking)

		if err != nil {
			log.Println("Find all caddie group err", err.Error())
		}

		caddieCodes := GetCaddieCode(listCaddies)

		// Lấy data xếp nốt
		var applyDate string

		if dayNow == 1 {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -3).Unix())
		} else {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -1).Unix())
		}

		caddieSlot := models.CaddieWorkingSlot{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			ApplyDate:  applyDate,
		}

		err = caddieSlot.FindFirst(db)

		if err != nil {
			caddieWorking = append(caddieWorking, caddieCodes...)
		} else {
			caddieMerge := MergeCaddieCode(caddieSlot.CaddieSlot, caddieCodes, caddieLeave)

			caddieWorking = append(caddieWorking, caddieMerge...)
		}
	}

	if len(dataGroupWorking) > 0 && (dayNow == 6 || dayNow == 0) {
		listCaddiesFull, err := caddies.FindAllCaddieGroup(db, constants.CADDIE_CONTRACT_STATUS_FULLTIME, dataGroupWorking)

		if err != nil {
			log.Println("Find all caddie group err", err.Error())
		}

		listCaddiesPart, err := caddies.FindAllCaddieGroup(db, constants.CADDIE_CONTRACT_STATUS_PARTTIME, dataGroupWorking)

		if err != nil {
			log.Println("Find all caddie group err", err.Error())
		}

		caddieCodes := append(listCaddiesFull, listCaddiesPart...)

		caddieSortSlots := GetCaddieCode(caddieCodes)

		// Lấy data xếp nốt
		var applyDate string

		if dayNow == 6 {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -6).Unix())
		} else {
			applyDate, _ = utils.GetBookingDateFromTimestamp(dateConvert.AddDate(0, 0, -1).Unix())
		}

		caddieSlot := models.CaddieWorkingSlot{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			ApplyDate:  applyDate,
		}

		err = caddieSlot.FindFirst(db)

		if err != nil {
			caddieWorking = append(caddieWorking, caddieSortSlots...)
		} else {
			caddieMerge := MergeCaddieCode(caddieSlot.CaddieSlot, caddieSortSlots, caddieLeave)

			caddieWorking = append(caddieWorking, caddieMerge...)
		}
	}

	slotCaddie := GetListCaddie(caddiePrioritize, caddieWork, caddieWorking)

	caddieSlot := models.CaddieWorkingSlot{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
		ApplyDate:  dateNow,
	}

	if !caddieSlot.IsDuplicated(db) {
		caddieSlot.CaddieSlot = slotCaddie

		err = caddieSlot.Create(db)
		if err != nil {
			log.Println("Create report caddie err", err.Error())
		}

		for _, caddieCode := range slotCaddie {
			caddie := models.Caddie{
				PartnerUid: "CHI-LINH",
				CourseUid:  "CHI-LINH-01",
				Code:       caddieCode,
			}

			if err = caddie.FindFirst(db); err == nil {
				caddie.IsWorking = 1
				caddie.Update(db)
			}
		}
	}

}

func ContainsCaddie(s []models.CaddieWorkingSchedule, e string) bool {
	for _, v := range s {
		if v.CaddieGroupCode == e {
			return true
		}
	}
	return false
}

func getIdGroup(s []models.CaddieGroup, e string) int64 {
	for _, v := range s {
		if v.Code == e {
			return v.Id
		}
	}
	return 0
}

func GetCaddieCodeFromVacation(s []models.CaddieVacationCalendar) []string {
	var caddies []string
	for _, v := range s {
		caddies = append(caddies, v.CaddieCode)
	}
	return caddies
}

func GetCaddieCode(s []models.Caddie) []string {
	var caddies []string
	for _, v := range s {
		caddies = append(caddies, v.Code)
	}
	return caddies
}

func MergeCaddieCode(x, y, z []string) []string {
	var caddies []string
	var caddieNew []string

	// Sort caddie with old slot
	for _, v := range x {
		if utils.Contains(y, v) && !utils.Contains(z, v) {
			caddies = append(caddies, v)
		}
	}

	// Add caddie new without slot
	for _, v := range y {
		if !utils.Contains(x, v) && !utils.Contains(z, v) {
			caddieNew = append(caddieNew, v)
		}
	}

	caddies = append(caddies, caddieNew...)

	return caddies
}

func GetListCaddie(x, y, z []string) []string {
	var caddies []string

	caddies = append(caddies, x...)

	for _, v := range y {
		if !utils.Contains(caddies, v) {
			caddies = append(caddies, v)
		}
	}

	for _, v := range z {
		if !utils.Contains(caddies, v) {
			caddies = append(caddies, v)
		}
	}

	return caddies
}
