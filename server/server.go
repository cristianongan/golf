package server

import (
	"log"
	"start/config"
	"start/datasources"
	// "start/datasources/aws"
	// "start/datasources/elasticsearch"
)

func Init() {
	log.Println("server init")

	config := config.GetConfig()
	// cron.CronStart()
	// cron.InitCronJobCallApi()

	//datasources.MinioConnect()

	datasources.MySqlConnect()
	MigrateDb()
	// ============ Use redis
	datasources.MyRedisConnect()

	// ======== Connect elasticsearch/
	// elasticsearch.ElasticSearchInit()

	r := NewRouter()
	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
}
