package mysql

import (
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"reflect"
)

//WARNING when update with struct, GORM will only update those fields that with non blank value
//For below Update, nothing will be updated as "", 0, false are blank values of their types

// NOTE When query with struct, GORM will only query with those fields has non-zero value,
// that means if your field’s value is 0, ”, false or other zero values, it won’t be used to build query conditions
type DB struct {
	Config *Config
	*gorm.DB
}

func (db *DB) Create(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, _ := db.QueryBuilder(model, opts...)
	if err := query.Create(model).Error; err != nil {
		if db.IsUniqueIndexError(err) {
			return GetUniqueIndexError(model, err.Error())
		}
		return ErrorQuery(err)
	}
	return nil
}

func (db *DB) Count(model interface{}, opts ...Option) (int, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return 0, ErrorModel()
	}
	query, _ := db.QueryBuilder(model, opts...)
	var count int64 = 0
	if err := db.CountBuilder(query).Count(&count).Error; err != nil {
		return 0, ErrorQuery(err)
	}
	return int(count), nil
}

func (db *DB) FindById(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}

	query, queryOpt := db.QueryBuilder(model, opts...)

	if _, err := db.validatePK(model, queryOpt.PrimaryKey); err != nil {
		return err
	}

	if err := query.Take(model).Error; err != nil {
		if db.IsRecordNotFoundError(err) {
			if queryOpt.IgnoreNotFound {
				return nil
			} else {
				if err := queryOpt.ErrorNotFound; err != nil {
					return err
				}
				return GetRecordNotFoundError(model)
			}
		} else {
			return ErrorQuery(err)
		}
	}
	return nil
}

func (db *DB) FindOne(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}

	query, queryOpt := db.QueryBuilder(model, opts...)

	list := reflect.New(reflect.SliceOf(reflect.TypeOf(model).Elem()))
	if err := query.Find(list.Interface()).Error; err != nil {
		return ErrorQuery(err)
	}

	count := list.Elem().Len()
	if count == 0 {
		if queryOpt.IgnoreNotFound {
			return nil
		} else {
			if err := queryOpt.ErrorNotFound; err != nil {
				return err
			}
			return GetRecordNotFoundError(model)
		}
	}
	if count > 1 {
		if err := queryOpt.ErrorNotSingle; err != nil {
			return err
		}
		return ErrorRecordNotUnique()
	}

	reflect.ValueOf(model).Elem().Set(list.Elem().Index(0))
	return nil
}

func (db *DB) CloneById(model interface{}, opts ...Option) (interface{}, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, ErrorModel()
	}
	clone := reflect.New(reflect.TypeOf(model).Elem()).Interface()

	pk, err := db.validatePK(model)
	if err != nil {
		return nil, err
	}
	if err := db.setPKValue(clone, pk.Name, pk.Value); err != nil {
		return nil, err
	}
	if err := db.FindById(clone, opts...); err != nil {
		return nil, err
	}
	return clone, nil
}

func (db *DB) CloneOne(model interface{}, opts ...Option) (interface{}, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, ErrorModel()
	}
	clone := reflect.New(reflect.TypeOf(model).Elem()).Interface()
	if err := copier.Copy(clone, model); err != nil {
		return nil, err
	}
	if err := db.FindOne(clone, opts...); err != nil {
		return nil, err
	}
	return clone, nil
}

func (db *DB) delete(model interface{}, query *gorm.DB, queryOpt *QueryOption) error {
	query = query.Delete(model)
	if err := query.Error; err != nil {
		return ErrorQuery(err)
	}
	if query.RowsAffected == 0 && queryOpt.MustAffected {
		if err := queryOpt.ErrorNotAffected; err != nil {
			return err
		}
		return GetRecordNotAffectedError(model)
	}
	return nil
}

func (db *DB) DeleteAll(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	return db.delete(model, query, queryOpt)
}

func (db *DB) DeleteById(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	if _, err := db.validatePK(model, queryOpt.PrimaryKey); err != nil {
		return err
	}
	return db.delete(model, query, queryOpt)
}

