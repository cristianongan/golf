package constants

const (
	ERROR_PLAY_COUNT_INVALID      = 10 // Báo lỗi số lần chơi đã hết của khách
	ERROR_CREATE_FLIGHT_INVALID   = 11 // Bag Status chưa ở trạng thái WAITING để ghép flight
	ERROR_OUT_CADDIE              = 12 // Booking đã out caddie
	ERROR_BOOKING_OTA_LOCK        = 13 // Booking lỗi do có khóa từ OTA
	ERROR_DELETE_LOCK_OTA         = 14 // Lỗi unlock tee/turn time từ OTA
	ERROR_DELETE_LOCK_NOT_PERMISS = 15 // Lỗi unlock tee/turn time - KHÔNG CÓ QUYỀN
)
