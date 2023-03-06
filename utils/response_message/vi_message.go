package response_message

var ViLanguage = map[string]string{
	// COMMON
	"SUCCESS":                    "Success",
	"SYSTEM_ERROR":               "Có lỗi trong quá trình xử lý. Vui lòng thử lại sau",
	"CANNOT_REGISTER":            "Cannot register",
	"ERROR_REQUEST_DATA":         "Có lỗi trong quá trình xử lý. Vui lòng thử lại sau",
	"ERROR_VALIDATE_DATA":        "Data is invalidate",
	"ERROR_NOT_FOUND":            "Record not found",
	"ERROR_DUP_RECORD":           "Dữ liệu bị trùng",
	"ERROR_NOT_FIND_USER":        "Cannot find user",
	"ERROR_UPDATE_USER":          "Cannot update user",
	"ERROR_SEND_EMAIL":           "Cannot send email",
	"ERROR_UPDATE_PROFILE":       "Cannot update profile",
	"EXPIRE_TIME_VALIDATE_EMAIL": "Time to validate email is expired",
	"ERROR_VALIDATE_EMAIL":       "Link verify email is wrong or expired",
	"PERMISSION_DENY":            "Bạn không có quyền cho tính năng này",
	"UNAUTHORIZED_LOGIN":         "Lỗi đăng nhập",
	"USER_BE_LOCKED":             "Tài khoản bị khóa",
	"PHONE_INVALID":              "Số điện thoại không hợp lệ",
	"TABLE_PRICE_DEL_NOTE":       "Không hỗ trợ xoá bảng giá",

	// USER
	"JWT_TOKEN_INVALID":             "Đăng nhập không thành công",
	"VALIDATE_SOURCE_INVALID":       "Start point is not available",
	"VALIDATE_DESTINATION_INVALID":  "Destination point is not available",
	"USER_VALIDATE_PASSWORD_POLICY": "Mật khẩu ít nhất 8 ký tự, kết hợp các ký tự: Chữ, Số, Ký tự đặc biệt",
	"USER_VALIDATE_PASSWORD_WEEK":   "Mật khẩu yếu",

	//BAG
	"BAG_NOT_FOUND":                 "Không tìm thấy Bag này",
	"PLAY_COUNT_INVALID":            "Số lần chơi đã hết",
	"BOOKING_NOT_FOUND":             "Không tìm thấy booking này",
	"NOTI_NOT_FOUND":                "Không tìm thấy notification này",
	"UPDATE_BOOKING_ERROR":          "Update booking lỗi",
	"MAIN_BAG_NOT_FOUND":            "Bag đang là main bag",
	"ROUND_NOT_FOUND":               "Không tìm thấy round này",
	"GUEST_STYLE_NOT_FOUND":         "Không tìm thấy guest style này",
	"UPDATE_ERROR":                  "Update error",
	"BAG_NOT_IN_COURSE":             "Bag status is not in course",
	"MERGE_ROUND_NOT_ENOUGH":        "Cần tối thiểu 2 round để merge",
	"MEMBER_CARD_INACTIVE":          "Member Card Inactive",
	"ANNUAL_TYPE_SLEEP_NOT_CHECKIN": "Member Card là thẻ ngủ",
	"DUPLICATE_BAG":                 "Bag này đã được sử dụng",
	"OUT_CADDIE_ERROR":              "Booking chưa ghép caddie",
	"INVENTORY_NOT_FOUND":           "Không tìm thấy kho",
	"TEE_TIME_SLOT_FULL":            "Tee Time đã đủ 4 slot",
	"BAG_BE_LOCK":                   "Bag bị lock",

	//Agency
	"AGENCY_DUPLI_CONTRACT_NO": "Bị trùng contract no",
	"AGENCY_DUPLI_AGENCY_ID":   "Bị trùng agency id",
}
