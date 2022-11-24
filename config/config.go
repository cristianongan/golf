package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

var config *viper.Viper

var blacklistPass []string

func GetBlacklistPass() []string {
	return blacklistPass
}

func ReadConfigFile(env string) {
	config = viper.New()

	pwd, err := os.Getwd()
	if err != nil {
		glog.Fatalf("Error get current path, %s", err)
	}

	config.SetConfigFile(pwd + "/config/" + env + ".json")
	config.AddConfigPath(pwd + "/config")

	config.SetConfigName(env)
	config.SetConfigType("json")
	// Searches for config file in given paths and read it
	if err := config.ReadInConfig(); err != nil {
		glog.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	//log.Println("Using config: ", config.ConfigFileUsed())
	//log.Println("name", config.GetString("name"))

}

func ReadBlackListPassword() {
	pwd, err := os.Getwd()
	file, err := os.Open(pwd + "/config/blacklist_pass.txt")
	if err != nil {
		blacklistPass = []string{}
		// return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() == nil {
		blacklistPass = lines
	}
	// return lines, scanner.Err()
}

func GetConfig() *viper.Viper {
	return config
}

func GetUrlRoot() string {
	return config.GetString("url_root")
}

func GetUrlBackendApi() string {
	return config.GetString("url_root") + "/" + strings.Replace(config.GetString("module_name"), "_", "-", -1)
}

func GetEnvironmentName() string {
	return config.GetString("name")
}

// ===============================================
func GetCronBackupOrderRunning() bool {
	return config.GetBool("cron_backup_order_is_running")
}

func GetCronIsRunning() bool {
	return config.GetBool("cron_is_running")
}

func GetModuleName() string {
	return config.GetString("module_name")
}

func GetJwtSecret() string {
	return config.GetString("jwt_secret")
}

// =============== Get database config ========================
func GetDbName() string {
	return config.GetString("mysql.db_name")
}

func GetDbUser() string {
	return config.GetString("mysql.user")
}

func GetDbPassword() string {
	return config.GetString("mysql.password")
}

func GetDbHost() string {
	return config.GetString("mysql.host")
}

func GetDbPort() string {
	return config.GetString("mysql.port")
}

func GetDbDebug() bool {
	return config.GetBool("mysql.debug")
}

func GetIsMigrated() bool {
	return config.GetBool("mysql.is_migrated")
}

// =============== Get database2 config ========================
func GetDbName2() string {
	return config.GetString("mysql2.db_name")
}

func GetDbUser2() string {
	return config.GetString("mysql2.user")
}

func GetDbPassword2() string {
	return config.GetString("mysql2.password")
}

func GetDbHost2() string {
	return config.GetString("mysql2.host")
}

func GetDbPort2() string {
	return config.GetString("mysql2.port")
}

func GetDbDebug2() bool {
	return config.GetBool("mysql2.debug")
}

func GetIsMigrated2() bool {
	return config.GetBool("mysql2.is_migrated")
}

// =============== Get database Auth config ========================
func GetDbAuthName() string {
	return config.GetString("mysql_auth.db_name")
}

func GetDbAuthUser() string {
	return config.GetString("mysql_auth.user")
}

func GetDbAuthPassword() string {
	return config.GetString("mysql_auth.password")
}

func GetDbAuthHost() string {
	return config.GetString("mysql_auth.host")
}

func GetDbAuthPort() string {
	return config.GetString("mysql_auth.port")
}

func GetDbAuthDebug() bool {
	return config.GetBool("mysql_auth.debug")
}

func GetDbAuthIsMigrated() bool {
	return config.GetBool("mysql_auth.is_migrated")
}

// ================================================
func GetKibanaLog() bool {
	return config.GetBool("kibana_log")
}

func GetElasticSearchUrl() string {
	return config.GetString("elasticsearch.url")
}

// ============ Fluentd ==========================
func GetFluentdUrl() string {
	return config.GetString("fluentd.url")
}
func GetFluentdUser() string {
	return config.GetString("fluentd.user_name")
}
func GetFluentdPass() string {
	return config.GetString("fluentd.password")
}

// ============ Minio ==========================
func GetMinioEndpoint() string {
	return config.GetString("minio.endpoint")
}
func GetMinioBucket() string {
	return config.GetString("minio.bucket")
}
func GetMinioAccessKey() string {
	return config.GetString("minio.access_key")
}
func GetMinioSecretKey() string {
	return config.GetString("minio.secret_key")
}
func GetMinioSsl() bool {
	return config.GetBool("minio.ssl")
}
func GetMinioGetDataHost() string {
	return config.GetString("minio.get_data_host")
}

// ============ Cron job Key ===============
func GetCronJobSecretKey() string {
	return config.GetString("cronjob_secret_key")
}
func GetCronJobPageLimit() int {
	return config.GetInt("cronjob_page_limit")
}

// =============== Get Payment SecretKey ========================
func GetPaymentSecretKey() string {
	return config.GetString("payment_secret_key")
}

// =============== Get URL Golf Partner ========================
func GetGolfPartnerURL() string {
	return config.GetString("golf_partner.url")
}

// =============== Get Pass SecretKey ========================
func GetPassSecretKey() string {
	return config.GetString("aes_pass_secret_key")
}

// =============== Có mở socket không ============
func GetIsOpenSocket() bool {
	return config.GetBool("is_open_socket")
}
