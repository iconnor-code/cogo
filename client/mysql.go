package client

import (
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDB struct {
	*gorm.DB
	conf   core.IConfig
	logger core.ILogger
}

type MysqlDBOption func(db *MysqlDB) error

func NewMysqlDB(config core.IConfig, logger core.ILogger) (*MysqlDB, error) {
	mysqlDB := &MysqlDB{
		conf:   config,
		logger: logger,
	}

	gormLogger := NewGormZapLogger(mysqlDB.logger)

	mysqlConf := mysqlDB.conf.GetMySQL()
	dsn := mysqlConf.DSN

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, cerrs.Wrap(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, cerrs.Wrap(err)
	}

	maxOpenConns := mysqlConf.Pool.MaxOpenConns
	maxIdleConns := mysqlConf.Pool.MaxIdleConns
	maxLifetime := mysqlConf.Pool.MaxLifetime
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)

	mysqlDB.DB = db
	return mysqlDB, nil
}
