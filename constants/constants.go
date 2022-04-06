package constants

/*
	Member Card Type:
	Member Card Base Type
*/
const (
	MEMBER_CARD_BASE_TYPE_FRIENDLY       = "FRIENDLY"
	MEMBER_CARD_BASE_TYPE_INSIDE_MEMBER  = "INSIDE_MEMBER"
	MEMBER_CARD_BASE_TYPE_OUTSIDE_MEMBER = "OUTSIDE_MEMBER"
	MEMBER_CARD_BASE_TYPE_PROMOTION      = "PROMOTION"
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

const MAX_LIMIT = 9999999999

// const USER_PROFILE_KEY = "USER_PROFILE_KEY"
const CMS_USER_PROFILE_KEY = "CMS_USER_PROFILE_KEY"
const UNAUTHORIZED_MESSAGE = "Unauthorized"
const UNAUTHORIZED_LOGIN_MESSAGE = "Unauthorized, please login again"
const URL_CHECK_CRON = "cron-job/check-cron"
const URL_CRONJOB_BACKUP_ORDER = "cron-job/backup-order"

const CRONJOB_PREFIX = "CRONJOB:"
