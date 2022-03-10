package config

import (
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

var config *viper.Viper

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

func GetConfig() *viper.Viper {
	return config
}

func GetUrlRoot() string {
	return config.GetString("url_root")
}

func GetUrlBackendApi() string {
	return config.GetString("url_root") + "/" + strings.Replace(config.GetString("module_name"), "_", "-", -1)
}

func GetEnviromentName() string {
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
	return config.GetString("minio.enpoint")
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
