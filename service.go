package mysql

import (
	"encoding/json"
	"reflect"
	"time"
)

type IdReq struct {
	Id json.Number `json:"id" validate:"required"`
}

type FilterReq struct {
	Filters map[string]interface{}
}

type Pageable struct {
	Page int    `json:"page"`
	Size int    `json:"size"`
	Sort string `json:"sort"`
}

type PageReq struct {
	*Pageable `validate:"required"`
	Filters   map[string]interface{} `json:"filters"`
}

type PageRes struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
}

type TitleRes struct {
	Title string `json:"title"`
}

type BaseService struct {
	DB       *DB
	Model    interface{}
	GetTitle func(interface{}) string
	Pk       string
}

func (b *BaseService) GetPk() (string, error) {
	if b.Pk != "" {
		return b.Pk, nil
	}
	pkField := GetPKField(b.Model)
	pkName := pkField.Name
	if pkName == "" {
		return "", ErrorPrimaryKeyUnset()
	}
	b.Pk = pkName
	return b.Pk, nil
}

func (b *BaseService) NewModel() (interface{}, error) {
	model := reflect.New(reflect.TypeOf(b.Model).Elem())
	return model.Interface(), nil
}

func (b *BaseService) NewModelWithId(id interface{}) (interface{}, error) {
	model := reflect.New(reflect.TypeOf(b.Model).Elem())
	pk, err := b.GetPk()
	if err != nil {
		return nil, err
	}
	modelV := model.Elem()
	modelFieldV := modelV.FieldByName(pk)
	modelFieldV.Set(reflect.ValueOf(id).Convert(modelFieldV.Type()))
	return model.Interface(), nil
}

func (b *BaseService) NewModelWithValue(value interface{}) (interface{}, error) {
	model := reflect.New(reflect.TypeOf(b.Model).Elem())
	pk, err := b.GetPk()
	if err != nil {
		return nil, err
	}
	modelV := model.Elem()
	modelFieldV := modelV.FieldByName(pk)

	valueV := reflect.ValueOf(value)
	if valueV.Kind() == reflect.Ptr {
		valueV = valueV.Elem()
	}
	valueFieldV := valueV.FieldByName(pk)

	modelFieldV.Set(valueFieldV.Convert(modelFieldV.Type()))
	return model.Interface(), nil
}

func (b *BaseService) Create(value interface{}) (interface{}, error) {
	err := b.DB.Create(value)
	return value, err
}

func (b *BaseService) CreateWithUserId(value interface{}, userId int) (interface{}, error) {
	SetCreatedBy(value, userId)
	return b.Create(value)
}

func (b *BaseService) Update(value interface{}, filters ...map[string]interface{}) (interface{}, error) {
	if err := b.DB.UpdateById(
		value,
		value,
		b.DB.WithOmit("created_at", "created_by"),
		b.DB.WithFilters(filters...),
	); err != nil {
		return nil, err
	}

	if err := b.DB.FindById(
		value,
	); err != nil {
		return nil, err
	}
	return value, nil
}

func (b *BaseService) UpdateWithUserId(value interface{}, userId int, filters ...map[string]interface{}) (interface{}, error) {
	SetUpdatedBy(value, userId)
	return b.Update(value, filters...)
}

func (b *BaseService) UpdateOrCreateWithUserId(value interface{}, userId int, filters ...map[string]interface{}) (interface{}, error) {
	SetUpdatedBy(value, userId)
	r, err := b.Update(value, filters...)
	if err != nil && IsRecordNotFoundError(err) {
		return b.CreateWithUserId(value, userId)
	}
	return r, err
}

func (b *BaseService) Remove(value interface{}, filters ...map[string]interface{}) error {
	SetDeleted(value)
	return b.DB.UpdateById(
		value,
		value,
		b.DB.WithAttend("updated_at", "updated_by", "deleted"),
		b.DB.WithFilters(filters...),
	)
}

