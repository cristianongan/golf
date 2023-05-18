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
	"TABLE_PRICE_DEL_NOTE":       "Không hỗ trợ xoá bảng giá",

	// USER
	"JWT_TOKEN_INVALID":             "Đăng nhập không thành công",
	"VALIDATE_SOURCE_INVALID":       "Start point is not available",
	"VALIDATE_DESTINATION_INVALID":  "Destination point is not available",
	"PHONE_INVALID":                 "Invalid phone number",
	"USER_VALIDATE_PASSWORD_POLICY": "Mật khẩu ít nhất 8 ký tự, kết hợp các ký tự: Chữ, Số, Ký tự đặc biệt",
	"USER_VALIDATE_PASSWORD_WEEK":   "Mật khẩu yếu",

	//BAG
	"BAG_NOT_FOUND":                 "Cannot found this bag",
	"PLAY_COUNT_INVALID":            "Play count remain over",
	"BOOKING_NOT_FOUND":             "Booking not found",
	"NOTI_NOT_FOUND":                "Notification not found",
	"UPDATE_BOOKING_ERROR":          "Update booking error",
	"MAIN_BAG_NOT_FOUND":            "Bag was main bag",
	"ROUND_NOT_FOUND":               "Round not found",
	"GUEST_STYLE_NOT_FOUND":         "Guest style not found",
	"UPDATE_ERROR":                  "Update error",
	"BAG_NOT_IN_COURSE":             "Bag status is not in course",
	"MERGE_ROUND_NOT_ENOUGH":        "Need minimum 2 round to merge",
	"MEMBER_CARD_INACTIVE":          "Member Card Inactive",
	"ANNUAL_TYPE_SLEEP_NOT_CHECKIN": "Annual type is Sleep",
	"DUPLICATE_BAG":                 "Duplicate bag",
	"OUT_CADDIE_ERROR":              "Booking have not caddie",
	"INVENTORY_NOT_FOUND":           "Inventory not found",
	"TEE_TIME_SLOT_FULL":            "Tee Time is full 4 slot",
	"BAG_BE_LOCK":                   "Bag bị lock",
	"LOCKER_RETURNED":               "Locker returned",
	"LOCKER_UNRETURNED":             "Locker unreturned",

	//Agency
	"AGENCY_DUPLI_CONTRACT_NO": "Bị trùng contract no",
	"AGENCY_DUPLI_AGENCY_ID":   "Bị trùng agency id",
}
