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

	// USER
	"JWT_TOKEN_INVALID":            "Đăng nhập không thành công",
	"VALIDATE_SOURCE_INVALID":      "Start point is not available",
	"VALIDATE_DESTINATION_INVALID": "Destination point is not available",

	// Order
	"INVALID_SERVICE_HOUR": "Không thể sử dụng dịch vụ trong khung giờ này",
	"INVALID_BOOKING_TIME": "Thời gian đặt đơn không hợp lệ",
	"VALIDATE_MIN_PATH":    "Không tìm thấy Điểm giao",
	"INVALID_DISTANCE":     "Rất tiếc, dịch vụ giao hàng của chúng tôi chưa triển khai tại địa điểm này",
	"ERROR_ORDER_UPDATE":   "Đơn hàng không thể cập nhật",
	"ERROR_FINISH_ORDER":   "Không thể kết thúc chuyến đi",
	"ERROR_CANCEL_ORDER":   "Không thể huỷ đơn hàng",
	"BOOKING_RATED":        "Cuốc đặt đã được đánh giá",

	//NEW
	"ERROR_MAX_LIST_SAVED": "Số địa điểm đã lưu đã đạt tối đa, vui lòng xóa để lưu địa điểm mới",
}
