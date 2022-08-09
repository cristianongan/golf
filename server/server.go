package server

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"start/config"
	"start/datasources"
	"start/logger"
	"time"
	// "start/datasources/aws"
	// "start/datasources/elasticsearch"
)

func Init() {
	log.Println("server init")

	config := config.GetConfig()
	// cron.CronStart()
	// cron.InitCronJobCallApi()

	datasources.MinioConnect()

	datasources.MySqlConnect()
	MigrateDb()
	// ============ Use redis
	datasources.MyRedisConnect()

	// ======== Connect elasticsearch/
	// elasticsearch.ElasticSearchInit()

	// Init Cron
	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), ">>> [INIT_CRON]")
	c := cron.New(cron.WithSeconds())
	//c.AddFunc("* * * * * *", func() {
	//	fmt.Println("one second")
	//})
	c.Start()

	r := NewRouter()

	routers := r.Routes()

	// Init authority
	initAuthority(routers)

	logger.InitLogger()

	log.Println("Server is running ...", "listen", config.GetString("backend_port"))
	r.Run(config.GetString("backend_port"))
}
