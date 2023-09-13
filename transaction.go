package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func (db *DB) Transaction(f func(tx *gorm.DB) error) (err error) {
	tx := db.DB.Begin()
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
