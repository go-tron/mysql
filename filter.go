package mysql

import (
	"fmt"
	"github.com/go-tron/types/fieldUtil"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

const (
	SymbolEquals            = "eq"
	SymbolNotEquals         = "ne"
	SymbolGreatThanOrEquals = "gte"
	SymbolGreatThan         = "gt"
	SymbolLessThanOrEquals  = "lte"
	SymbolLessThan          = "lt"
	SymbolLike              = "like"
	SymbolNotLike           = "notLike"
	SymbolIn                = "in"
	SymbolNotIn             = "notIn"
)

var Symbol = map[string]string{
	SymbolEquals:            "=",
	SymbolNotEquals:         "!=",
	SymbolGreatThanOrEquals: ">=",
	SymbolGreatThan:         ">",
	SymbolLessThanOrEquals:  "<=",
	SymbolLessThan:          "<",
	SymbolLike:              "like",
	SymbolNotLike:           "not like",
	SymbolIn:                "in",
	SymbolNotIn:             "not in",
}

type FilterKey struct {
	Column          string
	Operator        string
	IgnoreZeroValue bool
}

func KeyFormat(key string) *FilterKey {
	var filterKey FilterKey

	if strings.HasPrefix(key, "?") {
		filterKey.IgnoreZeroValue = true
		key = strings.Replace(key, "?", "", 1)
	}

	nIndex := strings.Index(key, "#")
	if nIndex != -1 {
		key = key[:nIndex]
	}

	oIndex := strings.Index(key, "$")
	if oIndex == -1 {
		filterKey.Column = key
	} else {
		filterKey.Column = key[:oIndex]
		filterKey.Operator = key[oIndex+1:]
	}
	return &filterKey
}

func (db *DB) Filters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {

	var wheres [][]interface{}
	for key, val := range filters {
		if val == nil {
			continue
		}

		filterKey := KeyFormat(key)
		if filterKey.IgnoreZeroValue && fieldUtil.IsEmpty(val) {
			continue
		}

		switch filterKey.Operator {
		case SymbolEquals:
			wheres = append(wheres, []interface{}{filterKey.Column + " = ?", val})
		case SymbolNotEquals:
			wheres = append(wheres, []interface{}{filterKey.Column + " != ?", val})
		case SymbolGreatThanOrEquals:
			wheres = append(wheres, []interface{}{filterKey.Column + " >= ?", val})
		case SymbolGreatThan:
			wheres = append(wheres, []interface{}{filterKey.Column + " > ?", val})
		case SymbolLessThanOrEquals:
			wheres = append(wheres, []interface{}{filterKey.Column + " <= ?", val})
		case SymbolLessThan:
			wheres = append(wheres, []interface{}{filterKey.Column + " < ?", val})
		case SymbolLike:
			wheres = append(wheres, []interface{}{filterKey.Column + " like ?", fmt.Sprintf("%%%s%%", val)})
		case SymbolNotLike:
			wheres = append(wheres, []interface{}{filterKey.Column + " not like ?", fmt.Sprintf("%%%s%%", val)})
		case SymbolIn:
			wheres = append(wheres, []interface{}{filterKey.Column + " in (?)", val})
		case SymbolNotIn:
			wheres = append(wheres, []interface{}{filterKey.Column + " not in (?)", val})
		default:
			if reflect.ValueOf(val).Kind() == reflect.Slice {
				wheres = append(wheres, []interface{}{filterKey.Column + " in (?)", val})
			} else {
				wheres = append(wheres, []interface{}{filterKey.Column + " = ?", val})
			}
			//return nil,ErrorSymbol()
		}
	}

	for _, where := range wheres {
		query = query.Where(where[0], where[1:]...)
	}
	return query
}
