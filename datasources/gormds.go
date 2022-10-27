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
var dbRole *gorm.DB

func MySqlConnect() {
	var err error
	var errDbRole error

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
	sqlDB, err := db.DB()

	// https://github.com/go-sql-driver/mysql#important-settings
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(5)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(25)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(4 * time.Minute)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if config.GetDbDebug() {
		db.Debug()
	}
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{mysql.Open(args2)},
	}, config.GetDbName2()))

	/// Database Role
	argsDBRole := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.GetDbAuthUser(),
		config.GetDbAuthPassword(),
		config.GetDbAuthHost(),
		config.GetDbAuthPort(),
		config.GetDbAuthName(), params)

	dbRole, errDbRole = gorm.Open(mysql.Open(argsDBRole), &gorm.Config{})

	if errDbRole != nil {
		panic(fmt.Sprintf("failed to connect database @ %s:%s", config.GetDbAuthHost(), config.GetDbAuthPort()))
	}

}

func GetDatabase() *gorm.DB {
	return db
}

func GetDatabaseAuth() *gorm.DB {
	return dbRole
}

func GetDatabaseWithPartner(partnerUid string) *gorm.DB {
	if partnerUid == "FLC" {
		return db.Clauses(dbresolver.Use(config.GetDbName2()))
	}
	return db
}
