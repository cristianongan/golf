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

type CmsUser struct {
	Model

	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"`
	UserName   string `json:"user_name" gorm:"type:varchar(20);uniqueIndex"`

	FullName string `json:"full_name" gorm:"type:varchar(256)"`
	Password string `json:"-" gorm:"type:varchar(256)"`
	LoggedIn bool   `json:"logged_in"`

	Email    string `json:"email" gorm:"type:varchar(100)"`
	Phone    string `json:"phone" gorm:"type:varchar(20)"`
	BirthDay int64  `json:"birth_day"`
}

type CmsUserResponse struct {
	Model

	UserName string `json:"user_name"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
}

func (item *CmsUserResponse) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item CmsUserResponse) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type CmsUserBaseInfo struct {
	Uid        string `json:"uid"`
	PartnerUid string `json:"partner_uid"`
	UserName   string `json:"user_name"`
	Status     string `json:"status"`
}

type CmsUserProfile struct {
	CmsUserBaseInfo
	jwt.StandardClaims
}

func (item *CmsUser) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	item.Model.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CmsUser) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CmsUser) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CmsUser) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CmsUser{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CmsUser) FindList(page Page) ([]CmsUser, int64, error) {
	db := datasources.GetDatabase().Model(CmsUser{})
	list := []CmsUser{}
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

func (item *CmsUser) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
