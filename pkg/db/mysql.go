package db

import (
	"time"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDB struct {
	*gorm.DB
}

func NewMysqlDB(config *config.Conf, logger *logger.Logger) *MysqlDB {

	gormLogger := NewGormZapLogger(logger.Log())

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       config.Mysql.DSN,
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Log().Error("数据库连接错误", zap.Error(err))
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Log().Error("数据库连接错误", zap.Error(err))
		return nil
	}

	sqlDB.SetMaxOpenConns(config.Mysql.Pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Mysql.Pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Mysql.Pool.MaxLifetime) * time.Second)

	return &MysqlDB{DB: db}
}
