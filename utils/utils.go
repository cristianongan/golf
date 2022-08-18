package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"regexp"
	"sort"
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
	"github.com/ttacon/libphonenumber"
)

func GetCurrentYear() string {
	currentYearStr, _ := GetDateFromTimestampWithFormat(time.Now().Unix(), constants.YEAR_FORMAT)
	return currentYearStr
}

func GetCurrentDay() string {
	currentDayStr, _ := GetDateFromTimestampWithFormat(time.Now().Unix(), constants.DATE_FORMAT)
	return currentDayStr
}

func GetCurrentDay1() string {
	currentDayStr, _ := GetDateFromTimestampWithFormat(time.Now().Unix(), constants.DATE_FORMAT_1)
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
	t := time.Now().In(loc)
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
	return uuid.New().String() + "-" + NumberToString(time.Now().UnixNano())
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
	return uid.String() + "-" + NumberToString(time.Now().UnixNano())
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

func CheckDow(dow string, timeCheck time.Time) bool {
	if dow == "" {
		return false
	}

	list := strings.Split(dow, "")
	log.Println("Check Dow ", list, len(list))

	if len(list) == 0 {
		return false
	}
	isOk := false
	for _, v := range list {
		dayInt, err := strconv.Atoi(v)
		if err != nil {
			log.Println("CheckDow err", err.Error())
		}
		dayInt = dayInt - 1 // Vì Dow 0 là ngày lễ
		if dayInt == int(timeCheck.Weekday()) {
			isOk = true
		}
	}

	return isOk
}

func GetFeeFromListFee(feeList ListGolfHoleFee, hole int) int64 {
	fee := int64(0)

	for _, feeModel := range feeList {
		if feeModel.Hole == hole {
			fee = feeModel.Fee
		}
	}

	return fee
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

func GetFeeWidthHolePrice(feeList ListGolfHoleFee) int64 {
	re := regexp.MustCompile(`(gia)\w+`)

	confifFeeRaw := re.FindAllString("gia18+gia18/18*(ho-18)", -1)

	confifFees := removeDuplicateStr(confifFeeRaw)

	expression, err := govaluate.NewEvaluableExpression("gia18+gia18/18*(ho-18)")

	if err != nil {
		log.Println("NewEvaluableExpression err", err.Error())
		return 0
	}

	parameters := make(map[string]interface{}, 8)

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
