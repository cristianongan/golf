package cron

import (
	"log"
	"start/config"
	"start/constants"

	"github.com/robfig/cron/v3"
)

func CronStart() {
	c := cron.New()

	//c.AddFunc("@every 5s", checkOrderPayment)

	if config.GetCronIsRunning() {
		log.Println("==== CronStart =====")
		c.Start()
	}
}

func testCron() {
	url := config.GetUrlBackendApi() + constants.URL_CHECK_CRON
	err, statusCode, bResponse := requestToCron(url)
	if err != nil {
		log.Println(err, statusCode, string(bResponse))
	}
}
