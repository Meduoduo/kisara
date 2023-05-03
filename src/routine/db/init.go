package db

import (
	"fmt"

	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
	ORM for SQLite
*/

var kisaraDB *gorm.DB

var (
	ErrNotFound = gorm.ErrRecordNotFound
)

func InitKisaraDB(db_path string) {
	db, err := gorm.Open(sqlite.Open(db_path), &gorm.Config{})
	if err != nil {
		log.Panic("[Kisara] Failed to init Kisara DB: " + err.Error())
	}

	kisaraDB = db

	registerModel()
}

func registerModel() {
	kisaraDB.AutoMigrate(&types.DBContainer{})
	kisaraDB.AutoMigrate(&types.DBService{})
	kisaraDB.AutoMigrate(&types.DBImage{})
}

func CreateGeneric[T any](data *T) error {
	return kisaraDB.Create(data).Error
}

func UpdateGeneric[T any](data *T) error {
	return kisaraDB.Save(data).Error
}

func DeleteGeneric[T any](data *T) error {
	return kisaraDB.Delete(data).Error
}

type genericComparableConstraint interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

type genericEqualConstraint interface {
	genericComparableConstraint | string
}

type genericQuery func(tx *gorm.DB) *gorm.DB

func GenericEqual[T genericEqualConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s = ?", field), value)
	}
}

func GenericNotEqual[T genericEqualConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s != ?", field), value)
	}
}

func GenericGreaterThan[T genericComparableConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s > ?", field), value)
	}
}

func GenericGreaterThanOrEqual[T genericComparableConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s >= ?", field), value)
	}
}

func GenericLessThan[T genericComparableConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s < ?", field), value)
	}
}

func GenericLessThanOrEqual[T genericComparableConstraint](field string, value T) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s <= ?", field), value)
	}
}

func GenericInArray(field string, value []interface{}) genericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s IN ?", field), value)
	}
}

func GetGenericOne[T any](query ...genericQuery) (T /* data */, error) {
	var data T
	tmp := kisaraDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.First(&data).Error
	return data, err
}

func GetGenericAll[T any](query ...genericQuery) ([]T /* data */, error) {
	var data []T
	tmp := kisaraDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.Find(&data).Error
	return data, err
}
