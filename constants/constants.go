package constants

const (
	CADDIE_CURRENT_STATUS_WORKING_ONLY = "WORKING_ONLY"
	CADDIE_CURRENT_STATUS_JOB          = "JOB"
	CADDIE_CURRENT_STATUS_READY        = "READY"
	CADDIE_CURRENT_STATUS_IN_COURSE    = "IN_COURSE"
	CADDIE_CURRENT_STATUS_IN_COURSE_R2 = "IN_COURSE_R2"
	CADDIE_CURRENT_STATUS_IN_COURSE_R3 = "IN_COURSE_R3"
	CADDIE_CURRENT_STATUS_FINISH       = "FINISH"
	CADDIE_CURRENT_STATUS_FINISH_R2    = "FINISH_R2"
	CADDIE_CURRENT_STATUS_FINISH_R3    = "FINISH_R3"
	CADDIE_CURRENT_STATUS_LOCK         = "LOCK"
)

const (
	BUGGY_CURRENT_STATUS_ACTIVE      = "ACTIVE"      // Trạng thái hoạt động sẵn sàng ghép khách
	BUGGY_CURRENT_STATUS_IN_COURSE   = "IN_COURSE"   // Đang được cho khách thuê
	BUGGY_CURRENT_STATUS_MAINTENANCE = "MAINTENANCE" // Đang bảo hành không ghép khách
	BUGGY_CURRENT_STATUS_IN_ACTIVE   = "INACTIVE"    // Không sử dụng nữa
	BUGGY_CURRENT_STATUS_LOCK        = "LOCK"        // Buggy đã được ghép với khách nhưng chưa được ghép Flight
	BUGGY_CURRENT_STATUS_FIX         = "FIX"         // Đang sửa chữa không ghép khách
)

const (
	CADDIE_WORKING_STATUS_ACTIVE   = "ACTIVE"
	CADDIE_WORKING_STATUS_INACTIVE = "INACTIVE"
)

const (
	STATUS_DELETE = "DELETE"
	STATUS_IN     = "IN"
	STATUS_OUT    = "OUT"
)

const (
	GORM_API_LOG_RECORD_NOT_FOUND = "record not found"
)

/*
Bag note Type
*/
const (
	BAGS_NOTE_TYPE_BOOKING = "BOOKING"
	BAGS_NOTE_TYPE_BAG     = "BAG"
)

/*
Để phân biệt bag booking được tạo từ single book, hay từ check in lễ tân tạo booking luôn
*/
const (
	BOOKING_INIT_TYPE_BOOKING = "BOOKING" // được tạo từ booking single book
	BOOKING_INIT_TYPE_CHECKIN = "CHECKIN" // Tạo từ check in lễ tân
)

/*
Trạng thái Kiosk Inventory
*/
const (
	KIOSK_BILL_INVENTORY_PENDING  = "PENDING"  // Đơn nhập đang chờ duyệt
	KIOSK_BILL_INVENTORY_ACCEPT   = "ACCEPT"   // Đơn nhập đã chấp nhận thêm vào kho
	KIOSK_BILL_INVENTORY_RETURN   = "RETURN"   // Đơn nhập bị trả lại
	KIOSK_BILL_INVENTORY_SELL     = "SELL"     // Đơn xuất đang chờ bán
	KIOSK_BILL_INVENTORY_TRANSFER = "TRANSFER" // Đơn xuất đã xuất thành công
)

const (
	KIOSK_BILL_INVENTORY_IMPORT = "IMPORT"
	KIOSK_BILL_INVENTORY_EXPORT = "EXPORT"
)

/*
Bag status
*/
const (
	BAG_STATUS_BOOKING       = "BOOKING"       // Tạo Booking xong: Khách đặt booking
	BAG_STATUS_WAITING       = "WAITING"       // Waiting, Đã check in chưa ghép flight
	BAG_STATUS_IN_COURSE     = "IN_COURSE"     // Đã checkin và ghép Flight
	BAG_STATUS_TIMEOUT       = "TIMEOUT"       // Đã out flight(không được ghép flight nào)
	BAG_STATUS_CHECK_OUT     = "CHECK_OUT"     // Đã check out
	BAG_STATUS_CANCEL        = "CANCEL"        // Cancel booking
	BAG_STATUS_GUEST_NO_SHOW = "GUEST_NO_SHOW" // Khách đặt booking nhưng không đến
)

/*
Caddie status on booking
trạng thái Caddie của Booking
Dùng cho cả log caddie in out booking
*/
const (
	BOOKING_CADDIE_STATUS_IN   = "IN"   // Bag đươc gán caddie
	BOOKING_CADDIE_STATUS_OUT  = "OUT"  // Bag đã out caddie
	BOOKING_CADDIE_STATUS_INIT = "INIT" // Bag mới khởi tạo chưa gán caddie
)

