package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type User struct {
	Model

	Phone         string `json:"phone" gorm:"index;type:varchar(20)"`
	Name          string `json:"name" gorm:"type:varchar(100)"`
	Password      string `json:"-"`
	FirebaseToken string `json:"firebase_token"`
	Language      string `json:"language" gorm:"type:varchar(20)"`
}

type UserResponse struct {
	Model
	Phone string `json:"phone" gorm:"index"`
	Name  string `json:"name"`
	OsUid string `json:"os_uid"` // OrderSource Uid
}

func (item *UserResponse) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item UserResponse) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type UserBaseInfo struct {
	Uid    string `json:"uid"`
	Phone  string `json:"phone" gorm:"index"`
	OsCode string `json:"os_code"` // OrderSource Uid
}

type UserProfile struct {
	UserBaseInfo
	jwt.StandardClaims
}

func (item *User) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	item.Model.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *User) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *User) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *User) Count() (int64, error) {
	db := datasources.GetDatabase().Model(User{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *User) FindList(page Page) ([]User, int64, error) {
	db := datasources.GetDatabase().Model(User{})
	list := []User{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *User) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
