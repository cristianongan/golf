package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"start/config"
	"start/constants"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gitee.com/mirrors/govaluate"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/google/uuid"
	"github.com/leekchan/accounting"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/ttacon/libphonenumber"
)

/*
  dùng redis để check single payment đã dc tạo chưa(Check ở sql chưa đủ realtime)
*/
func GetRedisKeySinglePaymentCreated(partnerUid, courseUid, billCode string) string {
	singlePaymentRedisKey := config.GetEnvironmentName() + ":" + "single_payment:" + "_" + partnerUid + "_" + courseUid + "_" + billCode

	return singlePaymentRedisKey
}

/*
  dùng redis để check agency payment đã dc tạo chưa(Check ở sql chưa đủ realtime)
*/
func GetRedisKeyAgencyPaymentCreated(partnerUid, courseUid, bookCode string) string {
	agencyPaymentRedisKey := config.GetEnvironmentName() + ":" + "agency_payment:" + "_" + partnerUid + "_" + courseUid + "_" + bookCode

	return agencyPaymentRedisKey
}

func GetDateLocal() time.Time {
	dateDisplay, _ := GetBookingDateFromTimestamp(time.Now().Unix())
	applyDate, _ := time.Parse(constants.DATE_FORMAT_1, dateDisplay)
	return applyDate
}

func GetCurrentYear() string {
	currentYearStr, _ := GetDateFromTimestampWithFormat(GetTimeNow().Unix(), constants.YEAR_FORMAT)
	return currentYearStr
}

func GetCurrentDay() string {
	currentDayStr, _ := GetDateFromTimestampWithFormat(GetTimeNow().Unix(), constants.DATE_FORMAT)
	return currentDayStr
}

func GetCurrentDay1() string {
	currentDayStr, _ := GetDateFromTimestampWithFormat(GetTimeNow().Unix(), constants.DATE_FORMAT_1)
	return currentDayStr
}

func HashCodeUuid(uid string) string {
	return NumberToString(uuid.MustParse(uid).ID())
}

func FormatPhone(input string) (string, error) {
	num, errPhone := libphonenumber.Parse(input, "VN")
	if errPhone != nil {
		return "", errPhone
	}
	strPhoneNumber := "+" + fmt.Sprint(num.GetCountryCode()) + fmt.Sprint(num.GetNationalNumber())
	return strPhoneNumber, nil
}

func Sum(v []int) int {
	output := 0
	for _, num := range v {
		output += num
	}
	return output
}

func GetTimeStampFromLocationTime(location, formatTime, localTime string) int64 { //
	if location == "" {
		location = constants.LOCATION_DEFAULT
	}
	if formatTime == "" {
		formatTime = constants.DATE_TIME_FORMAT
	}
	loc, errLoc := time.LoadLocation(location)
	if errLoc != nil {
		log.Println(errLoc)
		return 0
	}
	time, errParse := time.ParseInLocation(formatTime, localTime, loc)
	if errParse != nil {
		log.Println(errParse)
		return 0
	}
	return time.Unix()
}

func GetLocalTimeFromTimeStamp(location, format string, timeStamp int64) (string, error) {
	loc, errLoc := time.LoadLocation(location)
	if errLoc != nil {
		log.Println(errLoc)
		return "", errLoc
	}
	tm := time.Unix(timeStamp, 0).In(loc)
	return tm.Format(format), nil
}

func GetBookingTimeFrom(timeStr string) (string, error) {
	location := constants.LOCATION_DEFAULT
	formatTime := constants.DATE_FORMAT
	loc, errLoc := time.LoadLocation(location)
	if errLoc != nil {
		log.Println(errLoc)
		return "", errLoc
	}
	time, errParse := time.ParseInLocation(formatTime, timeStr, loc)
	if errParse != nil {
		log.Println(errParse)
		return "", errParse
	}

	localTime, errLocalTime := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, time.Unix())
	if errLocalTime != nil {
		return "", errLocalTime
	}
	return localTime, nil
}

