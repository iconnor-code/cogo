package database

import (
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDB struct {
	*gorm.DB
	conf   map[string]any
	logger core.ILogger
}

type MysqlDBOption func(db *MysqlDB) error

func WithMysqlConfig(conf core.IConfig) MysqlDBOption {
	return func(db *MysqlDB) error {
		confMap := conf.Get("mysql").(map[string]any)
		db.conf = confMap
		return nil
	}
}

func WithMysqlLogger(logger core.ILogger) MysqlDBOption {
	return func(db *MysqlDB) error {
		db.logger = logger
		return nil
	}
}

func NewMysqlDB(opts ...MysqlDBOption) (*MysqlDB, error) {
	mysqlDB := &MysqlDB{}
	for _, opt := range opts {
		opt(mysqlDB)
	}

	gormLogger := NewGormZapLogger(mysqlDB.logger)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       mysqlDB.conf["dsn"].(string),
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, cerrs.Wrap("failed to open mysql db", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, cerrs.Wrap("failed to get sql db", err)
	}

	sqlDB.SetMaxOpenConns(mysqlDB.conf["max_open_conns"].(int))
	sqlDB.SetMaxIdleConns(mysqlDB.conf["max_idle_conns"].(int))
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlDB.conf["max_lifetime"].(int)) * time.Second)

	mysqlDB.DB = db
	return mysqlDB, nil
}
