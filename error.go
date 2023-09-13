package mysql

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/go-tron/base-error"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"reflect"
)

var (
	ErrorQuery = baseError.WrapFactoryStack(3, "1100")

	ErrorModel             = baseError.SystemFactoryStack(3, "1101", "model is not ptr or mismatch")
	ErrorPrimaryKeyUnset   = baseError.SystemFactoryStack(3, "1102", "model primary key is undefined")
	ErrorPrimaryKeyInvalid = baseError.SystemFactoryStack(3, "1103", "model primary key is invalid")
	ErrorPrimaryKeyEmpty   = baseError.SystemFactoryStack(3, "1104", "model primary key is empty")
	ErrorRecordNotUnique   = baseError.SystemFactoryStack(3, "1110", "find duplicate record")
	ErrorRecordNotFound    = baseError.SystemFactoryStack(3, "1111")
	ErrorRecordNotAffected = baseError.SystemFactoryStack(3, "1112")
	ErrorPluck             = baseError.SystemFactoryStack(3, "1113", "pluck not supplied")
	ErrorSymbol            = baseError.SystemFactoryStack(3, "1114", "symbol not exists")
	ErrorValue             = baseError.SystemFactoryStack(3, "1115")

	UniqueIndexErrorCodes        = []string{"1120", "1121", "1122", "1123", "1124", "1125", "1126", "1127"}
	ErrorUniqueIndex             = baseError.FactoryStack(3, "1120")
	ErrorUniqueIndexUnset        = baseError.SystemFactoryStack(3, "1121", "data duplicate(01)")
	ErrorUniqueIndexEmpty        = baseError.SystemFactoryStack(3, "1122", "data duplicate(02)")
	ErrorUniqueIndexMisMatch     = baseError.SystemFactoryStack(3, "1123", "data duplicate(03)")
	ErrorUniqueIndexNameUnset    = baseError.SystemFactoryStack(3, "1124", "data duplicate(04)")
	ErrorUniqueIndexNameEmpty    = baseError.SystemFactoryStack(3, "1125", "data duplicate(05)")
	ErrorUniqueIndexMessageUnset = baseError.SystemFactoryStack(3, "1126", "data duplicate(06)")
	ErrorUniqueIndexColumnUnset  = baseError.SystemFactoryStack(3, "1127", "data duplicate(07)")
)

func (db *DB) IsUniqueIndexError(err error) bool {
	return IsUniqueIndexError(err)
}

func (db *DB) IsNotSingleError(err error) bool {
	return IsNotSingleError(err)
}

func (db *DB) IsRecordNotFoundError(err error) bool {
	return IsRecordNotFoundError(err)
}

func (db *DB) IsRecordNotAffectedError(err error) bool {
	return IsRecordNotAffectedError(err)
}

func IsUniqueIndexError(err error) bool {
	errType := reflect.TypeOf(err).String()
	if errType == "*mysql.MySQLError" && err.(*mysql.MySQLError).Number == 1062 {
		return true
	}
	if errType == "*baseError.Error" && funk.IndexOf(UniqueIndexErrorCodes, err.(*baseError.Error).Code) != -1 {
		return true
	}
	return false
}

func IsNotSingleError(err error) bool {
	if reflect.TypeOf(err).String() == "*baseError.Error" && err.(*baseError.Error).Code == "1110" {
		return true
	}
	return false
}

func IsRecordNotFoundError(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	if reflect.TypeOf(err).String() == "*baseError.Error" && err.(*baseError.Error).Code == "1111" {
		return true
	}
	return false
}

func IsRecordNotAffectedError(err error) bool {
	if reflect.TypeOf(err).String() == "*baseError.Error" && err.(*baseError.Error).Code == "1112" {
		return true
	}
	return false
}