func GetYearMonthDateFromTimestamp(timeStamp int64) (string, error) {
	// date, errDate := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT, timeStamp)
	// month, errMonth := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.MONTH_FORMAT, timeStamp)
	// year, errYear := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.YEAR_FORMAT, timeStamp)
	localTime, errLocalTime := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_TIME_FORMAT, timeStamp)
	// if errDate != nil {
	// 	return "", "", "", "", errDate
	// }
	// if errMonth != nil {
	// 	return "", "", "", "", errMonth
	// }
	// if errYear != nil {
	// 	return "", "", "", "", errYear
	// }
	if errLocalTime != nil {
		return "", errLocalTime
	}
	return localTime, nil
}

func GetBookingDateFromTimestamp(timeStamp int64) (string, error) {
	localTime, errLocalTime := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, timeStamp)
	if errLocalTime != nil {
		return "", errLocalTime
	}
	return localTime, nil
}

func GetDateFromTimestampWithFormat(timeStamp int64, format string) (string, error) {
	localTime, errLocalTime := GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, format, timeStamp)
	if errLocalTime != nil {
		return "", errLocalTime
	}
	return localTime, nil
}

// ================================
func convertUtf8ToUnicode(s string) string {
	myFunc := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(myFunc), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func TimeStampMilisecond(nanoTimeStamp int64) int64 {
	return nanoTimeStamp / int64(time.Millisecond)
}

// ==================================
func NumberToString(number interface{}) string {
	return fmt.Sprint(number)
}

func ValueToFloat64(number interface{}) float64 {
	return number.(float64)
}

func StringToInt64(number string) (int64, error) {
	i, err := strconv.ParseInt(number, 10, 64)
	return i, err
}

func GetStartDayByLocation(location string) int64 {
	loc, _ := time.LoadLocation(location)
	t := GetTimeNow().In(loc)
	year, month, day := t.Date()
	rounded := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return rounded.Unix()
}

func GetStartDayByTimeStamp(timestamp int64, location string) int64 {
	loc, _ := time.LoadLocation(location)
	t := time.Unix(timestamp, 0).In(loc)
	year, month, day := t.Date()
	rounded := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return rounded.Unix()
}

func UuidElasticSearch() string {
	return uuid.New().String() + "-" + NumberToString(GetTimeNow().UnixNano())
}

func RandStringBytes(n int, letter string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
func StructToJson(a interface{}) string {
	b, _ := json.Marshal(a)
	return string(b)
}

func CheckListAEqualListB(listA, listB []int64) bool {
	sort.Slice(listA, func(i, j int) bool {
		return listA[i] < listA[j]
	})
	sort.Slice(listB, func(i, j int) bool {
		return listB[i] < listB[j]
	})

	if len(listA) != len(listB) {
		return false
	}
	for index, item := range listA {
		if listB[index] != item {
			return false
		}
	}
	return true
}

func StringInList(a string, list []string) int {
	for index, b := range list {
		if b == a {
			return index
		}
	}
	return -1
}

// output:
// Xh, Yh, H trong đoạn AB-> true,
func GetPerpendicularProjection(Xa, Ya, Xb, Yb, Xo, Yo float64) (float64, float64, bool) {
	Xba := Xb - Xa
	Yba := Yb - Ya
	Yoa := Yo - Ya
	Xh := (Yba*Yba*Xa + Xba*Xba*Xo + Yoa*Xba*Yba) / (Xba*Xba + Yba*Yba)
	Yh := (Xba*Xo + Yba*Yo - Xba*Xh) / Yba

	Xha := Xh - Xa
	Yha := Yh - Ya
	Xhb := Xh - Xb
	Yhb := Yh - Yb
	tichVoHuong := Xha*Xhb + Yha*Yhb

	return Xh, Yh, tichVoHuong < 0
}

func CalculateDistance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515 // miles
	dist = dist * 1.609344    // km
	dist = dist * 1000        // m

	return dist
}

func FormatMoney(inputMoney float64) string {
	acc := accounting.Accounting{Precision: 0}
	return acc.FormatMoneyFloat64(inputMoney)
}

func GenerateUidTimestamp() string {
	uid := uuid.New()
	return uid.String() + "-" + NumberToString(GetTimeNow().UnixNano())
}

// ================================================================================
func GeneratePassword(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	return string(hash), err
}

func ComparePassword(hashPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}

func ConvertHourToTime(hourStr string) (time.Time, error) {
	t, err := time.Parse(constants.HOUR_FORMAT, hourStr)
	if err != nil {
		log.Println("ConvertHourToTime err", err.Error())
		return t, err
	}
	return t, nil
}

/*
Check ngày và h
*/
func CheckDow(dow, hour string, timeCheck time.Time) bool {
	if dow == "" {
		return false
	}

	list := strings.Split(dow, "")
	// log.Println("Check Dow ", list, len(list))

	if len(list) == 0 {
		return false
	}
	isOk := false
	for _, v := range list {
		dayInt, err := strconv.Atoi(v)
		if err != nil {
			log.Println("CheckDow err", err.Error())
		}
		if dayInt == int(timeCheck.Weekday()+1) {
			if hour != "" {
				if CheckHour(hour, timeCheck) {
					isOk = true
				}
			} else {
				isOk = true
			}
		}
	}

	return isOk
}

/*
Check giờ: format 13:00,23:00
*/
func CheckHour(hour string, timeCheck time.Time) bool {

	currentHour := timeCheck.Hour()
	currentMinute := timeCheck.Minute()

	// Parse Hour
	fromHour := -1
	fromMinute := -1
	toHour := -1
	toMinute := -1
	if strings.Contains(hour, ",") {
		listH := strings.Split(hour, ",")
		for i, v := range listH {
			if i == 0 {
				timeHour, err := ConvertHourToTime(v)
				if err == nil {
					fromHour = timeHour.Hour()
					fromMinute = timeHour.Minute()
				} else {
					log.Println("CheckHour err0", err.Error())
				}
			} else if i == 1 {
				timeHour, err := ConvertHourToTime(v)
				if err == nil {
					toHour = timeHour.Hour()
					toMinute = timeHour.Minute()
				} else {
					log.Println("CheckHour err1", err.Error())
				}
			}
		}
	}

	if fromHour >= 0 && toHour == -1 {
		if currentHour > fromHour {
			return true
		}
		if currentHour == fromHour && currentMinute >= fromMinute {
			return true
		}
	}

	if fromHour == -1 && toHour >= 0 {
		if currentHour < toHour {
			return true
		}
		if currentHour == toHour && currentMinute <= toMinute {
			return true
		}
	}
	if fromHour >= 0 && toHour >= 0 {
		if fromHour <= currentHour && currentHour <= toHour {
			return true
		}

	}
	return false
}

func GetFeeFromListFee(feeList ListGolfHoleFee, hole int) int64 {
	fee := int64(0)

	roundedHole := RoundHole(hole)
	for _, feeModel := range feeList {
		if feeModel.Hole == roundedHole {
			fee = feeModel.Fee
		}
	}

	return fee
}

func RoundHole(hole int) int {
	if hole >= 0 && hole <= 2 {
		return 0
	} else if hole > 2 && hole <= 9 {
		return 9
	} else if hole > 9 && hole <= 18 {
		return 18
	} else if hole > 18 && hole <= 27 {
		return 27
	} else if hole > 27 && hole <= 36 {
		return 36
	} else if hole > 36 && hole <= 45 {
		return 45
	} else if hole > 45 && hole <= 54 {
		return 54
	} else if hole > 54 && hole <= 63 {
		return 63
	}
	return 72
}

func IsDateValue(stringDate string) bool {
	_, err := time.Parse("01/02/2006", stringDate)
	return err == nil
}

func IsWeekend(ti int64) bool {
	t := time.Unix(ti, 0).Local()
	switch t.Weekday() {
	// case time.Friday:
	//     h, _, _ := t.Clock()
	//     if h >= 12+10 {
	//         return true
	//     }
	case time.Saturday:
		return true
	case time.Sunday:
		return true
	}
	return false
}
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func IndexOf[T comparable](s []T, e T) int {
	for k, v := range s {
		if v == e {
			return k
		}
	}
	return -1
}

func Remove[T comparable](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func CheckDupArray[T comparable](arr []T) []T {
	visited := make(map[T]bool, 0)
	var listDup []T

	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] {
			listDup = append(listDup, arr[i])
		} else {
			visited[arr[i]] = true
		}
	}
	return listDup
}

