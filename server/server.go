package server

import (
	"log"
	"start/config"
	"start/datasources"
	"start/logger"

	ccron "start/cron"
	// "start/datasources/aws"
	// "start/datasources/elasticsearch"
)

func Init() {
	log.Println("server init")

	config := config.GetConfig()

	// --- Cron ---
	ccron.CronStart()

	datasources.MinioConnect()

	datasources.MySqlConnect()
	MigrateDb()
	// ============ Use redis
	datasources.MyRedisConnect()

	// ======== Connect elasticsearch/
	// elasticsearch.ElasticSearchInit()

	r := NewRouter()

	routers := r.Routes()

	// Init authority
	initAuthority(routers)

	logger.InitLogger()

	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
	// socket.RunSocket()
}