func (b *BaseService) RemoveById(id interface{}, filters ...map[string]interface{}) error {
	value, err := b.NewModelWithId(id)
	if err != nil {
		return err
	}
	return b.Remove(value, filters...)
}

func (b *BaseService) RemoveByIdWithUserId(id interface{}, userId int, filters ...map[string]interface{}) error {
	value, err := b.NewModelWithId(id)
	if err != nil {
		return err
	}
	SetUpdatedBy(value, userId)
	return b.Remove(value, filters...)
}

func (b *BaseService) FindTitle(id interface{}, filters ...map[string]interface{}) (*TitleRes, error) {
	model, err := b.FindById(id, filters...)
	if err != nil {
		return nil, err
	}
	return &TitleRes{Title: b.GetTitle(model)}, nil
}

func (b *BaseService) FindById(id interface{}, filters ...map[string]interface{}) (interface{}, error) {
	model, err := b.NewModelWithId(id)
	if err != nil {
		return nil, err
	}
	if err := b.DB.FindById(
		model,
		b.DB.WithFilters(filters...),
	); err != nil {
		return nil, err
	}
	return model, nil
}

func (b *BaseService) FindAll(filters ...map[string]interface{}) (interface{}, error) {
	model, err := b.NewModel()
	if err != nil {
		return nil, err
	}
	list, err := b.DB.FindAllWithModel(
		model,
		b.DB.WithFilters(filters...),
	)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (b *BaseService) FindPage(pageable *Pageable, filters ...map[string]interface{}) (*PageRes, error) {
	model, err := b.NewModel()
	if err != nil {
		return nil, err
	}
	list, total, err := b.DB.FindPageWithModel(
		model,
		b.DB.WithFilters(filters...),
		b.DB.WithPageable(pageable),
	)
	if err != nil {
		return nil, err
	}
	return &PageRes{Total: total, List: list}, nil
}

func (b *BaseService) FindOne(filters ...map[string]interface{}) (interface{}, error) {
	model, err := b.NewModel()
	if err != nil {
		return nil, err
	}

	if err := b.DB.FindOne(
		model,
		b.DB.WithFilters(filters...),
		b.DB.WithIgnoreNotFound(),
	); err != nil {
		return nil, err
	}
	return model, nil
}

func SetCreatedBy(value interface{}, userId int) {
	valueV := reflect.ValueOf(value)
	if valueV.Kind() == reflect.Ptr {
		valueV = valueV.Elem()
	}
	createdByFieldV := valueV.FieldByName("CreatedBy")
	if createdByFieldV.IsValid() && createdByFieldV.CanSet() {
		createdByFieldV.Set(reflect.ValueOf(NewUserId(userId)))
	}
	updatedByFieldV := valueV.FieldByName("UpdatedBy")
	if updatedByFieldV.IsValid() && updatedByFieldV.CanSet() {
		updatedByFieldV.Set(reflect.ValueOf(NewUserId(userId)))
	}
}

func SetUpdatedBy(value interface{}, userId int) {
	valueV := reflect.ValueOf(value)
	if valueV.Kind() == reflect.Ptr {
		valueV = valueV.Elem()
	}

	updatedByFieldV := valueV.FieldByName("UpdatedBy")
	if updatedByFieldV.IsValid() && updatedByFieldV.CanSet() {
		updatedByFieldV.Set(reflect.ValueOf(NewUserId(userId)))
	}
}

func SetDeleted(value interface{}) {
	valueV := reflect.ValueOf(value)
	if valueV.Kind() == reflect.Ptr {
		valueV = valueV.Elem()
	}

	updatedByFieldV := valueV.FieldByName("Deleted")
	if updatedByFieldV.IsValid() && updatedByFieldV.CanSet() {
		updatedByFieldV.Set(reflect.ValueOf(time.Now().Unix()).Convert(updatedByFieldV.Type()))
	}
}
