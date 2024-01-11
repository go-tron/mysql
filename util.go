package mysql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func (db *DB) structToMap(obj interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	if obj == nil {
		return m
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return m
		}
		v = v.Elem()
	}
	t := v.Type()

	var keys []string
	for k := 0; k < t.NumField(); k++ {
		keys = append(keys, t.Field(k).Name)
	}

	for _, key := range keys {
		fieldV := v.FieldByName(key)
		name := db.Config.NamingStrategy.ColumnName("", key)
		if fieldT, ok := t.FieldByName(key); ok {
			name = db.getColumnName(fieldT)
		}
		if fieldV.Kind() == reflect.Ptr && !fieldV.IsNil() {
			fieldV = fieldV.Elem()
		}
		m[name] = fieldV.Interface()
	}
	return m
}

func (db *DB) checkUpdateField(valueField reflect.Value, fieldName string, modelV reflect.Value, primaryKey string) error {
	fieldType := valueField.Type().String()
	debugInfo := fmt.Sprintf("%s[%s]:%v", fieldName, fieldType, valueField.Interface())
	if strings.ToLower(primaryKey) == strings.ToLower(fieldName) {
		return errors.New(debugInfo + "!!!skip primaryKey")
	}
	if (fieldType == "localTime.Time" || fieldType == "*localTime.Time") && valueField.IsZero() {
		return errors.New(debugInfo + "!!!skip time zero")
	}
	modelField := modelV.FieldByName(fieldName)
	if !modelField.IsValid() {
		return errors.New(debugInfo + "!!!skip invalid fieldName")
	}
	if modelField.IsZero() && valueField.IsZero() {
		return errors.New(debugInfo + "!!!skip same zero")
	}

	debugInfo += fmt.Sprintf("\nmodel %s:%s:%v", modelField.Type(), modelField.Kind(), modelField.Interface())

	if valueField.Kind() == reflect.Interface && !valueField.IsNil() {
		valueField = valueField.Elem()
	}
	if valueField.Kind() == reflect.Ptr && !valueField.IsNil() {
		valueField = valueField.Elem()
	}

	if modelField.Kind() == reflect.Ptr && !modelField.IsNil() {
		modelField = modelField.Elem()
	}

	if modelField.Interface() == valueField.Interface() {
		return errors.New(debugInfo + "!!!skip same interface value")
	}
	if fmt.Sprintf("%v", modelField) == fmt.Sprintf("%v", valueField) {
		return errors.New(debugInfo + "!!!skip same fmt value")
	}
	return nil
}

func (db *DB) getUpdateValue(model interface{}, value interface{}) (map[string]interface{}, error) {

	var m = map[string]interface{}{}

	if value == nil {
		return m, nil
	}

	valueV := reflect.ValueOf(value)
	valueT := valueV.Type()

	//value类型必须是struct,map及struct的指针
	if valueV.Kind() == reflect.Ptr {
		if reflect.ValueOf(value).Elem().Kind() == reflect.Map {
			return nil, ErrorValue("value is ptr type can't be a map")
		}
		if reflect.ValueOf(value).Elem().Kind() != reflect.Struct {
			return nil, ErrorValue("value is ptr type must be a struct")
		}
		valueV = valueV.Elem()
		valueT = valueT.Elem()
	} else if !(valueV.Kind() == reflect.Struct || valueV.Kind() == reflect.Map) {
		return nil, ErrorValue("value isn't ptr type must be a struct or map")
	}

	modelV := reflect.ValueOf(model)
	if modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}
	modelT := reflect.TypeOf(model)
	if modelT.Kind() == reflect.Ptr {
		modelT = modelT.Elem()
	}
	primaryKey := db.getPKName(model)

	//如果是struct则转换为map
	//移除值与数据库值相同的字段
	//移除时间值为空的字段
	if valueV.Kind() == reflect.Map {
		for _, mapKey := range valueV.MapKeys() {
			valueField := valueV.MapIndex(mapKey)
			fieldName := db.Config.NamingStrategy.SchemaName(mapKey.Interface().(string))
			modelField, found := modelT.FieldByName(fieldName)
			if found && modelField.Tag.Get("gorm") == "-" {
				continue
			}
			if err := db.checkUpdateField(valueField, fieldName, modelV, primaryKey); err != nil {
				//fmt.Println(err)
				continue
			}
			m[fieldName] = valueField.Interface()
		}
	} else {
		for i := 0; i < valueV.NumField(); i++ {
			valueField := valueV.Field(i)
			fieldName := valueT.Field(i).Name
			modelField, found := modelT.FieldByName(fieldName)
			if found && modelField.Tag.Get("gorm") == "-" {
				continue
			}
			if err := db.checkUpdateField(valueField, fieldName, modelV, primaryKey); err != nil {
				//fmt.Println(err)
				continue
			}
			m[fieldName] = valueField.Interface()
		}
	}
	//fmt.Printf("m:%v\n", m)
	return m, nil
}

