package datasources

import (
	"fmt"
	"log"
	"os"
	"start/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var db *gorm.DB

func MySqlConnect() {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			// LogLevel:                  logger.Info, // Log level: show log
			LogLevel:                  logger.Silent, // Log level: disable log
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	params := "charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True"
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.GetDbUser(),
		config.GetDbPassword(),
		config.GetDbHost(),
		config.GetDbPort(),
		config.GetDbName(), params)
	args2 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.GetDbUser2(),
		config.GetDbPassword2(),
		config.GetDbHost2(),
		config.GetDbPort2(),
		config.GetDbName2(), params)
	db, err = gorm.Open(mysql.New(mysql.Config{DSN: args,
		DefaultStringSize: 256,
	}), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to connect database @ %s:%s", config.GetDbHost(), config.GetDbPort()))
	}

	if config.GetDbDebug() {
		db.Debug()
	}
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{mysql.Open(args2)},
	}, config.GetDbName2()))
}

func GetDatabase() *gorm.DB {
	return db
}

func GetDatabaseWithPartner(partnerUid string) *gorm.DB {
	if partnerUid == "FLC" {
		return db.Clauses(dbresolver.Use(config.GetDbName2()))
	}
	return db
}
