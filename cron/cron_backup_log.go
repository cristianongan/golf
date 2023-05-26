package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"github.com/bsm/redislock"
)

func runAddLogBackupJob() {
	// Để xử lý cho chạy nhiều instance Server
	redisKey := datasources.GetRedisKeyLockerAddLogBackup()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 10*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[BACKUP LOG] runAddLogBackupJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())
	// Logic chạy cron bên dưới
	addLogBackup()
}

// Add log thao tác
func addLogBackup() {
	// Get data by page
	page := models.Page{
		Limit: 100,
		Page:  1,
	}
	// Danh sách tháng thao tác
	listMonth := make(map[string][]models.OperationLog)
	listDelete := make(map[string][]int64)

	// get date to find
	currentYear, currentMonth, _ := time.Now().Date()
	currentLocation := time.Now().Location()
	date := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).Unix()

	// get log
	logReq := models.OperationLog{}

	logReq.CreatedAt = date

	listLogs, _, err := logReq.FindList(page)
	if err != nil {
		log.Println("Find list log", err.Error())
	}

	// Phân chia data vào tháng phù hợp
	for _, log := range listLogs {
		month := time.Unix(log.CreatedAt, 0).Format(constants.MONTH_FORMAT_LOG)

		listMonth[month] = append(listMonth[month], log)
		listDelete[month] = append(listDelete[month], log.Id)
	}

	// Create batch data
	for month := range listMonth {
		tbName := "operation_logs_" + month
		err := BatchInsertLog(tbName, listMonth[month])
		if err == nil {
			// Delete data sau khi add vào bảng khác
			logDelete := models.OperationLog{}

			errD := logDelete.BatchDeleteLog(listDelete[month])
			if errD != nil {
				log.Println("Table log batch delete err: ", errD.Error())
			}
		}
	}
}

// Kiểm data theo key trong redis
func checkTableExistByTBName(tbName string) bool {
	strData, err := datasources.GetCache(datasources.GetTableLogRedisKey(tbName))
	if err != nil || strData == "" {
		return false
	}
	return true
}

// Tạo data cho table log mới
func BatchInsertLog(tbName string, list []models.OperationLog) error {
	db := datasources.GetDatabaseAuth()
	var err error
	//Check table is Exits from redis
	if !checkTableExistByTBName(tbName) {
		datasources.SetCache(datasources.GetTableLogRedisKey(tbName), tbName, 246060)
		err = db.Table(tbName).AutoMigrate(&models.OperationLog{})
		err = db.Table(tbName).Create(&list).Error
	} else {
		err = db.Table(tbName).Create(&list).Error
	}

	if err != nil {
		log.Println("Table log batch insert err: ", err.Error())
	}
	return err
}
