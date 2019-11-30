package db

import (
	"github.com/go-web/go-banner-web/conf"
	"github.com/go-web/go-banner-web/db/gorm"
)

type DBInterface interface {
	RegisterTable(modules ...interface{}) error
	Insert(data interface{}) (int64, error)
	Update(data interface{}, newData interface{}) error
	QueryByLimit(limit string, limitData interface{}, data interface{}) error
	Query(data interface{}) error
}

func NewDBInstance(prefix string) (DBInterface, error) {
	runMode := conf.AppConfig.String("runmode")
	// (可选)设置最大空闲连接
	maxIdle := conf.AppConfig.DefaultInt(runMode+".db.maxIdle", 60)
	// (可选) 设置最大数据库连接 (go >= 1.2)
	maxConn := conf.AppConfig.DefaultInt(runMode+".db.maxConn", 320)
	dbLink := conf.AppConfig.DefaultString(runMode+".db.link", "")

	db, err := gorm.RegisterDB(maxIdle, maxConn, dbLink, prefix)
	if err != nil {
		return nil, err
	}
	return db, nil
}
