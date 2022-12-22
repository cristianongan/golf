package datasources

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"start/config"
	"start/constants"
	"time"

	// https://godoc.org/github.com/go-redis/redis
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v9"
)

type DataCache struct {
	Key            string
	JsonStringData string
	Tll            time.Duration
}

var redisdb *redis.Client
var redisLocker *redislock.Client

var ctx = context.Background()

func MyRedisConnect() {
	log.Println("redis connect")
	config := config.GetConfig()
	// log.Println(config.GetString("redis.host"), config.GetString("redis.port"), config.GetString("redis.password"))

	redisOption := &redis.Options{
		Addr:     config.GetString("redis.host") + ":" + config.GetString("redis.port"),
		Password: config.GetString("redis.password"),
		DB:       0, // use default DB
	}
	if config.GetBool("redis.insecure_skip_verify") {
		redisOption.TLSConfig = &tls.Config{InsecureSkipVerify: config.GetBool("redis.insecure_skip_verify")}
	}
	redisdb = redis.NewClient(redisOption)
	data := redisdb.Ping(ctx)
	log.Println(data)
	// pong, err := redisdb.Ping().Result()
	// log.Println("pong: "+pong, "redis error: "+err.Error())

	redisLocker = redislock.New(redisdb)
}

// / =================== Redis locker ===================
func GetLockerRedis() *redislock.Client {
	return redisLocker
}

// / Check đạt được lock mới xử lý tiếp
func GetLockerRedisObtainWith(key string, timeSecond time.Duration) bool {
	lock, err := redisLocker.Obtain(ctx, key, timeSecond*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("Could not obtain lock", key)
		return false
	}

	defer lock.Release(ctx)
	// Logic chạy cron bên dưới
	return true
}

func GetRedisKeyLockerResetDataMemberCard() string {
	return config.GetEnvironmentName() + "_" + "haicv_redis_locker_reset_data_member_card"
}

func GetRedisKeyLockerReportCaddieFeeToDay() string {
	return config.GetEnvironmentName() + "_" + "anhnq_redis_locker_report_caddie_fee_today"
}

func GetRedisKeyLockerReportInventoryStatisticItem() string {
	return config.GetEnvironmentName() + "_" + "tuandd_redis_locker_report_inventory_statistic_item"
}

func GetRedisKeySystemLogout() string {
	return config.GetEnvironmentName() + "_" + "tuandd_redis_key_system_log_out"
}

func GetRedisKeyLockerCreateGuestName() string {
	return config.GetEnvironmentName() + "_" + "redis_locker_create_guest_name"
}

func GetRedisKeyTeeTimeLock(teeTime string) string {
	return config.GetEnvironmentName() + "_" + "redis_tee_time_lock" + "_" + teeTime
}

func GetRedisKeyUserLogin(userName string) string {
	return config.GetEnvironmentName() + "_" + "redis_user_login" + "_" + userName
}

func GetRedis() *redis.Client {
	return redisdb
}

// ttl : second
func SetCache(key string, value interface{}, ttl int64) error {
	if redisdb == nil {
		return errors.New("redisdb is not connected")
	}

	return redisdb.Set(ctx, key, value, time.Duration(ttl*int64(time.Second))).Err()
}

func GetCache(key string) (string, error) {
	if redisdb == nil {
		return "", errors.New("redisdb is not connected")
	}
	return redisdb.Get(ctx, key).Result()
}

func ScanCache(keyPattern string, count int) ([]string, error) {
	var cursor uint64
	var keys []string
	var matchedKeys []string
	var cnt = int64(count)
	var err error

	for done := false; !done; {
		if len(matchedKeys) >= count {
			break
		}

		keys, cursor, err = redisdb.Scan(ctx, cursor, keyPattern, cnt).Result()
		if err != nil {
			// glog.ErrorDepth(1, err.Error())
		}

		if cursor == 0 {
			done = true
		}
		matchedKeys = append(matchedKeys, keys...)
	}

	if len(matchedKeys) == 0 {
		return nil, nil
	}

	return matchedKeys, err
}

func GetCaches(keys ...string) ([]interface{}, error) {
	if redisdb == nil {
		return nil, errors.New("redisdb is not connected")
	}
	return redisdb.MGet(ctx, keys...).Result()
}

func GetAllKeysWith(prefix string) ([]string, error) {
	if redisdb == nil {
		return []string{}, errors.New("redisdb is not connected")
	}
	keyCmds := redisdb.Keys(ctx, prefix+"*")
	return keyCmds.Val(), nil
}

func IncreaseFlagCounter(key string) (int, error) {
	if redisdb == nil {
		return 0, errors.New("redisdb is not connected")
	}
	result, err := redisdb.Incr(ctx, key).Result()
	return int(result), err
}

func Keys(pattern string) ([]string, error) {
	if redisdb == nil {
		return []string{}, errors.New("redisdb is not connected")
	}

	return redisdb.Keys(ctx, pattern).Result()
}

// =====================================
// PUSH and POP redis

func RPush(key string, value interface{}) (int64, error) {
	if redisdb == nil {
		return 0, errors.New("redisdb is not connected")
	}
	strCmd := redisdb.RPush(ctx, key, value)

	return strCmd.Result()
}

func LPop(key string) (string, error) {
	if redisdb == nil {
		return "", errors.New("redisdb is not connected")
	}
	strCmd := redisdb.LPop(ctx, key)
	// log.Println(strCmd.Result())

	return strCmd.Result()
}

func LTrim(key string, start, end int64) (string, error) {
	if redisdb == nil {
		return "", errors.New("redisdb is not connected")
	}
	strCmd := redisdb.LTrim(ctx, key, start, end)

	return strCmd.Result()
}

func LRange(key string, start, stop int64) ([]string, error) {
	if redisdb == nil {
		return []string{}, errors.New("redisdb is not connected")
	}
	strCmd := redisdb.LRange(ctx, key, start, stop)

	return strCmd.Result()
}

func DelCacheByKey(keys ...string) error {
	if redisdb == nil {
		return errors.New("redisdb is not connected")
	}
	redisdb.Del(ctx, keys...)
	return nil
}

func ExpireByKey(key string, ttl int) error {
	if redisdb == nil {
		return errors.New("redisdb is not connected")
	}
	return redisdb.Expire(ctx, key, time.Duration(ttl*1000000000)).Err()
}

// ========= Set User Jwt Cache ====================
func getKeyJwt(userUid string) string {
	return config.GetEnvironmentName() + ":" + constants.PREFIX_RKEY_JWT + ":" + userUid
}

func SetCacheJwt(userUid, jwtToken string, ttl int64) {
	key := getKeyJwt(userUid)
	SetCache(key, jwtToken, ttl)
}

func GetCacheJwt(userUid string) (string, error) {
	key := getKeyJwt(userUid)
	return GetCache(key)
}

func DelCacheJwt(userUid string) {
	key := getKeyJwt(userUid)
	DelCacheByKey(key)
}