/*
Main bag for Pay SUB Bag
*/
const (
	MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND = "FIRST_ROUND"
	MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS = "NEXT_ROUNDS"
	MAIN_BAG_FOR_PAY_SUB_RENTAL      = "RENTAL"
	MAIN_BAG_FOR_PAY_SUB_KIOSK       = "KIOSK"
	MAIN_BAG_FOR_PAY_SUB_RESTAURANT  = "RESTAURANT"
	MAIN_BAG_FOR_PAY_SUB_PROSHOP     = "PROSHOP"
	MAIN_BAG_FOR_PAY_SUB_OTHER_FEE   = "OTHER_FEE"
)

/*
Member Card Type:
Member Card Base Type
*/
const (
	MEMBER_CARD_BASE_TYPE_SHORT_TERM = "SHORT_TERM" // ngắn hạn
	MEMBER_CARD_BASE_TYPE_LONG_TERM  = "LONG_TERM"  // dài hạn
	MEMBER_CARD_BASE_TYPE_VIP        = "VIP"        // vip
	MEMBER_CARD_BASE_TYPE_FOREIGN    = "FOREIGN"    // nước ngoài
)

/*
Annual Type:
Không giới hạn
Chơi có giới hạn
Thẻ ngủ
*/
const (
	ANNUAL_TYPE_LIMITED    = "LIMITED"    // chơi giới hạn
	ANNUAL_TYPE_UN_LIMITED = "UN_LIMITED" // Chơi không giới hạn
	ANNUAL_TYPE_SLEEP      = "SLEEP"      // Thẻ ngủ
)

/*
Sân 18: Tee 1, Tee 10
Sân 27: Tee 1A, Tee 1B, Tee 1C
Sân 36: Tee 1A, Tee 10A, Tee 1B, Tee 10B
*/
const (
	TEE_TYPE_1   = "1"   // Sân 18
	TEE_TYPE_10  = "10"  // Sân 18
	TEE_TYPE_10A = "10A" // Sân 36
	TEE_TYPE_10B = "10B" // Sân 36
	TEE_TYPE_1A  = "1A"  // Sân 27 or 36
	TEE_TYPE_1B  = "1B"  // Sân 27 or 36
	TEE_TYPE_1C  = "1C"  // Sân 27
)

/*
Các dịch vụ của sân Golf: thuê đồ, shop, nhà hàng...
*/
const (
	GOLF_SERVICE_RENTAL     = "RENTAL"
	GOLF_SERVICE_PROSHOP    = "PROSHOP"
	GOLF_SERVICE_RESTAURANT = "RESTAURANT"
	GOLF_SERVICE_KIOSK      = "KIOSK"
)

/*
Các loại KIOSK
*/
const (
	KIOSK_SETTING      = "KIOSK"
	MINI_B_SETTING     = "MINI_B"
	MINI_R_SETTING     = "MINI_R"
	DRIVING_SETTING    = "DRIVING"
	RENTAL_SETTING     = "RENTAL"
	PROSHOP_SETTING    = "PROSHOP"
	RESTAURANT_SETTING = "RESTAURANT"
)

/*
Các loại Group Service
*/
const (
	GROUP_FB      = "FB"
	GROUP_RENTAL  = "RENTAL"
	GROUP_PROSHOP = "PROSHOP"
)

/*
Các dịch vụ của nhà hàng: mang theo, giao hàng, đặt bàn...
*/
const (
	RES_TYPE_BRING = "BRING"
	RES_TYPE_SHIP  = "SHIP"
	RES_TYPE_TABLE = "TABLE"
)

/*
Các trạng thái món ăn của nhà hàng
*/
const (
	RES_STATUS_BOOKING = "BOOKING" // Trạng thái món được booking
	RES_STATUS_ORDER   = "ORDER"   // Trạng thái là người đã đặt món và đang chờ đồ ăn
	RES_STATUS_PROCESS = "PROCESS" // Trạng thái món ăn đang được chế biến chưa được phục vụ
	RES_STATUS_DONE    = "DONE"    // Trạng thái món đã được phục vụ
	RES_STATUS_CANCEL  = "CANCEL"  // Trạng thái món đã bị hủy
)

