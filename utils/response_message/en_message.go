package response_message

var EnLanguage = map[string]string{
	// COMMON
	"SUCCESS":                    "Success",
	"SYSTEM_ERROR":               "An error accurred. Please try again later",
	"CANNOT_REGISTER":            "Cannot register",
	"ERROR_REQUEST_DATA":         "An error accurred. Please try again later",
	"ERROR_VALIDATE_DATA":        "Data is invalidate",
	"ERROR_NOT_FOUND":            "Record not found",
	"ERROR_DUP_RECORD":           "Data is duplicated",
	"ERROR_NOT_FIND_USER":        "Cannot find user",
	"ERROR_UPDATE_USER":          "Cannot update user",
	"ERROR_SEND_EMAIL":           "Cannot send email",
	"ERROR_UPDATE_PROFILE":       "Cannot update profile",
	"EXPIRE_TIME_VALIDATE_EMAIL": "Time to validate email is expired",
	"ERROR_VALIDATE_EMAIL":       "Link verify email is wrong or expired",
	"PERMISSION_DENY":            "Bạn không có quyền cho tính năng này",
	"UNAUTHORIZED_LOGIN":         "Unauthorized, please login again",
	"USER_BE_LOCKED":             "Account be deactived",

	// USER
	"JWT_TOKEN_INVALID":            "Đăng nhập không thành công",
	"VALIDATE_SOURCE_INVALID":      "Start point is not available",
	"VALIDATE_DESTINATION_INVALID": "Destination point is not available",
	"PHONE_INVALID":                "Invalid phone number",
}
