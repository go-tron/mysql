package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

func NewUser(userId int, name string) *UserId {
	return &UserId{
		Id:   userId,
		Name: name,
	}
}
func NewUserId(userId int) *UserId {
	return &UserId{
		Id: userId,
	}
}

type UserId struct {
	Id   int
	Name string
}

func (u UserId) Value() (driver.Value, error) {
	return strconv.Itoa(u.Id), nil
}

func (u *UserId) Scan(src interface{}) (err error) {
	var id = 0
	switch v := src.(type) {
	case int64:
		id = int(v)
	case int:
		id = v
	case string:
		id, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	case []uint8:
		id, err = strconv.Atoi(string(v))
		if err != nil {
			return err
		}
	default:
		id, err = strconv.Atoi(fmt.Sprintf(`%s`, v))
		if err != nil {
			return err
		}
	}
	*u = UserId{
		Id: id,
	}
	return nil
}

func (u *UserId) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 || string(data) == "\"\"" {
		return nil
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err == nil {
		if m["id"] != nil {
			id, ok := m["id"].(float64)
			if ok {
				name, _ := m["name"].(string)
				*u = *NewUser(int(id), name)
				return nil
			}
		}
	}
	id, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	*u = *NewUserId(id)
	return nil
}

func (u UserId) MarshalJSON() ([]byte, error) {
	var m = map[string]interface{}{
		"id":   u.Id,
		"name": u.Name,
	}
	if u.Name != "" || user == nil {
		return json.Marshal(m)
	}
	m["name"] = user.GetNameById(u.Id)
	return json.Marshal(m)
}

var user User

type User interface {
	GetNameById(userId int) string
	Init()
}

func InitUser(u User) {
	user = u
	user.Init()
}
