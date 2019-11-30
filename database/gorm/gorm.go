package gorm

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type GormInterface struct {
	gormDB *gorm.DB
}

func RegisterDB(maxIdle, maxConn int, dbLink string, prefix string) (*GormInterface, error) {
	if dbLink == "" {
		return nil, fmt.Errorf("no db link")
	}

	db, err := gorm.Open("mysql", dbLink)
	if err != nil {
		return nil, fmt.Errorf("RegisterDateBase error: " + err.Error())
	}

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxConn)

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return prefix + "_" + defaultTableName
	}

	return &GormInterface{gormDB: db}, nil
}

func (db *GormInterface) RegisterTable(modules ...interface{}) error {
	// db.gormDB.SingularTable(true)
	for _, module := range modules {
		if db.gormDB.HasTable(module) {
			continue
		}

		db.gormDB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(module)
	}
	return nil
}

func (db *GormInterface) Insert(data interface{}) (int64, error) {
	dbTemp := db.gormDB.Create(data)
	if dbTemp.Error != nil {
		return 0, dbTemp.Error
	}
	return dbTemp.RowsAffected, nil
}

func (db *GormInterface) Update(data interface{}, newData interface{}) error {
	dbTemp := db.gormDB.Model(data).Update(newData)
	if dbTemp.Error != nil {
		return dbTemp.Error
	}
	return nil
}

func (db *GormInterface) QueryByLimit(limit string, limitData interface{}, data interface{}) error {
	dbTemp := db.gormDB.Where(limit, limitData).Find(data)
	if dbTemp.Error != nil {
		return dbTemp.Error
	}
	return nil
}

// func (db *GormInterface) QueryBannerBySequence(seq int, data interface{}) error {
// 	dbTemp := db.gormDB.Where("sequence = ?", seq).First(data)
// 	if dbTemp.Error != nil {
// 		return dbTemp.Error
// 	}
// 	return nil
// }

func (db *GormInterface) Query(data interface{}) error {
	dbTemp := db.gormDB.Find(data)
	if dbTemp.Error != nil {
		return dbTemp.Error
	}
	return nil
}
