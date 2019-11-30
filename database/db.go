package db

import (
	"github.com/fengfenghuo/go-common-lib/database/gorm"
)

type DBInterface interface {
	RegisterTable(modules ...interface{}) error
	Insert(data interface{}) (int64, error)
	Update(data interface{}, newData interface{}) error
	QueryByLimit(limit string, limitData interface{}, data interface{}) error
	Query(data interface{}) error
}

func NewDBInstance(maxIdle, maxConn int, dbLink, prefix string) (DBInterface, error) {
	db, err := gorm.RegisterDB(maxIdle, maxConn, dbLink, prefix)
	if err != nil {
		return nil, err
	}
	return db, nil
}