type PK struct {
	Name  string
	Value interface{}
}

func (db *DB) validatePK(model interface{}, primaryKey ...string) (*PK, error) {
	modelV := reflect.ValueOf(model)
	if modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}

	var pkName string
	if len(primaryKey) > 0 && primaryKey[0] != "" {
		pkName = primaryKey[0]
	} else {
		pkField := GetPKField(model)
		if pkField.Name == "" {
			return nil, ErrorPrimaryKeyUnset()
		}
		pkName = pkField.Name
	}

	fieldV := modelV.FieldByName(pkName)
	if !fieldV.IsValid() {
		return nil, ErrorPrimaryKeyInvalid()
	}
	if fieldV.IsZero() {
		return nil, ErrorPrimaryKeyEmpty()
	}

	return &PK{
		Name:  pkName,
		Value: fieldV.Interface(),
	}, nil
}

func (db *DB) setPKValue(model interface{}, pkFieldName string, value interface{}) error {
	modelV := reflect.ValueOf(model)
	if modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}

	fieldV := modelV.FieldByName(pkFieldName)
	if !(fieldV.IsValid() && fieldV.CanSet()) {
		return ErrorPrimaryKeyInvalid()
	}
	fieldV.Set(reflect.ValueOf(value))
	return nil
}

func (db *DB) getColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("gorm")
	if tag == "" {
		return db.Config.NamingStrategy.ColumnName("", field.Name)
	}
	arr := strings.Split(tag, ",")
	for _, str := range arr {
		if strings.HasPrefix(str, "column:") {
			return db.Config.NamingStrategy.ColumnName("", strings.Replace(str, "column:", "", 1))
		}
	}
	return db.Config.NamingStrategy.ColumnName("", field.Name)
}

func (db *DB) getPKName(model interface{}) string {
	return db.getColumnName(GetPKField(model))
}

func GetPKField(model interface{}) reflect.StructField {
	modelT := reflect.TypeOf(model)
	if modelT.Kind() == reflect.Ptr {
		modelT = modelT.Elem()
	}

	for i := 0; i < modelT.NumField(); i++ {
		tag := modelT.Field(i).Tag.Get("gorm")
		if strings.Contains(tag, "primary_key") {
			return modelT.Field(i)
		}
	}
	return reflect.StructField{}
}

func ModelMethod(model interface{}, methodName string) (r interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	modelV := reflect.ValueOf(model)
	if modelV.Kind() != reflect.Ptr {
		modelV = modelV.Addr()
	}
	method := modelV.MethodByName(methodName)
	result := method.Call(make([]reflect.Value, 0))
	return result[0].Interface(), nil
}

func GetRecordNotFoundError(model interface{}) string {
	result, err := ModelMethod(model, "RecordNotFoundError")
	if err != nil {
		return "record not found"
	}
	return result.(string)
}

func GetRecordNotAffectedError(model interface{}) string {
	result, err := ModelMethod(model, "NoRecordAffectedError")
	if err != nil {
		return "record for update not found"
	}
	return result.(string)
}

func GetTableName(model interface{}) string {
	result, err := ModelMethod(model, "TableName")
	if err != nil {
		return ""
	}
	return result.(string)
}
