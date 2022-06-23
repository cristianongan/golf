module start

go 1.16

require (
	github.com/bsm/redislock v0.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/go-redis/redis/v8 v8.1.0
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529
	github.com/google/uuid v1.2.0
	github.com/leekchan/accounting v1.0.0
	github.com/minio/minio-go/v7 v7.0.12
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/cors/wrapper/gin v0.0.0-20220223021805-a4a5ce87d5a2
	github.com/spf13/viper v1.8.1
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/text v0.3.7
	gorm.io/datatypes v1.0.6
	gorm.io/driver/mysql v1.3.2
	gorm.io/gorm v1.23.2
)
