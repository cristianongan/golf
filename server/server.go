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
	// fs := http.FileServer(http.Dir("../public"))
	// http.Handle("/", fs)

	if config.GetBool("is_open_socket") {
		// Configure websocket route
		http.HandleFunc("/ws", socket.HandleConnections)

		// Start listening for incoming chat messages
		go socket.HandleMessages()

		// Start the server on localhost port 8000 and log any errors
		log.Println("socket http server started on :8000")
		a := func() {
			err := http.ListenAndServe(":8000", nil)
			log.Println("ListenAndServe", err)
		}
		go a()
	}

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

	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
}
