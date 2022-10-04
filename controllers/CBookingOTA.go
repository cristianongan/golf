package controllers

import (
	"start/controllers/request"
	"start/controllers/response"

	// "start/datasources"
	// model_booking "start/models/booking"
	// "start/utils"
	// "strconv"
	// "strings"
	// "time"

	"github.com/gin-gonic/gin"
)

type CBookingOTA struct{}

/*
Booking OTA
*/
func (cBooking *CBookingOTA) CreateBookingOTA(c *gin.Context) {
	body := request.CreateBookingOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	res := response.BookingOTARes{
		BookID: "121212",
	}

	okResponse(c, res)
}

/*
GetTeeTimeList
*/
// func (cBooking *CBookingOTA) GetTeeTimeList(c *gin.Context) {
// 	body := request.GetTeeTimeOTAList{}
// 	if bindErr := c.ShouldBind(&body); bindErr != nil {
// 		badRequest(c, bindErr.Error())
// 		return
// 	}

// 	cBookingSetting := CBookingSetting{}
// 	db := datasources.GetDatabase()
// 	form := request.GetListBookingSettingForm{
// 		CourseUid:  body.CourseCode,
// 	}
// 	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
// 	weekday := strconv.Itoa(int(time.Now().Weekday()))
// 	bookSetting := model_booking.BookingSetting{}

// 	teeTimeList := []response.TeeTimeOTA {}
// 	responseDate := response.GetTeeTimeOTAResponse{
// 		IsMainCourse: body.IsMainCourse,
// 		Token: nil,
// 		CourseCode: body.CourseCode,
// 		OTACode: body.OTA_Code,
// 		GuestCode: body.Guest_Code,
// 		Date: body.Date,
// 	}
// 	for _, data := range listSettingDetail {
// 		if strings.ContainsAny(data.Dow, weekday) {
// 			bookSetting = data
// 			break
// 		}
// 	}
// 	index := 0

// 	countPart := 0
// 	for {
// 		startTime, _ := utils.ConvertHourToTime(bookSetting.StartPart1)
// 		teeTimeFirst, _ := utils.ConvertHourToTime(bookSetting.StartPart1)
// 		for {
// 			index += 1
// 			teeTime := response.TeeTimeOTA {
// 				TeeOffStr: bookSetting.StartPart1,
// 				DateStr: body.Date,
// 				Part: 0,
// 				TimeIndex: int64(index),
// 				NumBook: 0,
// 				IsMainCourse: body.IsMainCourse,
// 				Tee: 1,
// 			}
// 			teeTimeList = append(teeTimeList, )

// 			if false {
// 				break
// 			}
// 		}

// 		if countPart > 2 {
// 			break
// 		}
// 		countPart += 1
// 	}

// 	okResponse(c, res)
// }
