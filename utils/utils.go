package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"start/constants"
	"strconv"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/google/uuid"
	"github.com/leekchan/accounting"
	"github.com/ttacon/libphonenumber"
)

func GetPathOrderBillCode(orderBillcode string, pathIndex int) string {
	return orderBillcode + "-" + strconv.Itoa(pathIndex)
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

func GetTimeStampFromLocationTime(location, localTime string) int64 { //
	loc, errLoc := time.LoadLocation(location)
	if errLoc != nil {
		log.Println(errLoc)
		return 0
	}
	time, errParse := time.ParseInLocation(constants.DATE_TIME_FORMAT, localTime, loc)
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

// func ConvertUidFromPartnerName(partnerName string) string {
// 	s := convertUtf8ToUnicode(partnerName)
// 	type StrConvert struct {
// 		Input   string
// 		Replace string
// 	}
// 	listConverts := []StrConvert{
// 		StrConvert{"Đ", "D"},
// 		StrConvert{" ", ""},
// 	}

// 	strUpperCase := strings.ToUpper(s)
// 	strRemoveSpace := strUpperCase
// 	for _, item := range listConverts {
// 		strRemoveSpace = strings.Replace(strRemoveSpace, item.Input, item.Replace, -1)
// 	}
// 	return strRemoveSpace
// }

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