func (db *DB) DeleteOne(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	list := reflect.New(reflect.SliceOf(reflect.TypeOf(model).Elem()))
	if err := query.Find(list.Interface()).Error; err != nil {
		return ErrorQuery(err)
	}
	count := list.Elem().Len()
	if count == 0 {
		if queryOpt.IgnoreNotFound {
			return nil
		} else {
			if err := queryOpt.ErrorNotFound; err != nil {
				return err
			}
			return GetRecordNotFoundError(model)
		}
	}
	if count > 1 {
		if err := queryOpt.ErrorNotSingle; err != nil {
			return err
		}
		return ErrorRecordNotUnique()
	}
	return db.delete(model, query, queryOpt)
}

func (db *DB) update(model interface{}, updates interface{}, query *gorm.DB, queryOpt *QueryOption) (int, error) {
	if len(queryOpt.Attend) > 0 {
		attends := make([]interface{}, 0)
		for _, attend := range queryOpt.Attend {
			attends = append(attends, attend)
		}
		if len(attends) == 1 {
			query = query.Select(attends[0])
		} else {
			query = query.Select(attends[0], attends[1:]...)
		}
	}
	query = query.Updates(updates)
	if err := query.Error; err != nil {
		if db.IsUniqueIndexError(err) {
			return 0, GetUniqueIndexError(model, err.Error())
		}
		return 0, ErrorQuery(err)
	}

	if query.RowsAffected == 0 && queryOpt.MustAffected {
		if err := queryOpt.ErrorNotAffected; err != nil {
			return 0, err
		}
		return 0, GetRecordNotAffectedError(model)
	}
	return int(query.RowsAffected), nil
}

func (db *DB) UpdateAll(model interface{}, updates interface{}, opts ...Option) (int, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return 0, ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	return db.update(model, updates, query, queryOpt)
}

