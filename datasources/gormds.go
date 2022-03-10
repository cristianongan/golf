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
)

var db *gorm.DB

func MySqlConnect() {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
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
}

func GetDatabase() *gorm.DB {
	return db
}
