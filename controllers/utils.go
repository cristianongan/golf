package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"start/constants"
	"start/models"
	"strconv"
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

// Caddie Fee, Green Fee, Buggy Fee string to List Model int
func golfFeeToList(feeText string) []models.GolfFeeText {
	listF := strings.Split(feeText, constants.FEE_SEPARATE_CHAR)
	listResult := []models.GolfFeeText{}
	if len(listF) == 0 {
		return listResult
	}

	for i, v := range listF {
		feeInt, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			golfFeeText := models.GolfFeeText{
				Hole: (i + 1) * 9,
				Fee:  feeInt,
			}
			listResult = append(listResult, golfFeeText)
		} else {
			log.Println("golfFeeToList err", err.Error())
		}
	}

	return listResult
}