func SwapValue[T comparable](s []T, o, n T) []T {
	indexOld := IndexOf(s, o)
	indexNew := IndexOf(s, n)

	s[indexOld], s[indexNew] = s[indexNew], s[indexOld]

	return s
}

func removeDuplicateStr(str []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, item := range str {
		if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func GetFeeWidthHolePrice(feeList ListGolfHoleFee, hole int, formula string) int64 {
	re := regexp.MustCompile(`(gia)\w+`)

	confifFeeRaw := re.FindAllString(formula, -1)

	confifFees := removeDuplicateStr(confifFeeRaw)

	expression, err := govaluate.NewEvaluableExpression(formula)

	if err != nil {
		log.Println("NewEvaluableExpression err", err.Error())
		return 0
	}

	parameters := make(map[string]interface{}, 8)

	parameters["ho"] = hole

	for _, item := range confifFees {
		hole, err := strconv.Atoi(item[3:])
		if err != nil {
			log.Println("Convert string to int err", err.Error())
			return 0
		}

		parameters[item] = GetFeeFromListFee(feeList, hole)
	}

	result, _ := expression.Evaluate(parameters)

	return int64(result.(float64))
}

func CalculateFeeByHole(hole int, fee int64, rateRaw string) int64 {
	re := regexp.MustCompile(`(\d[.]\d)|(\d)+`)
	roundedHole := RoundHole(hole)

	index := (roundedHole / 9) - 1
	listRate := re.FindAllString(rateRaw, -1)

	rate := listRate[index]

	parseRate, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		log.Println("Convert string to int64 err", err.Error())
		return fee
	}

	return int64(float64(fee) * parseRate)
}

/*
	 trong Go thì Sunday = 0
			// A Weekday specifies a day of the week (Sunday = 0, ...).
		type Weekday int

		const (
			Sunday Weekday = iota
			Monday
			Tuesday
			Wednesday
			Thursday
			Friday
			Saturday
		)
		Theo mockup dự án
		Note

# D.O.W được quy định như sau

là cấu hình bảng giá theo thứ

1/ Chủ nhật

2/Thứ 2

3/Thứ 3

4/Thứ 4

5/Thứ 5

6/ Thứ 6

7/ Thứ 7

0/ Ngày lễ, ngày nghỉ
*/
func GetCurrentDayStrWithMap() string {
	day := strconv.FormatInt(int64(GetLocalUnixTime().Weekday())+1, 10)
	// log.Println("GetCurrentDayStrWithMap ", day)
	return day
}

func GetDayOfWeek(strTime string) string {
	date, err := time.Parse("02/01/2006", strTime)
	if err != nil {
		return ""
	}

	day := strconv.FormatInt(int64(date.Weekday())+1, 10)
	// log.Println("GetCurrentDayStrWithMap ", day)
	return day
}

func RandomCharNumber(length int) string {
	id, err := gonanoid.Generate("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 5)
	if err != nil {
		return ""
	}
	return id
}

func RandomCharNumberV2(length int) string {
	id, err := gonanoid.Generate("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz", length)
	if err != nil {
		return ""
	}
	return id
}

func VerifyPassword(s string) (bool, bool, bool, bool) {
	eightOrMore := false
	number := false
	upper := false
	special := false

	if len(s) >= 8 {
		eightOrMore = true
	}

	letters := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}
	// eightOrMore = letters >= 8
	return eightOrMore, number, upper, special
}

func ConvertStringToIntArray(data string) ListInt {
	if data == "" {
		return ListInt{}
	}
	trimmed := strings.Trim(data, "[]")
	strings := strings.Split(trimmed, ",")
	ints := make([]int, len(strings))

	for i, s := range strings {
		ints[i], _ = strconv.Atoi(s)
	}

	return ints
}

func RemoveIndex(s []int, index int) []int {
	return append(s[:index], s[index+1:]...)
}

func GetTimeNow() time.Time {
	// hours, _ := GetDateFromTimestampWithFormat(time.Now().Add(time.Hour*time.Duration(-7)+
	// 	time.Minute*time.Duration(0)+
	// 	time.Second*time.Duration(0)).Unix(), constants.HOUR_FORMAT_1)
	// time, _ := time.Parse(constants.DATE_FORMAT_4, "11/02/2023 "+hours)
	return time.Now()
}

func GetLocalUnixTime() time.Time {
	loc, errLoc := time.LoadLocation(constants.LOCATION_DEFAULT)
	if errLoc != nil {
		return GetTimeNow()
	}
	tm := GetTimeNow().In(loc)
	return tm
}