func (db *DB) UpdateById(model interface{}, values interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	if db.isStruct(values) {
		_, err := db.UpdateByIdWithChangedValues(model, values, opts...)
		return err
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	if _, err := db.validatePK(model, queryOpt.PrimaryKey); err != nil {
		return err
	}
	_, err := db.update(model, values, query, queryOpt)
	return err
}

func (db *DB) UpdateByIdWithChangedValues(model interface{}, values interface{}, opts ...Option) (map[string]interface{}, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, ErrorModel()
	}
	clone, err := db.CloneById(model, opts...)
	if err != nil {
		return nil, err
	}
	if clone == nil {
		return nil, nil
	}
	updates, err := db.getUpdateValue(clone, values)
	if err != nil {
		return nil, err
	}
	_, err = db.UpdateAll(model, updates, opts...)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (db *DB) UpdateOne(model interface{}, values interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	if db.isStruct(values) {
		_, err := db.UpdateOneWithChangedValues(model, values, opts...)
		return err
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	list := reflect.New(reflect.SliceOf(reflect.TypeOf(model).Elem()))
	if err := query.Find(list.Interface()).Error; err != nil {
		return ErrorQuery(err)
	}
	count := list.Elem().Len()
	if count == 0 {
		if queryOpt.IgnoreNotFound {
			return nil
		} else {
			if err := queryOpt.ErrorNotFound; err != nil {
				return err
			}
			return GetRecordNotFoundError(model)
		}
	}
	if count > 1 {
		if err := queryOpt.ErrorNotSingle; err != nil {
			return err
		}
		return ErrorRecordNotUnique()
	}
	_, err := db.update(model, values, query, queryOpt)
	return err
}

func (db *DB) UpdateOneWithChangedValues(model interface{}, values interface{}, opts ...Option) (map[string]interface{}, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, ErrorModel()
	}
	clone, err := db.CloneOne(model, opts...)
	if err != nil {
		return nil, err
	}
	if clone == nil {
		return nil, nil
	}
	updates, err := db.getUpdateValue(clone, values)
	if err != nil {
		return nil, err
	}
	_, err = db.UpdateAll(model, updates, opts...)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (db *DB) Find(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)

	if queryOpt.First {
		query.First(model)
	} else if queryOpt.Last {
		query.Last(model)
	} else {
		query.Find(model)
	}

	if err := query.Error; err != nil {
		if db.IsRecordNotFoundError(err) {
			if queryOpt.IgnoreNotFound || queryOpt.First || queryOpt.Last {
				return nil
			} else {
				if err := queryOpt.ErrorNotFound; err != nil {
					return err
				}
				return GetRecordNotFoundError(model)
			}
		} else {
			return ErrorQuery(err)
		}
	}
	return nil
}

func (db *DB) FindAllWithModel(model interface{}, opts ...Option) (interface{}, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	query = db.DefaultSort(model, query, queryOpt)

	list := reflect.New(reflect.SliceOf(reflect.TypeOf(model).Elem()))
	if err := query.Find(list.Interface()).Error; err != nil {
		return nil, ErrorQuery(err)
	}
	return list.Elem().Interface(), nil
}

func (db *DB) FindPageWithModel(model interface{}, opts ...Option) (interface{}, int, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return nil, 0, ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	query = db.DefaultSort(model, query, queryOpt)

	list := reflect.New(reflect.SliceOf(reflect.TypeOf(model).Elem()))
	if err := query.Find(list.Interface()).Error; err != nil {
		return nil, 0, ErrorQuery(err)
	}

	var total int64 = 0
	if queryOpt.Pageable != nil {
		if err := db.CountBuilder(query).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}
	return list.Elem().Interface(), int(total), nil
}

func (db *DB) FindAll(list interface{}, opts ...Option) error {
	listT := reflect.TypeOf(list)
	if listT.Kind() != reflect.Ptr || listT.Elem().Kind() != reflect.Slice {
		return ErrorModel()
	}
	if !(listT.Elem().Elem().Kind() == reflect.Struct || (listT.Elem().Elem().Kind() == reflect.Ptr && listT.Elem().Elem().Elem().Kind() == reflect.Struct)) {
		return ErrorModel()
	}

	elem := listT.Elem().Elem()
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	model := reflect.New(elem).Interface()
	query, queryOpt := db.QueryBuilder(model, opts...)
	query = db.DefaultSort(model, query, queryOpt)

	//TODO: cache
	//sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
	//	return tx.Find(list)
	//})

	if err := query.Find(list).Error; err != nil {
		return ErrorQuery(err)
	}
	return nil
}

func (db *DB) FindPage(list interface{}, opts ...Option) (int, error) {
	listT := reflect.TypeOf(list)
	if listT.Kind() != reflect.Ptr || listT.Elem().Kind() != reflect.Slice {
		return 0, ErrorModel()
	}
	if !(listT.Elem().Elem().Kind() == reflect.Struct || (listT.Elem().Elem().Kind() == reflect.Ptr && listT.Elem().Elem().Elem().Kind() == reflect.Struct)) {
		return 0, ErrorModel()
	}

	elem := listT.Elem().Elem()
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	model := reflect.New(elem).Interface()
	query, queryOpt := db.QueryBuilder(model, opts...)
	query = db.DefaultSort(model, query, queryOpt)

	var total int64 = 0
	if err := query.Find(list).Error; err != nil {
		return 0, ErrorQuery(err)
	}
	if queryOpt.Pageable != nil {
		if err := db.CountBuilder(query).Count(&total).Error; err != nil {
			return 0, err
		}
	}
	return int(total), nil
}

func (db *DB) FindPluck(model interface{}, opts ...Option) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrorModel()
	}
	query, queryOpt := db.QueryBuilder(model, opts...)
	if queryOpt.Pluck == nil {
		return ErrorPluck()
	}
	if err := query.Pluck(queryOpt.Pluck[0].(string), queryOpt.Pluck[1]).Error; err != nil {
		return ErrorQuery(err)
	}
	return nil
}

func (db *DB) isStruct(value interface{}) bool {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		return reflect.ValueOf(value).Elem().Kind() == reflect.Struct
	} else {
		return v.Kind() == reflect.Struct
	}
}
