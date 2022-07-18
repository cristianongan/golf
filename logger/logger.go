package logger

import (
	"github.com/ivpusic/golog"
	"github.com/ivpusic/golog/appenders"
	"start/config"
)

var fileLogger, mysqlLogger *golog.Logger

var activityFileAppender *appenders.FileAppender

var activityMysqlAppender *ActivityMysqlAppender

func InitLogger() {
	fileLogger = golog.Default
	mysqlLogger = golog.Default

	activityFileAppender = appenders.File(golog.Conf{
		"path": "logs/system-activity/log.txt",
	})

	activityMysqlAppender = ActivityMysql(golog.Conf{
		"user":     config.GetDbUser(),
		"password": config.GetDbPassword(),
		"host":     config.GetDbHost(),
		"port":     config.GetDbPort(),
		"db_name":  config.GetDbName(),
	})
}

func GetActivityFileLogger() *golog.Logger {
	fileLogger.Enable(activityFileAppender)
	return fileLogger
}

func GetActivityMysqlLogger() *golog.Logger {
	mysqlLogger.Enable(activityMysqlAppender)
	return mysqlLogger
}
