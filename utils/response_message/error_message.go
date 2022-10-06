package response_message

import (
	"net/http"
	"start/config"
	"start/constants"

	"github.com/gin-gonic/gin"
)

var (
	HEADER_KEY_LANGUAGE = "language"
)

type ResponseMessage struct {
	Language string `json:"language"`
	Message  string `json:"message"`
}

type ErrorResponseData struct {
	Message    string `json:"message"`
	Log        string `json:"log"`
	StatusCode int    `json:"status_code"`
}

type ErrorResponseDataV2 struct {
	Message     string      `json:"message"`
	Log         string      `json:"log"`
	StatusCode  int         `json:"status_code"`
	ErrorDetail interface{} `json:"error_detail"`
	Code        string      `json:"code"`
}

var languages = map[string]map[string]string{
	"default": ViLanguage,
	"vi":      ViLanguage,
	"en":      EnLanguage,
}

func (item *ResponseMessage) SetLanguage(language string) {
	// keys := []string{}
	for k := range languages {
		if language == k {
			item.Language = k
			return
		}
	}
	item.Language = "default"
}

func (item *ResponseMessage) GetMessage(key string) string {
	if item.Language == "" {
		item.Language = "default"
	}
	item.Message = languages[item.Language][key]
	return item.Message
}

func GetMessageByLanguage(language, key string) string {
	if language == "" {
		language = "default"
	}
	return languages[language][key]
}

func GetErrorResponseData(language, key, log string, statusCode int) ErrorResponseData {
	if language != constants.LANGUAGE_DEFAULT && language != constants.LANGUAGE_EN {
		language = "default"
	}

	errData := ErrorResponseData{
		Message:    languages[language][key],
		Log:        log,
		StatusCode: statusCode,
	}

	if errData.Message == "" {
		errData.Message = languages[language]["SYSTEM_ERROR"]
	}

	config := config.GetConfig()
	if config.GetBool("system_log_response") == false {
		errData.Log = ""
	}
	return errData
}

func ErrorResponse(c *gin.Context, code int, key, log string, statusCode int) {
	//lang := c.Request.Header.Get(HEADER_KEY_LANGUAGE)
	lang := c.Request.Header.Get(constants.API_HEADER_KEY_LANGUAGE)
	internalErr := GetErrorResponseData(lang, key, log, statusCode)
	c.JSON(code, internalErr)
}

// ================ For response status 200, hanlde in Reponse status_code = 400, ... in response ==============
func BadRequestDynamicKey(c *gin.Context, key, log string) {
	if c == nil {
		return
	}
	ErrorResponse(c, http.StatusBadRequest, key, log, http.StatusBadRequest) //400
}

func BadRequest(c *gin.Context, log string) {
	if c == nil {
		return
	}
	ErrorResponse(c, http.StatusBadRequest, "ERROR_REQUEST_DATA", log, http.StatusBadRequest) //400
}

func InternalServerErrorWithKey(c *gin.Context, log, key string) {
	if c == nil {
		return
	}
	ErrorResponse(c, http.StatusInternalServerError, key, log, http.StatusInternalServerError) //500
}

func InternalServerError(c *gin.Context, log string) {
	if c == nil {
		return
	}
	ErrorResponse(c, http.StatusInternalServerError, "SYSTEM_ERROR", log, http.StatusInternalServerError) //500
}

func NotFound(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusNotFound, "ERROR_NOT_FOUND", log, http.StatusNotFound) // 404
}

func LoginFailed(c *gin.Context) {
	ErrorResponse(c, http.StatusNotFound, "LOGIN_FAILED", "", http.StatusNotFound) //404
}

func DuplicateRecord(c *gin.Context, log string) {
	if c == nil {
		return
	}
	ErrorResponse(c, http.StatusConflict, "ERROR_DUP_RECORD", log, http.StatusConflict) // 409
}

func PermissionDeny(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusMethodNotAllowed, "PERMISSION_DENY", log, http.StatusMethodNotAllowed) //405
}

func UnAuthorized(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED_LOGIN", log, http.StatusUnauthorized) //401
}

func UserLocked(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusUnauthorized, "USER_BE_LOCKED", log, http.StatusUnauthorized) //401
}

// =============== V1 ==================================
/*
func InternalServerError(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusInternalServerError, "SYSTEM_ERROR", log)
}

func BadRequest(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusBadRequest, "ERROR_REQUEST_DATA", log)
}

func NotFound(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusNotFound, "ERROR_NOT_FOUND", log)
}

func LoginFailed(c *gin.Context) {
	ErrorResponse(c, http.StatusNotFound, "LOGIN_FAILED", "")
}

func DuplicateRecord(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusConflict, "ERROR_DUP_RECORD", log)
}

func PermissionDeny(c *gin.Context, log string) {
	ErrorResponse(c, http.StatusMethodNotAllowed, "PERMISSION_DENY", log)
}
*/
