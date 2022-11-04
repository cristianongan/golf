package server

import (
	"log"
	"net/http"
	"start/config"
	"start/datasources"
	socket "start/socket"

	ccron "start/cron"
	// "start/datasources/aws"
	// "start/datasources/elasticsearch"
)

func Init() {
	log.Println("server init")

	config := config.GetConfig()

	// --- Socket ---
	// go socket.RunSocket(config.GetString("socket_port"))
	http.HandleFunc("/socket", socket.Echo)
	http.HandleFunc("/", socket.Home)
	go http.ListenAndServe(config.GetString("socket_port"), nil)

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

	// routers := r.Routes()
	// Init authority
	// initAuthority(routers)

	// logger.InitLogger()

	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
}