/*
Các trạng thái đơn của nhà hàng
*/
const (
	RES_BILL_STATUS_BOOKING = "BOOKING" // Trạng thái là người booking bàn nhưng chưa vào nhà hàng dùng món
	RES_BILL_STATUS_ORDER   = "ORDER"   // Trạng thái là người đã đặt món và đang chờ đồ ăn
	RES_BILL_STATUS_PROCESS = "PROCESS" // Trạng thái món ăn đang được chế biến chưa được phục vụ
	RES_BILL_STATUS_FINISH  = "FINISH"  // Trạng thái món đã được phục vụ
	RES_BILL_STATUS_CANCEL  = "CANCEL"  // Trạng thái món đã bị hủy
	RES_BILL_STATUS_OUT     = "OUT"     //Trạng thái khách đã dùng xong món ăn và out khỏi nhà hàng.
)

/*
Các trạng thái đơn của point of sale
*/
const (
	POS_BILL_STATUS_PENDING = "PENDING"   // Đơn hàng đang ở trạng thái chưa xác nhận
	POS_BILL_STATUS_ACTIVE  = "ACTIVE"    // Đơn hàng đã được xác nhận và chốt đơn
	POS_BILL_STATUS_OUT     = "CHECK_OUT" // Đơn hàng đã hủy
)

/*
Các loại thay đổi hố booking
*/
const (
	BOOKING_STOP_BY_SELF = "STOP_BY_SELF" // Dừng do khách
	BOOKING_STOP_BY_RAIN = "STOP_BY_RAIN" // Dừng do trời mua
	BOOKING_CHANGE_HOLE  = ""             // Đổi hố
)

/*
Các dịch vụ của sân Golf: thuê đồ, shop, nhà hàng...
*/
const (
	DAY_OFF_TYPE_AFTERNOON = "H_AFTERNOON"
	DAY_OFF_TYPE_MORNING   = "H_MORNING"
	DAY_OFF_TYPE_SICK      = "SICK"
)

const BOOKING_OTHER_FEE = "OTHER_FEE"

const FEE_SEPARATE_CHAR = "/"

const DB_ERR_RECORD_NOT_FOUND = "RECORD NOT FOUND"
const API_ERR_DUPLICATED_RECORD = "DUPLICATED RECORD"
const API_ERR_INVALID_BODY_DATA = "INVALID BODY DATA"

const CUSTOMER_TYPE_CUSTOMER = "CUSTOMER"
const CUSTOMER_TYPE_AGENCY = "AGENCY"
const CUSTOMER_TYPE_NONE_GOLF = "NONE_GOLF"
const CUSTOMER_TYPE_WALKING_FEE = "WALKING_FEE"

const TYPE_ADMIN = "ADMIN"

const DELETE_STR = "delete"

const TIMEOUT = 20

var MAX_SIZE_AVATAR_UPLOAD = int64(3000000)

const ENV_PROD = "prod" //TODO: set in config environment name prod.json
const LANGUAGE_DEFAULT = "vi"
const LANGUAGE_EN = "en"
const API_HEADER_KEY_LANGUAGE = "language"

const JWT_EXPIRED_TIME = 604800 // 1 tháng // 1 tuan: 604800 // 1 ngay: 86400

const STATUS_DELETED = "DELETED"
const STATUS_ENABLE = "ENABLE"
const STATUS_DISABLE = "DISABLE"
const STATUS_PENDING = "PENDING"
const STATUS_PROCESSING = "PROCESSING"
const STATUS_FAILED = "FAILED"
const STATUS_SUCCESS = "SUCCESS"

const TEE_TIME_LOCKED = "LOCKED"
const TEE_TIME_UNLOCK = "UNLOCK"
const TEE_TIME_DELETED = "DELETED"

const CADDIE_WORKING_CALENDAR_LABEL_READY = "READY"
const CADDIE_WORKING_CALENDAR_LABEL_IN_COURSE_R1 = "IN_COURSE_R1"
const CADDIE_WORKING_CALENDAR_LABEL_IN_COURSE_R2 = "IN_COURSE_R2"
const CADDIE_WORKING_CALENDAR_LABEL_IN_COURSE_R3 = "IN_COURSE_R3"
const CADDIE_WORKING_CALENDAR_LABEL_FINISH_R1 = "FINISH_R1"
const CADDIE_WORKING_CALENDAR_LABEL_FINISH_R2 = "FINISH_R2"
const CADDIE_WORKING_CALENDAR_LABEL_WORKING_ONLY = "WORKING_ONLY"

const MAX_LIMIT = 9999999999

// const USER_PROFILE_KEY = "USER_PROFILE_KEY"
const CMS_USER_PROFILE_KEY = "CMS_USER_PROFILE_KEY"
const UNAUTHORIZED_MESSAGE = "Unauthorized"
const UNAUTHORIZED_LOGIN_MESSAGE = "Unauthorized, please login again"
const URL_CHECK_CRON = "cron-job/check-cron"

const CRONJOB_PREFIX = "CRONJOB:"
