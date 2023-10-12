package mysql

import (
	"github.com/go-tron/types/pageable"
	"gorm.io/gorm"
	"reflect"
)

type Option func(*QueryOption)

type Select struct {
	Query string
	Args  []interface{}
}
type QueryOption struct {
	DB               *gorm.DB
	Table            string
	PrimaryKey       string
	Updates          map[string]interface{}
	Select           Select
	Omit             []string
	Attend           []string
	Join             [][]interface{}
	Where            [][]interface{}
	Or               [][]interface{}
	Filters          map[string]interface{}
	Group            string
	Limit            int
	Offset           int
	Pageable         *pageable.Pageable
	Sort             []string
	Pluck            []interface{}
	First            bool
	Last             bool
	WithDeleted      bool
	IgnoreNotFound   bool
	MustAffected     bool
	ErrorNotFound    error
	ErrorNotAffected error
	ErrorNotSingle   error
}

func (db *DB) WithDB(val *gorm.DB) Option {
	return func(opts *QueryOption) {
		opts.DB = val
	}
}
func (db *DB) WithTable(val string) Option {
	return func(opts *QueryOption) {
		opts.Table = val
	}
}
func (db *DB) WithPrimaryKey(val string) Option {
	return func(opts *QueryOption) {
		opts.PrimaryKey = val
	}
}
func (db *DB) WithSelect(query string, args ...interface{}) Option {
	return func(opts *QueryOption) {
		if opts.Select.Query != "" {
			opts.Select.Query += ","
		}
		opts.Select.Query += query
		opts.Select.Args = append(opts.Select.Args, args...)
	}
}
func (db *DB) WithSelectCondition(condition bool, query string, args ...interface{}) Option {
	return func(opts *QueryOption) {
		if condition {
			if opts.Select.Query != "" {
				opts.Select.Query += ","
			}
			opts.Select.Query += query
			opts.Select.Args = append(opts.Select.Args, args...)
		}
	}
}
func (db *DB) WithOmit(val ...string) Option {
	return func(opts *QueryOption) {
		opts.Omit = append(opts.Omit, val...)
	}
}
func (db *DB) WithAttend(val ...string) Option {
	return func(opts *QueryOption) {
		opts.Attend = append(opts.Attend, val...)
	}
}
func (db *DB) WithUpdates(val map[string]interface{}) Option {
	return func(opts *QueryOption) {
		if opts.Updates == nil {
			opts.Updates = map[string]interface{}{}
		}
		for k, v := range val {
			opts.Updates[k] = v
		}
	}
}
func (db *DB) WithJoin(query string, val ...interface{}) Option {
	return func(opts *QueryOption) {
		opts.Join = append(opts.Join, []interface{}{query, val})
	}
}
func (db *DB) WithJoinCondition(condition bool, query string, val ...interface{}) Option {
	return func(opts *QueryOption) {
		if condition {
			opts.Join = append(opts.Join, []interface{}{query, val})
		}
	}
}
func (db *DB) WithWhere(val ...interface{}) Option {
	return func(opts *QueryOption) {
		opts.Where = append(opts.Where, val)
	}
}
func (db *DB) WithWhereCondition(condition bool, val ...interface{}) Option {
	return func(opts *QueryOption) {
		if condition {
			opts.Where = append(opts.Where, val)
		}
	}
}
func (db *DB) WithWheres(val ...[]interface{}) Option {
	return func(opts *QueryOption) {
		for _, val := range val {
			opts.Where = append(opts.Where, val)
		}
	}
}
func (db *DB) WithWheresCondition(condition bool, val ...[]interface{}) Option {
	return func(opts *QueryOption) {
		if condition {
			for _, val := range val {
				opts.Where = append(opts.Where, val)
			}
		}
	}
}
func (db *DB) WithOr(val ...interface{}) Option {
	return func(opts *QueryOption) {
		opts.Or = append(opts.Or, val)
	}
}
func (db *DB) WithOrCondition(condition bool, val ...interface{}) Option {
	return func(opts *QueryOption) {
		if condition {
			opts.Or = append(opts.Or, val)
		}
	}
}
func (db *DB) WithFilters(val ...map[string]interface{}) Option {
	return func(opts *QueryOption) {
		if opts.Filters == nil {
			opts.Filters = map[string]interface{}{}
		}
		for _, m := range val {
			for k, v := range m {
				opts.Filters[k] = v
			}
		}
	}
}
func (db *DB) WithGroup(val string) Option {
	return func(opts *QueryOption) {
		opts.Group = val
	}
}
func (db *DB) WithLimit(val int) Option {
	return func(opts *QueryOption) {
		opts.Limit = val
	}
}
func (db *DB) WithOffset(val int) Option {
	return func(opts *QueryOption) {
		opts.Offset = val
	}
}
func (db *DB) WithPage(page int, size int, sort string) Option {
	return func(opts *QueryOption) {
		opts.Pageable = &pageable.Pageable{
			Page: page, Size: size, Sort: sort,
		}
	}
}
func (db *DB) WithPageable(val *pageable.Pageable) Option {
	return func(opts *QueryOption) {
		opts.Pageable = val
	}
}
func (db *DB) WithSort(val ...string) Option {
	return func(opts *QueryOption) {
		opts.Sort = append(opts.Sort, val...)
	}
}
func (db *DB) WithPluck(column string, val interface{}) Option {
	return func(opts *QueryOption) {
		opts.Pluck = []interface{}{column, val}
	}
}
func (db *DB) WithFirst() Option {
	return func(opts *QueryOption) {
		opts.First = true
	}
}
func (db *DB) WithLast() Option {
	return func(opts *QueryOption) {
		opts.Last = true
	}
}
func (db *DB) WithDeleted() Option {
	return func(opts *QueryOption) {
		opts.WithDeleted = true
	}
}
func (db *DB) WithIgnoreNotFound() Option {
	return func(opts *QueryOption) {
		opts.IgnoreNotFound = true
	}
}
func (db *DB) WithMustAffected() Option {
	return func(opts *QueryOption) {
		opts.MustAffected = true
	}
}
func (db *DB) WithErrorNotFound(val error) Option {
	return func(opts *QueryOption) {
		opts.ErrorNotFound = val
	}
}
func (db *DB) WithErrorNotAffected(val error) Option {
	return func(opts *QueryOption) {
		opts.ErrorNotAffected = val
	}
}
func (db *DB) WithErrorNotSingle(val error) Option {
	return func(opts *QueryOption) {
		opts.ErrorNotSingle = val
	}
}
func (db *DB) QueryBuilder(model interface{}, opts ...Option) (*gorm.DB, *QueryOption) {

	queryOption := &QueryOption{}
	for _, apply := range opts {
		if apply != nil {
			apply(queryOption)
		}
	}

	var query *gorm.DB
	if queryOption.DB != nil {
		query = queryOption.DB
	} else {
		query = db.DB
	}

	if queryOption.Table != "" {
		query = query.Table(queryOption.Table)
	} else if model != nil {
		query = query.Model(model)
	}

	if len(queryOption.Select.Query) > 0 {
		query = query.Select(queryOption.Select.Query, queryOption.Select.Args...)
	}

	if len(queryOption.Omit) > 0 {
		query = query.Omit(queryOption.Omit...)
	}

	if len(queryOption.Join) > 0 {
		for _, joins := range queryOption.Join {
			if joins == nil || joins[0] == nil {
				continue
			}
			if len(joins) == 1 {
				query = query.Joins(joins[0].(string))
			} else {
				query = query.Joins(joins[0].(string), joins[1:]...)
			}
		}
	}

	if queryOption.Group != "" {
		query = query.Group(queryOption.Group)
	}

	if len(queryOption.Where) > 0 {
		for _, where := range queryOption.Where {
			if where == nil || where[0] == nil {
				continue
			}
			if len(where) == 1 {
				//query = query.Where(wheres[0])
				if (reflect.TypeOf(where[0]).Kind() == reflect.Ptr && reflect.TypeOf(where[0]).Elem().Kind() == reflect.Struct) ||
					reflect.TypeOf(where[0]).Kind() == reflect.Struct {
					query = query.Where(db.structToMap(where[0]))
				} else {
					query = query.Where(where[0])
				}
			} else {
				query = query.Where(where[0], where[1:]...)
			}
		}
	}

	if len(queryOption.Or) > 0 {
		for _, or := range queryOption.Or {
			if or == nil || or[0] == nil {
				continue
			}
			if len(or) == 1 {
				//query = query.Or(wheres[0])
				if (reflect.TypeOf(or[0]).Kind() == reflect.Ptr && reflect.TypeOf(or[0]).Elem().Kind() == reflect.Struct) ||
					reflect.TypeOf(or[0]).Kind() == reflect.Struct {
					query = query.Or(db.structToMap(or[0]))
				} else {
					query = query.Or(or[0])
				}
			} else {
				query = query.Or(or[0], or[1:]...)
			}
		}
	}

	if len(queryOption.Sort) != 0 {
		for _, sort := range queryOption.Sort {
			if queryOption.Sort[0] != "" {
				query = query.Order(sort)
			}
		}
	}

	if queryOption.Pageable != nil && queryOption.Pageable.Size > 0 {
		query = query.Limit(queryOption.Pageable.Size).Offset((queryOption.Pageable.Page - 1) * queryOption.Pageable.Size)
		if queryOption.Pageable.Sort != "" {
			query = query.Order(queryOption.Pageable.Sort)
		}
	}
	if queryOption.Limit != 0 {
		query = query.Limit(queryOption.Limit)
	}
	if queryOption.Offset != 0 {
		query = query.Offset(queryOption.Offset)
	}

	if !queryOption.WithDeleted && model != nil {
		modelT := reflect.TypeOf(model)
		if modelT.Kind() == reflect.Ptr {
			modelT = modelT.Elem()
		}
		if _, ok := modelT.FieldByName("Deleted"); ok {
			tableName := db.Config.NamingStrategy.TableName(modelT.Name())
			if v := GetTableName(model); v != "" {
				tableName = v
			}
			query = query.Where(tableName + ".deleted = 0")
		}
	}

	if queryOption.Filters != nil {
		query = db.Filters(query, queryOption.Filters)
	}

	return query, queryOption
}

func (db *DB) CountBuilder(query *gorm.DB) *gorm.DB {
	return query.Select("*").Limit(-1).Offset(-1)
}

func (db *DB) DefaultSort(model interface{}, query *gorm.DB, queryOption *QueryOption) *gorm.DB {
	if (queryOption.Pageable == nil || queryOption.Pageable.Sort == "") && queryOption.Sort == nil {
		if primaryKey := db.getPKName(model); primaryKey != "" {
			query = query.Order(primaryKey + " desc")
		}
	}
	return query
}
