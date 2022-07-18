package logger

import (
	"fmt"
	"github.com/ivpusic/golog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"start/constants"
	"start/models"
	"time"
)

type ActivityLog struct {
	models.ModelId
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	UserUid    string `json:"user_uid"`
	Action     string `json:"action"`   //customer_update_info, agency_update_info
	Category   string `json:"category"` //customer, agency
	Label      string `json:"label"`    //create, update, delete
	Value      string `json:"value"`    //create_data, update_data, delete_id
}

const EVENT_CATEGORY_SYSTEM = "SYSTEM_ACTIVITY_LOG"
const EVENT_CATEGORY_CUSTOMER = "CUSTOMER_ACTIVITY_LOG"
const EVENT_CATEGORY_AGENCY = "AGENCY_ACTIVITY_LOG"

type ActivityMysqlAppender struct {
	db *gorm.DB
}

func (_ ActivityMysqlAppender) Id() string {
	return "activity-mysql-appender"
}

func (activityMysql ActivityMysqlAppender) Append(activityGoLog golog.Log) {
	activityGoLogData0, ok := activityGoLog.Data[0].(map[string]string)
	if !ok {
		panic(fmt.Sprint("activity_go_log_data_0 is invalid"))
	}

	now := time.Now()

	systemActivityLog := ActivityLog{
		ModelId: models.ModelId{
			CreatedAt: now.Unix(),
			UpdatedAt: now.Unix(),
			Status:    constants.STATUS_ENABLE,
		},
		PartnerUid: activityGoLogData0["partner_uid"],
		CourseUid:  activityGoLogData0["course_uid"],
		UserUid:    activityGoLogData0["user_uid"],
		Action:     activityGoLogData0["action"],
		Category:   activityGoLogData0["category"],
		Label:      activityGoLogData0["label"],
		Value:      activityGoLogData0["value"],
	}

	if err := activityMysql.db.Create(&systemActivityLog).Error; err != nil {
		panic(err.Error())
	}
}

func ActivityMysql(cnf golog.Conf) *ActivityMysqlAppender {
	params := "charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True"
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		cnf["user"],
		cnf["password"],
		cnf["host"],
		cnf["port"],
		cnf["db_name"], params)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:               args,
		DefaultStringSize: 256,
	}), &gorm.Config{})

	if err != nil {
		panic(fmt.Sprintf("failed to connect database @ %s:%s", cnf["host"], cnf["port"]))
	}

	return &ActivityMysqlAppender{
		db: db,
	}
}
