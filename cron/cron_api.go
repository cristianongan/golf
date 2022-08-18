package cron

import (
	"io/ioutil"
	"log"
	"net/http"
	"start/config"
	"start/constants"
	"time"

	"github.com/robfig/cron/v3"
)

func InitCronJobCallApi() {
	//create new cron
	c := cron.New()

	// Backup booking
	log.Println("CronJobApi can backup order: ", config.GetCronBackupOrderRunning())
	if config.GetCronBackupOrderRunning() {
		log.Println("CronJobApi backup order Started")
		//c.AddFunc("@every 5m", backupOrder) /// Backup booking
	}

	c.Start()
}

// func backupOrder() {
// 	requestCronJob("POST", constants.URL_CRONJOB_BACKUP_ORDER)
// }

func requestCronJob(method, endpoint string) {
	url := config.GetUrlRoot() + endpoint

	// =======================================
	req, errNewRequest := http.NewRequest(method, url, nil)
	if errNewRequest != nil {
		log.Println(errNewRequest.Error())
		return
	}
	req.Header.Set("Authorization", config.GetCronJobSecretKey())

	client := &http.Client{
		Timeout: time.Second * constants.TIMEOUT}
	resp, errRequest := client.Do(req)
	if errRequest != nil {
		log.Println("Error_SET_TIMEOUT: ", errRequest)
		return
	}
	defer resp.Body.Close()

	responseData, errReadBody := ioutil.ReadAll(resp.Body)
	if errReadBody != nil {
		log.Println(errReadBody)
		if responseData != nil {
			log.Println(string(responseData))
		}
	}
}
