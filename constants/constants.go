package constants

const TYPE_ADMIN = "ADMIN"

const VOUCHER_DELETE_PARAMS = "-1"

const DELETE_STR = "delete"

const TIMEOUT = 20
const DELIVERY_ERROR_SERVICE = "ERROR_SERVICE"

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

const MAX_LIMIT = 9999999999

// ================== Date time
const LOCATION_DEFAULT = "Asia/Ho_Chi_Minh"
const DATE_TIME_FORMAT = "2006-01-02 15:04:05"
const DATE_TIME_FORMAT_OTT = "020106150405" // ddMMyyhhmmss
const DATE_FORMAT = "2006-01-02"
const MONTH_FORMAT = "2006-01"
const YEAR_FORMAT = "2006"

const USER_PROFILE_KEY = "USER_PROFILE_KEY"
const CMS_USER_PROFILE_KEY = "CMS_USER_PROFILE_KEY"
const UNAUTHORIZED_MESSAGE = "Unauthorized"
const UNAUTHORIZED_LOGIN_MESSAGE = "Unauthorized, please login again"
const URL_CHECK_CRON = "cron-job/check-cron"
const URL_CRONJOB_BACKUP_ORDER = "cron-job/backup-order"

const CRONJOB_PREFIX = "CRONJOB:"

// Mobile banking transaction status
const (
	MbTransInit      = "INIT"
	MbTransInProcess = "IN_PROCESS"
	MbTransPaid      = "PAID" // Đã thanh toán
	MbTransCanceled  = "CANCELED"
)
