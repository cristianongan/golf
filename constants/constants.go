package constants

const (
	CADDIE_CURRENT_STATUS_WORKING_ONLY = "WORKING_ONLY"
	CADDIE_CURRENT_STATUS_JOB          = "JOB"
	CADDIE_CURRENT_STATUS_READY        = "READY"
	CADDIE_CURRENT_STATUS_IN_COURSE    = "IN_COURSE"
	CADDIE_CURRENT_STATUS_FINISH       = "FINISH"
	CADDIE_CURRENT_STATUS_LOCK         = "LOCK"
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

const (
	BAGS_NOTE_TYPE_BOOKING = "BOOKING"
	BAGS_NOTE_TYPE_BAG     = "BAG"
)

const (
	BOOKING_INIT_TYPE_BOOKING = "BOOKING"
	BOOKING_INIT_TYPE_CHECKIN = "CHECKIN"
)

/*
 Bag status
*/
const (
	BAG_STATUS_IN            = "IN"            // Đã check in( = Waiting ở doc)
	BAG_STATUS_OUT           = "OUT"           // Đã check out
	BAG_STATUS_INIT          = "INIT"          // Tạo Booking xong( = Booking ở doc)
	BAG_STATUS_CANCEL        = "CANCEL"        // Cancel booking
	BAG_STATUS_TIMEOUT       = "TIMEOUT"       // Đã checkin và out caddie
	BAG_STATUS_IN_COURSE     = "IN_COURSE"     // Đã checkin và ghép Flight
	BAG_STATUS_GUEST_NO_SHOW = "GUEST_NO_SHOW" // Khách đặt booking nhưng không đến
)

/*
 Caddie status on booking
*/
const (
	BOOKING_CADDIE_STATUS_IN   = "IN"
	BOOKING_CADDIE_STATUS_OUT  = "OUT"
	BOOKING_CADDIE_STATUS_INIT = "INIT"
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
	KIOSK_SETTING   = "KIOSK"
	MINI_B_SETTING  = "MINI_B"
	MINI_R_SETTING  = "MINI_R"
	DRIVING_SETTING = "DRIVING"
	RENTAL_SETTING  = "RENTAL"
	PROSHOP_SETTING = "PROSHOP"
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

const MAX_LIMIT = 9999999999

// const USER_PROFILE_KEY = "USER_PROFILE_KEY"
const CMS_USER_PROFILE_KEY = "CMS_USER_PROFILE_KEY"
const UNAUTHORIZED_MESSAGE = "Unauthorized"
const UNAUTHORIZED_LOGIN_MESSAGE = "Unauthorized, please login again"
const URL_CHECK_CRON = "cron-job/check-cron"
const URL_CRONJOB_BACKUP_ORDER = "cron-job/backup-order"

const CRONJOB_PREFIX = "CRONJOB:"
