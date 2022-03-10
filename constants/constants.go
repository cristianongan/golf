package constants

const TYPE_ADMIN = "ADMIN"

const VOUCHER_DELETE_PARAMS = "-1"

const DELETE_STR = "delete"

const (
	NOTI_WARNING_REDIRECT_TYPE_INTERNAL_LINK = "INTERNAL_LINK"
	NOTI_WARNING_REDIRECT_TYPE_EXTERNAL_LINK = "EXTERNAL_LINK"
)

const (
	NOTI_WARNING_REDIRECT_TO_HOME      = "HOME"
	NOTI_WARNING_REDIRECT_TO_HIS       = "HISTORY"
	NOTI_WARNING_REDIRECT_TO_PROMOTION = "PROMOTION"
)

const (
	VNPAY_OTT_SENDER_SYSTEM      = "SYSTEM"
	VNPAY_OTT_SENDER_BACK_OFFICE = "BACK-OFFICE"
)

const (
	VNPAY_OTT_RAW_DATA_TYPE_DAT_LAI    = "DAT-LAI"
	VNPAY_OTT_RAW_DATA_TYPE_THANH_TOAN = "THANH-TOAN"
)

const (
	OTT_TYPE_NOTI_ALERT_PAYMENT        = 1
	OTT_TYPE_NOTI_ALERT_PICKUP         = 2
	OTT_TYPE_NOTI_ORDER_COMPLETE       = 3
	OTT_TYPE_NOTI_ORDER_CANCELED       = 4
	OTT_TYPE_NOTI_ORDER_CSKH_CANCELED  = 5
	OTT_TYPE_NOTI_ORDER_REMIND_PAYMENT = 6
)

const VNPAY_OTT_AES_IV = "0123456789123456"

const INIT_APP_CHECK_ORDER = "INIT_APP_CHECK_ORDER"

const TIMEOUT = 20
const DELIVERY_ERROR_SERVICE = "ERROR_SERVICE"

var ORDER_CANCEL_REASON_FORCE = "Bắt buộc huỷ"
var ORDER_CANCEL_REASON_DEFAULT = "Khách hàng huỷ đơn"
var ORDER_CANCEL_REASON_DEFAULT_EN = "cancelation by customer"
var ORDER_CANCEL_REASON_PAYMENT_ORDERDUE = "Quá hạn thanh toán"
var ORDER_CANCEL_REASON_PAYMENT_ORDERDUE_EN = "payment timeout"
var MAX_SIZE_AVATAR_UPLOAD = int64(3000000)

const ENV_PROD = "prod" //TODO: set in config environment name prod.json
const LANGUAGE_DEFAULT = "vi"
const LANGUAGE_EN = "en"
const API_HEADER_KEY_LANGUAGE = "language"

const AHAMOVE_PARTNER_ID = "AHAMOVE"

const JWT_EXPIRED_TIME = 604800 // 1 tháng // 1 tuan: 604800 // 1 ngay: 86400

const STATUS_DELETED = "DELETED"
const STATUS_ENABLE = "ENABLE"
const STATUS_DISABLE = "DISABLE"
const STATUS_PENDING = "PENDING"
const STATUS_PROCESSING = "PROCESSING"
const STATUS_FAILED = "FAILED"
const STATUS_SUCCESS = "SUCCESS"

const MAX_LIMIT = 9999999999

// ================== Date time
const LOCATION_DEFAULT = "Asia/Ho_Chi_Minh"
const DATE_TIME_FORMAT = "2006-01-02 15:04:05"
const DATE_TIME_FORMAT_OTT = "020106150405" // ddMMyyhhmmss
const DATE_FORMAT = "2006-01-02"
const MONTH_FORMAT = "2006-01"
const YEAR_FORMAT = "2006"

// ============= URL ===================
const USER_PROFILE_KEY = "USER_PROFILE_KEY"
const UNAUTHORIZED_MESSAGE = "Unauthorized"
const UNAUTHORIZED_LOGIN_MESSAGE = "Unauthorized, please login again"
const URL_CHECK_CRON = "cron-job/check-cron"
const URL_CRONJOB_BACKUP_ORDER = "cron-job/backup-order"

const CRONJOB_PREFIX = "CRONJOB:"

// ===================== Order status ======================
const (
	OrderAssigning = "ASSIGNING"  // Đang tìm TX
	OrderAccepted  = "ACCEPTED"   // TX đã nhận đơn
	OrderBoarded   = "BOARDED"    // TX đã đến điểm nhận hàng
	OrderInProcess = "IN_PROCESS" // Đang gửi hàng
	OrderPartial   = "PARTIAL"    // Hoàn thành một phần
	OrderCompleted = "COMPLETED"  // Hoàn thành toàn bộ

	OrderIDLE           = "IDLE"            // Khách hàng khởi tạo đơn - IDLE
	OrderIDLECanceled   = "IDLE_CANCELED"   // Khách hàng Huỷ khởi tạo đơn
	OrderUserCanceled   = "USER_CANCELLED"  // Khách hàng hủy đơn
	OrderDriverCanceled = "DRIVER_CANCELED" // Toàn bộ điểm giao bị thất bại
	OrderCskhCanceled   = "CSKH_CANCELED"   // CSKH hủy đơn
	OrderAutoCanceled   = "AUTO_CANCELED"   // Quá 20 phút ko tìm thấy TX
)

// Mobile banking transaction status
const (
	MbTransInit      = "INIT"
	MbTransInProcess = "IN_PROCESS"
	MbTransPaid      = "PAID" // Đã thanh toán
	MbTransCanceled  = "CANCELED"
)

const ORDER_CANCEL_REASON_PAYMENT_CANCEL = "Huỷ thanh toán"

// ===================== Order Path ====================
const (
	OrderPathFail      = "FAILED"    // Giao lỗi
	OrderPathCompleted = "COMPLETED" // Giao thành công
)

const (
	OrderPathTypeSource      = "SOURCE"
	OrderPathTypeDestination = "DESTINATION"
)

const (
	PlaceTypeHome    = "HOME"
	PlaceTypeCompany = "COMPANY"
	PlaceTypeOther   = "OTHER"
)
