package server

import (
	"log"
	"start/config"
	"start/datasources"
	"start/logger"
	socket "start/socket"
	"start/utils"

	ccron "start/cron"
	// "start/datasources/aws"
	// "start/datasources/elasticsearch"
)

func Init() {

	log.Println("server init")

	config := config.GetConfig()

	// Init Logger
	logger.InitLogger()

	// go socket_room.Hub.Run()

	//Test Time
	log.Println("Time now server", utils.GetTimeNow())
	log.Println("Time now local", utils.GetLocalUnixTime())

	// --- Socket ---

	// socket.GetHubSocket() = socket.NewHub()
	socket.InitHubSocket()
	go socket.GetHubSocket().Run()

	// http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	socket.ServeWs(socket.GetHubSocket(), w, r)
	// })

	// listener := func() {
	// 	err := http.ListenAndServe(":8000", nil)
	// 	log.Println("ListenAndServe", err)
	// }
	// go listener()

	// --- Cron ---
	ccron.CronStart()

	datasources.MinioConnect()

	datasources.MySqlConnect()
	MigrateDb()
	// ============ Use redis
	datasources.MyRedisConnect()

	// ======== Connect elasticsearch/
	// elasticsearch.ElasticSearchInit()

	// IMPORT DATA

	r := NewRouter()

	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
}
